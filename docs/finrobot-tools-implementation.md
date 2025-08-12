# FinRobot Tools Implementation Guide

## Overview

This document provides detailed implementation guidelines for all tools required to match FinRobot's capabilities in Mosychlos. Based on analysis of the FinRobot codebase, we identify key tool categories and provide Go implementation strategies.

## Tool Categories Analysis

### 1. Financial Analysis Tools

#### **ReportAnalysisUtils** - Comprehensive Financial Statement Analysis

**Priority**: 🔥 Critical (core financial analysis)

```go
// internal/tools/financial_analysis/report_analysis_tool.go
type ReportAnalysisTool struct {
    secUtils    *SECUtils
    yfinance    *YFinanceUtils
    fmpUtils    *FMPUtils
    sharedBag   *bag.SharedBag
    cache       cache.Cache
}

// Income Statement Analysis
func (t *ReportAnalysisTool) AnalyzeIncomeStatement(ctx context.Context, input map[string]any) (*models.ToolResult, error) {
    symbol := input["symbol"].(string)
    fiscalYear := input["fiscal_year"].(string)

    // Get 10-K Section 7 (MD&A) from SEC
    sectionText, err := t.secUtils.Get10KSection(symbol, fiscalYear, 7)
    if err != nil {
        return nil, fmt.Errorf("failed to get SEC section: %w", err)
    }

    // Get income statement data from YFinance
    incomeStmt, err := t.yfinance.GetIncomeStatement(symbol)
    if err != nil {
        return nil, fmt.Errorf("failed to get income statement: %w", err)
    }

    // Financial Chain-of-Thought Analysis Prompt
    prompt := t.buildIncomeAnalysisPrompt(sectionText, incomeStmt)

    return &models.ToolResult{
        Status: "completed",
        Data: map[string]any{
            "analysis_type": "income_statement",
            "symbol": symbol,
            "fiscal_year": fiscalYear,
            "sec_section": sectionText,
            "financial_data": incomeStmt,
            "analysis_prompt": prompt,
        },
    }, nil
}

func (t *ReportAnalysisTool) buildIncomeAnalysisPrompt(sectionText string, incomeStmt *IncomeStatement) string {
    return fmt.Sprintf(`
As a Senior Financial Analyst, conduct a comprehensive income statement analysis:

FINANCIAL DATA:
%s

SEC 10-K MANAGEMENT DISCUSSION:
%s

ANALYSIS FRAMEWORK:
1. Revenue Analysis:
   - Examine total revenue trends over 3-5 years
   - Identify revenue drivers and segment performance
   - Analyze revenue quality and sustainability
   - Compare to industry growth rates and peers

2. Profitability Analysis:
   - Calculate and interpret gross margin trends
   - Assess operating margin expansion/contraction
   - Evaluate net profit margin and bottom-line growth
   - Identify margin pressures and improvement opportunities

3. Expense Analysis:
   - Review operating expense trends and efficiency
   - Analyze SG&A expenses relative to revenue
   - Assess R&D investment levels and strategy
   - Evaluate one-time charges and their impact

4. Earnings Quality:
   - Compare cash flow from operations to net income
   - Assess working capital changes and their sustainability
   - Evaluate revenue recognition and accounting policies
   - Identify any red flags or accounting concerns

5. Forward-Looking Assessment:
   - Synthesize management guidance and outlook
   - Assess competitive positioning and market dynamics
   - Evaluate strategic initiatives and their financial impact
   - Provide investment thesis and key risks

