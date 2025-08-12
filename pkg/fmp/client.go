package fmp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"log/slog"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// Client provides access to Financial Modeling Prep API
type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

// Config holds FMP client configuration
type Config struct {
	APIKey  string
	BaseURL string
	Timeout time.Duration
}

// NewClient creates a new FMP client
func NewClient(cfg Config) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api_key is required for FMP API")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://financialmodelingprep.com/api"
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &Client{
		apiKey:  cfg.APIKey,
		baseURL: baseURL,
		http: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     30 * time.Second,
			},
		},
	}, nil
}

// GetCompanyProfile retrieves company profile information
func (c *Client) GetCompanyProfile(ctx context.Context, symbol string) (*models.FMPCompanyProfile, error) {
	endpoint := fmt.Sprintf("/v3/profile/%s", symbol)
	params := url.Values{
		"apikey": {c.apiKey},
	}

	var response []models.FMPCompanyProfile
	if err := c.makeRequest(ctx, endpoint, params, &response); err != nil {
		return nil, fmt.Errorf("failed to get company profile for %s: %w", symbol, err)
	}

	if len(response) == 0 {
		return nil, fmt.Errorf("no company profile found for symbol %s", symbol)
	}

	return &response[0], nil
}

// GetFinancialStatements retrieves financial statements
func (c *Client) GetFinancialStatements(ctx context.Context, symbol string, statementType string) (*models.FMPFinancialStatements, error) {
	endpoint := fmt.Sprintf("/v3/%s/%s", statementType, symbol)
	params := url.Values{
		"apikey": {c.apiKey},
		"limit":  {"20"}, // Get last 20 periods
	}

	var response []models.FMPFinancialStatement
	if err := c.makeRequest(ctx, endpoint, params, &response); err != nil {
		return nil, fmt.Errorf("failed to get %s statements for %s: %w", statementType, symbol, err)
	}

	return &models.FMPFinancialStatements{
		Symbol:     symbol,
		Type:       statementType,
		Statements: response,
	}, nil
}

// GetKeyMetrics retrieves key financial metrics
func (c *Client) GetKeyMetrics(ctx context.Context, symbol string) (*models.FMPKeyMetrics, error) {
	endpoint := fmt.Sprintf("/v3/key-metrics/%s", symbol)
	params := url.Values{
		"apikey": {c.apiKey},
		"limit":  {"20"},
	}

	var response []models.FMPKeyMetric
	if err := c.makeRequest(ctx, endpoint, params, &response); err != nil {
		return nil, fmt.Errorf("failed to get key metrics for %s: %w", symbol, err)
	}

	return &models.FMPKeyMetrics{
		Symbol:  symbol,
		Metrics: response,
	}, nil
}

// GetAnalystEstimates retrieves analyst estimates
func (c *Client) GetAnalystEstimates(ctx context.Context, symbol string) (*models.FMPAnalystEstimates, error) {
	endpoint := fmt.Sprintf("/v3/analyst-estimates/%s", symbol)
	params := url.Values{
		"apikey": {c.apiKey},
		"limit":  {"20"},
	}

	var response []models.FMPAnalystEstimate
	if err := c.makeRequest(ctx, endpoint, params, &response); err != nil {
		return nil, fmt.Errorf("failed to get analyst estimates for %s: %w", symbol, err)
	}

	return &models.FMPAnalystEstimates{
		Symbol:    symbol,
		Estimates: response,
	}, nil
}

// GetStockPrice retrieves current stock price
func (c *Client) GetStockPrice(ctx context.Context, symbol string) (*models.FMPStockPrice, error) {
	endpoint := fmt.Sprintf("/v3/quote-short/%s", symbol)
	params := url.Values{
		"apikey": {c.apiKey},
	}

	var response []models.FMPStockPrice
	if err := c.makeRequest(ctx, endpoint, params, &response); err != nil {
		return nil, fmt.Errorf("failed to get stock price for %s: %w", symbol, err)
	}

	if len(response) == 0 {
		return nil, fmt.Errorf("no stock price found for symbol %s", symbol)
	}

	return &response[0], nil
}

// GetMarketCap retrieves market capitalization data
func (c *Client) GetMarketCap(ctx context.Context, symbol string) (*models.FMPMarketCap, error) {
	endpoint := fmt.Sprintf("/v3/market-capitalization/%s", symbol)
	params := url.Values{
		"apikey": {c.apiKey},
	}

	var response []models.FMPMarketCap
	if err := c.makeRequest(ctx, endpoint, params, &response); err != nil {
		return nil, fmt.Errorf("failed to get market cap for %s: %w", symbol, err)
	}

	if len(response) == 0 {
		return nil, fmt.Errorf("no market cap found for symbol %s", symbol)
	}

	return &response[0], nil
}

// makeRequest performs HTTP request with proper headers
func (c *Client) makeRequest(ctx context.Context, endpoint string, params url.Values, result any) error {
	fullURL := c.baseURL + endpoint
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mosychlos/1.0")

	slog.Debug("Making FMP API request",
		"url", fullURL,
	)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("FMP API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	slog.Debug("FMP API request successful",
		"url", fullURL,
		"status", resp.StatusCode,
	)

	return nil
}
