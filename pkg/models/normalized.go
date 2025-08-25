package models

import "time"

// ===== PORTFOLIO-FOCUSED NORMALIZATION =====

// NormalizedPortfolio represents ONLY core portfolio data for AI analysis.
// No market data, no complex performance calculations - just the portfolio itself.
type NormalizedPortfolio struct {
	// Core Portfolio Identity
	TotalValueUSD float64   `json:"total_value_usd" jsonschema_description:"Total portfolio value in USD"`
	BaseCurrency  string    `json:"base_currency" jsonschema_description:"Primary currency used for portfolio valuation"`
	AsOfDate      time.Time `json:"as_of_date" jsonschema_description:"Date when portfolio data was captured"`
	HoldingsCount int       `json:"holdings_count" jsonschema_description:"Total number of positions in the portfolio"`

	// Asset Allocation (percentages 0-100)
	AssetAllocations  map[string]float64 `json:"asset_allocations" jsonschema_description:"Percentage allocation by asset class (stocks, bonds, cash, etc.)"`
	RegionAllocations map[string]float64 `json:"region_allocations" jsonschema_description:"Geographic exposure as percentage by region (US, Europe, Asia, etc.)"`
	SectorAllocations map[string]float64 `json:"sector_allocations,omitempty" jsonschema_description:"Industry sector exposure as percentage (tech, healthcare, financials, etc.)"` // Only if we have sector data

	// Individual Holdings
	Holdings []NormalizedHolding `json:"holdings" jsonschema_description:"List of individual positions in the portfolio"`

	// Basic Risk Metrics (calculated from holdings)
	RiskMetrics NormalizedRisk `json:"risk_metrics" jsonschema_description:"Portfolio concentration and diversification risk measures"`
}

// NormalizedHolding represents a single position with clean, consistent fields
type NormalizedHolding struct {
	Symbol        string  `json:"symbol" jsonschema_description:"Stock ticker symbol or instrument identifier"`
	Name          string  `json:"name,omitempty" jsonschema_description:"Company or fund name"`
	WeightPercent float64 `json:"weight_percent" jsonschema_description:"Position size as percentage of total portfolio value"` // 0-100
	ValueUSD      float64 `json:"value_usd" jsonschema_description:"Current market value of holding in USD"`
	Quantity      float64 `json:"quantity" jsonschema_description:"Number of shares or units owned"`
	AssetClass    string  `json:"asset_class" jsonschema_description:"Investment type classification (stock, ETF, bond, cash, crypto)"` // "stock", "etf", "bond", "cash", "crypto"
	Region        string  `json:"region" jsonschema_description:"Geographic market exposure (US, Europe, Asia, Emerging, Global)"`      // "US", "Europe", "Asia", "Emerging", "Global"
	Sector        string  `json:"sector,omitempty" jsonschema_description:"Industry sector classification (technology, healthcare, financials, etc.)"`
	Currency      string  `json:"currency" jsonschema_description:"Currency denomination of the investment"`

	// Simple flags for AI analysis
	IsLargePosition bool `json:"is_large_position" jsonschema_description:"Indicates if position exceeds 5% of total portfolio"`        // >5% of portfolio
	IsForeign       bool `json:"is_foreign" jsonschema_description:"Indicates if investment is in a currency other than base currency"` // Outside base currency
}

// NormalizedRisk provides ONLY concentration and diversification metrics
// that can be calculated directly from portfolio holdings
type NormalizedRisk struct {
	// Concentration Risk
	HerfindahlIndex    float64 `json:"herfindahl_index" jsonschema_description:"Portfolio concentration measure, higher values indicate more concentrated holdings"` // 0-1
	LargestPositionPct float64 `json:"largest_position_pct" jsonschema_description:"Percentage weight of the single largest holding in the portfolio"`               // 0-100
	Top5PositionsPct   float64 `json:"top5_positions_pct" jsonschema_description:"Combined percentage weight of the five largest positions"`                         // 0-100
	EffectiveHoldings  float64 `json:"effective_holdings" jsonschema_description:"Number of effectively diversified holdings, calculated as 1/Herfindahl Index"`     // 1/Herfindahl

	// Geographic/Sector Concentration
	RegionConcentration float64 `json:"region_concentration" jsonschema_description:"Highest percentage allocation to any single geographic region"` // Max single region %
	SectorConcentration float64 `json:"sector_concentration" jsonschema_description:"Highest percentage allocation to any single industry sector"`   // Max single sector %
	ForeignCurrencyPct  float64 `json:"foreign_currency_pct" jsonschema_description:"Percentage of portfolio exposed to non-base currencies"`        // 0-100
}