Provide a comprehensive analysis with specific metrics, percentages, and year-over-year comparisons.
Support all conclusions with data from the financial statements and management discussion.
`, formatIncomeStatement(incomeStmt), sectionText)
}

// Balance Sheet Analysis
func (t *ReportAnalysisTool) AnalyzeBalanceSheet(ctx context.Context, input map[string]any) (*models.ToolResult, error) {
    // Similar structure for balance sheet analysis
    // Focus on liquidity, solvency, and capital structure
}

// Cash Flow Analysis
func (t *ReportAnalysisTool) AnalyzeCashFlow(ctx context.Context, input map[string]any) (*models.ToolResult, error) {
    // Focus on operating cash flow quality
    // Investment and financing activities analysis
    // Free cash flow generation capability
}

// Segment Analysis
func (t *ReportAnalysisTool) AnalyzeSegments(ctx context.Context, input map[string]any) (*models.ToolResult, error) {
    // Business segment performance analysis
    // Geographic segment analysis if applicable
    // Segment profitability and growth trends
}
```

#### **Quantitative Analysis Tools** - BackTrader Integration

**Priority**: 🔥 High (sophisticated backtesting capabilities)

```go
// internal/tools/quantitative/backtest_tool.go
type BacktestTool struct {
    dataProvider DataProvider
    strategies   map[string]Strategy
    sharedBag    *bag.SharedBag
}

func (t *BacktestTool) ExecuteBacktest(ctx context.Context, input map[string]any) (*models.ToolResult, error) {
    symbol := input["symbol"].(string)
    strategy := input["strategy"].(string)
    startDate := input["start_date"].(string)
    endDate := input["end_date"].(string)
    initialCash := input["initial_cash"].(float64)

    // Get historical data
    stockData, err := t.dataProvider.GetStockData(symbol, startDate, endDate)
    if err != nil {
        return nil, fmt.Errorf("failed to get stock data: %w", err)
    }

    // Initialize backtesting engine
    engine := NewBacktestEngine(initialCash)
    strategyInstance := t.strategies[strategy]

    // Run backtest
    results, err := engine.RunStrategy(strategyInstance, stockData)
    if err != nil {
        return nil, fmt.Errorf("backtest failed: %w", err)
    }

    // Generate performance metrics
    metrics := t.calculatePerformanceMetrics(results)

    // Generate charts
    chartPath, err := t.generatePerformanceChart(results, symbol, strategy)
    if err != nil {
        logger.Warn("Failed to generate chart", "error", err)
    }

    return &models.ToolResult{
        Status: "completed",
        Data: map[string]any{
            "symbol": symbol,
            "strategy": strategy,
            "period": fmt.Sprintf("%s to %s", startDate, endDate),
            "initial_cash": initialCash,
            "final_value": results.FinalValue,
            "total_return": metrics.TotalReturn,
            "annualized_return": metrics.AnnualizedReturn,
            "sharpe_ratio": metrics.SharpeRatio,
            "max_drawdown": metrics.MaxDrawdown,
            "win_rate": metrics.WinRate,
            "profit_factor": metrics.ProfitFactor,
            "total_trades": metrics.TotalTrades,
            "chart_path": chartPath,
            "trade_log": results.TradeLog,
        },
    }, nil
}

// Pre-built strategies
type SMACrossoverStrategy struct {
    FastPeriod int
    SlowPeriod int
}

type RSIMeanReversionStrategy struct {
    RSIPeriod   int
    OverboughtLevel float64
    OversoldLevel   float64
}

type BollingerBandStrategy struct {
    Period         int
    StandardDeviations float64
}

func (t *BacktestTool) calculatePerformanceMetrics(results *BacktestResults) *PerformanceMetrics {
    // Calculate comprehensive performance metrics
    // Sharpe ratio, Sortino ratio, Calmar ratio
    // Maximum drawdown, volatility
    // Win/loss ratio, profit factor
}
```

#### **Monte Carlo Simulation Tool**

**Priority**: 🟡 Medium (advanced risk analysis)

