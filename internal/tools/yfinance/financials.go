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

// YFinanceFinancialsTool provides financial statements and ratios
type YFinanceFinancialsTool struct {
	client    *yfinance.Client
	sharedBag bag.SharedBag
}

var _ models.Tool = &YFinanceFinancialsTool{}

// NewFinancialsFromConfig creates a new YFinanceFinancialsTool from config
func NewFinancialsFromConfig(cfg *config.YFinanceConfig, sharedBag bag.SharedBag) (*YFinanceFinancialsTool, error) {
	clientCfg := yfinance.Config{
		BaseURL: cfg.BaseURL,
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}

	client, err := yfinance.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create yfinance client: %w", err)
	}

	return &YFinanceFinancialsTool{
		client:    client,
		sharedBag: sharedBag,
	}, nil
}

// NewFinancials creates a new YFinanceFinancialsTool
func NewFinancials(baseURL string, timeout time.Duration) (*YFinanceFinancialsTool, error) {
	clientCfg := yfinance.Config{
		BaseURL: baseURL,
		Timeout: timeout,
	}

	client, err := yfinance.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create yfinance client: %w", err)
	}

	return &YFinanceFinancialsTool{
		client: client,
	}, nil
}

// Name returns the tool name for AI function calling
func (t *YFinanceFinancialsTool) Name() string {
	return "yfinance_financials"
}

// Key returns the unique tool key
func (t *YFinanceFinancialsTool) Key() bag.Key {
	return bag.YFinance
}

// Description returns the tool description for AI
func (t *YFinanceFinancialsTool) Description() string {
	return "Access complete financial statements: income statements, balance sheets, and cash flow reports"
}

func (t *YFinanceFinancialsTool) IsExternal() bool { return false }

// Definition returns the tool definition
func (t *YFinanceFinancialsTool) Definition() models.ToolDef {
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
					"statement_type": map[string]any{
						"type":        "string",
						"enum":        []string{"income", "balance", "cashflow"},
						"description": "Type of financial statement to retrieve. Defaults to 'income'.",
					},
					"frequency": map[string]any{
						"type":        "string",
						"enum":        []string{"annual", "quarterly"},
						"description": "Frequency of financial data. Defaults to 'annual'.",
					},
				},
				"required": []string{"symbol", "statement_type", "frequency"},
			},
		},
	}
}

// Tags returns tool tags for categorization
func (t *YFinanceFinancialsTool) Tags() []string {
	return []string{"financial", "market-data", "yahoo-finance", "stocks", "financials", "statements"}
}

// Run executes the tool with the given arguments
func (t *YFinanceFinancialsTool) Run(ctx context.Context, args string) (string, error) {
	slog.Debug("Running yfinance financials tool",
		"tool", t.Name(),
		"args", args,
	)

	// Parse arguments
	var params struct {
		Symbol        string `json:"symbol"`
		StatementType string `json:"statement_type,omitempty"`
		Frequency     string `json:"frequency,omitempty"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Validate required parameters
	if params.Symbol == "" {
		return "", fmt.Errorf("symbol is required")
	}

	// Set defaults
	if params.StatementType == "" {
		params.StatementType = "income"
	}
	if params.Frequency == "" {
		params.Frequency = "annual"
	}

	// Get financial data
	result, err := t.getFinancials(ctx, params.Symbol, params.StatementType, params.Frequency)
	if err != nil {
		slog.Error("Financial data retrieval failed",
			"tool", t.Name(),
			"symbol", params.Symbol,
			"statement_type", params.StatementType,
			"frequency", params.Frequency,
			"error", err,
		)
		return "", fmt.Errorf("failed to get financial data: %w", err)
	}

	// Return JSON response
	response, err := json.Marshal(map[string]any{
		"status":         "success",
		"symbol":         params.Symbol,
		"statement_type": params.StatementType,
		"frequency":      params.Frequency,
		"data":           result,
		"metadata": map[string]any{
			"timestamp": time.Now().UTC(),
			"source":    "yahoo_finance",
			"tool":      t.Name(),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	slog.Info("Financial data retrieved successfully",
		"tool", t.Name(),
		"symbol", params.Symbol,
		"statement_type", params.StatementType,
		"frequency", params.Frequency,
		"result_size", len(response),
	)

	return string(response), nil
}

// getFinancials retrieves financial statements from Yahoo Finance
func (t *YFinanceFinancialsTool) getFinancials(ctx context.Context, symbol, statementType, frequency string) (any, error) {
	financialsData, err := t.client.GetFinancials(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get financials for %s: %w", symbol, err)
	}

	// Check for API errors first
	if financialsData.QuoteSummary.Error != nil {
		return map[string]any{
			"error": map[string]any{
				"code":        financialsData.QuoteSummary.Error.Code,
				"description": financialsData.QuoteSummary.Error.Description,
			},
		}, nil
	}

	// Check if we have results
	if len(financialsData.QuoteSummary.Result) == 0 {
		return map[string]any{
			"error": map[string]any{
				"code":        "NO_DATA",
				"description": fmt.Sprintf("No financial data available for symbol %s", symbol),
			},
		}, nil
	}

	// Convert the entire response to map using JSON marshaling/unmarshaling
	// This automatically handles all fields without manual enumeration
	jsonData, err := json.Marshal(financialsData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal financial data: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	// Add metadata about the number of results available and requested parameters
	result["results_count"] = len(financialsData.QuoteSummary.Result)
	result["requested_statement_type"] = statementType
	result["requested_frequency"] = frequency

	return result, nil
}
