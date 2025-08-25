package yfinance

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/amaurybrisou/mosychlos/pkg/yfinance"
)

// YFinanceMarketDataTool provides market indices and sector information
type YFinanceMarketDataTool struct {
	client    *yfinance.Client
	sharedBag bag.SharedBag
}

var _ models.Tool = &YFinanceMarketDataTool{}

// newMarketDataFromConfig creates a new YFinanceMarketDataTool from config
func newMarketDataFromConfig(cfg *config.YFinanceConfig, sharedBag bag.SharedBag) (*YFinanceMarketDataTool, error) {
	clientCfg := yfinance.Config{
		BaseURL: cfg.BaseURL,
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}

	client, err := yfinance.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create yfinance client: %w", err)
	}

	return &YFinanceMarketDataTool{
		client:    client,
		sharedBag: sharedBag,
	}, nil
}

// NewMarketData creates a new YFinanceMarketDataTool
func NewMarketData(baseURL string, timeout time.Duration) (*YFinanceMarketDataTool, error) {
	clientCfg := yfinance.Config{
		BaseURL: baseURL,
		Timeout: timeout,
	}

	client, err := yfinance.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create yfinance client: %w", err)
	}

	return &YFinanceMarketDataTool{
		client: client,
	}, nil
}

// Name returns the tool name for AI function calling
func (t *YFinanceMarketDataTool) Name() string {
	return "yfinance_market_data"
}

// Key returns the unique tool key
func (t *YFinanceMarketDataTool) Key() bag.Key {
	return bag.YFinanceMarketData
}

// Description returns the tool description for AI
func (t *YFinanceMarketDataTool) Description() string {
	return "Monitor broad market indices, sector performance, and macroeconomic market indicators"
}

func (t *YFinanceMarketDataTool) IsExternal() bool { return false }

// Definition returns the tool definition
func (t *YFinanceMarketDataTool) Definition() models.ToolDef {
	return &models.CustomToolDef{
		Type: models.CustomToolDefType,
		FunctionDef: models.FunctionDef{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"symbols": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "Market index or sector symbols (e.g., ['^GSPC', '^DJI', '^IXIC', 'SPY', 'QQQ'])",
						"maxItems":    20,
					},
					"period": map[string]any{
						"type":        "string",
						"enum":        []string{"1d", "5d", "1mo", "3mo", "6mo", "1y", "2y", "5y"},
						"description": "Time period for market data (defaults to '1d' if not specified)",
					},
				},
				"required": []string{"symbols", "period"},
			},
		},
	}
}

// Tags returns tool tags for categorization
func (t *YFinanceMarketDataTool) Tags() []string {
	return []string{"financial", "market-data", "yahoo-finance", "indices", "sectors", "market"}
}

// Run executes the tool with the given arguments
func (t *YFinanceMarketDataTool) Run(ctx context.Context, args any) (any, error) {
	slog.Debug("Running yfinance market data tool",
		"tool", t.Name(),
		"args", args,
	)

	// Parse arguments
	var params struct {
		Symbols []string `json:"symbols"`
		Period  string   `json:"period,omitempty"`
	}

	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", args)), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Validate required parameters
	if len(params.Symbols) == 0 {
		return "", fmt.Errorf("symbols array is required and cannot be empty")
	}

	// Set default period
	if params.Period == "" {
		params.Period = "1d"
	}

	// Get market data using the client - more efficient than individual calls
	result, err := t.getMarketData(ctx, params.Symbols, params.Period)
	if err != nil {
		slog.Error("Market data retrieval failed",
			"tool", t.Name(),
			"symbols", params.Symbols,
			"error", err,
		)
		return "", fmt.Errorf("failed to get market data: %w", err)
	}

	// Return JSON response
	response, err := json.Marshal(map[string]any{
		"status":  "success",
		"symbols": params.Symbols,
		"period":  params.Period,
		"data":    result,
		"metadata": map[string]any{
			"timestamp": time.Now().UTC(),
			"source":    "yahoo_finance",
			"tool":      t.Name(),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	slog.Info("Market data retrieved successfully",
		"tool", t.Name(),
		"symbols", params.Symbols,
		"period", params.Period,
		"result_size", len(response),
	)

	return result, nil
}

// getMarketData retrieves market data for multiple symbols using the client
func (t *YFinanceMarketDataTool) getMarketData(ctx context.Context, symbols []string, period string) (any, error) {
	marketData, err := t.client.GetMarketData(ctx, symbols)
	if err != nil {
		return nil, fmt.Errorf("failed to get market data for symbols %v: %w", symbols, err)
	}

	// Check for API errors first
	if marketData.QuoteResponse.Error != nil {
		return map[string]any{
			"error": map[string]any{
				"code":        marketData.QuoteResponse.Error.Code,
				"description": marketData.QuoteResponse.Error.Description,
			},
		}, nil
	}

	// Check if we have results
	if len(marketData.QuoteResponse.Result) == 0 {
		return map[string]any{
			"error": map[string]any{
				"code":        "NO_DATA",
				"description": fmt.Sprintf("No market data available for symbols %v", symbols),
			},
		}, nil
	}

	// Convert the entire response to map using JSON marshaling/unmarshaling
	// This automatically handles all fields without manual enumeration
	// jsonData, err := json.Marshal(marketData)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to marshal market data: %w", err)
	// }

	// var result map[string]any
	// if err := json.Unmarshal(jsonData, &result); err != nil {
	// 	return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	// }

	// // Add metadata about the number of results available
	// result["results_count"] = len(marketData.QuoteResponse.Result)
	// result["requested_period"] = period

	return marketData, nil
}