// ===== MARKET DATA (FROM TOOLS) =====

// NormalizedMarketData aggregates market data from various tools.
// Each tool contributes its part, final context is assembled for AI.
type NormalizedMarketData struct {
	AsOfDate    time.Time `json:"as_of_date" jsonschema_description:"Date of market data snapshot"`
	LastUpdated time.Time `json:"last_updated" jsonschema_description:"Timestamp when market data was last refreshed"`

	// From FMP tool
	Indices  []NormalizedIndex `json:"indices,omitempty" jsonschema_description:"Major market index levels and performance (S&P 500, NASDAQ, etc.)"` // SPY, QQQ, etc.
	VIXLevel *float64          `json:"vix_level,omitempty" jsonschema_description:"CBOE Volatility Index level indicating market fear/complacency"`

	// From FRED tool
	InterestRates []NormalizedRate `json:"interest_rates,omitempty" jsonschema_description:"Key interest rates (Fed funds, 10-year treasury, etc.)"` // Fed funds, 10Y, etc.
	EconomicData  []NormalizedRate `json:"economic_data,omitempty" jsonschema_description:"Economic indicators (GDP, CPI, unemployment, etc.)"`      // GDP, CPI, etc.

	// From News API tool (derived)
	MarketSentiment string `json:"market_sentiment,omitempty" jsonschema_description:"Overall market sentiment derived from news analysis"` // "positive", "negative", "neutral"
	MarketRegime    string `json:"market_regime,omitempty" jsonschema_description:"Current market environment characterization"`            // "bull", "bear", "sideways"

	// Currency rates (from FMP or other)
	CurrencyRates []NormalizedCurrencyRate `json:"currency_rates,omitempty" jsonschema_description:"Exchange rates for major currency pairs"` // vs USD
}

type NormalizedCurrencyRate struct {
	FromCurrency string  `json:"from_currency" jsonschema_description:"Base currency code (e.g., EUR, GBP, JPY)"`
	ToCurrency   string  `json:"to_currency" jsonschema_description:"Quote currency code (typically USD)"`
	ExchangeRate float64 `json:"exchange_rate" jsonschema_description:"Current exchange rate from base to quote currency"`
}

// NormalizedIndex represents market index data from FMP tool
type NormalizedIndex struct {
	Symbol       string  `json:"symbol" jsonschema_description:"Market index ticker symbol (e.g., SPY, QQQ, VTI)"`
	CurrentLevel float64 `json:"current_level" jsonschema_description:"Current price or index level"`
	Change1Day   float64 `json:"change_1day" jsonschema_description:"One-day percentage change as decimal (-0.02 = -2%)"` // decimal (-0.02 = -2%)
	Change1Week  float64 `json:"change_1week" jsonschema_description:"One-week percentage change as decimal"`
	Change1Month float64 `json:"change_1month" jsonschema_description:"One-month percentage change as decimal"`
	Change1Year  float64 `json:"change_1year" jsonschema_description:"One-year percentage change as decimal"`
}

// NormalizedRate represents economic data from FRED tool
type NormalizedRate struct {
	Name         string    `json:"name" jsonschema_description:"Name of the economic indicator or interest rate"`
	CurrentValue float64   `json:"current_value" jsonschema_description:"Most recent data point value"`
	Units        string    `json:"units" jsonschema_description:"Unit of measurement (percent, index, billions_usd, etc.)"` // "percent", "index", "billions_usd"
	LastUpdated  time.Time `json:"last_updated" jsonschema_description:"Date when this data point was last updated"`
	Change1Month *float64  `json:"change_1month,omitempty" jsonschema_description:"Change from one month ago (absolute units)"`
	Change1Year  *float64  `json:"change_1year,omitempty" jsonschema_description:"Change from one year ago (absolute units)"`
	Trend        string    `json:"trend" jsonschema_description:"Directional trend characterization"` // "rising", "falling", "stable"
}

