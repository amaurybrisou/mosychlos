package llm

import (
	"context"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// ResponsesStrategyInterface defines the interface for response strategies
type ResponsesStrategyInterface interface {
	RegisterTool(t models.Tool)
	SetToolConsumer(c models.ToolConsumer)
	Ask(ctx context.Context, req models.PromptRequest) (*models.LLMResponse, error)
	AskStream(ctx context.Context, req models.PromptRequest) (<-chan models.StreamChunk, error)
}
