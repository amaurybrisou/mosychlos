// internal/llm/openai/strategy_chat.go
package openai

import (
	"context"
	"errors"
	"net/http"

	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	pkgopenai "github.com/amaurybrisou/mosychlos/pkg/openai"
)

// ChatStrategy calls the legacy Chat Completions REST endpoint via pkg/openai.Client.
type ChatStrategy struct {
	cli     *pkgopenai.Client
	baseURL string
	model   string

	toolRegistry map[keys.Key]models.Tool
	consumer     models.ToolConsumer
}

func NewChatStrategy(cli *pkgopenai.Client, baseURL, model string) *ChatStrategy {
	return &ChatStrategy{
		cli:          cli,
		baseURL:      baseURL,
		model:        model,
		toolRegistry: make(map[keys.Key]models.Tool),
	}
}

func (s *ChatStrategy) Name() string { return "openai-chat" }

type chatReq struct {
	Model       string           `json:"model"`
	Messages    []map[string]any `json:"messages"`
	MaxTokens   *int             `json:"max_tokens,omitempty"`
	Temperature *float64         `json:"temperature,omitempty"`
	// You can add: TopP, PresencePenalty, FrequencyPenalty, Stop, etc.
}

type chatResp struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *models.Usage `json:"usage,omitempty"`
}

func (s *ChatStrategy) RegisterTool(t models.Tool)            { s.toolRegistry[t.Key()] = t }
func (s *ChatStrategy) SetToolConsumer(c models.ToolConsumer) { s.consumer = c }

func (s *ChatStrategy) Ask(ctx context.Context, req models.PromptRequest) (*models.LLMResponse, error) {
	// messages in PromptRequest are already []map[string]any with {role, content}
	body := chatReq{
		Model:    firstNonEmpty(req.Model, s.model),
		Messages: filterChatSupported(req.Messages), // drop unsupported roles like "tool"
	}
	if req.MaxTokens > 0 {
		body.MaxTokens = &req.MaxTokens
	}
	if req.Temperature != nil {
		body.Temperature = req.Temperature
	}

	var out chatResp
	// Headers: your pkg/openai middleware already injects Authorization, etc.
	_, err := s.cli.DoJSON(ctx, http.MethodPost, s.baseURL+"/v1/chat/completions", nil, body, &out)
	if err != nil {
		return nil, err
	}
	content := ""
	if len(out.Choices) > 0 {
		content = out.Choices[0].Message.Content
	}
	return &models.LLMResponse{
		Model:   body.Model,
		Content: content,
		Usage:   out.Usage,
	}, nil
}

func (s *ChatStrategy) AskStream(ctx context.Context, _ models.PromptRequest) (<-chan models.StreamChunk, error) {
	// Implement only if you really need chat streaming; otherwise keep Responses for streaming.
	return nil, errors.New("streaming not supported for OpenAI chat strategy")
}

// filterChatSupported removes roles not supported by /v1/chat/completions
func filterChatSupported(ms []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(ms))
	for _, m := range ms {
		role, _ := m["role"].(string)
		switch role {
		case "system", "user", "assistant":
			out = append(out, m)
		default:
			// ignore "tool" and any unknown roles for Chat Completions
		}
	}
	return out
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
