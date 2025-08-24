package bag

// Key is a symbolic identifier used inside the Bag.
type Key string

// String returns the string representation of the Key.
func (k Key) String() string {
	return string(k)
}

const (
	// === APPLICATION STATE & CONFIGURATION ===
	KVerboseMode       Key = "verbose_mode"        // Whether verbose logging is enabled
	KBatchMode         Key = "batch_mode"          // Whether current run is in batch mode
	KRegionalConfig    Key = "localization_config" // Localization configuration
	KAnalysisConfig    Key = "analysis_config"     // Analysis configuration
	KJurisdiction      Key = "jurisdiction"        // Jurisdiction/regulatory context
	KJurisdictionNotes Key = "jurisdiction_notes"  // Jurisdiction-specific notes

	// === LOCALIZATION & REGION ===
	KLanguage               Key = "language"                // Current language setting
	KCountry                Key = "country"                 // Current country setting
	KCurrency               Key = "currency"                // Current currency setting
	KStrings                Key = "strings"                 // Localization strings
	KMarketContext          Key = "market_context"          // Market context data
	KTaxContext             Key = "tax_context"             // Tax context data
	KRegulatoryFocus        Key = "regulatory_focus"        // Regulatory focus field
	KInvestmentCulture      Key = "investment_culture"      // Investment culture field
	KPreferredThemes        Key = "preferred_themes"        // Preferred themes field
	KPrimaryExchanges       Key = "primary_exchanges"       // Primary exchanges field
	KMajorIndices           Key = "major_indices"           // Major indices field
	KPrimaryAccounts        Key = "primary_accounts"        // Primary accounts field
	KOptimizationStrategies Key = "optimization_strategies" // Optimization strategies field

	// === CORE DATA - PORTFOLIO & PROFILE ===
	KPortfolio         Key = "portfolio"          // Raw portfolio data
	KPortfolioFiltered Key = "portfolio_filtered" // Filtered/processed portfolio
	KProfile           Key = "profile"            // Investment profile

	// === PORTFOLIO RISK TOLERANCE ===
	KRiskToleranceConservative Key = "conservative"
	KRiskToleranceModerate     Key = "moderate"
	KRiskToleranceAggressive   Key = "aggressive"

	// === PORTFOLIO METADATA & TRACKING ===
	KPortfolioValidationTime   Key = "portfolio.validation_time"   // When portfolio was validated
	KPortfolioLastFetched      Key = "portfolio.last_fetched"      // Last fetch timestamp
	KPortfolioFetchSource      Key = "portfolio.fetch_source"      // Source of portfolio data
	KPortfolioValidationRecord Key = "portfolio.validation_record" // Validation record

	// === PORTFOLIO ANALYSIS DATA ===
	KHoldings                 Key = "holdings"                   // Portfolio holdings
	KCurrentAlloc             Key = "current_allocation"         // Current allocation breakdown
	KTopHoldings              Key = "top_holdings"               // Top portfolio holdings
	KPortfolioRiskMetrics     Key = "portfolio.risk_metrics"     // Risk analysis metrics
	KPortfolioAllocationData  Key = "portfolio.allocation_data"  // Allocation analysis
	KPortfolioPerformanceData Key = "portfolio.performance_data" // Performance metrics
	KPortfolioComplianceData  Key = "portfolio.compliance_data"  // Compliance analysis
	KPortfolioNormalizedForAI Key = "portfolio.normalized_ai"    // AI-normalized portfolio data

	// === MARKET & ECONOMIC DATA ===
	KMacro                Key = "macro"                 // Macroeconomic data
	KMacroDataNormalized  Key = "macro.normalized"      // Normalized macro data
	KMarketDataNormalized Key = "market.normalized"     // Normalized market data
	KMarketDataFreshness  Key = "market_data_freshness" // Age/quality of market data
	KFundamentals         Key = "fundamentals"          // Fundamental analysis data
	KStockAnalysis        Key = "stock_analysis"        // Individual stock analysis

	// === NEWS & INFORMATION ===
	KNewsRaw      Key = "news_raw"      // Raw news data
	KNewsAnalyzed Key = "news_analyzed" // Analyzed news
	KNewsScored   Key = "news_scored"   // Scored/rated news

	// === ANALYSIS RESULTS ===
	KThemes                   Key = "themes"                     // Investment themes
	KDrift                    Key = "drift"                      // Portfolio drift analysis
	KRisk                     Key = "risk"                       // Risk analysis
	KRiskMetrics              Key = "risk_metrics"               // Risk metrics
	KInsights                 Key = "insights"                   // Generated insights
	KDiagnostics              Key = "diagnostics"                // System diagnostics
	KAnalysisResults          Key = "analysis_results"           // AI analysis results
	KRiskAnalysisResult       Key = "risk_analysis_result"       // Risk analysis output
	KInvestmentResearchResult Key = "investment_research_result" // Investment research output

	// === EXECUTION & REPORTING ===
	KExecutionReport Key = "execution_report" // Execution report
	KPack            Key = "context_pack"     // Context package for AI

	// === PERFORMANCE & MONITORING ===
	KToolComputations   Key = "tool_computations"    // Tool computation tracking
	KToolMetrics        Key = "tool_metrics"         // Tool performance metrics
	KCacheStats         Key = "cache_stats"          // Cache statistics
	KCacheHealthStatus  Key = "cache_health_status"  // Cache health status
	KAPICallMetrics     Key = "api_call_metrics"     // External API call metrics
	KLastAPICallStatus  Key = "last_api_call_status" // Last API call status per tool
	KApplicationHealth  Key = "application_health"   // Overall system health status
	KPerformanceMetrics Key = "performance_metrics"  // Application performance metrics
	KExternalDataHealth Key = "external_data_health" // External data provider health

	ResponseFormatJSON Key = "json_schema" // JSON schema response format
)

// === EXTERNAL TOOLS & DATA SOURCES ===
const (
	// Web & Search
	WebSearch Key = "web_search_preview" // Web search functionality

	// News & Information
	NewsAPI Key = "news_api" // NewsAPI service

	// Economic Data
	Fred Key = "fred" // Federal Reserve Economic Data

	// Financial Market Data
	FMP                 Key = "fmp"                   // Financial Modeling Prep
	FMPAnalystEstimates Key = "fmp_analyst_estimates" // FMP analyst estimates

	// Yahoo Finance
	YFinance           Key = "yfinance"             // Yahoo Finance base
	YFinanceStockData  Key = "yfinance_stock_data"  // Stock data from Yahoo Finance
	YFinanceStockInfo  Key = "yfinance_stock_info"  // Stock info from Yahoo Finance
	YFinanceDividends  Key = "yfinance_dividends"   // Dividend data
	YFinanceFinancials Key = "yfinance_financials"  // Financial statements
	YFinanceMarketData Key = "yfinance_market_data" // Market data

	// SEC EDGAR
	SECFilings        Key = "sec_filings"         // SEC filing data
	SECCompanyFacts   Key = "sec_company_facts"   // SEC company facts
	SECInsiderTrading Key = "sec_insider_trading" // SEC insider trading data
	SECOwnership      Key = "sec_ownership"       // SEC ownership data
)
