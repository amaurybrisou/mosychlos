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

// YFinanceStockInfoTool provides company information and key metrics
type YFinanceStockInfoTool struct {
	client    *yfinance.Client
	sharedBag bag.SharedBag
}

var _ models.Tool = &YFinanceStockInfoTool{}

// newStockInfoFromConfig creates a new StockInfoTool from config
func newStockInfoFromConfig(cfg *config.YFinanceConfig, sharedBag bag.SharedBag) (*YFinanceStockInfoTool, error) {
	clientCfg := yfinance.Config{
		BaseURL: cfg.BaseURL,
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}

	client, err := yfinance.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create yfinance client: %w", err)
	}

	return &YFinanceStockInfoTool{
		client:    client,
		sharedBag: sharedBag,
	}, nil
}

// NewStockInfo creates a new StockInfoTool
func NewStockInfo(baseURL string, timeout time.Duration) (*YFinanceStockInfoTool, error) {
	clientCfg := yfinance.Config{
		BaseURL: baseURL,
		Timeout: timeout,
	}

	client, err := yfinance.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create yfinance client: %w", err)
	}

	return &YFinanceStockInfoTool{
		client: client,
	}, nil
}

// Name returns the tool name for AI function calling
func (t *YFinanceStockInfoTool) Name() string {
	return "yfinance_stock_info"
}

// Key returns the unique tool key
func (t *YFinanceStockInfoTool) Key() bag.Key {
	return bag.YFinance
}

// Description returns the tool description for AI
func (t *YFinanceStockInfoTool) Description() string {
	return "Fetch comprehensive company profiles, business summaries, key statistics, and financial metrics"
}

func (t *YFinanceStockInfoTool) IsExternal() bool { return false }

// Definition returns the tool definition
func (t *YFinanceStockInfoTool) Definition() models.ToolDef {
	return &models.CustomToolDef{
		Type: models.CustomToolDefType,
		FunctionDef: models.FunctionDef{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"symbol": map[string]any{
						"type":        "string",
						"description": "Stock symbol (e.g., 'AAPL', 'MSFT', 'GOOGL')",
					},
				},
				"required": []string{"symbol"},
			},
		},
	}
}

// Tags returns tool tags for categorization
func (t *YFinanceStockInfoTool) Tags() []string {
	return []string{"financial", "market-data", "yahoo-finance", "stocks", "company-info", "fundamentals"}
}

// Run executes the tool with the given arguments
func (t *YFinanceStockInfoTool) Run(ctx context.Context, args string) (string, error) {
	slog.Debug("Running yfinance stock info tool",
		"tool", t.Name(),
		"args", args,
	)

	// Parse arguments
	var params struct {
		Symbol string `json:"symbol"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Validate required parameters
	if params.Symbol == "" {
		return "", fmt.Errorf("symbol is required")
	}

	// Get stock info
	result, err := t.getStockInfo(ctx, params.Symbol)
	if err != nil {
		slog.Error("Stock info retrieval failed",
			"tool", t.Name(),
			"symbol", params.Symbol,
			"error", err,
		)
		return "", fmt.Errorf("failed to get stock info: %w", err)
	}

	// Return JSON response
	response, err := json.Marshal(map[string]any{
		"status": "success",
		"symbol": params.Symbol,
		"data":   result,
		"metadata": map[string]any{
			"timestamp": time.Now().UTC(),
			"source":    "yahoo_finance",
			"tool":      t.Name(),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	slog.Info("Stock info retrieved successfully",
		"tool", t.Name(),
		"symbol", params.Symbol,
		"result_size", len(response),
	)

	return string(response), nil
}

// getStockInfo retrieves company information from Yahoo Finance
func (t *YFinanceStockInfoTool) getStockInfo(ctx context.Context, symbol string) (any, error) {
	stockInfo, err := t.client.GetStockInfo(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock info for %s: %w", symbol, err)
	}

	// Check for API errors first
	if stockInfo.Chart.Error != nil {
		return map[string]any{
			"error": map[string]any{
				"code":        stockInfo.Chart.Error.Code,
				"description": stockInfo.Chart.Error.Description,
			},
		}, nil
	}

	// Check if we have results
	if len(stockInfo.Chart.Result) == 0 {
		return map[string]any{
			"error": map[string]any{
				"code":        "NO_DATA",
				"description": fmt.Sprintf("No data available for symbol %s", symbol),
			},
		}, nil
	}

	// Convert the entire response to map using JSON marshaling/unmarshaling
	// This automatically handles all fields without manual enumeration
	jsonData, err := json.Marshal(stockInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal stock info: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	// Add metadata about the number of results available
	result["results_count"] = len(stockInfo.Chart.Result)

	return result, nil
}
