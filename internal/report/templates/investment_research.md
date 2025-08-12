## 📈 Investment Research Analysis

### Executive Summary

**Market Outlook:** {{.ExecutiveSummary.MarketOutlook | toUpper}} | **Time Horizon:** {{.ExecutiveSummary.TimeHorizon | title}}

#### Key Takeaways

{{range .ExecutiveSummary.KeyTakeaways}}

- {{.}}
  {{end}}

#### Recommended Actions

{{range .ExecutiveSummary.RecommendedActions}}

- 📋 {{.}}
  {{end}}

---

### 🌍 Regional Context

**Location:** {{.RegionalContext.Country}} | **Currency Focus:** {{.RegionalContext.CurrencyFocus}} | **Language:** {{.RegionalContext.Language}}

#### Tax Optimization Strategies

{{if .RegionalContext.TaxOptimizations}}
{{range .RegionalContext.TaxOptimizations}}

- **{{.AccountType}}**: {{.Strategy}}
  - Implementation: {{.Implementation}}
    {{if .Constraints}}- Constraints: {{join .Constraints ", "}}{{end}}
    {{end}}
    {{else}}
    _No specific tax optimization strategies identified_
    {{end}}

#### Local Market Access

{{if .RegionalContext.LocalMarketAccess}}
{{range .RegionalContext.LocalMarketAccess}}

- **{{.Exchange}}**
  - Asset Classes: {{join .AssetClasses ", "}}
  - Trading Hours: {{.TradingHours}}
  - Settlement: T+{{.SettlementDays}}
  - Trading Costs: {{.TradingCosts}}
    {{end}}
    {{else}}
    _No specific market access information provided_
    {{end}}

---

### 🔍 Research Findings

{{range .ResearchFindings}}

#### {{.Title}}

**Asset Class:** {{.AssetClass}} | **Geographic Focus:** {{.GeographicFocus}} | **Theme:** {{.InvestmentTheme}}

##### Investment Details

- **Expected Return:** {{formatPercent .ExpectedReturn.BaseCase}} ({{.ExpectedReturn.TimeHorizon}})
- **Confidence:** {{.ExpectedReturn.Confidence}} | **Methodology:** {{.ExpectedReturn.Methodology}}
- **Volatility Estimate:** {{formatPercent .RiskProfile.VolatilityEstimate}}
- **Liquidity Risk:** {{.RiskProfile.LiquidityRisk}} | **Currency Risk:** {{.RiskProfile.CurrencyRisk}}

##### Market Drivers

{{range .MarketDrivers}}

- {{.}}
  {{end}}

##### Specific Instruments

{{range .SpecificInstruments}}

- **{{.Name}}** ({{.Type}})
  - Currency: {{.Currency}}
    {{if .Ticker}}- Ticker: {{.Ticker}}{{end}}
    {{if .Exchange}}- Exchange: {{.Exchange}}{{end}}
  - PEA Eligible: {{if .PEAEligible}}✅{{else}}❌{{end}}
    {{if .AccessibilityNotes}}- Notes: {{join .AccessibilityNotes ", "}}{{end}}
    {{end}}

##### Regional Relevance

{{.RegionalRelevance}}

{{if .TaxImplications}}
**Tax Implications:**
{{range .TaxImplications}}

- {{.}}
  {{end}}
  {{end}}

---

{{end}}

### 📊 Market Analysis

**Overall Sentiment:** {{.MarketAnalysis.OverallSentiment | title}} | **Volatility:** {{formatPercent .MarketAnalysis.MarketVolatility}}

**Valuation Levels:** {{.MarketAnalysis.ValuationLevels}}

**Economic Backdrop:** {{.MarketAnalysis.EconomicBackdrop}}

#### Sector Performance

{{range $sector, $performance := .MarketAnalysis.SectorPerformance}}

- **{{$sector | title}}:** {{ $performance }}
  {{end}}

**Currency Impact:** {{.MarketAnalysis.CurrencyImpact}}

**Liquidity Conditions:** {{.MarketAnalysis.LiquidityConditions}}

---

### 🎯 Investment Themes

{{range .InvestmentThemes}}

#### {{.Name}}

{{.Description}}

- **Growth Projection:** {{.GrowthProjection | title}}
- **Time Horizon:** {{.TimeHorizon | title}}
- **Recommended Allocation:** {{.RecommendedAllocation}}
- **Regulatory Support:** {{if .RegulatorySupport}}✅ Yes{{else}}❌ No{{end}}

**Key Drivers:**
{{range .KeyDrivers}}

- {{.}}
  {{end}}

**Regional Exposure:**
{{range $region, $percentage := .RegionalExposure}}

- {{$region}}: {{formatPercent $percentage}}
  {{end}}

{{if .LocalChampions}}
**Local Champions:**
{{range .LocalChampions}}

- {{.}}
  {{end}}
  {{end}}

**Access Methods:** {{join .AccessMethods ", "}}

---

{{end}}

### ⚠️ Risk Considerations

{{range .RiskConsiderations}}

#### {{.Type | title}} Risk

- **Severity:** {{.Severity | title}} | **Probability:** {{.Probability | title}}
- **Timeline:** {{.Timeline}}

**Impact:** {{.Impact}}

**Mitigation:** {{.Mitigation}}

---

{{end}}

### 🚀 Actionable Insights

{{range .ActionableInsights}}

#### {{.Priority | toUpper}} Priority: {{.Action | title}}

**Instrument:** {{.Instrument.Name}} ({{.Instrument.Type}})

- Currency: {{.Instrument.Currency}}
- PEA Eligible: {{if .Instrument.PEAEligible}}✅{{else}}❌{{end}}

**Target Allocation:** {{.TargetAllocation}} | **Position Size:** {{.PositionSize}}

**Rationale:** {{.Rationale}}

**Timeline:** {{.Timeline}} | **Entry Strategy:** {{.EntryStrategy}}

{{if .MonitoringPoints}}
**Monitoring Points:**
{{range .MonitoringPoints}}

- 👁️ {{.}}
  {{end}}
  {{end}}

{{if .ExitCriteria}}
**Exit Criteria:**
{{range .ExitCriteria}}

- 🚪 {{.}}
  {{end}}
  {{end}}

---

{{end}}

### 📚 Research Sources

{{range .Sources}}

- [{{.Title}}]({{.URL}}) - {{.Source}} (Relevance: {{formatPercent .RelevanceScore}})
  - Query: "{{.SearchQuery}}"
    {{end}}

---

**Analysis Generated:** {{.Metadata.GeneratedAt.Format "January 2, 2006 15:04 MST"}}

**Research Depth:** {{.Metadata.ResearchDepth | title}} | **Regional Context:** {{.Metadata.RegionalContext}}
