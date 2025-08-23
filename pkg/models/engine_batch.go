package models

import (
	"context"
	"fmt"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
)

//go:generate mockgen -destination=mocks/mock_batch_engine_hooks.go -package=mocks . BatchEngineHooks

// BatchEngineHooks defines the hooks for customizing batch engine behavior
type BatchEngineHooks interface {
	// GetInitialPrompt returns the initial prompt to start the batch process
	GetInitialPrompt(ctx context.Context) (string, error)

	// GenerateCustomID generates a unique custom ID for batch requests
	GenerateCustomID(iteration, jobIndex int) string

	// PreIteration is called before each batch iteration
	PreIteration(iteration int, jobs []BatchJob) error

	// PostIteration is called after each batch iteration with results
	PostIteration(iteration int, results *BatchResult) error

	// ProcessToolResult is called when a tool call result is processed
	ProcessToolResult(customID, toolName, result string, sharedBag bag.SharedBag) error

	// ProcessFinalResult is called when a final result (no more tool calls) is processed
	ProcessFinalResult(customID, content string, sharedBag bag.SharedBag) error

	// ShouldContinueIteration determines if the batch process should continue
	ShouldContinueIteration(iteration int, nextJobs []BatchJob) bool

	// ResultKey returns the key where final results should be stored
	ResultKey() bag.Key
}

// ToolConstraints is an interface for BaseToolConstraints
type ToolConstraints interface {
	// Core constraint methods - actually used
	GetMaxCalls(toolKey bag.Key) int
	GetMinCalls(toolKey bag.Key) int
	GetRequiredTools() []bag.Key
	GetToolsWithLimits() []bag.Key
	RemainingCalls(toolKey bag.Key, currentCalls int) int

	// Validation
	Validate() error
}

// BatchConstraints defines constraints and limits for batch processing operations
// Composes existing Constraints and adds batch-specific settings
type BatchConstraints struct {
	// Embed existing constraints for model limits, timeouts, etc.
	ToolConstraints

	// Batch-specific constraints
	MaxIterations    int           `json:"max_iterations" yaml:"max_iterations"`         // Maximum number of batch iterations
	MaxJobsPerBatch  int           `json:"max_jobs_per_batch" yaml:"max_jobs_per_batch"` // Maximum jobs in a single batch
	BatchTimeout     time.Duration `json:"batch_timeout" yaml:"batch_timeout"`           // Total timeout for entire batch process
	IterationTimeout time.Duration `json:"iteration_timeout" yaml:"iteration_timeout"`   // Timeout per iteration
	RetryAttempts    int           `json:"retry_attempts" yaml:"retry_attempts"`         // Number of retry attempts for failed jobs
	RetryDelay       time.Duration `json:"retry_delay" yaml:"retry_delay"`               // Delay between retries
	ConcurrentJobs   int           `json:"concurrent_jobs" yaml:"concurrent_jobs"`       // Number of jobs to process concurrently
	EnableEarlyStop  bool          `json:"enable_early_stop" yaml:"enable_early_stop"`   // Allow early termination based on results
	SuccessThreshold float64       `json:"success_threshold" yaml:"success_threshold"`   // Minimum success rate to continue (0.0-1.0)
}

// DefaultBatchConstraints returns sensible defaults for batch processing
func DefaultBatchConstraints() BatchConstraints {
	return BatchConstraints{
		ToolConstraints:  DefaultToolConstraints(), // Use existing defaults
		MaxIterations:    10,
		MaxJobsPerBatch:  100,
		BatchTimeout:     30 * time.Minute,
		IterationTimeout: 5 * time.Minute,
		RetryAttempts:    3,
		RetryDelay:       5 * time.Second,
		ConcurrentJobs:   5,
		EnableEarlyStop:  true,
		SuccessThreshold: 0.7, // 70% success rate required to continue
	}
}

// // RiskAnalysisBatchConstraints returns constraints optimized for risk analysis
// func RiskAnalysisBatchConstraints() BatchConstraints {
// 	constraints := DefaultBatchConstraints()

// 	// Risk analysis specific settings
// 	constraints.MaxIterations = 3    // Risk analysis usually needs fewer iterations
// 	constraints.MaxJobsPerBatch = 50 // Smaller batches for focused analysis
// 	constraints.BatchTimeout = 20 * time.Minute
// 	constraints.SuccessThreshold = 0.8 // Higher success threshold for risk analysis

// 	return constraints
// }

// // InvestmentResearchBatchConstraints returns constraints optimized for investment research
// func InvestmentResearchBatchConstraints() BatchConstraints {
// 	constraints := DefaultBatchConstraints()

// 	// Investment research specific settings
// 	constraints.MaxIterations = 5               // More iterations for comprehensive research
// 	constraints.MaxJobsPerBatch = 20            // Smaller batches for deep analysis
// 	constraints.BatchTimeout = 45 * time.Minute // Longer timeout for research
// 	constraints.SuccessThreshold = 0.6          // Lower threshold for exploratory research

// 	return constraints
// }

// Validate checks if the batch constraints are valid
func (bc *BatchConstraints) Validate() error {
	// Validate embedded constraints first
	if err := bc.ToolConstraints.Validate(); err != nil {
		return fmt.Errorf("base constraints validation failed: %w", err)
	}

	// Validate batch-specific constraints
	if bc.MaxIterations <= 0 {
		return fmt.Errorf("max_iterations must be greater than 0")
	}

	if bc.MaxJobsPerBatch <= 0 {
		return fmt.Errorf("max_jobs_per_batch must be greater than 0")
	}

	if bc.BatchTimeout <= 0 {
		return fmt.Errorf("batch_timeout must be greater than 0")
	}

	if bc.IterationTimeout <= 0 {
		return fmt.Errorf("iteration_timeout must be greater than 0")
	}

	if bc.BatchTimeout < bc.IterationTimeout {
		return fmt.Errorf("batch_timeout must be greater than iteration_timeout")
	}

	if bc.RetryAttempts < 0 {
		return fmt.Errorf("retry_attempts cannot be negative")
	}

	if bc.ConcurrentJobs <= 0 {
		return fmt.Errorf("concurrent_jobs must be greater than 0")
	}

	if bc.SuccessThreshold < 0.0 || bc.SuccessThreshold > 1.0 {
		return fmt.Errorf("success_threshold must be between 0.0 and 1.0")
	}

	return nil
}

// IsEarlyStopTriggered checks if early stopping should be triggered based on results
func (bc *BatchConstraints) IsEarlyStopTriggered(successRate float64, currentIteration int) bool {
	if !bc.EnableEarlyStop {
		return false
	}

	// Stop if success rate is too low after first iteration
	if currentIteration > 0 && successRate < bc.SuccessThreshold {
		return true
	}

	return false
}

// ShouldRetry determines if a failed job should be retried
func (bc *BatchConstraints) ShouldRetry(attemptCount int) bool {
	return attemptCount < bc.RetryAttempts
}
