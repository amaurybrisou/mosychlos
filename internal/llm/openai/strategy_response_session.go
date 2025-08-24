// internal/llm/openai/strategy_response_session.go
// Completely rewritten for the Responses API flow (create → tool calls → previous_response_id + function_call_output → loop).
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	llmutils "github.com/amaurybrisou/mosychlos/internal/llm/llm_utils"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	pkgopenai "github.com/amaurybrisou/mosychlos/pkg/openai"
	"github.com/openai/openai-go/v2/responses"
)

// ------------------------------ Provider (business-side) ------------------------------

type Provider struct {
	name      string
	cli       *pkgopenai.Client
	cfg       config.LLMConfig
	sharedBag bag.SharedBag

	consumer  models.ToolConsumer     // optional; if you have one
	toolByKey map[bag.Key]models.Tool // registered tools by name/key
}

func NewProvider(cli *pkgopenai.Client, cfg config.LLMConfig, sharedBag bag.SharedBag) *Provider {
	return &Provider{
		name:      "openai-responses",
		cli:       cli,
		cfg:       cfg,
		sharedBag: sharedBag,
		toolByKey: map[bag.Key]models.Tool{},
	}
}

func (p *Provider) Name() string                                                  { return p.name }
func (p *Provider) Embedding(ctx context.Context, text string) ([]float64, error) { return nil, nil }
func (p *Provider) RegisterTool(t ...models.Tool) {
	for _, tool := range t {
		p.toolByKey[tool.Key()] = tool
	}
}
func (p *Provider) SetToolConsumer(tc models.ToolConsumer) { p.consumer = tc }

// ------------------------------ Engine (pure transport) ------------------------------
//
// Engine does the HTTP JSON for /v1/responses “create” and “continue” (with previous_response_id).
// It does not know about your business logic or tool execution.

type Engine struct {
	http    *pkgopenai.Client
	cfg     config.LLMConfig
	baseURL string
}

func NewEngine(httpClient *pkgopenai.Client, cfg config.LLMConfig) *Engine {
	base := cfg.BaseURL
	if base == "" {
		base = "https://api.openai.com"
	}
	return &Engine{http: httpClient, cfg: cfg, baseURL: normalizeBase(base)}
}

func normalizeBase(base string) string {
	base = strings.TrimRight(base, "/")
	if strings.HasSuffix(base, "/v1") {
		base = strings.TrimSuffix(base, "/v1")
	}
	return base
}
func (e *Engine) build(path string) string {
	path = strings.TrimLeft(path, "/")
	return e.baseURL + "/v1/" + path
}

// ----- request bodies (explicit JSON; independent from SDK helpers) -----

type createReq struct {
	Model             string   `json:"model"`
	Input             any      `json:"input,omitempty"` // string or []messages
	Tools             []any    `json:"tools,omitempty"`
	MaxOutputTokens   int64    `json:"max_output_tokens,omitempty"`
	Temperature       *float64 `json:"temperature,omitempty"`
	ServiceTier       string   `json:"service_tier,omitempty"`
	ParallelToolCalls *bool    `json:"parallel_tool_calls,omitempty"`
	ResponseFormat    any      `json:"text,omitempty"`
	Metadata          any      `json:"metadata,omitempty"`
	Store             bool     `json:"store,omitempty"`
}

// function_call_output item for continuation
type funcCallOutputItem struct {
	Type   string `json:"type"`    // "function_call_output"
	CallID string `json:"call_id"` // the tool call's CallID
	Output string `json:"output"`  // stringified JSON or plain text
}

type continueReq struct {
	Model              string               `json:"model"`
	PreviousResponseID string               `json:"previous_response_id"`
	Input              []funcCallOutputItem `json:"input"`
}

func (e *Engine) doJSON(ctx context.Context, method, url string, body any) (*responses.Response, error) {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if e.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.cfg.APIKey)
	}
	if e.cfg.OpenAI.OrganizationID != "" {
		req.Header.Set("OpenAI-Organization", e.cfg.OpenAI.OrganizationID)
	}
	if e.cfg.OpenAI.ProjectID != "" {
		req.Header.Set("OpenAI-Project", e.cfg.OpenAI.ProjectID)
	}

	resp, err := e.http.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s %s -> %d: %s", method, url, resp.StatusCode, string(b))
	}

	var out responses.Response
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (e *Engine) Create(ctx context.Context, body createReq) (*responses.Response, error) {
	return e.doJSON(ctx, http.MethodPost, e.build("responses"), body)
}

// Continue: send the outputs of tool calls back to the model using previous_response_id
func (e *Engine) Continue(ctx context.Context, model string, previousResponseID string, outputs []funcCallOutputItem) (*responses.Response, error) {
	if previousResponseID == "" {
		return nil, fmt.Errorf("previousResponseID is required")
	}
	u := e.build("responses")
	body := continueReq{
		Model:              model,
		PreviousResponseID: url.PathEscape(previousResponseID),
		Input:              outputs,
	}
	return e.doJSON(ctx, http.MethodPost, u, body)
}

