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

// YFinanceDividendsTool provides dividend history and yield information
type YFinanceDividendsTool struct {
	client    *yfinance.Client
	sharedBag bag.SharedBag
}

var _ models.Tool = &YFinanceDividendsTool{}

// NewDividendsFromConfig creates a new DividendsTool from config
func NewDividendsFromConfig(cfg *config.YFinanceConfig, sharedBag bag.SharedBag) (*YFinanceDividendsTool, error) {
	clientCfg := yfinance.Config{
		BaseURL: cfg.BaseURL,
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}

	client, err := yfinance.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create yfinance client: %w", err)
	}

	return &YFinanceDividendsTool{
		client:    client,
		sharedBag: sharedBag,
	}, nil
}

// NewDividends creates a new DividendsTool
func NewDividends(baseURL string, timeout time.Duration) (*YFinanceDividendsTool, error) {
	clientCfg := yfinance.Config{
		BaseURL: baseURL,
		Timeout: timeout,
	}

	client, err := yfinance.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create yfinance client: %w", err)
	}

	return &YFinanceDividendsTool{
		client: client,
	}, nil
}

// Name returns the tool name for AI function calling
func (t *YFinanceDividendsTool) Name() string {
	return "yfinance_dividends"
}

// Key returns the unique tool key
func (t *YFinanceDividendsTool) Key() keys.Key {
	return keys.YFinance
}

// Description returns the tool description for AI
func (t *YFinanceDividendsTool) Description() string {
	return "Analyze dividend payments, yield calculations, and distribution history for income-focused analysis"
}

func (t *YFinanceDividendsTool) IsExternal() bool { return false }

// Definition returns the tool definition
func (t *YFinanceDividendsTool) Definition() models.ToolDef {
	return models.ToolDef{
		Type: "function",
		Function: models.FunctionDef{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"symbol": map[string]any{
						"type":        "string",
						"description": "Stock symbol (e.g., 'AAPL', 'MSFT', 'JNJ')",
					},
					"period": map[string]any{
						"type":        "string",
						"enum":        []string{"1y", "2y", "5y", "10y", "max"},
						"description": "Time period for dividend history (defaults to '5y' if not specified)",
					},
				},
				"required": []string{"symbol", "period"},
			},
		},
	}
}

// Tags returns tool tags for categorization
func (t *YFinanceDividendsTool) Tags() []string {
	return []string{"financial", "market-data", "yahoo-finance", "stocks", "dividends", "income"}
}

// Run executes the tool with the given arguments
func (t *YFinanceDividendsTool) Run(ctx context.Context, args string) (string, error) {
	slog.Debug("Running yfinance dividends tool",
		"tool", t.Name(),
		"args", args,
	)

	// Parse arguments
	var params struct {
		Symbol string `json:"symbol"`
		Period string `json:"period,omitempty"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Validate required parameters
	if params.Symbol == "" {
		return "", fmt.Errorf("symbol is required")
	}

	// Set default period
	if params.Period == "" {
		params.Period = "5y"
	}

	// Get dividend data
	result, err := t.getDividends(ctx, params.Symbol, params.Period)
	if err != nil {
		slog.Error("Dividend data retrieval failed",
			"tool", t.Name(),
			"symbol", params.Symbol,
			"period", params.Period,
			"error", err,
		)
		return "", fmt.Errorf("failed to get dividend data: %w", err)
	}

	// Return JSON response
	response, err := json.Marshal(map[string]any{
		"status": "success",
		"symbol": params.Symbol,
		"period": params.Period,
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

	slog.Info("Dividend data retrieved successfully",
		"tool", t.Name(),
		"symbol", params.Symbol,
		"period", params.Period,
		"result_size", len(response),
	)

	return string(response), nil
}

// getDividends retrieves dividend data from Yahoo Finance
func (t *YFinanceDividendsTool) getDividends(ctx context.Context, symbol, period string) (any, error) {
	// Convert period to date range
	endDate := time.Now()
	var startDate time.Time

	switch period {
	case "1d":
		startDate = endDate.AddDate(0, 0, -1)
	case "5d":
		startDate = endDate.AddDate(0, 0, -5)
	case "1mo":
		startDate = endDate.AddDate(0, -1, 0)
	case "3mo":
		startDate = endDate.AddDate(0, -3, 0)
	case "6mo":
		startDate = endDate.AddDate(0, -6, 0)
	case "1y":
		startDate = endDate.AddDate(-1, 0, 0)
	case "2y":
		startDate = endDate.AddDate(-2, 0, 0)
	case "5y":
		startDate = endDate.AddDate(-5, 0, 0)
	case "10y":
		startDate = endDate.AddDate(-10, 0, 0)
	case "max":
		startDate = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	default:
		// Default to 1 year if period is not recognized
		startDate = endDate.AddDate(-1, 0, 0)
	}

	dividendData, err := t.client.GetDividends(ctx, symbol, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get dividends for %s: %w", symbol, err)
	}

	// Check for API errors first
	if dividendData.Chart.Error != nil {
		return map[string]any{
			"error": map[string]any{
				"code":        dividendData.Chart.Error.Code,
				"description": dividendData.Chart.Error.Description,
			},
		}, nil
	}

	// Check if we have results
	if len(dividendData.Chart.Result) == 0 {
		return map[string]any{
			"error": map[string]any{
				"code":        "NO_DATA",
				"description": fmt.Sprintf("No dividend data available for symbol %s", symbol),
			},
		}, nil
	}

	// Convert the entire response to map using JSON marshaling/unmarshaling
	// This automatically handles all fields without manual enumeration
	jsonData, err := json.Marshal(dividendData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dividend data: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	// Add metadata about the number of results available
	result["results_count"] = len(dividendData.Chart.Result)

	return result, nil
}
