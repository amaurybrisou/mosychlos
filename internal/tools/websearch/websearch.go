package websearch

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// Provider implements a virtual tool for OpenAI's internal web_search_preview
// This tool is executed by OpenAI internally, not by our system
type Provider struct {
	sharedBag bag.SharedBag
}

var _ models.Tool = &Provider{}

// NewFromConfig creates a new web search provider
// Since this is an internal OpenAI tool, no specific configuration is needed
// The web search behavior is controlled by OpenAI's configuration
func NewFromConfig(cfg *config.OpenAIConfig, sharedBag bag.SharedBag) (*Provider, error) {
	return New(sharedBag)
}

// New constructs a web search provider
func New(sharedBag bag.SharedBag) (*Provider, error) {
	return &Provider{
		sharedBag: sharedBag,
	}, nil
}

// Name returns the exact tool name that OpenAI expects
func (p *Provider) Name() string {
	return bag.WebSearch.String()
}

// Key returns the unique tool key
func (p *Provider) Key() bag.Key {
	return bag.WebSearch
}

// Description returns the tool description for AI function calling
func (p *Provider) Description() string {
	return "Search the web for current market information, news, analysis, and real-time data to complement portfolio analysis"
}

// Definition returns the OpenAI tool definition
func (p *Provider) Definition() models.ToolDef {
	return &models.CustomToolDef{
		Type: models.CustomToolDefType,
		FunctionDef: models.FunctionDef{
			Name:        p.Name(),
			Description: p.Description(),
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "Search query for current market conditions, news, or analysis",
					},
				},
				"required": []string{"query"},
			},
		},
	}
}

// Tags returns tool tags for categorization
func (p *Provider) Tags() []string {
	return []string{"web", "search", "market", "news", "real-time", "openai-internal"}
}

// IsExternal indicates whether the tool is external
func (p *Provider) IsExternal() bool {
	return true
}

// Run executes the tool with the given arguments
func (p *Provider) Run(ctx context.Context, args string) (string, error) {
	// Parse arguments for logging/debugging purposes
	var params struct {
		Query string `json:"query"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse web search arguments: %w", err)
	}

	slog.Debug("Processing web search citations", "query", params.Query)

	// Note: In actual implementation, web search results would be provided by OpenAI
	// real result processing will be done at the session level

	// Log that citation processing infrastructure is ready
	slog.Info("Web search citation processing ready", "query", params.Query)

	return "Web search citation processing initialized", nil
}
