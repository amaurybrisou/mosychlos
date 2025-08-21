// pkg/models/ai_batch.go
package models

import (
	"context"
	"io"
	"time"
)

//go:generate mockgen -source=ai_batch.go -destination=mocks/ai_batch_mock.go -package=mocks

// PromptRequest represents a request for AI processing
type PromptRequest struct {
	Model          string            `json:"model"`
	Messages       []map[string]any  `json:"messages"`
	MaxTokens      int               `json:"max_tokens,omitempty"`
	Temperature    *float64          `json:"temperature,omitempty"`
	Tools          []ToolDef         `json:"tools,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	ResponseFormat *ResponseFormat   `json:"response_format,omitempty"` // optional, for structured responses
	CustomID       string            `json:"custom_id,omitempty"`       // for batch processing
}

// LLMResponse represents a response from AI processing
type LLMResponse struct {
	ID             string            `json:"id"`
	Model          string            `json:"model"`
	Content        string            `json:"content"`
	Usage          *Usage            `json:"usage,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	ResponseFormat *ResponseFormat   `json:"response_format,omitempty"`
}

type ModelClass string

const (
	ModelClassStandard  ModelClass = "standard"
	ModelClassReasoning ModelClass = "reasoning"
)

type BatchStatus string

const (
	BatchStatusValidating BatchStatus = "validating"
	BatchStatusFailed     BatchStatus = "failed"
	BatchStatusInProgress BatchStatus = "in_progress"
	BatchStatusFinalizing BatchStatus = "finalizing"
	BatchStatusCompleted  BatchStatus = "completed"
	BatchStatusExpired    BatchStatus = "expired"
	BatchStatusCancelled  BatchStatus = "cancelled"
)

type BatchRequest struct {
	CustomID string         `json:"custom_id"`
	Method   string         `json:"method"` // "POST"
	URL      string         `json:"url"`    // "/v1/chat/completions" or "/v1/responses"
	Body     map[string]any `json:"body"`
	// ModelClass ModelClass     `json:"-"` // Internal use only - exclude from JSON
}

type BatchOptions struct {
	CompletionWindow string            `json:"completion_window"` // e.g. "24h"
	Metadata         map[string]string `json:"metadata,omitempty"`
	Priority         string            `json:"priority,omitempty"` // "low|normal|high"
	CostOptimize     bool              `json:"cost_optimize"`
	ModelClass       ModelClass        `json:"model_class,omitempty"`
}

type BatchJob struct {
	ID            string      `json:"id"`
	Status        BatchStatus `json:"status"`
	InputFileID   string      `json:"input_file_id"`
	OutputFileID  *string     `json:"output_file_id"`
	ErrorFileID   *string     `json:"error_file_id"`
	CreatedAt     int64       `json:"created_at_unix"`
	CompletedAt   *int64      `json:"completed_at_unix"`
	RequestCounts struct {
		Total     int `json:"total"`
		Completed int `json:"completed"`
		Failed    int `json:"failed"`
	} `json:"request_counts"`
	Metadata     map[string]string `json:"metadata"`
	CostEstimate *CostEstimate     `json:"cost_estimate,omitempty"`

	Request  PromptRequest    `json:"request"`
	CustomID string           `json:"custom_id"`
	Messages []map[string]any `json:"messages"`
}

type CostEstimate struct {
	EstimatedCost      float64 `json:"estimated_cost"`
	SavingsVsSync      float64 `json:"savings_vs_sync"`
	EstimatedTokensIn  int     `json:"estimated_tokens_in"`
	EstimatedTokensOut int     `json:"estimated_tokens_out"`
}

type AiBatchClient interface {
	SubmitBatch(ctx context.Context, reqs []BatchRequest, opts BatchOptions) (*BatchJob, error)
	GetBatchStatus(ctx context.Context, jobID string) (*BatchJob, error)
	GetBatchResults(ctx context.Context, jobID string) (io.ReadCloser, error) // stream success JSONL
	GetBatchErrors(ctx context.Context, jobID string) (io.ReadCloser, error)  // stream error JSONL
	CancelBatch(ctx context.Context, jobID string) error
	ListBatches(ctx context.Context, filters map[string]string) ([]BatchJob, error)
}

type BatchManager interface {
	SetPollDelay(delay time.Duration)
	EstimateCost(requests []BatchRequest) *CostEstimate
	ProcessBatch(ctx context.Context, requests []BatchRequest, opts BatchOptions, waitForCompletion bool) (*BatchJob, error)
	GetJobStatus(ctx context.Context, jobID string) (*BatchJob, error)
	WaitForCompletion(ctx context.Context, jobID string) (*BatchJob, error)
	GetResults(ctx context.Context, jobID string) (*BatchResult, error)
	CancelJob(ctx context.Context, jobID string) error
	ListBatches(ctx context.Context, filters map[string]string) ([]BatchJob, error)
	GetError(ctx context.Context, jobID string) (map[string]string, error)
}

// BatchResult represents the result of aggregating batch processing results
type BatchResult struct {
	JobID     string                `json:"job_id"`
	Successes int                   `json:"successes"`
	Failures  int                   `json:"failures"`
	Items     map[string]any        `json:"items"`      // by CustomID: final content
	Errors    map[string]string     `json:"errors"`     // by CustomID: raw error JSON
	ToolCalls map[string][]ToolCall `json:"tool_calls"` // by CustomID: tool calls if any
	Content   map[string]string     `json:"content"`    // by CustomID: text content
	Usage     map[string]BatchUsage `json:"usage"`      // by CustomID: token usage stats
}

// BatchUsage represents token usage statistics from OpenAI Batch API responses
// This is separate from the regular Usage type since batch API responses
// may have different field structures than real-time API responses
type BatchUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