```go
// internal/tools/quantitative/monte_carlo_tool.go
type MonteCarloTool struct {
    randomGenerator *rand.Rand
    sharedBag      *bag.SharedBag
}

func (t *MonteCarloTool) RunPortfolioSimulation(ctx context.Context, input map[string]any) (*models.ToolResult, error) {
    portfolio := input["portfolio"].(*Portfolio)
    simulations := input["simulations"].(int) // e.g., 10,000
    timeHorizon := input["time_horizon"].(int) // days
    confidenceLevel := input["confidence_level"].(float64) // 0.95 for 95%

    results := make([]float64, simulations)

    for i := 0; i < simulations; i++ {
        portfolioValue := t.simulatePortfolioPath(portfolio, timeHorizon)
        results[i] = portfolioValue
    }

    // Calculate risk metrics
    valueAtRisk := t.calculateVaR(results, confidenceLevel)
    expectedShortfall := t.calculateES(results, confidenceLevel)

    // Generate distribution chart
    chartPath, err := t.generateDistributionChart(results, portfolio.Name)

    return &models.ToolResult{
        Status: "completed",
        Data: map[string]any{
            "simulations": simulations,
            "time_horizon": timeHorizon,
            "confidence_level": confidenceLevel,
            "value_at_risk": valueAtRisk,
            "expected_shortfall": expectedShortfall,
            "mean_return": calculateMean(results),
            "standard_deviation": calculateStdDev(results),
            "percentiles": calculatePercentiles(results),
            "chart_path": chartPath,
        },
    }, nil
}

func (t *MonteCarloTool) simulatePortfolioPath(portfolio *Portfolio, days int) float64 {
    // Implement geometric Brownian motion simulation
    // Account for correlations between assets
    // Apply portfolio weights
    // Return final portfolio value
}
```

### 2. Chart Generation Tools

#### **MplFinance-style Chart Tool**

**Priority**: 🔥 High (professional visualizations)

```go
// internal/tools/charting/financial_chart_tool.go
type FinancialChartTool struct {
    plotly     PlotlyRenderer // or use gonum/plot for native Go charts
    sharedBag  *bag.SharedBag
}

func (t *FinancialChartTool) GenerateStockChart(ctx context.Context, input map[string]any) (*models.ToolResult, error) {
    symbol := input["symbol"].(string)
    startDate := input["start_date"].(string)
    endDate := input["end_date"].(string)
    chartType := input["chart_type"].(string) // "candlestick", "ohlc", "line"
    indicators := input["indicators"].([]string) // ["SMA_20", "RSI", "MACD"]

    // Get stock data
    stockData, err := t.getStockData(symbol, startDate, endDate)
    if err != nil {
        return nil, fmt.Errorf("failed to get stock data: %w", err)
    }

    // Create chart based on type
    var chart Chart
    switch chartType {
    case "candlestick":
        chart = t.createCandlestickChart(stockData)
    case "ohlc":
        chart = t.createOHLCChart(stockData)
    default:
        chart = t.createLineChart(stockData)
    }

    // Add technical indicators
    for _, indicator := range indicators {
        t.addTechnicalIndicator(chart, stockData, indicator)
    }

    // Style the chart professionally
    t.applyProfessionalStyling(chart, symbol)

    // Save chart
    chartPath := fmt.Sprintf("charts/%s_%s_chart.png", symbol, chartType)
    err = chart.SaveAs(chartPath)
    if err != nil {
        return nil, fmt.Errorf("failed to save chart: %w", err)
    }

    return &models.ToolResult{
        Status: "completed",
        Data: map[string]any{
            "symbol": symbol,
            "chart_type": chartType,
            "chart_path": chartPath,
            "period": fmt.Sprintf("%s to %s", startDate, endDate),
            "indicators": indicators,
        },
    }, nil
}

func (t *FinancialChartTool) addTechnicalIndicator(chart Chart, data *StockData, indicator string) {
    switch indicator {
    case "SMA_20":
        sma := calculateSMA(data.Close, 20)
        chart.AddLine("SMA(20)", sma, "blue")
    case "SMA_50":
        sma := calculateSMA(data.Close, 50)
        chart.AddLine("SMA(50)", sma, "red")
    case "RSI":
        rsi := calculateRSI(data.Close, 14)
        chart.AddSubplot("RSI", rsi, 0, 100)
    case "MACD":
        macd, signal := calculateMACD(data.Close)
        chart.AddSubplot("MACD", macd, -1, 1)
        chart.AddLine("Signal", signal, "red")
    case "BOLLINGER":
        upper, middle, lower := calculateBollingerBands(data.Close, 20, 2.0)
        chart.AddLine("BB Upper", upper, "gray", true) // dashed
        chart.AddLine("BB Middle", middle, "gray")
        chart.AddLine("BB Lower", lower, "gray", true)
    }
}

// Portfolio Performance Chart
func (t *FinancialChartTool) GeneratePortfolioChart(ctx context.Context, input map[string]any) (*models.ToolResult, error) {
    // Generate portfolio performance vs benchmark
    // Show asset allocation pie charts
    // Risk-return scatter plots
    // Drawdown charts
}

// Risk Analysis Charts
func (t *FinancialChartTool) GenerateRiskCharts(ctx context.Context, input map[string]any) (*models.ToolResult, error) {
    // VaR distribution charts
    // Correlation heatmaps
    // Beta analysis charts
    // Sector allocation charts
}
```

