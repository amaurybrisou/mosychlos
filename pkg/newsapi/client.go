// Package newsapi
package newsapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// Client provides access to NewsAPI endpoints
type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

// NewClient creates a new NewsAPI client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: "https://newsapi.org/v2",
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetBaseURL allows overriding the default base URL
func (c *Client) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}

// TopHeadlinesParams contains parameters for the top headlines endpoint
type TopHeadlinesParams struct {
	Country  string
	Category string
	PageSize int
	Language string
}

// EverythingParams contains parameters for the everything endpoint
type EverythingParams struct {
	Query    string
	PageSize int
	Language string
}

// GetTopHeadlines retrieves top headlines from NewsAPI
func (c *Client) GetTopHeadlines(ctx context.Context, params TopHeadlinesParams) (*NewsAPIResponse, error) {
	u, err := url.Parse(c.baseURL + "/top-headlines")
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	if params.Country != "" {
		q.Set("country", params.Country)
	}
	if params.Category != "" {
		q.Set("category", params.Category)
	}
	if params.PageSize > 0 {
		q.Set("pageSize", strconv.Itoa(params.PageSize))
	}
	if params.Language != "" {
		q.Set("language", params.Language)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("User-Agent", "Mosychlos/1.0")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result NewsAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetEverything retrieves everything from NewsAPI
func (c *Client) GetEverything(ctx context.Context, params EverythingParams) (*NewsAPIResponse, error) {
	u, err := url.Parse(c.baseURL + "/everything")
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	q := u.Query()
	if params.Query != "" {
		q.Set("q", params.Query)
	}
	if params.PageSize > 0 {
		q.Set("pageSize", strconv.Itoa(params.PageSize))
	}
	if params.Language != "" {
		q.Set("language", params.Language)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("User-Agent", "Mosychlos/1.0")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result NewsAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// ToNewsData converts NewsAPIResponse to models.NewsData
func (r *NewsAPIResponse) ToNewsData() *models.NewsData {
	data := &models.NewsData{
		LastUpdated: time.Now().UTC(),
		Articles:    make([]models.NewsArticle, 0, len(r.Articles)),
	}

	for _, a := range r.Articles {
		var ts time.Time
		if t, err := time.Parse(time.RFC3339, a.PublishedAt); err == nil {
			ts = t
		}

		data.Articles = append(data.Articles, models.NewsArticle{
			Title:       a.Title,
			Source:      a.Source.Name,
			PublishedAt: ts,
			URL:         a.URL,
		})
	}

	return data
}

// NewsAPIResponse represents the response structure from NewsAPI
type NewsAPIResponse struct {
	Status       string        `json:"status"`
	TotalResults int           `json:"totalResults"`
	Articles     []NewsArticle `json:"articles"`
}

// NewsArticle represents a news article from NewsAPI
type NewsArticle struct {
	Source struct {
		Name string `json:"name"`
	} `json:"source"`
	Author      string `json:"author"`
	Content     string `json:"content"`
	Description string `json:"description"`
	Title       string `json:"title"`
	PublishedAt string `json:"publishedAt"`
	URL         string `json:"url"`
	URLToImage  string `json:"urlToImage"`
}
