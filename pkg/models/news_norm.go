package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

// MarshalJSON provides compact JSON representation optimized for AI agents
func (n NewsData) MarshalJSON() ([]byte, error) {
	// limit articles for token efficiency, prioritize by relevance
	maxArticles := 10
	articles := n.Articles
	if len(articles) > maxArticles {
		articles = articles[:maxArticles]
	}

	// calculate average sentiment and relevance
	var avgSentiment, avgRelevance float64
	sentimentCounts := map[string]int{"positive": 0, "negative": 0, "neutral": 0}

	for _, article := range articles {
		avgSentiment += article.Sentiment
		avgRelevance += article.Relevance

		if article.Sentiment > 0.1 {
			sentimentCounts["positive"]++
		} else if article.Sentiment < -0.1 {
			sentimentCounts["negative"]++
		} else {
			sentimentCounts["neutral"]++
		}
	}

	if len(articles) > 0 {
		avgSentiment /= float64(len(articles))
		avgRelevance /= float64(len(articles))
	}

	compact := map[string]any{
		"article_count":   len(n.Articles),
		"shown_articles":  len(articles),
		"avg_sentiment":   avgSentiment,
		"avg_relevance":   avgRelevance,
		"sentiment_split": sentimentCounts,
		"summary":         n.Summary,
		"last_updated":    n.LastUpdated.Format("2006-01-02 15:04"),
		"articles_sample": articles, // limited set for AI analysis
	}

	return json.Marshal(compact)
}

// String provides human-readable summary optimized for AI understanding
func (n NewsData) String() string {
	if len(n.Articles) == 0 {
		return "No news articles available"
	}

	// calculate sentiment distribution
	positive, negative, neutral := 0, 0, 0
	for _, article := range n.Articles {
		if article.Sentiment > 0.1 {
			positive++
		} else if article.Sentiment < -0.1 {
			negative++
		} else {
			neutral++
		}
	}

	sentimentDesc := "neutral"
	if positive > negative+neutral {
		sentimentDesc = "bullish"
	} else if negative > positive+neutral {
		sentimentDesc = "bearish"
	}

	return fmt.Sprintf("News[%d articles]: %s sentiment, updated %s",
		len(n.Articles), sentimentDesc, n.LastUpdated.Format("Jan 2"))
}

// MarshalJSON provides compact JSON representation for AI agents
func (a NewsArticle) MarshalJSON() ([]byte, error) {
	compact := map[string]any{
		"title":        a.Title,
		"source":       a.Source,
		"published_at": a.PublishedAt.Format("2006-01-02"),
		"sentiment":    fmt.Sprintf("%.2f", a.Sentiment),
		"relevance":    fmt.Sprintf("%.2f", a.Relevance),
	}

	// include tags if meaningful
	if len(a.Tags) > 0 && len(a.Tags) <= 5 {
		compact["tags"] = a.Tags
	}

	// include URL if available (useful for AI to reference)
	if a.URL != "" {
		compact["url"] = a.URL
	}

	return json.Marshal(compact)
}

// String provides human-readable article summary
func (a NewsArticle) String() string {
	sentimentIcon := "→"
	if a.Sentiment > 0.1 {
		sentimentIcon = "↗"
	} else if a.Sentiment < -0.1 {
		sentimentIcon = "↘"
	}

	relevanceDesc := "low"
	if a.Relevance > 0.7 {
		relevanceDesc = "high"
	} else if a.Relevance > 0.3 {
		relevanceDesc = "med"
	}

	return fmt.Sprintf("%s %s [%s] (%s relevance) - %s",
		sentimentIcon, a.Title, a.Source, relevanceDesc,
		a.PublishedAt.Format("Jan 2"))
}

// MarshalJSON provides compact JSON representation for AI agents
func (a AnalyzedNewsArticle) MarshalJSON() ([]byte, error) {
	compact := map[string]any{
		"index":                a.Index,
		"title":                a.Title,
		"source":               a.Source,
		"investment_relevance": a.InvestmentRelevance,
		"summary":              a.Summary,
		"analyzed":             a.Analyzed,
	}

	// include key analysis fields if available
	if len(a.MarketDrivers) > 0 {
		compact["market_drivers"] = a.MarketDrivers
	}

	if len(a.SectorImplications) > 0 {
		compact["sector_implications"] = a.SectorImplications
	}

	if len(a.RiskFactors) > 0 {
		compact["risk_factors"] = a.RiskFactors
	}

	if len(a.Opportunities) > 0 {
		compact["opportunities"] = a.Opportunities
	}

	return json.Marshal(compact)
}

// String provides human-readable analyzed article summary
func (a AnalyzedNewsArticle) String() string {
	status := "unanalyzed"
	if a.Analyzed {
		status = fmt.Sprintf("%s relevance", strings.ToLower(a.InvestmentRelevance))
	}

	return fmt.Sprintf("[%d] %s - %s (%s)", a.Index, a.Title, a.Source, status)
}

// MarshalJSON provides compact JSON representation for AI agents
func (a AnalyzedNewsData) MarshalJSON() ([]byte, error) {
	// use the base NewsData marshaling but add analysis info
	baseData, err := json.Marshal(a.NewsData)
	if err != nil {
		return nil, err
	}

	var base map[string]any
	if err := json.Unmarshal(baseData, &base); err != nil {
		return nil, err
	}

	// add analysis-specific fields
	base["analysis"] = a.Analysis
	base["analyzed"] = a.Analyzed

	// override with analyzed articles if available
	if len(a.AnalyzedArticles) > 0 {
		base["articles_sample"] = a.AnalyzedArticles
		base["analyzed_count"] = len(a.AnalyzedArticles)
	}

	return json.Marshal(base)
}

// String provides human-readable analyzed news summary
func (a AnalyzedNewsData) String() string {
	baseStr := a.NewsData.String()
	if a.Analyzed {
		return fmt.Sprintf("%s + AI analysis (%d analyzed)", baseStr, len(a.AnalyzedArticles))
	}
	return baseStr
}