#### **Report Chart Utilities**

**Priority**: 🔥 High (professional report generation)

```go
// internal/tools/charting/report_chart_tool.go
type ReportChartTool struct {
    chartGenerator *FinancialChartTool
    sharedBag     *bag.SharedBag
}

func (t *ReportChartTool) GenerateSharePerformanceChart(ctx context.Context, input map[string]any) (*models.ToolResult, error) {
    symbol := input["symbol"].(string)
    filingDate := input["filing_date"].(string)

    // Create 5-year stock performance chart for annual reports
    endDate := filingDate
    startDate := calculateDateYearsAgo(filingDate, 5)

    stockData, err := t.getStockData(symbol, startDate, endDate)
    if err != nil {
        return nil, fmt.Errorf("failed to get stock data: %w", err)
    }

    // Create professional line chart
    chart := t.createProfessionalLineChart(stockData, symbol)

    // Add benchmark comparison (S&P 500)
    benchmarkData, _ := t.getStockData("SPY", startDate, endDate)
    if benchmarkData != nil {
        t.addBenchmarkLine(chart, benchmarkData, "S&P 500")
    }

    // Style for report inclusion
    t.applyReportStyling(chart)

    chartPath := fmt.Sprintf("reports/charts/%s_performance.png", symbol)
    err = chart.SaveAs(chartPath)

    return &models.ToolResult{
        Status: "completed",
        Data: map[string]any{
            "chart_path": chartPath,
            "chart_type": "share_performance",
            "symbol": symbol,
            "period_years": 5,
        },
    }, nil
}

func (t *ReportChartTool) GeneratePEEPSChart(ctx context.Context, input map[string]any) (*models.ToolResult, error) {
    symbol := input["symbol"].(string)

    // Get 5-year P/E and EPS history
    peHistory, err := t.getPEHistory(symbol, 5)
    if err != nil {
        return nil, fmt.Errorf("failed to get P/E history: %w", err)
    }

    epsHistory, err := t.getEPSHistory(symbol, 5)
    if err != nil {
        return nil, fmt.Errorf("failed to get EPS history: %w", err)
    }

    // Create dual-axis chart
    chart := t.createDualAxisChart()
    chart.AddLeftYAxis("P/E Ratio", peHistory.Values, "blue")
    chart.AddRightYAxis("EPS ($)", epsHistory.Values, "green")

    // Style for report
    t.applyReportStyling(chart)
    chart.SetTitle(fmt.Sprintf("%s P/E Ratio and EPS History", symbol))

    chartPath := fmt.Sprintf("reports/charts/%s_pe_eps.png", symbol)
    err = chart.SaveAs(chartPath)

    return &models.ToolResult{
        Status: "completed",
        Data: map[string]any{
            "chart_path": chartPath,
            "chart_type": "pe_eps_performance",
            "symbol": symbol,
        },
    }, nil
}
```

### 3. Report Generation Tools