// ------------------------------ Runner (SDK-like loop) ------------------------------
//
// Mirrors the Python Agents SDK run loop using the Responses API chaining:
// 1) Create response
// 2) If tool calls present in resp.Output → run tools, send function_call_output + previous_response_id
// 3) Repeat until no tool calls remain, return final text.

type Runner struct {
	engine   *Engine
	provider *Provider
}

func NewRunner(engine *Engine, provider *Provider) *Runner {
	return &Runner{engine: engine, provider: provider}
}

func (r *Runner) RegisterTool(t ...models.Tool)          { r.provider.RegisterTool(t...) }
func (r *Runner) SetToolConsumer(tc models.ToolConsumer) { r.provider.SetToolConsumer(tc) }

func (r *Runner) Run(ctx context.Context, req models.PromptRequest) (*models.LLMResponse, error) {
	// Build create request
	isReasoning := llmutils.IsReasoningModel(r.provider.cfg.Model.String())
	create := createReq{
		Model: req.Model,
		Input: req.Messages, // your code already formats Responses "messages" style
	}
	if create.Model == "" {
		create.Model = r.provider.cfg.Model.String()
	}
	// knobs
	if max := r.provider.cfg.OpenAI.MaxCompletionTokens; max > 0 {
		create.MaxOutputTokens = max
	}
	if r.provider.cfg.OpenAI.Temperature != nil && !isReasoning {
		create.Temperature = r.provider.cfg.OpenAI.Temperature
	}
	if r.provider.cfg.OpenAI.ServiceTier != nil && *r.provider.cfg.OpenAI.ServiceTier != "auto" {
		create.ServiceTier = *r.provider.cfg.OpenAI.ServiceTier
	}
	if r.provider.cfg.OpenAI.ParallelToolCalls {
		t := true
		create.ParallelToolCalls = &t
	}
	if req.ResponseFormat != nil {
		create.ResponseFormat = req.ResponseFormat
	}
	if req.Tools != nil && len(req.Tools) > 0 {
		create.Tools = toAnyTools(req.Tools)
	}

	start := time.Now()

	var (
		last     *responses.Response
		err      error
		turns    int
		maxTurns = 32
	)
	for {
		turns++
		if turns > maxTurns {
			return nil, fmt.Errorf("max agent turns exceeded (%d)", maxTurns)
		}

		if last == nil {
			last, err = r.engine.Create(ctx, create)
			if err != nil {
				return nil, err
			}
		}

		// Parse model output for text + tool calls
		turn, err := processResponsesAPIResult(last, start)
		if err != nil {
			return nil, err
		}

		// If no tool calls → final
		if len(turn.ToolCalls) == 0 {
			return &models.LLMResponse{
				CreatedAt: time.Now(),
				Model:     create.Model,
				Content:   turn.Content,
				Usage:     &turn.Usage,
			}, nil
		}

		// There are tool calls: run each, then continue with previous_response_id
		items := make([]funcCallOutputItem, 0, len(turn.ToolCalls))
		for _, call := range turn.ToolCalls {
			// Find tool by name; fall back to consumer if provided
			key := bag.Key(call.Function.Name)
			var outStr string

			if tool, ok := r.provider.toolByKey[key]; ok {
				outStr, err = tool.Run(ctx, call.Function.Arguments)
				if err != nil {
					return nil, fmt.Errorf("tool %s failed: %w", call.Function.Name, err)
				}
				if r.provider.consumer != nil {
					// If your ToolConsumer returns a string, use that here.
					// Adjust if your interface differs.
					err = r.provider.consumer.ConsumeTools(ctx, bag.Key(call.Function.Name))
					if err != nil {
						return nil, fmt.Errorf("consumer failed for tool %s: %w", call.Function.Name, err)
					}
				}
			} else {
				return nil, fmt.Errorf("no tool or consumer registered for %s", call.Function.Name)
			}

			// Build function_call_output item
			items = append(items, funcCallOutputItem{
				Type:   "function_call_output",
				CallID: call.CallID, // IMPORTANT: use external CallID, not internal ID
				Output: outStr,      // JSON string or plain text
			})
		}

		// Continue the same response chain
		next, err := r.engine.Continue(ctx, create.Model, last.ID, items)
		if err != nil {
			return nil, err
		}
		last = next
		// loop continues
	}
}

// toAnyTools converts your ToolDef types into the Responses API function tool schema.
func toAnyTools(funcTools []models.ToolDef) []any {
	out := make([]any, 0, len(funcTools))
	for _, t := range funcTools {
		switch tool := t.(type) {
		case *models.CustomToolDef:
			out = append(out, map[string]any{
				"type":        "function",
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.Parameters,
			})
		case *models.FunctionToolDef:
			out = append(out, map[string]any{
				"type":        tool.Type, // "function"
				"name":        tool.Function.Name,
				"description": tool.Function.Description,
				"parameters":  tool.Function.Parameters,
			})
		default:
			out = append(out, tool)
		}
	}
	return out
}
