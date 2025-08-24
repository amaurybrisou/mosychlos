// internal/llm/client.go
package llm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/amaurybrisou/mosychlos/internal/config"
	llmutils "github.com/amaurybrisou/mosychlos/internal/llm/llm_utils"
	llmopenai "github.com/amaurybrisou/mosychlos/internal/llm/openai"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/fs"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	pkgopenai "github.com/amaurybrisou/mosychlos/pkg/openai"
)

type Client struct {
	config *config.LLMConfig

	// Batch path (unchanged)
	batchManager models.BatchManager

	// Sync path (Responses)
	engine   *llmopenai.Engine   // pure Responses API transport
	runner   *llmopenai.Runner   // SDK-like loop (create → tools → submit → continue)
	provider *llmopenai.Provider // holds cfg + shared bag

	toolRegistry map[bag.Key]models.Tool
	consumer     models.ToolConsumer
}

// NewLLMClient wires:
//   - Batch manager (existing)
//   - Responses Engine (low-level HTTP)
//   - Runner (Agent-style loop that uses the Engine)
func NewLLMClient(cfg *config.Config, sharedBag bag.SharedBag) (*Client, error) {
	if cfg.LLM.Provider == "" {
		return nil, fmt.Errorf("LLM provider not configured")
	}
	if sharedBag == nil {
		return nil, fmt.Errorf("shared bag not provided")
	}

	// --- Batch path (existing) ---
	filesystem := fs.OS{}
	factory := NewBatchServiceFactory(cfg, filesystem, sharedBag)
	batchManager, err := factory.CreateManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create batch manager: %w", err)
	}

	// --- Sync path (Responses) ---
	doer := pkgopenai.NewHTTPClient() // *http.Client with sane defaults
	po := pkgopenai.NewClient(doer, cfg.LLM.OpenAI)

	provider := llmopenai.NewProvider(po, cfg.LLM, sharedBag)
	engine := llmopenai.NewEngine(po, cfg.LLM)      // pure transport
	runner := llmopenai.NewRunner(engine, provider) // agent loop

	return &Client{
		config:       &cfg.LLM,
		batchManager: batchManager,
		engine:       engine,
		runner:       runner,
		provider:     provider,
		toolRegistry: map[bag.Key]models.Tool{},
	}, nil
}

func (c *Client) RegisterTool(t ...models.Tool) {
	for _, tool := range t {
		c.toolRegistry[tool.Key()] = tool
	}
	// Make them visible to the runner/provider
	c.runner.RegisterTool(t...)
	c.provider.RegisterTool(t...)
}

func (c *Client) SetToolConsumer(tc models.ToolConsumer) {
	c.consumer = tc
	c.runner.SetToolConsumer(tc)
	c.provider.SetToolConsumer(tc)
}

// Ask is the SDK-like sync run (mirrors Python Runner.run).
func (c *Client) Ask(ctx context.Context, req models.PromptRequest) (*models.LLMResponse, error) {
	if c.runner == nil {
		return nil, errors.New("responses runner not initialized")
	}
	return c.runner.Run(ctx, req)
}

// AskStream streams tokens (if your Engine enables streaming later).
func (c *Client) AskStream(ctx context.Context, req models.PromptRequest) (<-chan models.StreamChunk, error) {
	if c.runner == nil {
		return nil, errors.New("responses runner not initialized")
	}
	// return c.runner.Run(ctx, req)
	return nil, nil
}

// DoSync kept for back-compat (alias of Ask).
func (c *Client) DoSync(ctx context.Context, req models.PromptRequest) (*models.LLMResponse, error) {
	return c.Ask(ctx, req)
}

// Batch (unchanged; small cleanup).
func (c *Client) DoBatch(ctx context.Context, reqs []models.PromptRequest) (*models.BatchJob, error) {
	if c.batchManager == nil {
		return nil, errors.New("batch manager not set")
	}

	batchRequests := make([]models.BatchRequest, len(reqs))
	for i, req := range reqs {
		model := req.Model
		if model == "" {
			model = c.config.Model.String()
		}
		isReasoning := llmutils.IsReasoningModel(model)

		endpoint := "/v1/chat/completions"
		if isReasoning {
			endpoint = "/v1/responses"
		}

		body := map[string]any{}
		if isReasoning {
			body = map[string]any{
				"model": model,
				"input": req.Messages, // Responses API uses "input"
			}
			if req.MaxTokens > 0 {
				body["max_output_tokens"] = req.MaxTokens
			}
		} else {
			body = map[string]any{
				"model":    model,
				"messages": req.Messages,
			}
			if req.MaxTokens > 0 {
				body["max_tokens"] = req.MaxTokens
			}
			if req.Temperature != nil {
				body["temperature"] = *req.Temperature
			}
			if req.Tools != nil {
				batchTools := make([]any, 0, len(req.Tools))
				for _, toolDef := range req.Tools {
					switch t := toolDef.(type) {
					case *models.FunctionToolDef:
						batchTools = append(batchTools, t)
					case *models.CustomToolDef:
						batchTools = append(batchTools, &models.FunctionToolDef{
							Type: models.FunctionToolDefType,
							Function: models.FunctionDef{
								Name:        t.Name,
								Description: t.Description,
								Parameters:  t.Parameters,
							},
						})
					default:
						batchTools = append(batchTools, toolDef)
					}
				}
				body["tools"] = batchTools
				slog.Debug("Adding tools to batch request", "tool_count", len(batchTools))
			}
			if req.ResponseFormat != nil {
				body["response_format"] = req.ResponseFormat
			}
		}
		if req.Metadata != nil {
			body["metadata"] = req.Metadata
		}
		batchRequests[i] = models.BatchRequest{
			CustomID: req.CustomID,
			Method:   http.MethodPost,
			URL:      endpoint,
			Body:     body,
		}
	}

	slog.Info("Submitting batch with requests", "count", len(batchRequests))
	return c.batchManager.ProcessBatch(ctx, batchRequests, models.BatchOptions{}, false)
}

func (c *Client) BatchManager() models.BatchManager { return c.batchManager }
