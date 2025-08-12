package models

import (
	"time"
)

// NewsArticle represents a single news article
// Moved from internal/context/types.go
type NewsArticle struct {
	Title       string    `json:"title"`
	Source      string    `json:"source"`
	PublishedAt time.Time `json:"published_at"`
	Sentiment   float64   `json:"sentiment"`
	Relevance   float64   `json:"relevance"`
	Tags        []string  `json:"tags,omitempty"`
	URL         string    `json:"url,omitempty"`
}

// NewsData represents a collection of news articles with metadata
// Moved from internal/context/types.go
type NewsData struct {
	Articles    []NewsArticle `json:"articles"`
	Summary     string        `json:"summary"`
	LastUpdated time.Time     `json:"last_updated"`
}

// Theme represents a clustered theme from news items
// Shared across cluster and insights phases
type Theme struct {
	Label     string   `json:"label"`
	Tickers   []string `json:"tickers"`
	Rationale string   `json:"rationale"`
}

// NewsAnalysisResult represents structured LLM response for news analysis
type NewsAnalysisResult struct {
	Articles []AnalyzedNewsArticle `json:"articles"`
}

// AnalyzedNewsArticle embeds NewsArticle with analysis fields
type AnalyzedNewsArticle struct {
	NewsArticle                  // Embed original article
	Index               int      `json:"index"`
	MarketDrivers       []string `json:"market_drivers"`
	SectorImplications  []string `json:"sector_implications"`
	InvestmentRelevance string   `json:"investment_relevance"`
	RiskFactors         []string `json:"risk_factors"`
	Opportunities       []string `json:"opportunities"`
	Summary             string   `json:"summary"`
	Analyzed            bool     `json:"analyzed"`
}

// AnalyzedNewsData embeds NewsData with analysis metadata
type AnalyzedNewsData struct {
	NewsData                               // Embed original news data
	AnalyzedArticles []AnalyzedNewsArticle `json:"analyzed_articles"` // Override with analyzed articles
	Analysis         string                `json:"analysis"`
	Analyzed         bool                  `json:"analyzed"`
}
