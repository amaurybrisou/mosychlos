package models

import "time"

// InvestmentResearchResult represents the result of an investment research query
type InvestmentResearchResult struct {
	ExecutiveSummary   ExecutiveSummary    `json:"executive_summary"`
	RegionalContext    RegionalContext     `json:"regional_context"`
	ResearchFindings   []ResearchFinding   `json:"research_findings"`
	MarketAnalysis     MarketAnalysis      `json:"market_analysis"`
	InvestmentThemes   []InvestmentTheme   `json:"investment_themes"`
	RiskConsiderations []RiskFactor        `json:"risk_considerations"`
	ActionableInsights []ActionableInsight `json:"actionable_insights"`
	Sources            []SearchSource      `json:"sources"`
	Metadata           AnalysisMetadata    `json:"metadata"`
}

// ExecutiveSummary provides a summary of the investment research findings
type ExecutiveSummary struct {
	KeyTakeaways       []string `json:"key_takeaways"`
	MarketOutlook      string   `json:"market_outlook"` // "bullish", "bearish", "neutral"
	RecommendedActions []string `json:"recommended_actions"`
	TimeHorizon        string   `json:"time_horizon"` // "short_term", "medium_term", "long_term"
}

// RegionalContext provides context about the region being analyzed
type RegionalContext struct {
	Country           string            `json:"country"`
	Language          string            `json:"language"`
	CurrencyFocus     string            `json:"currency_focus"`
	TaxOptimizations  []TaxOptimization `json:"tax_optimizations"`
	LocalMarketAccess []MarketAccess    `json:"local_market_access"`
}

// ResearchFinding represents a key finding from the investment research
type ResearchFinding struct {
	Title           string `json:"title"`
	AssetClass      string `json:"asset_class"`      // "equities", "bonds", "alternatives", "crypto"
	GeographicFocus string `json:"geographic_focus"` // "domestic", "developed", "emerging", "global"
	InvestmentTheme string `json:"investment_theme"` // "ai", "clean_energy", "demographics", etc.

	// Investment details
	SpecificInstruments []InvestmentInstrument `json:"specific_instruments"`
	ExpectedReturn      ExpectedReturn         `json:"expected_return"`
	RiskProfile         RiskProfile            `json:"risk_profile"`
	TimeHorizon         string                 `json:"time_horizon"`

	// Research context
	MarketDrivers       []string               `json:"market_drivers"`
	CompetitivePosition string                 `json:"competitive_position"`
	ValuationMetrics    map[string]interface{} `json:"valuation_metrics"`

	// Regional relevance
	RegionalRelevance string   `json:"regional_relevance"` // Why relevant for this region
	LocalAvailability bool     `json:"local_availability"`
	TaxImplications   []string `json:"tax_implications"`
}

type InvestmentInstrument struct {
	Type     string `json:"type"` // "stock", "etf", "bond", "fund", "alternative"
	Ticker   string `json:"ticker,omitempty"`
	Name     string `json:"name"`
	Exchange string `json:"exchange,omitempty"`
	ISIN     string `json:"isin,omitempty"`
	Currency string `json:"currency"`

	// Metrics
	CurrentPrice float64  `json:"current_price,omitempty"`
	MarketCap    *int64   `json:"market_cap,omitempty"`
	ExpenseRatio *float64 `json:"expense_ratio,omitempty"`

	// Regional context
	PEAEligible  bool `json:"pea_eligible"`  // France
	ISAEligible  bool `json:"isa_eligible"`  // UK
	TFSAEligible bool `json:"tfsa_eligible"` // Canada
	IRAEligible  bool `json:"ira_eligible"`  // US

	AccessibilityNotes []string `json:"accessibility_notes"`
}

type InvestmentTheme struct {
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	MarketSize       *int64   `json:"market_size_usd,omitempty"`
	GrowthProjection string   `json:"growth_projection"` // "high", "medium", "low"
	TimeHorizon      string   `json:"time_horizon"`
	KeyDrivers       []string `json:"key_drivers"`

	// Regional adaptation
	RegionalExposure  map[string]string `json:"regional_exposure"` // region descriptions (primary, secondary, etc.)
	LocalChampions    []string          `json:"local_champions"`   // Regional leaders
	RegulatorySupport bool              `json:"regulatory_support"`

	// Implementation
	AccessMethods         []string `json:"access_methods"`         // "direct_stocks", "sector_etfs", "thematic_funds"
	RecommendedAllocation string   `json:"recommended_allocation"` // "2-5%", "5-10%", etc.
}

