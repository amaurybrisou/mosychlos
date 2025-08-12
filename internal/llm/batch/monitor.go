// internal/llm/batch/monitor.go
package batch

import (
	"context"
	"fmt"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// Monitor provides job monitoring and progress tracking capabilities
type Monitor struct {
	manager *Manager
}

// NewMonitor creates a new batch job monitor
func NewMonitor(manager *Manager) *Monitor {
	return &Monitor{
		manager: manager,
	}
}

// WatchJob monitors a job with progress updates
func (m *Monitor) WatchJob(ctx context.Context, jobID string, progressCallback func(*models.BatchJob)) (*models.BatchJob, error) {
	ticker := time.NewTicker(30 * time.Second) // Poll every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			job, err := m.manager.GetJobStatus(ctx, jobID)
			if err != nil {
				return nil, fmt.Errorf("failed to get job status: %w", err)
			}

			// Call progress callback if provided
			if progressCallback != nil {
				progressCallback(job)
			}

			// Check if job is finished
			switch job.Status {
			case models.BatchStatusCompleted:
				return job, nil
			case models.BatchStatusFailed, models.BatchStatusExpired, models.BatchStatusCancelled:
				return job, fmt.Errorf("job finished with status: %s", job.Status)
			default:
				// Continue monitoring
				continue
			}
		}
	}
}

// GetProgress calculates job completion percentage
func (m *Monitor) GetProgress(job *models.BatchJob) float64 {
	if job.RequestCounts.Total == 0 {
		return 0.0
	}
	return float64(job.RequestCounts.Completed) / float64(job.RequestCounts.Total) * 100.0
}

// GetEstimatedTimeRemaining estimates time remaining based on current progress
func (m *Monitor) GetEstimatedTimeRemaining(job *models.BatchJob) time.Duration {
	if job.RequestCounts.Total == 0 || job.RequestCounts.Completed == 0 {
		return time.Hour // Default estimate if no progress
	}

	elapsed := time.Since(time.Unix(job.CreatedAt, 0))
	progressRatio := float64(job.RequestCounts.Completed) / float64(job.RequestCounts.Total)

	if progressRatio == 0 {
		return time.Hour // Avoid division by zero
	}

	estimatedTotal := time.Duration(float64(elapsed) / progressRatio)
	remaining := estimatedTotal - elapsed

	if remaining < 0 {
		return 0
	}

	return remaining
}

// FormatStatus formats job status for display
func (m *Monitor) FormatStatus(job *models.BatchJob) string {
	progress := m.GetProgress(job)
	remaining := m.GetEstimatedTimeRemaining(job)

	statusMsg := fmt.Sprintf("Status: %s | Progress: %.1f%% (%d/%d)",
		job.Status, progress, job.RequestCounts.Completed, job.RequestCounts.Total)

	if job.Status == models.BatchStatusInProgress && remaining > 0 {
		statusMsg += fmt.Sprintf(" | Est. remaining: %s", remaining.Round(time.Minute))
	}

	return statusMsg
}
