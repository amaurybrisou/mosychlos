// internal/llm/batch/manager.go
package batch

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// Manager orchestrates batch processing workflows with engine support
type Manager struct {
	client        models.AiBatchClient
	aggregator    *ResultAggregator
	costOptimizer *CostOptimizer
	pollDelay     time.Duration
}

// NewManager creates a new batch processing manager
func NewManager(client models.AiBatchClient) *Manager {
	// Create aggregator using the same client (if it implements ResultsReader)
	var aggregator *ResultAggregator
	if reader, ok := client.(ResultsReader); ok {
		aggregator = NewResultAggregator(reader)
	}

	return &Manager{
		client:        client,
		aggregator:    aggregator,
		costOptimizer: NewCostOptimizer(),
		pollDelay:     30 * time.Second, // Default polling interval
	}
}

// SetPollDelay configures the polling interval for job status checks
func (m *Manager) SetPollDelay(delay time.Duration) {
	m.pollDelay = delay
}

// EstimateCost provides cost estimation for batch requests without submitting
func (m *Manager) EstimateCost(requests []models.BatchRequest) *models.CostEstimate {
	return m.costOptimizer.EstimateCost(requests)
}

// ProcessBatch submits a batch job and optionally waits for completion
func (m *Manager) ProcessBatch(ctx context.Context, requests []models.BatchRequest, opts models.BatchOptions, waitForCompletion bool) (*models.BatchJob, error) {
	// Add cost estimation if cost optimization is enabled
	if opts.CostOptimize {
		costEstimate := m.costOptimizer.EstimateCost(requests)

		// Add cost estimate to options metadata
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]string)
		}
		opts.Metadata["estimated_cost"] = fmt.Sprintf("%.4f", costEstimate.EstimatedCost)
		opts.Metadata["estimated_savings"] = fmt.Sprintf("%.2f%%", costEstimate.SavingsVsSync*100)
	}

	// Submit the batch job
	job, err := m.client.SubmitBatch(ctx, requests, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to submit batch: %w", err)
	}

	if !waitForCompletion {
		return job, nil
	}

	// Wait for completion
	finalJob, err := m.waitForCompletion(ctx, job.ID)
	if err != nil {
		return job, fmt.Errorf("batch submitted but completion failed: %w", err)
	}

	return finalJob, nil
}

// GetJobStatus retrieves the current status of a batch job
func (m *Manager) GetJobStatus(ctx context.Context, jobID string) (*models.BatchJob, error) {
	return m.client.GetBatchStatus(ctx, jobID)
}

// WaitForCompletion polls until the batch job completes
func (m *Manager) WaitForCompletion(ctx context.Context, jobID string) (*models.BatchJob, error) {
	return m.waitForCompletion(ctx, jobID)
}

// GetResults retrieves and aggregates results from a completed batch job
func (m *Manager) GetResults(ctx context.Context, jobID string) (*models.Aggregated, error) {
	if m.aggregator == nil {
		return nil, fmt.Errorf("result aggregation not supported by this client")
	}

	// Check if job is completed
	job, err := m.client.GetBatchStatus(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to check job status: %w", err)
	}

	if job.Status != models.BatchStatusCompleted {
		return nil, fmt.Errorf("batch job not completed (status: %s)", job.Status)
	}

	return m.aggregator.AggregateResults(ctx, jobID)
}

// CancelJob cancels a running or queued batch job
func (m *Manager) CancelJob(ctx context.Context, jobID string) error {
	return m.client.CancelBatch(ctx, jobID)
}

// ListBatches lists batch jobs with optional filters
func (m *Manager) ListBatches(ctx context.Context, filters map[string]string) ([]models.BatchJob, error) {
	return m.client.ListBatches(ctx, filters)
}

// GetError retrieves error information for a given job ID
func (m *Manager) GetError(ctx context.Context, jobID string) (map[string]string, error) {
	// Check if job exists and get its status
	job, err := m.client.GetBatchStatus(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to check job status: %w", err)
	}

	// Get the error stream from the client
	reader, err := m.client.GetBatchErrors(ctx, jobID)
	if err != nil {
		// If no errors file exists, provide information based on job status
		switch job.Status {
		case models.BatchStatusCompleted:
			return make(map[string]string), nil // No errors for completed job
		case models.BatchStatusFailed:
			return nil, fmt.Errorf("job failed but no detailed error information available (status: %s)", job.Status)
		case models.BatchStatusValidating, models.BatchStatusInProgress, models.BatchStatusFinalizing:
			return nil, fmt.Errorf("job is still running (status: %s), errors not yet available", job.Status)
		case models.BatchStatusCancelled:
			return nil, fmt.Errorf("job was cancelled (status: %s), no error details available", job.Status)
		case models.BatchStatusExpired:
			return nil, fmt.Errorf("job expired (status: %s), error details may not be available", job.Status)
		default:
			return nil, fmt.Errorf("unable to retrieve errors for job %s (status: %s): %w", jobID, job.Status, err)
		}
	}

	if reader == nil {
		return nil, nil
	}

	defer reader.Close()

	// Parse errors from JSONL format
	errors := make(map[string]string)
	if err := m.parseErrors(reader, errors); err != nil {
		return nil, fmt.Errorf("failed to parse errors for job %s: %w", jobID, err)
	}

	return errors, nil
}

// parseErrors parses JSONL format errors and populates the errors map
func (m *Manager) parseErrors(reader io.Reader, errors map[string]string) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		var item map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &item); err == nil {
			if customID, ok := item["custom_id"].(string); ok {
				if errorData, ok := item["error"]; ok {
					if errorBytes, err := json.Marshal(errorData); err == nil {
						errors[customID] = string(errorBytes)
					} else {
						errors[customID] = fmt.Sprintf("failed to marshal error: %v", errorData)
					}
				}
			}
		}
		// Silently skip malformed lines as per the implementation guide
	}
	return scanner.Err()
}

// waitForCompletion implements the polling logic for job completion
func (m *Manager) waitForCompletion(ctx context.Context, jobID string) (*models.BatchJob, error) {
	ticker := time.NewTicker(m.pollDelay)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			job, err := m.client.GetBatchStatus(ctx, jobID)
			if err != nil {
				return nil, fmt.Errorf("failed to check job status: %w", err)
			}

			switch job.Status {
			case models.BatchStatusCompleted:
				return job, nil
			case models.BatchStatusFailed, models.BatchStatusExpired, models.BatchStatusCancelled:
				return job, fmt.Errorf("batch job failed with status: %s", job.Status)
			case models.BatchStatusValidating, models.BatchStatusInProgress, models.BatchStatusFinalizing:
				// Continue polling
				continue
			default:
				return job, fmt.Errorf("unknown batch status: %s", job.Status)
			}
		}
	}
}
