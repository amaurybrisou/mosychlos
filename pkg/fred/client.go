package fred

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

// Client provides access to FRED API
type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

// Config holds FRED client configuration
type Config struct {
	APIKey  string
	BaseURL string
	Timeout time.Duration
}

// NewClient creates a new FRED client
func NewClient(cfg Config) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api_key is required for FRED API")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.stlouisfed.org/fred"
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

// GetSeries retrieves economic data series
func (c *Client) GetSeries(ctx context.Context, seriesID string) (*models.FREDSeries, error) {
	endpoint := "/series"
	params := url.Values{
		"series_id": {seriesID},
		"api_key":   {c.apiKey},
		"file_type": {"json"},
	}

	var response models.FREDSeriesResponse
	if err := c.makeRequest(ctx, endpoint, params, &response); err != nil {
		return nil, fmt.Errorf("failed to get series %s: %w", seriesID, err)
	}

	if len(response.Seriess) == 0 {
		return nil, fmt.Errorf("no series found for ID %s", seriesID)
	}

	return &response.Seriess[0], nil
}

// GetSeriesObservations retrieves data observations for a series
func (c *Client) GetSeriesObservations(ctx context.Context, seriesID string, startDate, endDate time.Time) (*models.FREDSeriesObservations, error) {
	endpoint := "/series/observations"
	params := url.Values{
		"series_id":  {seriesID},
		"api_key":    {c.apiKey},
		"file_type":  {"json"},
		"sort_order": {"desc"},
		"limit":      {"1000"},
	}

	if !startDate.IsZero() {
		params.Set("observation_start", startDate.Format("2006-01-02"))
	}
	if !endDate.IsZero() {
		params.Set("observation_end", endDate.Format("2006-01-02"))
	}

	var response models.FREDObservationsResponse
	if err := c.makeRequest(ctx, endpoint, params, &response); err != nil {
		return nil, fmt.Errorf("failed to get observations for series %s: %w", seriesID, err)
	}

	return &models.FREDSeriesObservations{
		SeriesID:     seriesID,
		Observations: response.Observations,
		Count:        response.Count,
		Offset:       response.Offset,
		Limit:        response.Limit,
		OrderBy:      response.OrderBy,
		SortOrder:    response.SortOrder,
	}, nil
}

// GetCategories retrieves economic data categories
func (c *Client) GetCategories(ctx context.Context, categoryID string) (*models.FREDCategories, error) {
	endpoint := "/category"
	params := url.Values{
		"api_key":   {c.apiKey},
		"file_type": {"json"},
	}

	if categoryID != "" {
		params.Set("category_id", categoryID)
	}

	var response models.FREDCategoriesResponse
	if err := c.makeRequest(ctx, endpoint, params, &response); err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return &models.FREDCategories{
		Categories: response.Categories,
	}, nil
}

// GetCategorySeries retrieves series in a category
func (c *Client) GetCategorySeries(ctx context.Context, categoryID string) (*models.FREDCategorySeries, error) {
	endpoint := "/category/series"
	params := url.Values{
		"category_id": {categoryID},
		"api_key":     {c.apiKey},
		"file_type":   {"json"},
		"limit":       {"1000"},
	}

	var response models.FREDSeriesResponse
	if err := c.makeRequest(ctx, endpoint, params, &response); err != nil {
		return nil, fmt.Errorf("failed to get series for category %s: %w", categoryID, err)
	}

	return &models.FREDCategorySeries{
		CategoryID: categoryID,
		Series:     response.Seriess,
		Count:      response.Count,
		Offset:     response.Offset,
		Limit:      response.Limit,
	}, nil
}

// SearchSeries searches for economic data series
func (c *Client) SearchSeries(ctx context.Context, searchText string, limit int) (*models.FREDSearchResults, error) {
	endpoint := "/series/search"
	params := url.Values{
		"search_text": {searchText},
		"api_key":     {c.apiKey},
		"file_type":   {"json"},
	}

	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	} else {
		params.Set("limit", "100")
	}

	var response models.FREDSeriesResponse
	if err := c.makeRequest(ctx, endpoint, params, &response); err != nil {
		return nil, fmt.Errorf("failed to search series with text '%s': %w", searchText, err)
	}

	return &models.FREDSearchResults{
		SearchText: searchText,
		Series:     response.Seriess,
		Count:      response.Count,
		Offset:     response.Offset,
		Limit:      response.Limit,
	}, nil
}

// GetGeoFREDRegionalData retrieves regional data from FRED GeoFRED API
func (c *Client) GetGeoFREDRegionalData(ctx context.Context, seriesGroup, date, regionType, units, frequency, season string) (*models.FREDGeoRegionalData, error) {
	// Construct URL for GeoFRED API (different base URL)
	geoURL := "https://api.stlouisfed.org/geofred/regional/data"

	params := url.Values{
		"api_key":      {c.apiKey},
		"series_group": {seriesGroup},
		"date":         {date},
		"region_type":  {regionType},
		"units":        {units},
		"frequency":    {frequency},
		"season":       {season},
		"file_type":    {"json"},
	}

	fullURL := geoURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mosychlos/1.0")

	slog.Debug("Making FRED GeoFRED API request",
		"url", fullURL,
	)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("FRED GeoFRED API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	var response models.FREDGeoRegionalData
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	slog.Debug("FRED GeoFRED API request successful",
		"url", fullURL,
		"status", resp.StatusCode,
	)

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

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Mosychlos/1.0")

	slog.Debug("Making FRED API request",
		"url", fullURL,
	)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("FRED API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	slog.Debug("FRED API request successful",
		"url", fullURL,
		"status", resp.StatusCode,
	)

	return nil
}
