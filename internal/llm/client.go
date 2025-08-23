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
	config       *config.LLMConfig
	batchManager models.BatchManager
	syncProvider models.Provider

	toolRegistry map[bag.Key]models.Tool
	consumer     models.ToolConsumer

	responsesStrat ResponsesStrategyInterface
	chatStrat      ChatStrategyInterface // optional
}

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
	doer := pkgopenai.NewHTTPClient()
	po := pkgopenai.NewClient(doer, cfg.LLM.OpenAI)
	respProvider := llmopenai.NewProvider(po, cfg.LLM, sharedBag)
	responsesStrat := llmopenai.NewResponsesStrategy(respProvider, &cfg.LLM)

	return &Client{
		config:         &cfg.LLM,
		batchManager:   batchManager,
		syncProvider:   respProvider,
		responsesStrat: responsesStrat,
		toolRegistry:   map[bag.Key]models.Tool{},
	}, nil
}

func (c *Client) RegisterTool(t ...models.Tool) {
	for _, tool := range t {
		c.toolRegistry[tool.Key()] = tool
		// also make them visible to the Responses strategy
		c.responsesStrat.RegisterTool(tool)
	}
}
func (c *Client) SetToolConsumer(tc models.ToolConsumer) {
	c.consumer = tc
	c.responsesStrat.SetToolConsumer(tc)
}

func (c *Client) DoSync(ctx context.Context, req models.PromptRequest) (*models.LLMResponse, error) {
	if c.responsesStrat == nil {
		return nil, errors.New("provider not set")
	}
	return c.responsesStrat.Ask(ctx, req)
}
func (c *Client) Ask(ctx context.Context, req models.PromptRequest) (*models.LLMResponse, error) {
	if c.responsesStrat == nil {
		return nil, errors.New("provider not set")
	}
	return c.responsesStrat.Ask(ctx, req)
}
func (c *Client) AskStream(ctx context.Context, req models.PromptRequest) (<-chan models.StreamChunk, error) {
	if c.responsesStrat == nil {
		return nil, errors.New("provider not set")
	}
	return c.responsesStrat.AskStream(ctx, req)
}

func (c *Client) DoBatch(ctx context.Context, reqs []models.PromptRequest) (*models.BatchJob, error) {
	// Convert PromptRequests to BatchRequests
	batchRequests := make([]models.BatchRequest, len(reqs))

	for i, req := range reqs {
		model := req.Model
		if model == "" {
			model = c.config.Model.String()
		}

		IsReasoningModel := llmutils.IsReasoningModel(model)

		endpoint := "/v1/chat/completions"
		if IsReasoningModel {
			endpoint = "/v1/responses"
		}

		// Build request body based on model type
		var body map[string]any

		if IsReasoningModel {
			// For reasoning models using /v1/responses endpoint
			body = map[string]any{
				"model": model,
				"input": req.Messages, // Use 'input' instead of 'messages' for reasoning models
			}

			if req.MaxTokens > 0 {
				body["max_completion_tokens"] = req.MaxTokens // Different parameter name
			}
			// Note: reasoning models don't support temperature, tools, etc.
		} else {
			// For standard models using /v1/chat/completions endpoint
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
				// Convert tools for batch compatibility with /v1/chat/completions
				batchTools := make([]any, 0, len(req.Tools))
				for _, toolDef := range req.Tools {
					// For batch requests, convert all tools to function format
					switch t := toolDef.(type) {
					case *models.FunctionToolDef:
						batchTools = append(batchTools, t)
					case *models.CustomToolDef:
						// Convert custom tools to function tools for batch requests
						functionTool := &models.FunctionToolDef{
							Type: models.FunctionToolDefType,
							Function: models.FunctionDef{
								Name:        t.Name,
								Description: t.Description,
								Parameters:  t.Parameters,
							},
						}
						batchTools = append(batchTools, functionTool)
					default:
						// Fallback - try to use as is
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
			CustomID: req.CustomID, // Use the CustomID from the request
			Method:   http.MethodPost,
			URL:      endpoint,
			Body:     body,
		}
	}

	slog.Info("Submitting batch with requests", "count", len(batchRequests))

	// Use batch manager to process
	return c.batchManager.ProcessBatch(ctx, batchRequests, models.BatchOptions{}, false)
}

func (c *Client) BatchManager() models.BatchManager { return c.batchManager }
