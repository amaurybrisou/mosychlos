// internal/llm/openai/strategy_responses.go
package openai

import (
	"context"
	"errors"

	"github.com/amaurybrisou/mosychlos/internal/config"
	toolsruntime "github.com/amaurybrisou/mosychlos/internal/llm/tools_runtime"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// ResponsesStrategy routes sync/stream calls to a models.Provider (Responses API)
// and augments tools with built-in hosted tools (web_search) when enabled in cfg.
type ResponsesStrategy struct {
	provider     models.Provider
	cfg          *config.LLMConfig
	toolRegistry map[keys.Key]models.Tool
	consumer     models.ToolConsumer
}

func NewResponsesStrategy(p models.Provider, cfg *config.LLMConfig) *ResponsesStrategy {
	return &ResponsesStrategy{
		provider:     p,
		cfg:          cfg,
		toolRegistry: map[keys.Key]models.Tool{},
	}
}

func (s *ResponsesStrategy) Name() string { return "openai-responses" }

func (s *ResponsesStrategy) RegisterTool(t models.Tool)            { s.toolRegistry[t.Key()] = t }
func (s *ResponsesStrategy) SetToolConsumer(c models.ToolConsumer) { s.consumer = c }

func (s *ResponsesStrategy) Ask(ctx context.Context, req models.PromptRequest) (*models.LLMResponse, error) {
	if s.provider == nil {
		return nil, errors.New("provider not set")
	}
	sess := s.provider.NewSession()

	for _, m := range req.Messages { // []map[string]any
		role, _ := m["role"].(string)
		content, _ := m["content"].(string)
		if role != "" && content != "" {
			sess.Add(models.Role(role), content)
		}
	}

	// gather function tools from registry
	funcToolDefs := make([]models.ToolDef, 0, len(s.toolRegistry))
	for _, t := range s.toolRegistry {
		funcToolDefs = append(funcToolDefs, t.Definition())
	}

	// optional built-in web_search_preview tool
	var extraTools []any
	if s.cfg != nil && s.cfg.OpenAI.WebSearch {
		ws := map[string]any{"type": keys.WebSearch}
		// Context size hint if present: "low" | "medium" | "high"
		if sz := s.cfg.OpenAI.WebSearchContextSize; sz != "" {
			ws[keys.WebSearch.String()] = map[string]any{"context_size": sz}
		}

		if s.cfg.OpenAI.WebSearchUserLocation.Country != nil {
			ws["country"] = *s.cfg.OpenAI.WebSearchUserLocation.Country
		}
		if s.cfg.OpenAI.WebSearchUserLocation.City != nil {
			ws["city"] = *s.cfg.OpenAI.WebSearchUserLocation.City
		}
		if s.cfg.OpenAI.WebSearchUserLocation.Region != nil {
			ws["region"] = *s.cfg.OpenAI.WebSearchUserLocation.Region
		}
		if s.cfg.OpenAI.WebSearchUserLocation.Timezone != nil {
			ws["timezone"] = *s.cfg.OpenAI.WebSearchUserLocation.Timezone
		}
		extraTools = append(extraTools, ws)
	}

	turn, err := toolsruntime.RunConversation(
		ctx,
		sess,
		s.provider,
		funcToolDefs, // function tools (your own tools)
		extraTools,   // built-in hosted tools (web_search)
		s.toolRegistry,
		s.consumer,
		toolsruntime.Options{MaxRounds: 6},
		req.ResponseFormat,
	)
	if err != nil {
		return nil, err
	}
	return &models.LLMResponse{Model: req.Model, Content: turn.Content}, nil
}

func (s *ResponsesStrategy) AskStream(ctx context.Context, req models.PromptRequest) (<-chan models.StreamChunk, error) {
	sess := s.provider.NewSession()
	for _, m := range req.Messages {
		role, _ := m["role"].(string)
		content, _ := m["content"].(string)
		if role != "" && content != "" {
			sess.Add(models.Role(role), content)
		}
	}
	// NOTE: streaming + tools requires parsing streamed tool frames; start with no tools:
	return sess.NextStream(ctx, nil, req.ResponseFormat)
}
