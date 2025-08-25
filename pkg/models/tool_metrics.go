package models

import "time"

// ToolComputation represents a single tool execution with full context
type ToolComputation struct {
	ToolName     string        `json:"tool_name"`
	CallID       string        `json:"call_id,omitempty"`
	Arguments    any           `json:"arguments"`
	Result       any           `json:"result"`
	StartTime    time.Time     `json:"start_time"`
	Duration     time.Duration `json:"duration"`
	Success      bool          `json:"success"`
	Error        string        `json:"error,omitempty"`
	TokensUsed   int           `json:"tokens_used,omitempty"`
	Cost         float64       `json:"cost,omitempty"`
	DataConsumed []string      `json:"data_consumed,omitempty"` // Bag keys read
	DataProduced []string      `json:"data_produced,omitempty"` // Bag keys written
}

// ToolMetrics holds aggregated metrics for tool usage
type ToolMetrics struct {
	TotalCalls      int                  `json:"total_calls"`
	TotalDuration   time.Duration        `json:"total_duration"`
	TotalCost       float64              `json:"total_cost"`
	TotalTokens     int                  `json:"total_tokens"`
	SuccessCount    int                  `json:"success_count"`
	ErrorCount      int                  `json:"error_count"`
	SuccessRate     float64              `json:"success_rate"`
	AverageDuration time.Duration        `json:"average_duration"`
	ByTool          map[string]ToolStats `json:"by_tool"`
	Errors          []string             `json:"errors,omitempty"`
	LastUpdated     time.Time            `json:"last_updated"`
}

// ToolStats holds per-tool statistics
type ToolStats struct {
	Calls           int           `json:"calls"`
	Duration        time.Duration `json:"total_duration"`
	AverageDuration time.Duration `json:"average_duration"`
	Successes       int           `json:"successes"`
	Errors          int           `json:"errors"`
	Cost            float64       `json:"total_cost"`
	Tokens          int           `json:"total_tokens"`
}
