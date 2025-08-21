// internal/llm/batch/result_aggregator.go
package batch

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// batchResult represents one line of the JSONL file
type batchResult struct {
	CustomID string         `json:"custom_id"`
	Response *batchResponse `json:"response,omitempty"`
}

// batchResponse represents the "response" object
type batchResponse struct {
	Body *batchBody `json:"body,omitempty"`
}

// batchBody represents the "body" inside response
type batchBody struct {
	Choices []choice      `json:"choices,omitempty"`
	Usage   *models.Usage `json:"usage,omitempty"`
}

// choice represents one choice in the response
type choice struct {
	Message *message `json:"message,omitempty"`
}

// message represents the content and tool calls
type message struct {
	Content   string            `json:"content,omitempty"`
	ToolCalls []models.ToolCall `json:"tool_calls,omitempty"`
}

// ResultsReader interface for reading batch results and errors
// This should be implemented by batch clients that support result streaming
type ResultsReader interface {
	GetBatchResults(ctx context.Context, jobID string) (io.ReadCloser, error)
	GetBatchErrors(ctx context.Context, jobID string) (io.ReadCloser, error)
}

// ResultParser handles aggregation of batch processing results
type ResultParser struct {
	client ResultsReader
}

// NewBatchResultParser creates a new result aggregator
func NewBatchResultParser(client ResultsReader) *ResultParser {
	return &ResultParser{
		client: client,
	}
}

// AggregateBatchResult processes batch results and errors into a unified view
func (ra *ResultParser) AggregateBatchResult(ctx context.Context, jobID string) (*models.BatchResult, error) {
	result := &models.BatchResult{
		JobID:     jobID,
		Items:     make(map[string]any),
		Errors:    make(map[string]string),
		ToolCalls: make(map[string][]models.ToolCall),
		Content:   make(map[string]string),
		Usage:     make(map[string]models.BatchUsage),
	}

	// Process success results
	if err := ra.processResults(ctx, jobID, result); err != nil {
		return nil, fmt.Errorf("failed to process results: %w", err)
	}

	// Process errors (if any)
	if err := ra.processErrors(ctx, jobID, result); err != nil {
		// Log error but don't fail - errors file might not exist
		// TODO: Add proper logging
		slog.Error("failed to process errors", "jobID", jobID, "error", err)
	}

	result.Successes = len(result.Items)
	result.Failures = len(result.Errors)

	return result, nil
}

// processResults reads and processes the success results JSONL file
func (ra *ResultParser) processResults(ctx context.Context, jobID string, result *models.BatchResult) error {
	reader, err := ra.client.GetBatchResults(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get results: %w", err)
	}
	defer reader.Close()

	return scanJSONLLines(reader, func(item map[string]any) {
		// Marshal -> Unmarshal into our typed struct
		data, err := json.Marshal(item)
		if err != nil {
			slog.Warn("Failed to marshal raw item", "err", err)
			return
		}

		var br batchResult
		if err := json.Unmarshal(data, &br); err != nil {
			slog.Warn("Failed to unmarshal batch result", "err", err)
			return
		}

		if br.CustomID == "" {
			slog.Debug("No custom_id found in item")
			return
		}

		result.Items[br.CustomID] = item
		slog.Debug("Processing batch result item", "custom_id", br.CustomID)

		if br.Response == nil || br.Response.Body == nil {
			slog.Debug("No body/response found", "custom_id", br.CustomID)
			return
		}

		body := br.Response.Body

		// Extract content
		if len(body.Choices) > 0 && body.Choices[0].Message != nil {
			msg := body.Choices[0].Message
			if msg.Content != "" {
				result.Content[br.CustomID] = msg.Content
				slog.Debug("Extracted content", "custom_id", br.CustomID, "content_length", len(msg.Content))
			}

			if len(msg.ToolCalls) > 0 {
				// Normalize type
				for i := range msg.ToolCalls {
					if msg.ToolCalls[i].Type == "" {
						msg.ToolCalls[i].Type = "function"
					}
				}
				result.ToolCalls[br.CustomID] = msg.ToolCalls
				slog.Debug("Extracted tool calls", "custom_id", br.CustomID, "tool_calls_count", len(msg.ToolCalls))
			}
		}

		// Extract usage
		if body.Usage != nil && body.Usage.TotalTokens > 0 {
			result.Usage[br.CustomID] = models.BatchUsage{
				PromptTokens:     body.Usage.PromptTokens,
				CompletionTokens: body.Usage.CompletionTokens,
				TotalTokens:      body.Usage.TotalTokens,
			}
			slog.Debug("Extracted token usage", "custom_id", br.CustomID,
				"prompt_tokens", body.Usage.PromptTokens,
				"completion_tokens", body.Usage.CompletionTokens,
				"total_tokens", body.Usage.TotalTokens)
		}
	})
}

// processErrors reads and processes the error results JSONL file
func (ra *ResultParser) processErrors(ctx context.Context, jobID string, result *models.BatchResult) error {
	reader, err := ra.client.GetBatchErrors(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get errors: %w", err)
	}

	// If reader is nil, it means there's no error file (which is normal for successful batches)
	if reader == nil {
		slog.Debug("No errors to process - batch completed successfully", "jobID", jobID)
		return nil
	}
	defer reader.Close()

	return scanJSONLLines(reader, func(item map[string]any) {
		if customID, ok := item["custom_id"].(string); ok {
			if errorData, ok := item["error"]; ok {
				if errorBytes, err := json.Marshal(errorData); err == nil {
					result.Errors[customID] = string(errorBytes)
					slog.Error("Batch request failed", "custom_id", customID, "error", string(errorBytes))
				} else {
					result.Errors[customID] = "Failed to marshal error"
					slog.Error("Batch request failed with unmarshalable error", "custom_id", customID, "error_raw", errorData)
				}
			}
		}
	})
}

// scanJSONLLines scans JSONL format and calls fn for each parsed line
func scanJSONLLines(r io.Reader, fn func(map[string]any)) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		var item map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &item); err == nil {
			fn(item)
		}
		// Note: Silently skip malformed lines as per the implementation guide
	}
	return scanner.Err()
}
