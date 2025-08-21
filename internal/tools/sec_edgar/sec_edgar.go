package sec_edgar

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/amaurybrisou/mosychlos/pkg/sec"
)

// SecEdgarTool implements the Tool interface for SEC Edgar
type SecEdgarTool struct {
	client    *sec.Client
	sharedBag bag.SharedBag
}

var _ models.Tool = &SecEdgarTool{}

// NewFromConfig creates a new Provider from config
func NewFromConfig(cfg *config.SECEdgarConfig, sharedBag bag.SharedBag) (*SecEdgarTool, error) {
	return New(cfg.UserAgent, cfg.BaseURL, sharedBag)
}

// New constructs a provider with the given configuration
func New(userAgent, baseURL string, sharedBag bag.SharedBag) (*SecEdgarTool, error) {
	if userAgent == "" {
		return nil, fmt.Errorf("sec_edgar: missing user agent")
	}

	if baseURL == "" {
		baseURL = "https://data.sec.gov"
	}

	client, err := sec.NewClient(sec.Config{
		UserAgent: userAgent,
		BaseURL:   baseURL,
		Timeout:   30 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("sec_edgar: failed to create SEC client: %w", err)
	}

	return &SecEdgarTool{
		client:    client,
		sharedBag: sharedBag,
	}, nil
}

// Name returns the tool name for AI function calling
func (p *SecEdgarTool) Name() string {
	return "sec_edgar_filings"
}

// Key returns the unique tool key
func (p *SecEdgarTool) Key() keys.Key {
	return keys.SECFilings
}

// Description returns the tool description for AI
func (p *SecEdgarTool) Description() string {
	return "Access SEC EDGAR database for company filings, financial facts, and regulatory information. Provides company tickers, 10-K/10-Q filings, insider transactions, and financial data from official SEC sources."
}

func (t *SecEdgarTool) IsExternal() bool { return false }

// Definition returns the OpenAI tool definition
func (p *SecEdgarTool) Definition() models.ToolDef {
	return &models.CustomToolDef{
		Type: models.CustomToolDefType,
		FunctionDef: models.FunctionDef{
			Name:        p.Name(),
			Description: p.Description(),
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"action": map[string]any{
						"type":        "string",
						"enum":        []string{"tickers", "facts", "filings", "insider_transactions"},
						"description": "Type of SEC data to retrieve: tickers (company ticker info), facts (financial facts by CIK), filings (company filings), insider_transactions (insider trading data)",
					},
					"cik": map[string]any{
						"type":        "string",
						"description": "Central Index Key (CIK) for the company. Required for facts, filings, and insider_transactions actions. Can be with or without leading zeros.",
					},
					"ticker": map[string]any{
						"type":        "string",
						"description": "Company ticker symbol. Used to lookup CIK when CIK is not provided.",
					},
					"form_type": map[string]any{
						"type":        "string",
						"description": "SEC form type filter for filings (e.g., '10-K', '10-Q', '8-K', 'DEF 14A')",
					},
					"date_before": map[string]any{
						"type":        "string",
						"description": "Only include filings before this date (YYYY-MM-DD format)",
					},
					"count": map[string]any{
						"type":        "number",
						"description": "Number of filings to return (default 10, max 100)",
						"minimum":     1,
						"maximum":     100,
					},
				},
				"required":             []string{"action", "cik", "ticker", "form_type", "date_before", "count"},
				"additionalProperties": false,
			},
		},
	}
}

// Tags returns tool tags for categorization
func (p *SecEdgarTool) Tags() []string {
	return []string{"financial", "regulatory", "sec", "filings", "external-api"}
}

