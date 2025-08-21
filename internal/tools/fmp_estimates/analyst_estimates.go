package fmpestimates

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/fmp"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// NewAnalystEstimatesFromConfig creates a new AnalystEstimatesProvider from config
func NewAnalystEstimatesFromConfig(cfg *config.FMPAnalystEstimatesConfig, sharedBag bag.SharedBag) (*FMPAnalystEstimatesProvider, error) {
	return NewAnalystEstimatesProvider(cfg.APIKey, cfg.CacheDir, sharedBag)
}

// FMPAnalystEstimatesProvider implements a Financial Modeling Prep analyst estimates source
type FMPAnalystEstimatesProvider struct {
	client    *fmp.Client
	cacheDir  string // cache directory for caching wrapper
	sharedBag bag.SharedBag
}

// Ensure AnalystEstimatesProvider implements models.Tool interface
var _ models.Tool = &FMPAnalystEstimatesProvider{}

func NewAnalystEstimatesProvider(key, cacheDir string, sharedBag bag.SharedBag) (*FMPAnalystEstimatesProvider, error) {
	if key == "" {
		return nil, fmt.Errorf("fmp_analyst_estimates: missing FMP_API_KEY")
	}

	// Create FMP client
	clientCfg := fmp.Config{
		APIKey:  key,
		BaseURL: "https://financialmodelingprep.com/api/v3",
		Timeout: 10 * time.Second,
	}

	client, err := fmp.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create FMP client: %w", err)
	}

	p := &FMPAnalystEstimatesProvider{
		client:    client,
		cacheDir:  cacheDir,
		sharedBag: sharedBag,
	}
	return p, nil
}

// models.Tool interface implementation

// Name returns the tool name
func (p *FMPAnalystEstimatesProvider) Name() string {
	return keys.FMPAnalystEstimates.String()
}

// Key returns the tool key
func (p *FMPAnalystEstimatesProvider) Key() keys.Key {
	return keys.FMPAnalystEstimates
}

// Tags returns the tool tags
func (p *FMPAnalystEstimatesProvider) Tags() []string {
	return []string{"finance", "fundamentals", "stocks", "analyst-estimates", "forecasts"}
}

// Description returns the tool description
func (p *FMPAnalystEstimatesProvider) Description() string {
	return "Fetches analyst estimates and projections for stocks including revenue, EBITDA, EPS forecasts, and analyst coverage from Financial Modeling Prep API"
}

func (t *FMPAnalystEstimatesProvider) IsExternal() bool { return false }

// Definition returns the tool definition for AI systems
func (p *FMPAnalystEstimatesProvider) Definition() models.ToolDef {
	return &models.CustomToolDef{
		Type: models.CustomToolDefType,
		FunctionDef: models.FunctionDef{
			Name:        p.Name(),
			Description: p.Description(),
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"tickers": map[string]any{
						"type":        "array",
						"description": "Array of stock ticker symbols to fetch analyst estimates for (e.g., ['AAPL', 'GOOGL', 'MSFT']), not for cryptocurrencies",
						"items": map[string]any{
							"type": "string",
						},
						"minItems": 1,
						"maxItems": 50, // Based on batch size limitation
					},
				},
				"required": []string{"tickers"},
			},
		},
	}
}

// Run executes the tool with given arguments
func (p *FMPAnalystEstimatesProvider) Run(ctx context.Context, args string) (string, error) {
	// Parse input arguments
	var input struct {
		Tickers []string `json:"tickers"`
	}

	if args != "" {
		if err := json.Unmarshal([]byte(args), &input); err != nil {
			return "", fmt.Errorf("invalid JSON arguments: %w", err)
		}
	}

	// Validate input
	if len(input.Tickers) == 0 {
		return "", fmt.Errorf("at least one ticker symbol is required")
	}

	// Fetch analyst estimates using the new client-based approach
	result, err := p.fetchWithClient(ctx, input.Tickers)
	if err != nil {
		return "", fmt.Errorf("failed to fetch analyst estimates: %w", err)
	}

	// Convert the entire response to map using JSON marshaling/unmarshaling
	// This automatically handles all fields without manual enumeration
	jsonData, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	var response map[string]any
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	// Add metadata
	response["metadata"] = map[string]any{
		"timestamp": time.Now().UTC(),
		"source":    "fmp_analyst_estimates",
		"tickers":   input.Tickers,
		"count":     len(input.Tickers),
	}

	// Convert back to JSON string for AI response
	resultJSON, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal final response: %w", err)
	}

	return string(resultJSON), nil
}

// fetchWithClient uses the FMP provider client to fetch analyst estimates
func (p *FMPAnalystEstimatesProvider) fetchWithClient(ctx context.Context, tickers []string) (map[string]any, error) {
	result := make(map[string]any)

	for _, ticker := range tickers {
		// Clean the ticker
		ticker = strings.ToUpper(strings.TrimSpace(ticker))
		if ticker == "" {
			continue
		}

		// Use the FMP client to get analyst estimates
		estimates, err := p.client.GetAnalystEstimates(ctx, ticker)
		if err != nil {
			// Log error but continue with other tickers
			slog.Warn("Failed to get analyst estimates", "ticker", ticker, "error", err)
			result[ticker] = map[string]any{
				"error": fmt.Sprintf("failed to get estimates: %v", err),
			}
			continue
		}

		// Store the complete estimates data
		result[ticker] = estimates
	}

	return result, nil
}
