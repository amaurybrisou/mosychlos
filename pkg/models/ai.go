package models

//go:generate mockgen -source=ai.go -destination=mocks/mock_ai.go -package=mocks

import (
	"context"
	"fmt"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleDeveloper Role = "developer"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"

	FunctionToolDefType = "function"
	CustomToolDefType   = "custom"
)

type ToolDef interface {
	ToAny() any
}

type FunctionDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type FunctionToolDef struct {
	Type     string      `json:"type" default:"function"`
	Function FunctionDef `json:"function"`
}

func (f *FunctionToolDef) ToAny() any {
	return f
}

type CustomToolDef struct {
	Type        string `json:"type" default:"custom"`
	FunctionDef `json:",inline"`
}

func (c *CustomToolDef) ToAny() any {
	return c
}

type Tool interface {
	Name() string
	Key() keys.Key
	Description() string
	Definition() ToolDef
	Tags() []string
	IsExternal() bool
	Run(ctx context.Context, args string) (string, error)
}

type Format struct {
	Type      keys.Key       `json:"type"` // e.g. "json_schema"
	Name      string         `json:"name"`
	Schema    map[string]any `json:"schema"`
	Verbosity string         `json:"verbosity"`
}

type ResponseFormat struct {
	Format Format `json:"format,omitempty"`
}

type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolCall struct {
	ID       string           `json:"id"`
	CallID   string           `json:"call_id,omitempty"` // Internal field, not sent to API
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
	Status   string           `json:"status,omitempty"` // e.g. "completed", "failed"
}

type ToolChoiceFunction struct {
	Name string `json:"name"`
}

type ToolChoice struct {
	Name     string             `json:"name"`
	Type     string             `json:"type"`
	Function ToolChoiceFunction `json:"function"`
}

type AssistantTurn struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"function_call,omitempty"`
}

// StreamChunk represents a chunk of streaming response
type StreamChunk struct {
	Content      string     `json:"content,omitempty"`
	ToolCalls    []ToolCall `json:"function_call,omitempty"`
	IsComplete   bool       `json:"is_complete"`
	Error        error      `json:"error,omitempty"`
	Usage        *Usage     `json:"usage,omitempty"`
	FinishReason *string    `json:"finish_reason,omitempty"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	InputTokens      int `json:"input_tokens"`
	OutputTokens     int `json:"output_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type RoleContent interface {
	GetRole() Role
	GetContent() string
}

type Session interface {
	Add(role Role, content string)
	AddToolResult(toolCall ToolCall, content string)
	AddFunctionCallResult(toolCall ToolCall, content string)
	Next(ctx context.Context, tools []ToolDef, rf *ResponseFormat) (*AssistantTurn, error)
	NextStream(ctx context.Context, tools []ToolDef, rf *ResponseFormat) (<-chan StreamChunk, error)
	SetToolChoice(t *ToolChoice)
}

type Provider interface {
	Name() string
	NewSession() Session
	Embedding(ctx context.Context, text string) ([]float64, error)
}

type AiClient interface {
	DoSync(ctx context.Context, req PromptRequest) (*LLMResponse, error)
	Ask(ctx context.Context, req PromptRequest) (*LLMResponse, error)
	AskStream(ctx context.Context, req PromptRequest) (<-chan StreamChunk, error)
	DoBatch(ctx context.Context, reqs []PromptRequest) (*BatchJob, error)

	RegisterTool(t ...Tool)
	SetToolConsumer(consumer ToolConsumer)
	BatchManager() BatchManager
}

// ToolConstructor is a function that creates a tool from config
type ToolConstructor func(cfg any, sharedBag bag.SharedBag) (Tool, error)

// ToolsRateLimit holds rate limiting configuration
type ToolsRateLimit struct {
	RequestsPerSecond int
	RequestsPerDay    int
	Burst             int
}

// Validate validates the ToolsRateLimit configuration
func (trl *ToolsRateLimit) Validate() error {
	// RequestsPerSecond must be non-negative
	if trl.RequestsPerSecond < 0 {
		return fmt.Errorf("RequestsPerSecond must be non-negative, got: %d", trl.RequestsPerSecond)
	}

	// RequestsPerDay must be non-negative
	if trl.RequestsPerDay < 0 {
		return fmt.Errorf("RequestsPerDay must be non-negative, got: %d", trl.RequestsPerDay)
	}

	// Burst must be non-negative
	if trl.Burst < 0 {
		return fmt.Errorf("Burst must be non-negative, got: %d", trl.Burst)
	}

	// Burst should not exceed RequestsPerSecond if both are positive
	if trl.RequestsPerSecond > 0 && trl.Burst > 0 && trl.Burst > trl.RequestsPerSecond*10 {
		return fmt.Errorf("Burst (%d) should not be excessively larger than RequestsPerSecond (%d)", trl.Burst, trl.RequestsPerSecond)
	}

	// If RequestsPerDay is set, it should be reasonable compared to RequestsPerSecond
	if trl.RequestsPerSecond > 0 && trl.RequestsPerDay > 0 {
		dailyLimit := trl.RequestsPerSecond * 86400 // seconds in a day
		if trl.RequestsPerDay > dailyLimit*2 {
			return fmt.Errorf("RequestsPerDay (%d) is inconsistent with RequestsPerSecond (%d)", trl.RequestsPerDay, trl.RequestsPerSecond)
		}
	}

	return nil
}

type ChatMessage struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// ToolConfig holds configuration for a tool
type ToolConfig struct {
	Key          keys.Key
	Constructor  ToolConstructor
	Config       any
	CacheEnabled bool
	CacheTTL     time.Duration
	RateLimit    *ToolsRateLimit
}

// Validate validates the ToolConfig
func (tc *ToolConfig) Validate() error {
	// Key cannot be empty
	if tc.Key == "" {
		return fmt.Errorf("Key cannot be empty")
	}

	// Constructor is required
	if tc.Constructor == nil {
		return fmt.Errorf("Constructor cannot be nil")
	}

	// CacheTTL must be non-negative
	if tc.CacheTTL < 0 {
		return fmt.Errorf("CacheTTL must be non-negative, got: %v", tc.CacheTTL)
	}

	// Validate rate limit if provided
	if tc.RateLimit != nil {
		if err := tc.RateLimit.Validate(); err != nil {
			return fmt.Errorf("RateLimit validation failed: %w", err)
		}
	}

	return nil
}