#### **ReportLab-style PDF Generator**

**Priority**: 🔥 High (institutional reports)

```go
// internal/tools/reporting/pdf_report_tool.go
type PDFReportTool struct {
    pdfGenerator PDFGenerator // using gofpdf or unidoc
    chartTool    *ReportChartTool
    sharedBag    *bag.SharedBag
}

func (t *PDFReportTool) GenerateAnnualReport(ctx context.Context, input map[string]any) (*models.ToolResult, error) {
    symbol := input["symbol"].(string)
    operatingResults := input["operating_results"].(string)
    marketPosition := input["market_position"].(string)
    businessOverview := input["business_overview"].(string)
    riskAssessment := input["risk_assessment"].(string)
    competitorsAnalysis := input["competitors_analysis"].(string)
    sharePerformanceChart := input["share_performance_image_path"].(string)
    peEpsChart := input["pe_eps_performance_image_path"].(string)
    filingDate := input["filing_date"].(string)

    // Initialize PDF document
    pdf := t.pdfGenerator.NewDocument()

    // Set up professional styles
    t.setupProfessionalStyles(pdf)

    // Get company information
    companyInfo, err := t.getCompanyInfo(symbol)
    if err != nil {
        return nil, fmt.Errorf("failed to get company info: %w", err)
    }

    // Build report structure
    t.addCoverPage(pdf, symbol, companyInfo.Name, filingDate)
    t.addExecutiveSummary(pdf, businessOverview, operatingResults)
    t.addBusinessAnalysis(pdf, marketPosition, businessOverview)
    t.addFinancialAnalysis(pdf, operatingResults, symbol)
    t.addPerformanceCharts(pdf, sharePerformanceChart, peEpsChart)
    t.addRiskAnalysis(pdf, riskAssessment)
    t.addCompetitiveAnalysis(pdf, competitorsAnalysis)
    t.addKeyFinancialData(pdf, symbol, filingDate)

    // Save PDF
    reportPath := fmt.Sprintf("reports/%s_annual_report_%s.pdf", symbol, filingDate)
    err = pdf.SaveAs(reportPath)
    if err != nil {
        return nil, fmt.Errorf("failed to save PDF: %w", err)
    }

    return &models.ToolResult{
        Status: "completed",
        Data: map[string]any{
            "report_path": reportPath,
            "report_type": "annual_report",
            "symbol": symbol,
            "filing_date": filingDate,
            "page_count": pdf.PageCount(),
        },
    }, nil
}

func (t *PDFReportTool) addFinancialAnalysis(pdf PDFDocument, operatingResults, symbol string) {
    // Add financial metrics table
    keyData, err := t.getKeyFinancialData(symbol)
    if err != nil {
        logger.Warn("Failed to get key financial data", "symbol", symbol, "error", err)
        return
    }

    // Create professional financial table
    table := pdf.NewTable([]string{"Metric", "Value"})
    for metric, value := range keyData {
        table.AddRow(metric, formatFinancialValue(value))
    }

    pdf.AddSection("Financial Highlights")
    pdf.AddTable(table)

    // Add 5-year financial metrics comparison
    historicalData, err := t.getHistoricalFinancials(symbol, 5)
    if err == nil {
        t.addHistoricalFinancialsTable(pdf, historicalData)
    }

    // Add operating results analysis
    pdf.AddSection("Operating Results Analysis")
    pdf.AddParagraph(operatingResults)
}

func (t *PDFReportTool) addPerformanceCharts(pdf PDFDocument, shareChart, peEpsChart string) {
    pdf.AddPageBreak()
    pdf.AddSection("Performance Analysis")

    // Add share performance chart
    pdf.AddSubsection("Share Price Performance")
    pdf.AddImage(shareChart, ImageOptions{
        Width:  400,
        Height: 250,
        Center: true,
    })

    // Add P/E and EPS chart
    pdf.AddSubsection("Valuation and Earnings Metrics")
    pdf.AddImage(peEpsChart, ImageOptions{
        Width:  400,
        Height: 250,
        Center: true,
    })
}
```

