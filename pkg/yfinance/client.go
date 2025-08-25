// Package yfinance provides a client for accessing the Yahoo Finance API
package yfinance

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

// Client provides access to Yahoo Finance API
type Client struct {
	baseURL string
	http    *http.Client
}

// Config holds YFinance client configuration
type Config struct {
	BaseURL string
	Timeout time.Duration
}

// NewClient creates a new Yahoo Finance client
func NewClient(cfg Config) (*Client, error) {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://query1.finance.yahoo.com"
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &Client{
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

// GetStockData retrieves historical stock data
func (c *Client) GetStockData(ctx context.Context, symbol string, period string, interval string) (*models.StockDataResponse, error) {
	params := url.Values{
		"symbol":   {symbol},
		"period1":  {"0"},
		"period2":  {fmt.Sprintf("%d", time.Now().Unix())},
		"interval": {interval},
	}

	if period != "" {
		params.Set("range", period)
	}

	endpoint := "/v8/finance/chart/" + symbol

	var response models.StockDataResponse
	if err := c.makeRequest(ctx, endpoint, params, &response); err != nil {
		return nil, fmt.Errorf("failed to get stock data for %s: %w", symbol, err)
	}

	return &response, nil
}

// GetStockInfo retrieves company information and key statistics
func (c *Client) GetStockInfo(ctx context.Context, symbol string) (*models.StockInfoResponse, error) {
	endpoint := "/v8/finance/chart/" + symbol
	params := url.Values{
		"range":          {"1d"},
		"interval":       {"1d"},
		"includePrePost": {"false"},
	}

	var response models.StockInfoResponse
	if err := c.makeRequest(ctx, endpoint, params, &response); err != nil {
		return nil, fmt.Errorf("failed to get stock info for %s: %w", symbol, err)
	}

	return &response, nil
}

// GetDividends retrieves dividend history
func (c *Client) GetDividends(ctx context.Context, symbol string, startDate, endDate time.Time) (*models.DividendsResponse, error) {
	params := url.Values{
		"period1":  {fmt.Sprintf("%d", startDate.Unix())},
		"period2":  {fmt.Sprintf("%d", endDate.Unix())},
		"events":   {"div"},
		"interval": {"1d"},
	}

	endpoint := "/v8/finance/chart/" + symbol

	var response models.DividendsResponse
	if err := c.makeRequest(ctx, endpoint, params, &response); err != nil {
		return nil, fmt.Errorf("failed to get dividends for %s: %w", symbol, err)
	}

	return &response, nil
}

// GetFinancials retrieves financial statements
func (c *Client) GetFinancials(ctx context.Context, symbol string) (*models.FinancialsResponse, error) {
	endpoint := "/v10/finance/quoteSummary/" + symbol
	params := url.Values{
		"modules": {"financialData,defaultKeyStatistics,incomeStatementHistory,balanceSheetHistory,cashflowStatementHistory"},
	}

	var response models.FinancialsResponse
	if err := c.makeRequest(ctx, endpoint, params, &response); err != nil {
		return nil, fmt.Errorf("failed to get financials for %s: %w", symbol, err)
	}

	return &response, nil
}

// GetMarketData retrieves market data for multiple symbols
func (c *Client) GetMarketData(ctx context.Context, symbols []string) (*models.MarketDataResponse, error) {
	if len(symbols) == 0 {
		return nil, fmt.Errorf("no symbols provided")
	}

	// Yahoo Finance chart endpoint doesn't support multiple symbols in one request
	// We need to make separate requests and combine the results
	var allResults []models.ChartResult
	var responseError *models.APIError

	for _, symbol := range symbols {
		endpoint := "/v8/finance/chart/" + symbol
		params := url.Values{
			"range":    {"1d"},
			"interval": {"1d"},
		}

		var singleResponse models.StockDataResponse
		if err := c.makeRequest(ctx, endpoint, params, &singleResponse); err != nil {
			return nil, fmt.Errorf("failed to get market data for symbol %s: %w", symbol, err)
		}

		if singleResponse.Chart.Error != nil {
			responseError = singleResponse.Chart.Error
			continue
		}

		// Add all results from this symbol
		allResults = append(allResults, singleResponse.Chart.Result...)
	}

	// Create combined response
	response := &models.MarketDataResponse{
		QuoteResponse: models.QuoteResponse{
			Result: make([]models.QuoteResult, len(allResults)),
			Error:  responseError,
		},
	}

	// Convert ChartResults to QuoteResults (simplified mapping)
	for i, chartResult := range allResults {
		response.QuoteResponse.Result[i] = models.QuoteResult{
			Symbol:             chartResult.Meta.Symbol,
			RegularMarketPrice: chartResult.Meta.RegularMarketPrice,
			Currency:           chartResult.Meta.Currency,
			Exchange:           chartResult.Meta.ExchangeName,
		}
	}

	return response, nil
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

	// Yahoo Finance requires proper headers to avoid blocking
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	// Remove gzip encoding to avoid decompression issues
	// req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("Referer", "https://finance.yahoo.com/")
	req.Header.Set("Origin", "https://finance.yahoo.com")

	slog.Debug("Making Yahoo Finance API request",
		"url", fullURL,
	)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("yahoo finance API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	slog.Debug("Yahoo Finance API request successful",
		"url", fullURL,
		"status", resp.StatusCode,
	)

	return nil
}
