package models

import "time"

// ===== PORTFOLIO-FOCUSED NORMALIZATION =====

// NormalizedPortfolio represents ONLY core portfolio data for AI analysis.
// No market data, no complex performance calculations - just the portfolio itself.
type NormalizedPortfolio struct {
	// Core Portfolio Identity
	TotalValueUSD float64   `json:"total_value_usd"`
	BaseCurrency  string    `json:"base_currency"`
	AsOfDate      time.Time `json:"as_of_date"`
	HoldingsCount int       `json:"holdings_count"`

	// Asset Allocation (percentages 0-100)
	AssetAllocations  map[string]float64 `json:"asset_allocations"`
	RegionAllocations map[string]float64 `json:"region_allocations"`
	SectorAllocations map[string]float64 `json:"sector_allocations,omitempty"` // Only if we have sector data

	// Individual Holdings
	Holdings []NormalizedHolding `json:"holdings"`

	// Basic Risk Metrics (calculated from holdings)
	RiskMetrics NormalizedRisk `json:"risk_metrics"`
}

// NormalizedHolding represents a single position with clean, consistent fields
type NormalizedHolding struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name,omitempty"`
	WeightPercent float64 `json:"weight_percent"` // 0-100
	ValueUSD      float64 `json:"value_usd"`
	Quantity      float64 `json:"quantity"`
	AssetClass    string  `json:"asset_class"` // "stock", "etf", "bond", "cash", "crypto"
	Region        string  `json:"region"`      // "US", "Europe", "Asia", "Emerging", "Global"
	Sector        string  `json:"sector,omitempty"`
	Currency      string  `json:"currency"`

	// Simple flags for AI analysis
	IsLargePosition bool `json:"is_large_position"` // >5% of portfolio
	IsForeign       bool `json:"is_foreign"`        // Outside base currency
}

// NormalizedRisk provides ONLY concentration and diversification metrics
// that can be calculated directly from portfolio holdings
type NormalizedRisk struct {
	// Concentration Risk
	HerfindahlIndex    float64 `json:"herfindahl_index"`     // 0-1
	LargestPositionPct float64 `json:"largest_position_pct"` // 0-100
	Top5PositionsPct   float64 `json:"top5_positions_pct"`   // 0-100
	EffectiveHoldings  float64 `json:"effective_holdings"`   // 1/Herfindahl

	// Geographic/Sector Concentration
	RegionConcentration float64 `json:"region_concentration"` // Max single region %
	SectorConcentration float64 `json:"sector_concentration"` // Max single sector %
	ForeignCurrencyPct  float64 `json:"foreign_currency_pct"` // 0-100
}

// ===== MARKET DATA (FROM TOOLS) =====

// NormalizedMarketData aggregates market data from various tools.
// Each tool contributes its part, final context is assembled for AI.
type NormalizedMarketData struct {
	AsOfDate    time.Time `json:"as_of_date"`
	LastUpdated time.Time `json:"last_updated"`

	// From FMP tool
	Indices  map[string]NormalizedIndex `json:"indices,omitempty"` // SPY, QQQ, etc.
	VIXLevel *float64                   `json:"vix_level,omitempty"`

	// From FRED tool
	InterestRates map[string]NormalizedRate `json:"interest_rates,omitempty"` // Fed funds, 10Y, etc.
	EconomicData  map[string]NormalizedRate `json:"economic_data,omitempty"`  // GDP, CPI, etc.

	// From News API tool (derived)
	MarketSentiment string `json:"market_sentiment,omitempty"` // "positive", "negative", "neutral"
	MarketRegime    string `json:"market_regime,omitempty"`    // "bull", "bear", "sideways"

	// Currency rates (from FMP or other)
	CurrencyRates map[string]float64 `json:"currency_rates,omitempty"` // vs USD
}

// NormalizedIndex represents market index data from FMP tool
type NormalizedIndex struct {
	Symbol       string  `json:"symbol"`
	CurrentLevel float64 `json:"current_level"`
	Change1Day   float64 `json:"change_1day"` // decimal (-0.02 = -2%)
	Change1Week  float64 `json:"change_1week"`
	Change1Month float64 `json:"change_1month"`
	Change1Year  float64 `json:"change_1year"`
}

// NormalizedRate represents economic data from FRED tool
type NormalizedRate struct {
	Name         string    `json:"name"`
	CurrentValue float64   `json:"current_value"`
	Units        string    `json:"units"` // "percent", "index", "billions_usd"
	LastUpdated  time.Time `json:"last_updated"`
	Change1Month *float64  `json:"change_1month,omitempty"`
	Change1Year  *float64  `json:"change_1year,omitempty"`
	Trend        string    `json:"trend"` // "rising", "falling", "stable"
}

// ===== NEWS DATA (FROM NEWS TOOL) =====

// NormalizedNewsContext provides AI-optimized news analysis
type NormalizedNewsContext struct {
	AsOfDate    time.Time `json:"as_of_date"`
	LastUpdated time.Time `json:"last_updated"`

	// Aggregated insights
	OverallSentiment string   `json:"overall_sentiment"` // "positive", "negative", "neutral"
	KeyThemes        []string `json:"key_themes"`        // ["inflation", "fed_policy", "earnings"]
	MarketMood       string   `json:"market_mood"`       // "optimistic", "cautious", "fearful"

	// Recent headlines (filtered for relevance)
	RecentHeadlines []NormalizedNewsItem `json:"recent_headlines"`
}

// NormalizedNewsItem represents a single news item
type NormalizedNewsItem struct {
	Title       string    `json:"title"`
	Source      string    `json:"source"`
	PublishedAt time.Time `json:"published_at"`
	Sentiment   string    `json:"sentiment"`         // "positive", "negative", "neutral"
	Relevance   float64   `json:"relevance"`         // 0-1, how relevant to portfolio/markets
	Summary     string    `json:"summary,omitempty"` // Brief AI-generated summary if available
}

// ===== MACRO/ECONOMIC DATA (FROM FRED TOOL) =====

// NormalizedMacroData provides standardized economic indicators from FRED tool
type NormalizedMacroData struct {
	AsOfDate     time.Time                   `json:"as_of_date"`
	EconomicData map[string]NormalizedSeries `json:"economic_data"`
	LastUpdated  time.Time                   `json:"last_updated"`
}

// NormalizedSeries represents an economic data series from FRED
type NormalizedSeries struct {
	Name           string    `json:"name"`
	CurrentValue   float64   `json:"current_value"`
	Units          string    `json:"units"` // "percent", "index", "millions_usd"
	LastUpdated    time.Time `json:"last_updated"`
	Change1Month   *float64  `json:"change_1month,omitempty"`   // Change vs 1 month ago
	Change1Quarter *float64  `json:"change_1quarter,omitempty"` // Change vs 1 quarter ago
	Change1Year    *float64  `json:"change_1year,omitempty"`    // Change vs 1 year ago
	Trend          string    `json:"trend"`                     // "rising", "falling", "stable"
}