#### **Investment Committee Report Tool**

**Priority**: 🔥 High (multi-agent decision documentation)

```go
// internal/tools/reporting/committee_report_tool.go
type CommitteeReportTool struct {
    pdfGenerator *PDFReportTool
    sharedBag    *bag.SharedBag
}

func (t *CommitteeReportTool) GenerateCommitteeReport(ctx context.Context, input map[string]any) (*models.ToolResult, error) {
    committeeDecision := input["committee_decision"].(string)
    individualAnalyses := input["individual_analyses"].(map[string]string)
    finalRecommendation := input["final_recommendation"].(string)
    confidenceLevel := input["confidence_level"].(float64)
    keyRisks := input["key_risks"].([]string)
    expectedReturn := input["expected_return"].(float64)
    timeHorizon := input["time_horizon"].(string)

    // Create committee report
    pdf := t.pdfGenerator.NewDocument()

    // Header
    t.addCommitteeHeader(pdf, time.Now())

    // Executive Summary
    pdf.AddSection("Investment Committee Decision")
    pdf.AddParagraph(committeeDecision)

    // Individual Analyst Contributions
    pdf.AddSection("Analyst Contributions")
    for analystType, analysis := range individualAnalyses {
        pdf.AddSubsection(analystType)
        pdf.AddParagraph(analysis)
    }

    // Final Recommendation
    pdf.AddSection("Final Recommendation")
    pdf.AddParagraph(finalRecommendation)

    // Risk Assessment
    pdf.AddSection("Key Risks Identified")
    for _, risk := range keyRisks {
        pdf.AddBulletPoint(risk)
    }

    // Investment Metrics
    t.addInvestmentMetrics(pdf, expectedReturn, confidenceLevel, timeHorizon)

    reportPath := fmt.Sprintf("reports/committee_report_%s.pdf", time.Now().Format("20060102_150405"))
    err := pdf.SaveAs(reportPath)

    return &models.ToolResult{
        Status: "completed",
        Data: map[string]any{
            "report_path": reportPath,
            "report_type": "investment_committee",
            "recommendation": finalRecommendation,
            "confidence_level": confidenceLevel,
        },
    }, nil
}
```

### 4. Text Analysis Tools

#### **Text Length Checker**

**Priority**: 🟡 Low (utility function)

```go
// internal/tools/text/text_utils_tool.go
type TextUtilsTool struct {
    sharedBag *bag.SharedBag
}

func (t *TextUtilsTool) CheckTextLength(ctx context.Context, input map[string]any) (*models.ToolResult, error) {
    text := input["text"].(string)

    wordCount := len(strings.Fields(text))
    charCount := len(text)
    paragraphCount := len(strings.Split(text, "\n\n"))

    return &models.ToolResult{
        Status: "completed",
        Data: map[string]any{
            "word_count": wordCount,
            "character_count": charCount,
            "paragraph_count": paragraphCount,
            "text_length_category": categorizeTextLength(wordCount),
        },
    }, nil
}

func categorizeTextLength(wordCount int) string {
    switch {
    case wordCount < 100:
        return "short"
    case wordCount < 500:
        return "medium"
    case wordCount < 1000:
        return "long"
    default:
        return "very_long"
    }
}
```

### 5. Integration Tools

#### **Display Tool for Jupyter-style Output**

**Priority**: 🟡 Medium (development/debugging aid)