// Run executes the tool with the given arguments
func (p *SecEdgarTool) Run(ctx context.Context, args string) (string, error) {
	slog.Debug("Running SEC Edgar tool",
		"tool", p.Name(),
		"args", args,
	)

	// Parse arguments
	var params struct {
		Action     string `json:"action"`
		CIK        string `json:"cik,omitempty"`
		Ticker     string `json:"ticker,omitempty"`
		FormType   string `json:"form_type,omitempty"`
		DateBefore string `json:"date_before,omitempty"`
		Count      int    `json:"count,omitempty"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Validate required parameters
	if params.Action == "" {
		return "", fmt.Errorf("action is required")
	}

	// Set default count
	if params.Count == 0 {
		params.Count = 10
	}

	// Execute action
	var result any
	var err error

	switch params.Action {
	case "tickers":
		result, err = p.getCompanyTickers(ctx)
	case "facts":
		if params.CIK == "" && params.Ticker == "" {
			return "", fmt.Errorf("cik or ticker is required for facts action")
		}
		result, err = p.getCompanyFacts(ctx, params.CIK, params.Ticker)
	case "filings":
		if params.CIK == "" && params.Ticker == "" {
			return "", fmt.Errorf("cik or ticker is required for filings action")
		}
		result, err = p.getCompanyFilings(ctx, params.CIK, params.Ticker, params.FormType, params.DateBefore, params.Count)
	case "insider_transactions":
		if params.CIK == "" && params.Ticker == "" {
			return "", fmt.Errorf("cik or ticker is required for insider_transactions action")
		}
		result, err = p.getInsiderTransactions(ctx, params.CIK, params.Ticker)
	default:
		return "", fmt.Errorf("unsupported action: %s", params.Action)
	}

	if err != nil {
		slog.Error("SEC Edgar tool execution failed",
			"tool", p.Name(),
			"action", params.Action,
			"error", err,
		)
		return "", fmt.Errorf("SEC Edgar tool execution failed: %w", err)
	}

	// Create response
	response := map[string]any{
		"status": "success",
		"action": params.Action,
		"data":   result,
		"metadata": map[string]any{
			"timestamp": time.Now().UTC(),
			"source":    "sec_edgar",
			"tool":      p.Name(),
		},
	}

	// Return JSON response
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	slog.Info("SEC Edgar tool executed successfully",
		"tool", p.Name(),
		"action", params.Action,
		"response_size", len(responseJSON),
	)

	return string(responseJSON), nil
}

// getCompanyTickers retrieves company ticker information
func (p *SecEdgarTool) getCompanyTickers(ctx context.Context) (any, error) {
	response, err := p.client.GetCompanyTickers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get company tickers: %w", err)
	}

	return response, nil
}

// getCompanyFacts retrieves company facts by CIK
func (p *SecEdgarTool) getCompanyFacts(ctx context.Context, cik, ticker string) (any, error) {
	// If ticker is provided but no CIK, try to lookup CIK first
	if cik == "" && ticker != "" {
		var err error
		cik, err = p.lookupCIKByTicker(ctx, ticker)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup CIK for ticker %s: %w", ticker, err)
		}
	}

	response, err := p.client.GetCompanyFacts(ctx, cik)
	if err != nil {
		return nil, fmt.Errorf("failed to get company facts for CIK %s: %w", cik, err)
	}

	return response, nil
}

// getCompanyFilings retrieves company filings
func (p *SecEdgarTool) getCompanyFilings(ctx context.Context, cik, ticker, formType, dateBefore string, count int) (any, error) {
	// If ticker is provided but no CIK, try to lookup CIK first
	if cik == "" && ticker != "" {
		var err error
		cik, err = p.lookupCIKByTicker(ctx, ticker)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup CIK for ticker %s: %w", ticker, err)
		}
	}

	params := models.FilingsParams{
		CIK:        cik,
		Type:       formType,
		DateBefore: dateBefore,
		Count:      count,
	}

	response, err := p.client.GetCompanyFilings(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get company filings: %w", err)
	}

	return response, nil
}

// getInsiderTransactions retrieves insider transactions
func (p *SecEdgarTool) getInsiderTransactions(ctx context.Context, cik, ticker string) (any, error) {
	// If ticker is provided but no CIK, try to lookup CIK first
	if cik == "" && ticker != "" {
		var err error
		cik, err = p.lookupCIKByTicker(ctx, ticker)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup CIK for ticker %s: %w", ticker, err)
		}
	}

	response, err := p.client.GetInsiderTransactions(ctx, cik)
	if err != nil {
		return nil, fmt.Errorf("failed to get insider transactions for CIK %s: %w", cik, err)
	}

	return response, nil
}

// lookupCIKByTicker attempts to find CIK for a given ticker
func (p *SecEdgarTool) lookupCIKByTicker(ctx context.Context, ticker string) (string, error) {
	ticker = strings.ToUpper(strings.TrimSpace(ticker))

	// Get company tickers to find CIK
	tickers, err := p.client.GetCompanyTickers(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get tickers for CIK lookup: %w", err)
	}

	// Search through tickers data
	if tickers.Data != nil {
		for _, row := range tickers.Data {
			if len(row) >= 3 {
				if tickerStr, ok := row[0].(string); ok && strings.EqualFold(tickerStr, ticker) {
					if cikFloat, ok := row[1].(float64); ok {
						return fmt.Sprintf("%010d", int(cikFloat)), nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("ticker %s not found in SEC database", ticker)
}