type ActionableInsight struct {
	Priority         string               `json:"priority"` // "high", "medium", "low"
	Action           string               `json:"action"`   // "buy", "sell", "hold", "research_further"
	Instrument       InvestmentInstrument `json:"instrument"`
	TargetAllocation string               `json:"target_allocation"` // "3-5%", "immediate", etc.
	Rationale        string               `json:"rationale"`
	Timeline         string               `json:"timeline"` // "immediate", "next_quarter", "within_year"

	// Implementation details
	EntryStrategy string   `json:"entry_strategy"` // "dollar_cost_average", "lump_sum", "wait_for_dip"
	StopLoss      *float64 `json:"stop_loss,omitempty"`
	ProfitTarget  *float64 `json:"profit_target,omitempty"`

	// Risk management
	PositionSize     string   `json:"position_size"` // "small", "medium", "large"
	MonitoringPoints []string `json:"monitoring_points"`
	ExitCriteria     []string `json:"exit_criteria"`
}

type SearchSource struct {
	URL            string     `json:"url"`
	Title          string     `json:"title"`
	SearchQuery    string     `json:"search_query"`
	RelevanceScore float64    `json:"relevance_score"`
	PublishedDate  *time.Time `json:"published_date,omitempty"`
	Source         string     `json:"source"` // "financial_times", "reuters", etc.
}

// Missing supporting model types

type MarketAnalysis struct {
	OverallSentiment    string         `json:"overall_sentiment"`  // "bullish", "bearish", "neutral"
	SectorPerformance   map[string]any `json:"sector_performance"` // sector -> YTD performance %
	ValuationLevels     string         `json:"valuation_levels"`   // "expensive", "fair", "cheap"
	TechnicalIndicators map[string]any `json:"technical_indicators"`
	EconomicBackdrop    string         `json:"economic_backdrop"`
	MarketVolatility    float64        `json:"market_volatility"` // VIX or equivalent
	CurrencyImpact      string         `json:"currency_impact"`
	LiquidityConditions string         `json:"liquidity_conditions"`
}

type RiskFactor struct {
	Type        string `json:"type"`        // "market", "credit", "liquidity", "operational", "regulatory"
	Severity    string `json:"severity"`    // "low", "medium", "high"
	Probability string `json:"probability"` // "unlikely", "possible", "likely"
	Impact      string `json:"impact"`      // Description of potential impact
	Mitigation  string `json:"mitigation"`  // How to mitigate this risk
	Timeline    string `json:"timeline"`    // When this risk might manifest
}

type TaxOptimization struct {
	Strategy       string   `json:"strategy"`                 // "tax_loss_harvesting", "asset_location", etc.
	AccountType    string   `json:"account_type"`             // "pea", "isa", "tfsa", "ira", "taxable"
	BenefitAmount  *float64 `json:"benefit_amount,omitempty"` // Estimated tax savings
	Implementation string   `json:"implementation"`           // How to implement
	Constraints    []string `json:"constraints"`              // Limitations or requirements
}

type MarketAccess struct {
	Exchange          string   `json:"exchange"`
	AssetClasses      []string `json:"asset_classes"` // Available asset classes
	TradingHours      string   `json:"trading_hours"`
	SettlementDays    int      `json:"settlement_days"`
	AccessMethod      string   `json:"access_method"` // "direct", "via_depositary_receipts", "via_funds"
	MinimumInvestment *float64 `json:"minimum_investment,omitempty"`
	TradingCosts      string   `json:"trading_costs"`
}

type ExpectedReturn struct {
	BaseCase    float64  `json:"base_case"`           // Expected return percentage
	BullCase    *float64 `json:"bull_case,omitempty"` // Optimistic scenario
	BearCase    *float64 `json:"bear_case,omitempty"` // Pessimistic scenario
	TimeHorizon string   `json:"time_horizon"`        // Period for these returns
	Methodology string   `json:"methodology"`         // How returns were calculated
	Confidence  string   `json:"confidence"`          // "high", "medium", "low"
}

type RiskProfile struct {
	VolatilityEstimate     float64  `json:"volatility_estimate"`      // Annual volatility %
	MaxDrawdown            *float64 `json:"max_drawdown,omitempty"`   // Historical max decline
	BetaToMarket           *float64 `json:"beta_to_market,omitempty"` // Market beta
	CorrelationToPortfolio *float64 `json:"correlation_to_portfolio,omitempty"`
	LiquidityRisk          string   `json:"liquidity_risk"` // "low", "medium", "high"
	ConcentrationRisk      string   `json:"concentration_risk"`
	CurrencyRisk           string   `json:"currency_risk"`
}

type AnalysisMetadata struct {
	GeneratedAt     time.Time `json:"generated_at"`
	ResearchDepth   string    `json:"research_depth"`   // "basic", "standard", "comprehensive"
	RegionalContext string    `json:"regional_context"` // Country/language context
}