```go
// internal/tools/integration/display_tool.go
type DisplayTool struct {
    outputPath string
    sharedBag  *bag.SharedBag
}

func (t *DisplayTool) DisplayImage(ctx context.Context, input map[string]any) (*models.ToolResult, error) {
    imagePath := input["image_path"].(string)

    // Verify image exists
    if _, err := os.Stat(imagePath); os.IsNotExist(err) {
        return nil, fmt.Errorf("image file does not exist: %s", imagePath)
    }

    // For CLI usage, we might copy to a display directory
    // or generate a markdown link for report inclusion
    displayPath := filepath.Join(t.outputPath, "display", filepath.Base(imagePath))
    err := copyFile(imagePath, displayPath)
    if err != nil {
        return nil, fmt.Errorf("failed to copy image for display: %w", err)
    }

    return &models.ToolResult{
        Status: "completed",
        Data: map[string]any{
            "original_path": imagePath,
            "display_path": displayPath,
            "markdown_link": fmt.Sprintf("![Chart](%s)", displayPath),
        },
    }, nil
}
```

## Tool Integration Strategy

### 1. Tool Registration System Enhancement

```go
// internal/tools/registry.go
type ToolRegistry struct {
    tools      map[string]Tool
    categories map[string][]string
    sharedBag  *bag.SharedBag
}

func (r *ToolRegistry) RegisterFinancialTools() {
    // Financial Analysis Tools
    r.RegisterTool("analyze_income_statement", &ReportAnalysisTool{})
    r.RegisterTool("analyze_balance_sheet", &ReportAnalysisTool{})
    r.RegisterTool("analyze_cash_flow", &ReportAnalysisTool{})
    r.RegisterTool("analyze_segments", &ReportAnalysisTool{})

    // Quantitative Tools
    r.RegisterTool("backtest_strategy", &BacktestTool{})
    r.RegisterTool("monte_carlo_simulation", &MonteCarloTool{})

    // Chart Generation
    r.RegisterTool("generate_stock_chart", &FinancialChartTool{})
    r.RegisterTool("generate_portfolio_chart", &FinancialChartTool{})
    r.RegisterTool("generate_share_performance_chart", &ReportChartTool{})
    r.RegisterTool("generate_pe_eps_chart", &ReportChartTool{})

    // Report Generation
    r.RegisterTool("generate_annual_report", &PDFReportTool{})
    r.RegisterTool("generate_committee_report", &CommitteeReportTool{})

    // Utility Tools
    r.RegisterTool("check_text_length", &TextUtilsTool{})
    r.RegisterTool("display_image", &DisplayTool{})
}

func (r *ToolRegistry) GetToolsByCategory(category string) []Tool {
    toolNames := r.categories[category]
    tools := make([]Tool, len(toolNames))
    for i, name := range toolNames {
        tools[i] = r.tools[name]
    }
    return tools
}
```

### 2. Tool Toolkit System (FinRobot-style)

```go
// internal/engine/toolkit.go
type ToolKit struct {
    Name        string   `yaml:"name"`
    Description string   `yaml:"description"`
    Tools       []string `yaml:"tools"`
    Category    string   `yaml:"category"`
}

var FinancialAnalysisToolkit = ToolKit{
    Name:        "financial_analysis",
    Description: "Comprehensive financial statement analysis tools",
    Tools: []string{
        "analyze_income_statement",
        "analyze_balance_sheet",
        "analyze_cash_flow",
        "analyze_segments",
    },
    Category: "analysis",
}

var QuantitativeToolkit = ToolKit{
    Name:        "quantitative_analysis",
    Description: "Quantitative modeling and backtesting tools",
    Tools: []string{
        "backtest_strategy",
        "monte_carlo_simulation",
        "calculate_var",
        "optimize_portfolio",
    },
    Category: "quantitative",
}

var ReportingToolkit = ToolKit{
    Name:        "professional_reporting",
    Description: "Professional report and chart generation",
    Tools: []string{
        "generate_annual_report",
        "generate_committee_report",
        "generate_stock_chart",
        "generate_portfolio_chart",
    },
    Category: "reporting",
}

// Persona-specific tool assignments
var MarketAnalystToolkits = []ToolKit{
    NewsAnalysisToolkit,
    SentimentAnalysisToolkit,
    ChartingToolkit,
}

var FinancialAnalystToolkits = []ToolKit{
    FinancialAnalysisToolkit,
    ValuationToolkit,
    PeerComparisonToolkit,
}

var QuantitativeAnalystToolkits = []ToolKit{
    QuantitativeToolkit,
    BacktestingToolkit,
    RiskAnalysisToolkit,
}
```

