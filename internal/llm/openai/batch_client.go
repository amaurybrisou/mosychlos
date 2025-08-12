// internal/llm/openai/batch_client.go
package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strconv"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	oa "github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/openai/openai-go/v2/shared"
)

// BatchClient interface for OpenAI batch operations (internal interface)
type BatchClient interface {
	SubmitBatch(ctx context.Context, reqs []models.BatchRequest, opts models.BatchOptions) (*models.BatchJob, error)
	GetBatchStatus(ctx context.Context, jobID string) (*models.BatchJob, error)
	GetBatchResults(ctx context.Context, jobID string) (io.ReadCloser, error)
	GetBatchErrors(ctx context.Context, jobID string) (io.ReadCloser, error)
	CancelBatch(ctx context.Context, jobID string) error
	ListBatches(ctx context.Context, filters map[string]string) ([]models.BatchJob, error)
}

// batchClient implements BatchClient using OpenAI client
type batchClient struct {
	client oa.Client
	config config.LLMConfig
}

// NewBatchClient creates a new OpenAI batch client
func NewBatchClient(cfg config.LLMConfig, sharedBag bag.SharedBag) (models.AiBatchClient, error) {
	opts := []option.RequestOption{}

	if cfg.APIKey != "" {
		opts = append(opts, option.WithAPIKey(cfg.APIKey))
	}

	if cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	}

	return &batchClient{
		client: oa.NewClient(opts...),
		config: cfg,
	}, nil
}

