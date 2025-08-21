package yfinance

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/amaurybrisou/mosychlos/pkg/yfinance"
)

// YFinanceStockDataTool provides real-time and historical stock price data
type YFinanceStockDataTool struct {
	client    *yfinance.Client
	sharedBag bag.SharedBag
}

var _ models.Tool = &YFinanceStockDataTool{}

// NewStockDataFromConfig creates a new StockDataTool from config
func NewStockDataFromConfig(cfg *config.YFinanceConfig, sharedBag bag.SharedBag) (*YFinanceStockDataTool, error) {
	clientCfg := yfinance.Config{
		BaseURL: cfg.BaseURL,
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}

	client, err := yfinance.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create yfinance client: %w", err)
	}

	return &YFinanceStockDataTool{
		client:    client,
		sharedBag: sharedBag,
	}, nil
}

// NewStockData creates a new StockDataTool
func NewStockData(baseURL string, timeout time.Duration) (*YFinanceStockDataTool, error) {
	clientCfg := yfinance.Config{
		BaseURL: baseURL,
		Timeout: timeout,
	}

	client, err := yfinance.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create yfinance client: %w", err)
	}

	return &YFinanceStockDataTool{
		client: client,
	}, nil
}

// Name returns the tool name for AI function calling
func (t *YFinanceStockDataTool) Name() string {
	return "yfinance_stock_data"
}

// Key returns the unique tool key
func (t *YFinanceStockDataTool) Key() keys.Key {
	return keys.YFinance
}

// Description returns the tool description for AI
func (t *YFinanceStockDataTool) Description() string {
	return "Retrieve real-time quotes and historical OHLCV price data for individual stocks and securities"
}

func (t *YFinanceStockDataTool) IsExternal() bool { return false }

// Definition returns the tool definition
func (t *YFinanceStockDataTool) Definition() models.ToolDef {
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
						"description": "Stock symbol (e.g., 'AAPL', 'MSFT', 'SPY')",
					},
					"period": map[string]any{
						"type":        "string",
						"enum":        []string{"1d", "5d", "1mo", "3mo", "6mo", "1y", "2y", "5y", "10y", "ytd", "max"},
						"description": "Time period for historical data. Defaults to '1y'.",
					},
					"interval": map[string]any{
						"type":        "string",
						"enum":        []string{"1m", "2m", "5m", "15m", "30m", "60m", "90m", "1h", "1d", "5d", "1wk", "1mo", "3mo"},
						"description": "Data interval. Defaults to '1d'.",
					},
				},
				"required": []string{"symbol", "period", "interval"},
			},
		},
	}
}

// Tags returns tool tags for categorization
func (t *YFinanceStockDataTool) Tags() []string {
	return []string{"finance", "stocks", "data"}
}

// Run executes the tool with the given arguments
func (t *YFinanceStockDataTool) Run(ctx context.Context, args string) (string, error) {
	slog.Debug("Running yfinance stock data tool",
		"tool", t.Name(),
		"args", args,
	)

	// Parse arguments
	var params struct {
		Symbol   string `json:"symbol"`
		Period   string `json:"period,omitempty"`
		Interval string `json:"interval,omitempty"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Validate required parameters
	if params.Symbol == "" {
		return "", fmt.Errorf("symbol is required")
	}

	// Set defaults
	if params.Period == "" {
		params.Period = "1y"
	}
	if params.Interval == "" {
		params.Interval = "1d"
	}

	// Get stock data
	result, err := t.getStockData(ctx, params.Symbol, params.Period, params.Interval)
	if err != nil {
		slog.Error("Stock data retrieval failed",
			"tool", t.Name(),
			"symbol", params.Symbol,
			"period", params.Period,
			"interval", params.Interval,
			"error", err,
		)
		return "", fmt.Errorf("failed to get stock data: %w", err)
	}

	// Return JSON response
	response, err := json.Marshal(map[string]any{
		"status":   "success",
		"symbol":   params.Symbol,
		"period":   params.Period,
		"interval": params.Interval,
		"data":     result,
		"metadata": map[string]any{
			"timestamp": time.Now().UTC(),
			"source":    "yahoo_finance",
			"tool":      t.Name(),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	slog.Info("Stock data retrieved successfully",
		"tool", t.Name(),
		"symbol", params.Symbol,
		"period", params.Period,
		"interval", params.Interval,
		"result_size", len(response),
	)

	return string(response), nil
}

// getStockData retrieves stock price data from Yahoo Finance
func (t *YFinanceStockDataTool) getStockData(ctx context.Context, symbol, period, interval string) (any, error) {
	stockData, err := t.client.GetStockData(ctx, symbol, period, interval)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock data for %s: %w", symbol, err)
	}

	// Check for API errors first
	if stockData.Chart.Error != nil {
		return map[string]any{
			"error": map[string]any{
				"code":        stockData.Chart.Error.Code,
				"description": stockData.Chart.Error.Description,
			},
		}, nil
	}

	// Check if we have results
	if len(stockData.Chart.Result) == 0 {
		return map[string]any{
			"error": map[string]any{
				"code":        "NO_DATA",
				"description": fmt.Sprintf("No data available for symbol %s", symbol),
			},
		}, nil
	}

	// Convert the entire response to map using JSON marshaling/unmarshaling
	// This automatically handles all fields without manual enumeration
	jsonData, err := json.Marshal(stockData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal stock data: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	// Add metadata about the number of results available
	result["results_count"] = len(stockData.Chart.Result)

	return result, nil
}
