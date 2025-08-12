// internal/llm/router.go
package llm

import (
	"context"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

type Strategy interface {
	Ask(ctx context.Context, req models.PromptRequest) (*models.LLMResponse, error)
	AskStream(ctx context.Context, req models.PromptRequest) (<-chan models.StreamChunk, error)
	Name() string
}