// SubmitBatch submits a batch of requests to OpenAI
func (bc *batchClient) SubmitBatch(ctx context.Context, reqs []models.BatchRequest, opts models.BatchOptions) (*models.BatchJob, error) {
	slog.Info("Submitting batch to OpenAI",
		"request_count", len(reqs),
		"completion_window", opts.CompletionWindow,
	)

	// Convert batch requests to JSONL format
	jsonlData, err := bc.requestsToJSONL(reqs)
	if err != nil {
		return nil, fmt.Errorf("failed to convert requests to JSONL: %w", err)
	}

	// Determine endpoint based on request URLs
	endpoint := oa.BatchNewParamsEndpointV1ChatCompletions // default
	if len(reqs) > 0 {
		// Check if any request uses responses endpoint
		for _, req := range reqs {
			if req.URL == "/v1/responses" {
				endpoint = oa.BatchNewParamsEndpointV1Responses
				break
			}
		}
	}

	// Upload the JSONL file
	uploadResp, err := bc.client.Files.New(ctx, oa.FileNewParams{
		File:    oa.File(bytes.NewReader(jsonlData), "batch_input.jsonl", "application/jsonl"),
		Purpose: oa.FilePurposeBatch,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload batch file: %w", err)
	}

	slog.Info("Batch file uploaded",
		"file_id", uploadResp.ID,
		"size", len(jsonlData),
		"endpoint", endpoint,
	)

	// Create the batch job
	batchResp, err := bc.client.Batches.New(ctx, oa.BatchNewParams{
		InputFileID:      uploadResp.ID,
		Endpoint:         endpoint, // Use determined endpoint
		CompletionWindow: oa.BatchNewParamsCompletionWindow24h,
		Metadata:         shared.Metadata(opts.Metadata),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create batch: %w", err)
	}

	slog.Info("Batch created successfully",
		"batch_id", batchResp.ID,
		"status", string(batchResp.Status),
	)

	// Convert OpenAI batch to our model
	job := &models.BatchJob{
		ID:          batchResp.ID,
		Status:      models.BatchStatus(batchResp.Status),
		InputFileID: batchResp.InputFileID,
		CreatedAt:   batchResp.CreatedAt,
		RequestCounts: struct {
			Total     int `json:"total"`
			Completed int `json:"completed"`
			Failed    int `json:"failed"`
		}{
			Total:     int(batchResp.RequestCounts.Total),
			Completed: int(batchResp.RequestCounts.Completed),
			Failed:    int(batchResp.RequestCounts.Failed),
		},
		Metadata: opts.Metadata,
	}

	if batchResp.OutputFileID != "" {
		job.OutputFileID = &batchResp.OutputFileID
	}
	if batchResp.ErrorFileID != "" {
		job.ErrorFileID = &batchResp.ErrorFileID
	}
	if batchResp.CompletedAt != 0 {
		job.CompletedAt = &batchResp.CompletedAt
	}

	return job, nil
}

// requestsToJSONL converts batch requests to JSONL format
func (bc *batchClient) requestsToJSONL(reqs []models.BatchRequest) ([]byte, error) {
	var buf bytes.Buffer

	for _, req := range reqs {
		reqData, err := json.Marshal(req)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		buf.Write(reqData)
		buf.WriteByte('\n')
	}

	return buf.Bytes(), nil
}

// GetBatchStatus retrieves the current status of a batch job
func (bc *batchClient) GetBatchStatus(ctx context.Context, jobID string) (*models.BatchJob, error) {
	slog.Debug("Getting batch status", "batch_id", jobID)

	// Get batch status from OpenAI
	batchResp, err := bc.client.Batches.Get(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch status: %w", err)
	}

	// Convert OpenAI batch to our model
	job := &models.BatchJob{
		ID:          batchResp.ID,
		Status:      models.BatchStatus(batchResp.Status),
		InputFileID: batchResp.InputFileID,
		CreatedAt:   batchResp.CreatedAt,
		RequestCounts: struct {
			Total     int `json:"total"`
			Completed int `json:"completed"`
			Failed    int `json:"failed"`
		}{
			Total:     int(batchResp.RequestCounts.Total),
			Completed: int(batchResp.RequestCounts.Completed),
			Failed:    int(batchResp.RequestCounts.Failed),
		},
		Metadata: map[string]string(batchResp.Metadata),
	}

	if batchResp.OutputFileID != "" {
		job.OutputFileID = &batchResp.OutputFileID
	}
	if batchResp.ErrorFileID != "" {
		job.ErrorFileID = &batchResp.ErrorFileID
	}
	if batchResp.CompletedAt != 0 {
		job.CompletedAt = &batchResp.CompletedAt
	}

	slog.Debug("Batch status retrieved",
		"batch_id", jobID,
		"status", string(job.Status),
		"completed", job.RequestCounts.Completed,
		"total", job.RequestCounts.Total,
	)

	return job, nil
}

// GetBatchResults downloads and returns the results file from OpenAI
func (bc *batchClient) GetBatchResults(ctx context.Context, jobID string) (io.ReadCloser, error) {
	slog.Debug("Getting batch results", "batch_id", jobID)

	// First get batch info to get the output file ID
	batch, err := bc.client.Batches.Get(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch info: %w", err)
	}

	// Check if batch is actually completed
	if batch.Status != oa.BatchStatusCompleted {
		return nil, fmt.Errorf("batch %s is not completed (status: %s)", jobID, batch.Status)
	}

	if batch.OutputFileID == "" {
		// If no output file but batch is completed, check if there are errors or all requests failed
		if batch.ErrorFileID != "" {
			slog.Warn("No output file but error file exists, all requests may have failed",
				"batch_id", jobID,
				"error_file_id", batch.ErrorFileID,
				"failed_requests", batch.RequestCounts.Failed,
			)

			// Let's try to get the errors first so the user can see what went wrong
			errorReader, err := bc.GetBatchErrors(ctx, jobID)
			if err != nil {
				slog.Error("Failed to get error details", "batch_id", jobID, "error", err)
			} else if errorReader != nil {
				// Read a few lines to show the errors
				scanner := bufio.NewScanner(errorReader)
				lineCount := 0
				for scanner.Scan() && lineCount < 3 {
					slog.Error("Batch error sample", "batch_id", jobID, "error_line", scanner.Text())
					lineCount++
				}
				errorReader.Close()
			}

			return nil, fmt.Errorf("batch %s completed but has no output file - all %d requests failed (error file: %s)",
				jobID, batch.RequestCounts.Failed, batch.ErrorFileID)
		}

		// Check request counts
		if batch.RequestCounts.Completed == 0 {
			return nil, fmt.Errorf("batch %s completed but no requests were successfully processed (total: %d, failed: %d)",
				jobID, batch.RequestCounts.Total, batch.RequestCounts.Failed)
		}

		return nil, fmt.Errorf("batch %s completed but OpenAI has not generated output file yet (this may take a few more minutes)", jobID)
	}

	// Download the results file
	fileContent, err := bc.client.Files.Content(ctx, batch.OutputFileID)
	if err != nil {
		return nil, fmt.Errorf("failed to download results file: %w", err)
	}

	slog.Info("Batch results downloaded",
		"batch_id", jobID,
		"output_file_id", batch.OutputFileID,
		"completed_requests", batch.RequestCounts.Completed,
	)

	return fileContent.Body, nil
}

// GetBatchErrors downloads and returns the errors file from OpenAI
func (bc *batchClient) GetBatchErrors(ctx context.Context, jobID string) (io.ReadCloser, error) {
	slog.Debug("Getting batch errors", "batch_id", jobID)

	// First get batch info to get the error file ID
	batch, err := bc.client.Batches.Get(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get batch info: %w", err)
	}

	switch batch.Status {
	case oa.BatchStatusCompleted:
		// No error file means no errors - this is normal for successful batches
		slog.Debug("No error file found for batch - this is normal for successful batches", "batch_id", jobID)
		return nil, nil
	case oa.BatchStatusFailed:
		errorsBytes, err := json.Marshal(batch)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal batch errors: %w", err)
		}
		slog.Debug("Batch failed with errors", "batch_id", jobID, "errors", string(errorsBytes))
		return nil, fmt.Errorf("batch %s failed with errors: %s", jobID, string(errorsBytes))
	}

	if batch.ErrorFileID == "" {
		// No error file means no errors - this is normal for successful batches
		slog.Debug("No error file found for batch - this is normal for successful batches", "batch_id", jobID)
		return nil, nil
	}

	// Download the errors file
	fileContent, err := bc.client.Files.Content(ctx, batch.ErrorFileID)
	if err != nil {
		return nil, fmt.Errorf("failed to download errors file: %w", err)
	}

	slog.Info("Batch errors downloaded",
		"batch_id", jobID,
		"error_file_id", batch.ErrorFileID,
	)

	return fileContent.Body, nil
}

// CancelBatch cancels a batch job in OpenAI
func (bc *batchClient) CancelBatch(ctx context.Context, jobID string) error {
	slog.Info("Cancelling batch", "batch_id", jobID)

	// Cancel the batch via OpenAI API
	_, err := bc.client.Batches.Cancel(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to cancel batch: %w", err)
	}

	slog.Info("Batch cancelled successfully", "batch_id", jobID)
	return nil
}

// ListBatches lists batch jobs from OpenAI with optional filters
func (bc *batchClient) ListBatches(ctx context.Context, filters map[string]string) ([]models.BatchJob, error) {
	slog.Debug("Listing batches", "filters", filters)

	// Create list params
	params := oa.BatchListParams{}

	// Apply filters if provided
	if after, ok := filters["after"]; ok {
		params.After = oa.String(after)
	}
	if _, ok := filters["limit"]; ok {
		// Use default limit of 20
		v, err := strconv.ParseInt(filters["limit"], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid limit filter: %w", err)
		}
		params.Limit = oa.Int(v)
	}

	// List batches from OpenAI
	page, err := bc.client.Batches.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list batches: %w", err)
	}

	// Convert to our model
	var jobs []models.BatchJob
	for i, batch := range page.Data {
		if int64(i) > params.Limit.Value {
			break
		}
		job := models.BatchJob{
			ID:          batch.ID,
			Status:      models.BatchStatus(batch.Status),
			InputFileID: batch.InputFileID,
			CreatedAt:   batch.CreatedAt,
			RequestCounts: struct {
				Total     int `json:"total"`
				Completed int `json:"completed"`
				Failed    int `json:"failed"`
			}{
				Total:     int(batch.RequestCounts.Total),
				Completed: int(batch.RequestCounts.Completed),
				Failed:    int(batch.RequestCounts.Failed),
			},
			Metadata: map[string]string(batch.Metadata),
		}

		if batch.OutputFileID != "" {
			job.OutputFileID = &batch.OutputFileID
		}
		if batch.ErrorFileID != "" {
			job.ErrorFileID = &batch.ErrorFileID
		}
		if batch.CompletedAt != 0 {
			job.CompletedAt = &batch.CompletedAt
		}

		jobs = append(jobs, job)
	}

	slog.Info("Batches listed", "count", len(jobs))
	return jobs, nil
}
