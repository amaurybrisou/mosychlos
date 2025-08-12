package sec

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"log/slog"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// Client provides access to SEC EDGAR API
type Client struct {
	baseURL   string
	userAgent string
	http      *http.Client
}

// Config holds SEC client configuration
type Config struct {
	UserAgent string
	BaseURL   string
	Timeout   time.Duration
}

// NewClient creates a new SEC EDGAR client
func NewClient(cfg Config) (*Client, error) {
	if cfg.UserAgent == "" {
		return nil, fmt.Errorf("user_agent is required for SEC API")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://data.sec.gov"
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &Client{
		baseURL:   baseURL,
		userAgent: cfg.UserAgent,
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

// GetCompanyTickers retrieves company ticker information
func (c *Client) GetCompanyTickers(ctx context.Context) (*models.CompanyTickersResponse, error) {
	endpoint := "/files/company_tickers.json"

	var response models.CompanyTickersResponse
	if err := c.makeRequest(ctx, endpoint, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get company tickers: %w", err)
	}

	return &response, nil
}

// GetCompanyFacts retrieves company facts by CIK
func (c *Client) GetCompanyFacts(ctx context.Context, cik string) (*models.CompanyFactsResponse, error) {
	// Ensure CIK is properly formatted (10 digits with leading zeros)
	cik = strings.TrimSpace(cik)
	if len(cik) < 10 {
		cik = fmt.Sprintf("%010s", cik)
	}

	endpoint := fmt.Sprintf("/api/xbrl/companyfacts/CIK%s.json", cik)

	var response models.CompanyFactsResponse
	if err := c.makeRequest(ctx, endpoint, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get company facts for CIK %s: %w", cik, err)
	}

	return &response, nil
}

// GetCompanyFilings retrieves company filings
func (c *Client) GetCompanyFilings(ctx context.Context, params models.FilingsParams) (*models.CompanyFilingsResponse, error) {
	// Build query parameters
	queryParams := url.Values{}
	if params.CIK != "" {
		cik := strings.TrimSpace(params.CIK)
		if len(cik) < 10 {
			cik = fmt.Sprintf("%010s", cik)
		}
		queryParams.Set("CIK", cik)
	}
	if params.Type != "" {
		queryParams.Set("type", params.Type)
	}
	if params.DateBefore != "" {
		queryParams.Set("dateb", params.DateBefore)
	}
	if params.Owner != "" {
		queryParams.Set("owner", params.Owner)
	}
	if params.Count > 0 {
		queryParams.Set("count", fmt.Sprintf("%d", params.Count))
	}
	if params.Start > 0 {
		queryParams.Set("start", fmt.Sprintf("%d", params.Start))
	}

	endpoint := "/api/xbrl/submissions"
	if len(queryParams) > 0 {
		endpoint += "?" + queryParams.Encode()
	}

	var response models.CompanyFilingsResponse
	if err := c.makeRequest(ctx, endpoint, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get company filings: %w", err)
	}

	return &response, nil
}

// GetInsiderTransactions retrieves insider trading transactions
func (c *Client) GetInsiderTransactions(ctx context.Context, cik string) (*models.InsiderTransactionsResponse, error) {
	// Format CIK
	cik = strings.TrimSpace(cik)
	if len(cik) < 10 {
		cik = fmt.Sprintf("%010s", cik)
	}

	endpoint := fmt.Sprintf("/api/xbrl/submissions/CIK%s.json", cik)

	var response models.InsiderTransactionsResponse
	if err := c.makeRequest(ctx, endpoint, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get insider transactions for CIK %s: %w", cik, err)
	}

	return &response, nil
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

	// SEC requires proper User-Agent header
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")

	slog.Debug("Making SEC API request",
		"url", fullURL,
		"user_agent", c.userAgent,
	)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SEC API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	slog.Debug("SEC API request successful",
		"url", fullURL,
		"status", resp.StatusCode,
	)

	return nil
}
