package newsapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/amaurybrisou/mosychlos/pkg/newsapi"
)

// NewsApiTool implements the Tool interface for NewsAPI
type NewsApiTool struct {
	client    *newsapi.Client
	locale    string
	sharedBag bag.SharedBag
}

var _ models.Tool = &NewsApiTool{}

// NewFromConfig creates a new Provider from config
func NewFromConfig(cfg *config.NewsAPIConfig, sharedBag bag.SharedBag) (*NewsApiTool, error) {
	return New(cfg.APIKey, cfg.BaseURL, cfg.Locale, sharedBag)
}

// New constructs a NewsAPI provider using NEWSAPI_API_KEY from env or config API keys.
func New(apiKey, baseUrl, locale string, sharedBag bag.SharedBag) (*NewsApiTool, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("newsapi: missing NEWSAPI_API_KEY")
	}

	if locale == "" {
		locale = "en"
	}

	client := newsapi.NewClient(apiKey)
	if baseUrl != "" && baseUrl != "https://newsapi.org/v2" {
		client.SetBaseURL(baseUrl) // We'll add this method to the client
	}

	return &NewsApiTool{
		client:    client,
		locale:    locale,
		sharedBag: sharedBag,
	}, nil
}

func (p *NewsApiTool) Name() string {
	return keys.NewsApi.String()
}

func (p *NewsApiTool) Key() keys.Key {
	return keys.NewsApi
}

func (p *NewsApiTool) Description() string {
	return "NewsAPI provider for fetching news articles"
}

func (p *NewsApiTool) Tags() []string {
	return []string{"news", "newsapi"}
}

func (t *NewsApiTool) IsExternal() bool { return false }

func (p *NewsApiTool) Definition() models.ToolDef {
	return models.ToolDef{
		Type: "function",
		Function: models.FunctionDef{
			Name:        "news_api",
			Description: "Fetch current news headlines for given topics. Supports categories like business, technology, health, sports, etc.",
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"topics": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "News topics or categories to search for (e.g., 'business', 'technology', 'AAPL', 'Tesla')",
					},
				},
				"required": []string{"topics"},
			},
		},
	}
}

func (p *NewsApiTool) Run(ctx context.Context, args string) (string, error) {
	slog.Debug("Running NewsAPI tool",
		"tool", p.Name(),
		"args", args,
	)

	var params struct {
		Topics []string `json:"topics"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}

	if len(params.Topics) == 0 {
		return "", fmt.Errorf("topics parameter is required")
	}

	newsData, err := p.fetch(ctx, params.Topics)
	if err != nil {
		slog.Error("NewsAPI fetch failed",
			"tool", p.Name(),
			"error", err,
			"topics", params.Topics,
		)
		return "", fmt.Errorf("newsapi fetch failed: %w", err)
	}

	// Convert to map for Marshal/Unmarshal pattern
	dataMap, err := json.Marshal(newsData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal news data: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(dataMap, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	response, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	slog.Info("NewsAPI tool executed successfully",
		"tool", p.Name(),
		"topics", params.Topics,
		"articles_count", len(newsData.Articles),
	)

	return string(response), nil
}

// fetch retrieves news data for the given topics using NewsAPI
func (p *NewsApiTool) fetch(ctx context.Context, topics []string) (*models.NewsData, error) {
	slog.Debug("Fetching news",
		"topics", topics,
		"locale", p.locale,
	)

	// Determine if topics are categories or search queries
	var topHeadlinesResponse, everythingResponse *newsapi.NewsAPIResponse

	// Check if topics contain known categories
	categories := map[string]bool{
		"business":      true,
		"entertainment": true,
		"general":       true,
		"health":        true,
		"science":       true,
		"sports":        true,
		"technology":    true,
	}

	hasCategories := false
	hasQueries := false

	for _, topic := range topics {
		if categories[topic] {
			hasCategories = true
		} else {
			hasQueries = true
		}
	}

	// Use top headlines for categories
	if hasCategories && !hasQueries {
		for _, topic := range topics {
			if categories[topic] {
				params := newsapi.TopHeadlinesParams{
					Category: topic,
					PageSize: 20,
					Language: p.locale,
				}

				response, err := p.client.GetTopHeadlines(ctx, params)
				if err != nil {
					slog.Error("Failed to get top headlines",
						"error", err,
						"category", topic,
					)
					continue
				}

				if topHeadlinesResponse == nil {
					topHeadlinesResponse = response
				} else {
					// Merge responses
					topHeadlinesResponse.Articles = append(topHeadlinesResponse.Articles, response.Articles...)
					topHeadlinesResponse.TotalResults += response.TotalResults
				}
			}
		}
	}

	// Use everything endpoint for search queries
	if hasQueries {
		for _, topic := range topics {
			if !categories[topic] {
				params := newsapi.EverythingParams{
					Query:    topic,
					PageSize: 20,
					Language: p.locale,
				}

				response, err := p.client.GetEverything(ctx, params)
				if err != nil {
					slog.Error("Failed to get everything",
						"error", err,
						"query", topic,
					)
					continue
				}

				if everythingResponse == nil {
					everythingResponse = response
				} else {
					// Merge responses
					everythingResponse.Articles = append(everythingResponse.Articles, response.Articles...)
					everythingResponse.TotalResults += response.TotalResults
				}
			}
		}
	}

	// Combine results
	var finalResponse *newsapi.NewsAPIResponse
	if topHeadlinesResponse != nil && everythingResponse != nil {
		// Merge both
		finalResponse = topHeadlinesResponse
		finalResponse.Articles = append(finalResponse.Articles, everythingResponse.Articles...)
		finalResponse.TotalResults += everythingResponse.TotalResults
	} else if topHeadlinesResponse != nil {
		finalResponse = topHeadlinesResponse
	} else if everythingResponse != nil {
		finalResponse = everythingResponse
	} else {
		return nil, fmt.Errorf("no news data retrieved for topics: %v", topics)
	}

	// Convert to models.NewsData
	newsData := finalResponse.ToNewsData()

	slog.Info("NewsAPI fetch completed",
		"topics", topics,
		"articles_retrieved", len(newsData.Articles),
	)

	return newsData, nil
}
