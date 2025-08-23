package fmp

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
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// NewFromConfig creates a new Provider from config
func NewFromConfig(cfg *config.FMPConfig, sharedBag bag.SharedBag) (*FMPTool, error) {
	return New(cfg.APIKey, cfg.CacheDir, sharedBag)
}

// FMPTool implements a Financial Modeling Prep-backed fundamentals source.
type FMPTool struct {
	client    *fmp.Client
	cacheDir  string // cache directory for caching wrapper
	sharedBag bag.SharedBag
}

// Ensure FMPTool implements models.Tool interface
var _ models.Tool = &FMPTool{}

func New(key, cacheDir string, sharedBag bag.SharedBag) (*FMPTool, error) {
	if key == "" {
		return nil, fmt.Errorf("fmp: missing FMP_API_KEY")
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

	p := &FMPTool{
		client:    client,
		cacheDir:  cacheDir,
		sharedBag: sharedBag,
	}
	return p, nil
}

// Name returns the tool name
func (p *FMPTool) Name() string {
	return bag.FMP.String()
}

// Key returns the tool key
func (p *FMPTool) Key() bag.Key {
	return bag.FMP
}

// Tags returns the tool tags
func (p *FMPTool) Tags() []string {
	return []string{"finance", "fundamentals", "stocks", "metrics", "company-data"}
}

// Description returns the tool description
func (p *FMPTool) Description() string {
	return "Fetches fundamental financial data for stocks including company profiles, financial metrics, sector information, and key ratios from Financial Modeling Prep API"
}

func (t *FMPTool) IsExternal() bool { return false }

// Definition returns the tool definition for AI systems
func (p *FMPTool) Definition() models.ToolDef {
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
						"description": "Array of stock ticker symbols to fetch fundamental data for (e.g., ['AAPL', 'GOOGL', 'MSFT']), not for cryptocurrencies",
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
func (p *FMPTool) Run(ctx context.Context, args string) (string, error) {
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

	// Fetch fundamental data using the new client-based approach
	result, err := p.fetchWithClient(ctx, input.Tickers)
	if err != nil {
		return "", fmt.Errorf("failed to fetch fundamental data: %w", err)
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
		"source":    "fmp",
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

// fetchWithClient uses the FMP provider client to fetch company profiles
func (p *FMPTool) fetchWithClient(ctx context.Context, tickers []string) (map[string]any, error) {
	result := make(map[string]any)

	for _, ticker := range tickers {
		// Clean the ticker
		ticker = strings.ToUpper(strings.TrimSpace(ticker))
		if ticker == "" {
			continue
		}

		// Use the FMP client to get company profile
		profile, err := p.client.GetCompanyProfile(ctx, ticker)
		if err != nil {
			// Log error but continue with other tickers
			slog.Warn("Failed to get company profile", "ticker", ticker, "error", err)
			result[ticker] = map[string]any{
				"error": fmt.Sprintf("failed to get profile: %v", err),
			}
			continue
		}

		// Store the complete profile data
		result[ticker] = profile
	}

	return result, nil
}
