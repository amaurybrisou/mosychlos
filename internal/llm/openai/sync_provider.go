// internal/llm/openai/sync_provider.go
// File: internal/llm/openai/sync_provider.go
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	llmutils "github.com/amaurybrisou/mosychlos/internal/llm/llm_utils"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	pkgopenai "github.com/amaurybrisou/mosychlos/pkg/openai"
	"github.com/openai/openai-go/v2/responses"
)

type Provider struct {
	name      string
	cli       *pkgopenai.Client
	cfg       config.LLMConfig
	sharedBag bag.SharedBag
}

func NewProvider(cli *pkgopenai.Client, cfg config.LLMConfig, sharedBag bag.SharedBag) *Provider {
	return &Provider{name: "openai-responses", cli: cli, cfg: cfg, sharedBag: sharedBag}
}

func (p *Provider) Name() string                                                  { return p.name }
func (p *Provider) Embedding(ctx context.Context, text string) ([]float64, error) { return nil, nil }

func (p *Provider) NewSession() models.Session {
	return &session{p: p, messages: make([]map[string]any, 0, 8)}
}

type session struct {
	p          *Provider
	messages   []map[string]any // [{role, content}, ...]
	toolChoice *models.ToolChoice
}

func (s *session) Add(role models.Role, content string) {
	s.messages = append(s.messages, map[string]any{"role": string(role), "content": content})
}

func (s *session) AddToolResult(toolCallID, content string) {
	s.messages = append(s.messages, map[string]any{
		"role":         "tool",
		"tool_call_id": toolCallID,
		"content":      content,
	})
}

func (s *session) SetToolChoice(t *models.ToolChoice) { s.toolChoice = t }

// ---- Responses API payloads (minimal & correct) ----

type responsesReq struct {
	Model             string                 `json:"model"`
	Input             any                    `json:"input,omitempty"`             // string OR []message
	MaxOutputTokens   int64                  `json:"max_output_tokens,omitempty"` // omit when nil
	Temperature       *float64               `json:"temperature,omitempty"`
	Tools             []any                  `json:"tools,omitempty"`               // supports function tools and built-ins like {"type":"web_search"}
	ParallelToolCalls *bool                  `json:"parallel_tool_calls,omitempty"` // from cfg if you want
	ServiceTier       string                 `json:"service_tier,omitempty"`        // "auto"|"default"|"flex"|"priority"
	ResponseFormat    *models.ResponseFormat `json:"response_format,omitempty"`     // structured outputs
}

type responsesErr struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Param   string `json:"param,omitempty"`
		Code    string `json:"code,omitempty"`
	} `json:"error"`
}

func (s *session) Next(ctx context.Context, tools []models.ToolDef, rf *models.ResponseFormat) (*models.AssistantTurn, error) {
	body := responsesReq{
		Model: s.p.cfg.Model.String(),
		Input: s.messages,
		Tools: toAnyTools(tools, nil), // built-ins are appended by strategy via runtime (see runtime.go)
	}

	// map cfg → request
	if max := s.p.cfg.OpenAI.MaxCompletionTokens; max > 0 {
		body.MaxOutputTokens = max
	}
	if s.p.cfg.OpenAI.Temperature != nil && !llmutils.IsReasoningModel(s.p.cfg.Model.String()) {
		body.Temperature = s.p.cfg.OpenAI.Temperature
	}
	if rf != nil {
		body.ResponseFormat = rf
	}
	if s.p.cfg.OpenAI.ServiceTier != nil && *s.p.cfg.OpenAI.ServiceTier != "auto" {
		body.ServiceTier = *s.p.cfg.OpenAI.ServiceTier
	}
	if s.p.cfg.OpenAI.ParallelToolCalls {
		t := true
		body.ParallelToolCalls = &t
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}

	baseURL := s.p.cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	req, err := http.NewRequest(http.MethodPost, baseURL+"/v1/responses", &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if s.p.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+s.p.cfg.APIKey)
	}

	resp, err := s.p.cli.Do(ctx, req)
	if err != nil {
		slog.Error("responses request failed", slog.Any("err", err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp responsesErr
		err = json.NewDecoder(resp.Body).Decode(&errResp)
		if err == nil {
			return nil, fmt.Errorf("API error: %s (type: %s, param: %s, code: %s)", errResp.Error.Message, errResp.Error.Type, errResp.Error.Param, errResp.Error.Code)
		}
	}

	// TODO delete
	f, err := os.OpenFile("/home/amaury/Documents/mosychlos-v2/testdata/openai-response-api.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := io.TeeReader(resp.Body, f)

	// this dependency is intended since it has way too much body parsing helpers
	var out *responses.Response
	if err := json.NewDecoder(r).Decode(&out); err != nil {
		return nil, err
	}

	return s.processResponsesAPIResult(out, time.Now())
}

func toAnyTools(funcTools []models.ToolDef, extra []any) []any {
	out := make([]any, 0, len(funcTools)+len(extra))
	for _, t := range funcTools {
		out = append(out, t)
	}
	out = append(out, extra...)
	return out
}