// ===== NEWS DATA (FROM NEWS TOOL) =====

// NormalizedNewsContext provides AI-optimized news analysis
type NormalizedNewsContext struct {
	AsOfDate    time.Time `json:"as_of_date" jsonschema_description:"Date of news analysis snapshot"`
	LastUpdated time.Time `json:"last_updated" jsonschema_description:"Timestamp when news data was last refreshed"`

	// Aggregated insights
	OverallSentiment string   `json:"overall_sentiment" jsonschema_description:"Aggregate sentiment across all news sources"`  // "positive", "negative", "neutral"
	KeyThemes        []string `json:"key_themes" jsonschema_description:"Primary themes and topics dominating financial news"` // ["inflation", "fed_policy", "earnings"]
	MarketMood       string   `json:"market_mood" jsonschema_description:"Characterization of overall investor psychology"`    // "optimistic", "cautious", "fearful"

	// Recent headlines (filtered for relevance)
	RecentHeadlines []NormalizedNewsItem `json:"recent_headlines" jsonschema_description:"Most relevant recent financial news headlines"`
}

// NormalizedNewsItem represents a single news item
type NormalizedNewsItem struct {
	Title       string    `json:"title" jsonschema_description:"News article headline or title"`
	Source      string    `json:"source" jsonschema_description:"News publication or source name"`
	PublishedAt time.Time `json:"published_at" jsonschema_description:"Publication timestamp of the news article"`
	Sentiment   string    `json:"sentiment" jsonschema_description:"Sentiment classification of the news content"`                   // "positive", "negative", "neutral"
	Relevance   float64   `json:"relevance" jsonschema_description:"Relevance score to portfolio and financial markets (0-1 scale)"` // 0-1, how relevant to portfolio/markets
	Summary     string    `json:"summary,omitempty" jsonschema_description:"Brief summary of the news article content"`              // Brief AI-generated summary if available
}

// ===== MACRO/ECONOMIC DATA (FROM FRED TOOL) =====

// NormalizedMacroData provides standardized economic indicators from FRED tool
type NormalizedMacroData struct {
	AsOfDate     time.Time                   `json:"as_of_date" jsonschema_description:"Date of macroeconomic data snapshot"`
	EconomicData map[string]NormalizedSeries `json:"economic_data" jsonschema_description:"Economic indicators mapped by series identifier"`
	LastUpdated  time.Time                   `json:"last_updated" jsonschema_description:"Timestamp when macroeconomic data was last updated"`
}

// NormalizedSeries represents an economic data series from FRED
type NormalizedSeries struct {
	Name           string    `json:"name" jsonschema_description:"Full name of the economic data series"`
	CurrentValue   float64   `json:"current_value" jsonschema_description:"Most recent data point value"`
	Units          string    `json:"units" jsonschema_description:"Unit of measurement (percent, index, millions_usd, etc.)"` // "percent", "index", "millions_usd"
	LastUpdated    time.Time `json:"last_updated" jsonschema_description:"Date when this data series was last updated"`
	Change1Month   *float64  `json:"change_1month,omitempty" jsonschema_description:"Change from one month ago (absolute units)"`     // Change vs 1 month ago
	Change1Quarter *float64  `json:"change_1quarter,omitempty" jsonschema_description:"Change from one quarter ago (absolute units)"` // Change vs 1 quarter ago
	Change1Year    *float64  `json:"change_1year,omitempty" jsonschema_description:"Change from one year ago (absolute units)"`       // Change vs 1 year ago
	Trend          string    `json:"trend" jsonschema_description:"Directional trend characterization based on recent data"`          // "rising", "falling", "stable"
}
