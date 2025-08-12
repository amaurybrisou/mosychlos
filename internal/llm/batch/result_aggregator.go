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

// ResultsReader interface for reading batch results and errors
// This should be implemented by batch clients that support result streaming
type ResultsReader interface {
	GetBatchResults(ctx context.Context, jobID string) (io.ReadCloser, error)
	GetBatchErrors(ctx context.Context, jobID string) (io.ReadCloser, error)
}

// ResultAggregator handles aggregation of batch processing results
type ResultAggregator struct {
	client ResultsReader
}

// NewResultAggregator creates a new result aggregator
func NewResultAggregator(client ResultsReader) *ResultAggregator {
	return &ResultAggregator{
		client: client,
	}
}

// AggregateResults processes batch results and errors into a unified view
func (ra *ResultAggregator) AggregateResults(ctx context.Context, jobID string) (*models.Aggregated, error) {
	result := &models.Aggregated{
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
func (ra *ResultAggregator) processResults(ctx context.Context, jobID string, result *models.Aggregated) error {
	reader, err := ra.client.GetBatchResults(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get results: %w", err)
	}
	defer reader.Close()

	return scanJSONLLines(reader, func(item map[string]any) {
		if customID, ok := item["custom_id"].(string); ok {
			result.Items[customID] = item

			slog.Debug("Processing batch result item", "custom_id", customID)

			// Extract tool calls and content from the response
			if responseData, hasResponse := item["response"].(map[string]any); hasResponse {
				slog.Debug("Found response data", "custom_id", customID)
				if bodyData, hasBody := responseData["body"].(map[string]any); hasBody {
					slog.Debug("Found body data", "custom_id", customID)
					if choices, hasChoices := bodyData["choices"].([]any); hasChoices && len(choices) > 0 {
						slog.Debug("Found choices", "custom_id", customID, "choices_count", len(choices))
						if choice, isChoice := choices[0].(map[string]any); isChoice {
							slog.Debug("Processing first choice", "custom_id", customID)
							if message, hasMessage := choice["message"].(map[string]any); hasMessage {
								slog.Debug("Found message", "custom_id", customID)

								// Extract content
								if contentStr, hasContent := message["content"].(string); hasContent {
									result.Content[customID] = contentStr
									slog.Debug("Extracted content", "custom_id", customID, "content_length", len(contentStr))
								} else {
									slog.Debug("No content found in message", "custom_id", customID)
								}

								// Extract tool calls
								if toolCallsData, hasToolCalls := message["tool_calls"].([]any); hasToolCalls {
									var toolCalls []models.ToolCall
									for _, tcData := range toolCallsData {
										if tc, isTc := tcData.(map[string]any); isTc {
											toolCall := models.ToolCall{}

											if id, hasId := tc["id"].(string); hasId {
												toolCall.ID = id
											}
											if name, hasName := tc["name"].(string); hasName {
												toolCall.Function.Name = name
											}
											if function, hasFunction := tc["function"].(map[string]any); hasFunction {
												if name, hasName := function["name"].(string); hasName {
													toolCall.Function.Name = name
												}
												if args, hasArgs := function["arguments"].(string); hasArgs {
													toolCall.Function.Arguments = args
												}
											}

											if toolCall.ID != "" && toolCall.Function.Name != "" {
												toolCall.Type = "function" // Set the type for OpenAI compatibility
												toolCalls = append(toolCalls, toolCall)
											}
										}
									}

									if len(toolCalls) > 0 {
										result.ToolCalls[customID] = toolCalls
										slog.Debug("Extracted tool calls", "custom_id", customID, "tool_calls_count", len(toolCalls))
									} else {
										slog.Debug("No tool calls found", "custom_id", customID)
									}
								}
							} else {
								slog.Debug("No message found in choice", "custom_id", customID)
							}
						} else {
							slog.Debug("First choice is not a map", "custom_id", customID)
						}
					} else {
						slog.Debug("No choices found or empty choices", "custom_id", customID)
					}

					// Extract token usage from body data
					if usage, hasUsage := bodyData["usage"].(map[string]any); hasUsage {
						batchUsage := models.BatchUsage{}

						if promptTokens, hasPrompt := usage["prompt_tokens"].(float64); hasPrompt {
							batchUsage.PromptTokens = int(promptTokens)
						}
						if completionTokens, hasCompletion := usage["completion_tokens"].(float64); hasCompletion {
							batchUsage.CompletionTokens = int(completionTokens)
						}
						if totalTokens, hasTotal := usage["total_tokens"].(float64); hasTotal {
							batchUsage.TotalTokens = int(totalTokens)
						}

						if batchUsage.TotalTokens > 0 {
							result.Usage[customID] = batchUsage
							slog.Debug("Extracted token usage", "custom_id", customID,
								"prompt_tokens", batchUsage.PromptTokens,
								"completion_tokens", batchUsage.CompletionTokens,
								"total_tokens", batchUsage.TotalTokens)
						} else {
							slog.Debug("No valid token usage found", "custom_id", customID)
						}
					} else {
						slog.Debug("No usage data found in body", "custom_id", customID)
					}
				} else {
					slog.Debug("No body data found in response", "custom_id", customID)
				}
			} else {
				slog.Debug("No response data found", "custom_id", customID)
			}
		} else {
			slog.Debug("No custom_id found in item")
		}
	})
}

// processErrors reads and processes the error results JSONL file
func (ra *ResultAggregator) processErrors(ctx context.Context, jobID string, result *models.Aggregated) error {
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