### 3. Caching and Performance Optimization

```go
// internal/tools/cache/tool_cache.go
type ToolCache struct {
    cache     cache.Cache
    ttlMap    map[string]time.Duration
}

func (c *ToolCache) CacheToolResult(toolName string, input map[string]any, result *models.ToolResult) error {
    key := c.generateCacheKey(toolName, input)
    ttl := c.ttlMap[toolName]
    if ttl == 0 {
        ttl = 1 * time.Hour // default TTL
    }

    return c.cache.Set(key, result, ttl)
}

func (c *ToolCache) GetCachedResult(toolName string, input map[string]any) (*models.ToolResult, bool) {
    key := c.generateCacheKey(toolName, input)
    result, exists := c.cache.Get(key)
    if !exists {
        return nil, false
    }

    return result.(*models.ToolResult), true
}

// Cache TTL configuration
var ToolCacheTTLs = map[string]time.Duration{
    "analyze_income_statement":    24 * time.Hour, // SEC data changes infrequently
    "generate_stock_chart":        1 * time.Hour,  // Market data changes frequently
    "backtest_strategy":           7 * 24 * time.Hour, // Historical backtests rarely change
    "generate_annual_report":      30 * 24 * time.Hour, // Annual reports change rarely
    "get_company_news":            15 * time.Minute, // News updates frequently
}
```

## Implementation Priority Matrix

### **Phase 1 - Core Financial Analysis (Weeks 1-2)**

1. 🔥 **ReportAnalysisTool** - Income statement, balance sheet, cash flow analysis
2. 🔥 **FinancialChartTool** - Basic stock charts with technical indicators
3. 🔥 **PDFReportTool** - Basic professional report generation
4. 🟡 **TextUtilsTool** - Supporting text analysis functions

### **Phase 2 - Advanced Analytics (Weeks 3-4)**

1. 🔥 **BacktestTool** - Strategy backtesting with performance metrics
2. 🟡 **MonteCarloTool** - Risk simulation and VaR calculation
3. 🔥 **ReportChartTool** - Professional charts for report inclusion
4. 🔥 **CommitteeReportTool** - Multi-agent decision documentation

### **Phase 3 - Professional Features (Weeks 5-6)**

1. 🟡 **Advanced charting** - Portfolio performance, risk analysis charts
2. 🟡 **DisplayTool** - Integration with report generation pipeline
3. 🟡 **Tool caching** - Performance optimization for expensive operations
4. 🟡 **Error handling** - Robust error recovery and user feedback

### **Phase 4 - Enhancement & Optimization (Weeks 7-8)**

1. 🔶 **Advanced backtesting strategies** - Multiple strategy types
2. 🔶 **Interactive charts** - Web-based chart generation
3. 🔶 **Report customization** - Template-based report generation
4. 🔶 **Tool performance monitoring** - Metrics and optimization

## Expected Tool Usage in Personas

### **Financial Analyst Persona**

Primary Tools:

- `analyze_income_statement`
- `analyze_balance_sheet`
- `analyze_cash_flow`
- `generate_annual_report`
- `generate_pe_eps_chart`

### **Quantitative Analyst Persona**

Primary Tools:

- `backtest_strategy`
- `monte_carlo_simulation`
- `generate_stock_chart` (with technical indicators)
- `calculate_correlation`

### **Investment Committee Chairperson**

Primary Tools:

- `generate_committee_report`
- `generate_portfolio_chart`
- `display_image` (for presentation)

### **Market Analyst Persona**

Primary Tools:

- `generate_stock_chart`
- `analyze_news_sentiment` (future)
- `get_market_overview` (future)

This comprehensive tool implementation provides Mosychlos with institutional-grade analytical capabilities equivalent to FinRobot's sophisticated toolkit while leveraging Go's performance advantages and the existing architecture.
