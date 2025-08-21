# Mosychlos — Deep, Granular Roadmap (Engines • Tools • NEXT_PATTERN • Prompts)

## Index

1. Vision & Principles
2. System Diagrams (Engine Chains & Dependencies)
3. Engines (per‑engine SRP, I/O, StructuredOutput, impl notes, NEXT_PATTERN)
   3.1 PortfolioIngest
   3.2 MarketDataIngest (prices, ohlc, corp actions)
   3.3 FXNormalize (currency/base)
   3.4 TimeAlign (timeseries alignment)
   3.5 FundamentalsIngest (FMP/YF/EDGAR)
   3.6 MacroIngest (FRED)
   3.7 FactorBuild (quality, value, momentum)
   3.8 Risk (cov, VaR/ES, betas)
   3.9 Scenario & Stress (rates, inflation, drawdown)
   3.10 ComplianceTax (regional, PEA, TLH)
   3.11 OpportunityScan (screeners, ideas)
   3.12 NewsIngest & TextNormalize (WaterCrawl/Batch optional)
   3.13 NewsAnalyze (summary, entities, sentiment)
   3.14 Attribution (perf/sector contribution)
   3.15 Optimize (allocation, trades)
   3.16 Reporting (compose, narrative, IC conclusion)
4. Tools (per‑tool normalization & Go types)
   4.1 SEC EDGAR
   4.2 Yahoo Finance (your current tools)
   4.3 Financial Modeling Prep (FMP)
   4.4 FRED (macro)
5. Tool Chaining (analysis & design)
6. The ##NEXT_PATTERN## (engines & tools)
7. Prompt Management (libraries & org)
8. Prompt Set (tied to StructuredOutputs)
9. WaterCrawl + OpenAI Batch — where they fit
10. Testing Strategy (unit, integration, golden, contracts)
11. Migration to SaaS (notes you’ll be glad you had)

---

## 1) Vision & Principles

- **Single Responsibility Engines**: each engine does one thing, returns a typed `StructuredOutput`.
- **Normalize once, reuse everywhere**: currency, calendars, identifiers.
- **Deterministic chains** with **dynamic NEXT**: orchestrator defines safe defaults; engines/tools can propose next steps via ##NEXT_PATTERN##.
- **Reporting is its own phase**: treat reports as the “cherry on top” of analyze results.
- **Model-agnostic prompts** with observable changes (versioned, testable).

---

## 2) System Diagrams (Engine Chains & Dependencies)

### A) Weekly **Analyze** Phase (default)

```
PortfolioIngest
   ↓
MarketDataIngest → FXNormalize → TimeAlign
   ↓
FundamentalsIngest      MacroIngest
   ↓                        ↓
         FactorBuild  ←─────┘
               ↓
Risk  ←────── TimeAlign (uses aligned returns)
  │
  ├─→ Scenario&Stress (uses Risk + Macro)
  ├─→ ComplianceTax (uses Portfolio + Region rules)
  └─→ Attribution (if perf window available)
        ↓
NewsIngest → NewsAnalyze (optional, per-holdings)
        ↓
Reporting (Analyze Report: facts, risks, context)
```

### B) **Optimize** Phase (after Analyze or standalone)

```
[Analyze outputs available]
      ↓
Optimize (objectives + constraints + risk + factors)
      ↓
Reporting (Investment Committee Conclusion)
```

### C) Engine data fan‑out/fan‑in

```
[TextForAnalysis] ─┬─→ NewsSummary
                   ├─→ NewsEntities
                   ├─→ NewsSentiment
                   └─→ TopicClass
             (fan-in to) NewsSynthesis
```

---

## 3) Engines (SRP • I/O • StructuredOutput • Impl • NEXT)

> Tip: Put all shared types in `pkg/types` and per‑engine outputs in `pkg/engines/<engine>/types.go`.

### 3.1 PortfolioIngest (SRP: load your portfolio spec)

**Input**: none or `PortfolioSpecRequest{Source string}`
**Output (Go)**:

```go
type AssetClass string
const (
  Equity AssetClass = "equity"; Bond="bond"; ETF="etf"; Crypto="crypto"; Commodity="commodity"; Cash="cash"
)
type AssetID struct {
  Ticker string  `json:"ticker"`
  ISIN   string  `json:"isin,omitempty"`
  Mic    string  `json:"mic,omitempty"`   // exchange MIC
  Class  AssetClass `json:"class"`
  Currency string `json:"currency"`       // native
}
type Holding struct {
  ID       AssetID  `json:"id"`
  Qty      float64  `json:"qty"`
  Weight   float64  `json:"weight"`       // optional; recompute from prices
  CostBase float64  `json:"cost_base"`    // in base currency, optional
  LotDate  string   `json:"lot_date,omitempty"`
  Account  string   `json:"account,omitempty"` // taxable/PEA/retirement
}
type Portfolio struct {
  BaseCurrency string     `json:"base_currency"`
  AsOf         time.Time  `json:"as_of"`
  Holdings     []Holding  `json:"holdings"`
}
```

**Impl**: parse from file/db; validate unique IDs; infer missing `Class` when possible.
**NEXT_PATTERN**: if any `Class==Crypto` → propose `CryptoRisk` engine later; if weights missing → propose `MarketDataIngest` to compute.

---

### 3.2 MarketDataIngest (SRP: quotes/ohlc/corp actions)

**Input**: `Portfolio`
**Output**:

```go
type PriceBar struct {
  T time.Time `json:"t"`
  O,H,L,C float64 `json:"o","h","l","c"`
  V float64 `json:"v"`
  Currency string `json:"currency"`
}
type Quote struct {
  Last float64 `json:"last"`
  Bid  float64 `json:"bid"`
  Ask  float64 `json:"ask"`
  Time time.Time `json:"time"`
  Currency string `json:"currency"`
}
type CorpAction struct {
  Type string `json:"type"` // dividend, split
  ExDate string `json:"ex_date"`
  Amount float64 `json:"amount,omitempty"`
  Ratio  float64 `json:"ratio,omitempty"`
  Currency string `json:"currency,omitempty"`
}
type MarketSnapshot struct {
  Quotes map[string]Quote                `json:"quotes"` // key: AssetKey()
  OHLC   map[string][]PriceBar           `json:"ohlc"`   // uniform bar size (e.g., Daily)
  Actions map[string][]CorpAction        `json:"actions"`
  Window struct{ From, To time.Time }    `json:"window"`
}
```

**Impl**: call YF/FMP (details in Tools); normalize timezones; unify bar granularity (Daily).
**NEXT_PATTERN**: if any missing OHLC → propose backfill job; if currency!=portfolio base → propose `FXNormalize`.

---

### 3.3 FXNormalize (SRP: convert to base currency)

**Input**: `MarketSnapshot`, base currency
**Output**:

```go
type FXRate struct {
  Pair string    `json:"pair"` // e.g., EURUSD
  T    time.Time `json:"t"`
  Rate float64   `json:"rate"`
}
type FXBook struct {
  Base string               `json:"base"`
  Series map[string][]FXRate `json:"series"` // per pair
}
type MarketInBase struct {
  Quotes map[string]Quote
  OHLC   map[string][]PriceBar // all C in base
  Actions map[string][]CorpAction // amounts in base
  FX     FXBook
}
```

**Impl**: fetch FX (ECB/FRED or YF); align by date (use prior close when missing).
**NEXT_PATTERN**: if FX series gaps > threshold → propose `MacroIngest` (ECB FX) or fallback tool.

---

### 3.4 TimeAlign (SRP: align series & returns)

**Input**: `MarketInBase`
**Output**:

```go
type AlignedSeries struct {
  Dates []time.Time            `json:"dates"`
  Close map[string][]float64   `json:"close"`
  Returns map[string][]float64 `json:"returns"` // log or simple
  Weights map[string]float64   `json:"weights"` // current inferred
}
```

**Impl**: union calendar → choose business days; forward‑fill small gaps; compute returns; recompute current weights.
**NEXT_PATTERN**: if high missingness → propose `MarketDataIngest` backfill or restrict window.

---

### 3.5 FundamentalsIngest (SRP: financials)

**Input**: `Portfolio`
**Output**:

```go
type TTM struct { Revenue, EBITDA, EBIT, NetIncome, FCF float64; EPS float64 }
type MRQ struct { Assets, Liabilities, Equity, Cash, Debt float64 }
type Ratios struct { PE, PS, PB, ROE, ROIC, GrossMargin, OpMargin float64 }
type Fundamentals struct {
  Ticker string `json:"ticker"`
  Currency string `json:"currency"`
  PeriodEnd string `json:"period_end"`
  TTM TTM `json:"ttm"`
  MRQ MRQ `json:"mrq"`
  Ratios Ratios `json:"ratios"`
}
type FundamentalsBook struct {
  Items map[string]Fundamentals `json:"items"`
  AsOf  time.Time               `json:"as_of"`
}
```

**Impl**: pull from FMP/YF; currency convert; prefer standardized fields; store provenance.
**NEXT_PATTERN**: if missing EPS/TTM for >X% holdings → propose `EDGAR` filings extract.

---

### 3.6 MacroIngest (SRP: macro series, including FX, rates)

**Input**: config of series ids (FRED keys, etc.)
**Output**:

```go
type MacroPoint struct { T time.Time; Value float64 }
type MacroSeries struct {
  ID string `json:"id"`
  Name string `json:"name"`
  Freq string `json:"freq"`   // D/W/M/Q
  Units string `json:"units"` // %, index, level
  Data []MacroPoint `json:"data"`
}
type MacroBook struct {
  Series map[string]MacroSeries `json:"series"`
  AsOf time.Time `json:"as_of"`
}
```

**Impl**: FRED for US; add ECB/Eurostat later; standardize freq and units; resample if needed.
**NEXT_PATTERN**: if region=EU and ECB series not present → propose ECB fetch.

---

### 3.7 FactorBuild (SRP: compute factors)

**Input**: `AlignedSeries`, `FundamentalsBook`
**Output**:

```go
type FactorScores struct {
  Ticker string `json:"ticker"`
  Value  map[string]float64 `json:"value"`  // e.g., "PE_inv", "PB_inv"
  Quality map[string]float64 `json:"quality"` // ROE, ROIC
  Momentum map[string]float64 `json:"momentum"` // 12-1, 6M
  Volatility float64 `json:"volatility"`
}
type FactorBook struct {
  Items map[string]FactorScores `json:"items"`
  ZScores map[string]map[string]float64 `json:"z_scores"` // per factor
}
```

**Impl**: compute 12‑1 momentum, volatility; inverse valuation ratios; z‑score within peer bucket (sector/region).
**NEXT_PATTERN**: low data coverage → propose fallback categories; or skip weight in optimizer.

---

### 3.8 Risk (SRP: covariances & summary risk)

**Input**: `AlignedSeries`, portfolio weights
**Output**:

```go
type Covar struct {
  Tickers []string `json:"tickers"`
  Matrix  [][]float64 `json:"matrix"`
}
type RiskSummary struct {
  AnnualVol float64
  VaR95 float64
  ES95 float64
  BetaMkt float64
  TopRiskContrib []struct{ Ticker string; Contribution float64 }
}
type RiskBook struct {
  Cov Covar
  Summary RiskSummary
}
```

**Impl**: compute covariance (Ledoit‑Wolf shrinkage optional); VaR/ES (historical); beta to chosen index.
**NEXT_PATTERN**: if `AnnualVol>threshold` → propose `HedgingIdeas` or `Scenario&Stress`.

---

### 3.9 Scenario & Stress (SRP: what‑ifs)

**Input**: `RiskBook`, `MacroBook`, `AlignedSeries`
**Output**:

```go
type Shock struct { Name string; HorizonDays int; Factor string; ShockValue float64 }
type ScenarioImpact struct {
  Shock Shock
  PnL float64
  Drawdown float64
  Notes string
}
type ScenarioBook struct { Impacts []ScenarioImpact }
```

**Impl**: simple mapping (e.g., +100bps rates → equity −x%, bonds −y%); extend with regression to macro factors.
**NEXT_PATTERN**: if extreme impact in ‘RatesUp’ → propose bond duration hedge.

---

### 3.10 ComplianceTax (SRP: regional rules)

**Input**: `Portfolio`, region config (FR PEA, TLH flags)
**Output**:

```go
type ComplianceIssue struct { Code string; Severity string; Message string; Holding Optional[string] }
type TaxOpportunity struct { Type string; Holding string; EstimatedBenefit float64; Rationale string }
type ComplianceTaxReport struct {
  Issues []ComplianceIssue
  Opportunities []TaxOpportunity
}
```

**Impl**: PEA eligibility, position caps, wash‑sale/TLH candidates; mark account types.
**NEXT_PATTERN**: if TLH candidates exist → propose `Optimize` (harvest trades).

---

### 3.11 OpportunityScan (SRP: screen ideas)

**Input**: `FactorBook`, `MacroBook`, market universe (optional)
**Output**:

```go
type Idea struct {
  ID AssetID
  Thesis string
  Factors map[string]float64
  RiskNotes string
}
type IdeasList struct { Items []Idea }
```

**Impl**: rank by composite factor; filter by macro regime; optional LLM to write thesis from numbers.
**NEXT_PATTERN**: if top ideas improve diversification → propose to optimizer.

---

### 3.12 NewsIngest & TextNormalize (SRP: gather texts)

**Input**: `Portfolio.Holdings`, optional keywords
**Output**:

```go
type RawDoc struct { URL string; Title string; Published time.Time; Lang string; Body string }
type DocBatch struct { Docs []RawDoc }
```

**Impl**: WaterCrawl / News API / web fetch; strip HTML; dedupe; language detect.

---

### 3.13 NewsAnalyze (SRP: summarize/entities/sentiment)

**Input**: `DocBatch`
**Output**:

```go
type DocEntities struct { People, Orgs, Locations, Tickers []string; Dates []string }
type DocAnalysis struct {
  URL string
  Summary string
  Entities DocEntities
  Sentiment string // Positive/Neutral/Negative
  Topic string
}
type NewsBook struct { Items []DocAnalysis }
```

**Impl**: use Responses or Batch (see §9); structured extraction (JSON).
**NEXT_PATTERN**: if negative sentiment cluster on a holding → propose risk review.

---

### 3.14 Attribution (SRP: performance contribution)

**Input**: `AlignedSeries`, `Portfolio` with historical weights (if available)
**Output**:

```go
type Contribution struct { Ticker string; ReturnPct float64; ContributionPct float64 }
type AttributionReport struct {
  TotalReturn float64
  Contributions []Contribution
  BySector map[string]float64
}
```

**Impl**: Brinson basic; sector roll‑ups.
**NEXT_PATTERN**: if single contribution > threshold → flag concentration.

---

### 3.15 Optimize (SRP: target weights & trades)

**Input**: `RiskBook`, `FactorBook`, constraints (min/max weights, turnover, tax hints)
**Output**:

```go
type Constraint struct { Ticker string; Min, Max float64 }
type Objective struct { Type string; Params map[string]float64 } // e.g., "max_sharpe"
type Trade struct { Ticker string; Action string; Units float64; EstCost float64 }
type OptimizePlan struct {
  TargetWeights map[string]float64
  Trades []Trade
  Objective Objective
  Constraints []Constraint
  Notes string
}
```

**Impl**: mean‑variance or risk‑parity; penalty for turnover/tax; integer lots optional.
**NEXT_PATTERN**: if solution unstable → propose simplify (sector ETF).

---

### 3.16 Reporting (SRP: compose final docs)

**Input**: _All prior books (Risk, Macro, Factors, News, Optimize…)_
**Output**:

````go
type ReportSection struct { Title string; Markdown string }
type Report struct {
  Title string
  GeneratedAt time.Time
  Sections []ReportSection
  Locale string
}
````
**Impl**: assemble Analyze Report; for Optimize, an **IC Conclusion** section with rationale + proposed trades.

---

## 4) Tools (Normalization & Go Types)

> Keep tools SRP. Normalize to shared domain types above.

### 4.1 SEC EDGAR (filings → normalized financials & docs)
**Focus Endpoints**: company facts (XBRL), submissions, 10‑K/10‑Q text/PDF.
**Normalize**:
```go
type Filing struct { Accession string; Type string; Filed time.Time; PeriodEnd time.Time; URL string }
type FilingExtract struct { Sections map[string]string } // MD&A, RiskFactors, Liquidity
type EdgarPack struct {
  Ticker string
  Filings []Filing
  Extracts map[string]FilingExtract // by Accession
}
````

**Mapping**: XBRL → `Fundamentals` (TTM/MRQ); text sections → `FilingExtract`.
**Notes**: cache; rate‑limit; firm symbol mapping.

---

### 4.2 Yahoo Finance (your tools)

**Focus**: `quoteSummary` modules (price, summaryProfile, financialData), `chart` (OHLC), dividends/splits, options chain.
**Normalize** to:

- `Quote`, `PriceBar`, `CorpAction`
- Fundamentals: map `financialData` to `Fundamentals` (EPS, margins) with provenance.

---

### 4.3 FMP (Financial Modeling Prep)

**Focus**: Key metrics (TTM), income/balance/cashflow, ratios, company profile.
**Normalize** directly to `Fundamentals` / `Ratios` plus:

```go
type AnalystEstimate struct { Ticker string; FwdEPS float64; RevGrowth float64; Updated time.Time }
```

Use as **primary** for fundamentals; YF as **fallback**.

---

### 4.4 FRED (Macro)

**Focus**: CPI, Unemployment, 10Y yield, 2s10s, FX if needed.
**Normalize** to `MacroSeries` (freq, units); resample to monthly or daily as needed.

---

## 5) Tool Chaining (from standalone → composable)

- Current: each tool is called standalone by an engine.
- Plan: introduce a **Tool wrapper** that:

  - Calls tool `Run(input)`
  - Reads tool `GetNext(prevToolName string, prevOutput any) []string`
  - Dispatches to subsequent tools in order until no next remains (or hits a max depth).

- Benefits: same NEXT pattern for **tools** as **engines**; enables fallbacks and enrichments (e.g., `YF.Quote → FMP.Fundamentals if missing EPS → EDGAR.Extract`).

---

## 6) The ##NEXT_PATTERN## (Engines & Tools)

### Interface (engines)

```go
type NextSpec struct {
  NextEngines []string         // names/ids
  Reason      string
}
type Engine interface {
  Name() string
  Run(ctx context.Context, in any) (out any, err error)
  GetNext(prevName string, prevOut any, myOut any) NextSpec
}
```

**Rules** (examples):

- `PortfolioIngest`: if any Class==Crypto → add `"CryptoRisk"`.
- `MarketDataIngest`: if non‑base currencies detected → add `"FXNormalize"`.
- `Risk`: if AnnualVol>σ\* → add `"Scenario"`, `"HedgingIdeas"`.
- `NewsAnalyze`: if Negative sentiment cluster for ticker T → add `"Risk"` re‑run on that subset.

### Interface (tools)

```go
type Tool interface {
  Name() string
  Run(ctx context.Context, in any) (out any, err error)
  GetNext(prevTool string, prevOut any, myOut any) []string
}
```

**Wrapper**:

```go
func RunChain(ctx context.Context, start Tool, in any, registry map[string]Tool) ([]any, error) {
  out, err := start.Run(ctx, in); if err != nil { return nil, err }
  outputs := []any{out}
  q := start.GetNext("", in, out)
  for len(q) > 0 {
    name := q[0]; q = q[1:]
    t := registry[name]
    o, err := t.Run(ctx, out); if err != nil { return outputs, err }
    outputs = append(outputs, o)
    q = append(q, t.GetNext(start.Name(), out, o)...)
    out = o
  }
  return outputs, nil
}
```

**Orchestrator**: still owns **global** guardrails (max depth, cycles, SLAs). Engines/tools only **propose** next; orchestrator **decides**.

---

## 7) Prompt Management (libraries & approach)

| Option                              | What it gives                                    | When to use                                        |
| ----------------------------------- | ------------------------------------------------ | -------------------------------------------------- |
| **Go `text/template` + `embed`**    | Simple, fast, versionable; no deps               | Start here; perfect for SRP engines                |
| **LangChainGo (prompts & parsers)** | Prompt templates, output parsing helpers         | When you want JSON parsing helpers & few-shot mgmt |
| **CloudWeGo Eino**                  | Typed workflow DAG + node composition            | When orchestrating complex multi-engine flows      |
| **Langfuse**                        | Prompt & response observability, versions, evals | Add when you want analytics & A/B                  |
| **Pezzo / PromptLayer**             | Centralized prompt registry                      | SaaS/multitenant later                             |

**Recommendation**:

- Now: `text/template` in repo (`/prompts/<engine>/v1.tmpl`), with `embed.FS`.
- Add **Langfuse** when you want telemetry.
- Keep JSON outputs well‑defined; use a tiny parser per engine (no heavy framework needed).

---

## 8) Prompt Set (tied to outputs)

> Use `{{ }}` placeholders; keep each under `/prompts/<engine>/<name>.tmpl`.

- **NewsSummary.tmpl**
  “Summarize the article in 3–4 sentences, factual, no speculation.\n\nArticle:\n{{ .Body }}”
- **NewsEntities.tmpl**
  “Extract entities (people, orgs, locations, dates, tickers) from the text.\nReturn strict JSON: {"people":\[],"organizations":\[],"locations":\[],"dates":\[],"tickers":\[]}\n\nText:\n{{ .Body }}”
- **NewsSentiment.tmpl**
  “Overall sentiment of the article (Positive/Neutral/Negative). One word only.\n\nText:\n{{ .Body }}”
- **MacroNarrative.tmpl**
  “Given these macro series ({{ .SeriesList }}) up to {{ .AsOf }}, write a concise market context (<=150 words) and bullet implications (<=4) for EU‑centric investor.”
- **RiskCommentary.tmpl**
  “Given: AnnualVol={{ .Vol }}, VaR95={{ .VaR }}, TopContrib={{ .Top }}, write 4–6 sentences highlighting key risks and concentration.”
- **FactorCommentary.tmpl**
  “Given per‑ticker factors (value, quality, momentum z‑scores), identify top/bottom names and 3 takeaways.”
- **OptimizeRationale.tmpl**
  “Objective: {{ .Objective }}; Constraints: {{ .Constraints }}; Proposed target weights: {{ .Weights }}; Trades: {{ .Trades }}.\nWrite Investment Committee rationale (<=180 words), clear and actionable.”
- **AnalyzeReport.tmpl**
  “Compose a report with sections: Portfolio Overview, Market Context, Risk, Factors, News Highlights, Attribution.\nUse provided markdown fragments; keep under {{ .MaxWords }} words.”
- **ICConclusion.tmpl**
  “Draft the ‘Investment Committee Conclusion’ using OptimizePlan and key risks; list 3–6 explicit actions (rebalance %, hedge, TLH).”

> For JSON prompts, enforce **strict** JSON and validate. For narrative prompts, set low temperature and word caps.

---

## 9) WaterCrawl + OpenAI Batch

**Where WaterCrawl fits**:

- Weekly fetch → **NewsIngest**. It gives you URLs & bodies rapidly.
- Run **OpenAI Batch** to: summarize & extract entities for N articles tool‑free (cheap). Save results into `NewsBook`.
- In live Analyze, you **don’t** call tools; you **read cached** NewsBook → deterministic & cheap.

**Batch JSONL sample (tool‑free)**:

```json
{
  "custom_id": "news:sum:{{url}}",
  "method": "POST",
  "url": "/v1/chat/completions",
  "body": {
    "model": "gpt-4o-mini",
    "temperature": 0.1,
    "max_tokens": 200,
    "messages": [
      { "role": "system", "content": "Summarize in 3-4 sentences." },
      { "role": "user", "content": "{{ body }}" }
    ]
  }
}
```

Repeat similarly for entities (strict JSON). Store outputs keyed by URL hash.

---

## 10) Testing Strategy

- **Unit**:

  - Engines: table‑driven tests (valid → output shape).
  - Tools: fixtures (recorded JSON → normalized types).
  - Parsers: JSON strictness; error branches.

- **Contract** (per tool): JSON schema → Go struct round‑trip; field coverage %; currency correctness.
- **Property**: returns invariants, weight sums ≈ 1, z‑score mean≈0.
- **Golden**: reports snapshots (markdown) for canonical portfolio; change requires review.
- **E2E**: small portfolio → full Analyze + Optimize; assert sections exist; sanity bounds (vol < 100%, etc.).
- **Load**: parallel article summaries via Batch (dry run limits).
- **NEXT_PATTERN**: simulate outputs that trigger branches; verify orchestrator honors guardrails.

---

## 11) Migration to SaaS (quick notes)

- Keep engines **stateless**; pass state via inputs.
- Introduce user boundary (OrgID/UserID) into `Portfolio`, caching, prompts.
- Secrets per tenant; rate limiter.
- Observability: Langfuse; structured logs (engine name, duration, tokens).

---

# Tool Details (Normalization & Focus)

## SEC EDGAR

**Use**: company facts (XBRL) for TTM/MRQ; filings metadata; text sections.
**Go output**: `EdgarPack` + map to `Fundamentals`.
**Normalization**:

- Currency → portfolio base at period end.
- Preferred metrics: TTM EPS/Revenue, MRQ Debt/Cash, margins, ROE/ROIC (compute).
- Text: extract MD\&A, RiskFactors; store in `FilingExtract.Sections`.
  **Gotchas**: symbol mapping; amended filings; units scaling (millions).

## Yahoo Finance (your tools)

**Focus**:

- `chart` → `[]PriceBar` (daily), split/div adjustments into `CorpAction`.
- `quoteSummary` → `Quote`, profile (sector/industry), `financialData` (margins, EPS).
- Options (optional now) for risk hedging later.
  **Normalization**:
- Coerce currencies; timezones; ensure close vs adjusted close clarity.

## FMP

**Focus**: Key metrics, ratios; analyst estimates.
**Normalization**:

- Populate `Fundamentals.Ratios`, `TTM`, `MRQ`.
- Prefer FMP as primary; tag provenance.

## FRED

**Focus**: CPI, 10Y, 2s10s, unemployment, FX if used.
**Normalization**:

- `MacroSeries` with `Freq`, `Units`; resample to monthly; align as of Analyze date.

---

# Engines Chaining Order & E2E Data Dependencies (recap)

1. PortfolioIngest → MarketDataIngest → FXNormalize → TimeAlign
2. FundamentalsIngest + MacroIngest (parallel)
3. FactorBuild (needs TimeAlign + Fundamentals)
4. Risk (needs TimeAlign)
5. Scenario\&Stress (Risk + Macro)
6. ComplianceTax (Portfolio + region config)
7. Attribution (TimeAlign + Portfolio history)
8. NewsIngest → NewsAnalyze (parallel LLM)
9. Reporting (compose Analyze)
10. Optimize (Risk + Factors + Constraints) → IC Conclusion Reporting

---

# Optimize Command Orchestration

```
Load Analyze cache
   ↓
ConstraintsBuild (region/tax/position caps)
   ↓
Optimize (objective: max Sharpe / min Vol / target mix)
   ↓
SanityCheck (turnover, liquidity)
   ↓
IC Conclusion Report (rationale + trades)
```

**ConstraintsBuild Output**:

```go
type ConstraintsPack struct {
  MaxTurnover float64
  PositionCaps map[string]float64
  SectorCaps map[string]float64
  TLHEnabled bool
}
```

**SanityCheck Output**:

```go
type SanityReport struct { Issues []string; EstCosts float64 }
```

---

# Prompts — Library Choice & Organization

- **Now**: `text/template` + `embed.FS` (fast, simple).
- **Add**: Langfuse for logging prompts & responses, token use, versions.
- **Later**: LangChainGo only if you want built‑in output parsers & few‑shot mgmt.
- **Storage**: `/prompts/<engine>/<template>.tmpl`, version in filename (`v1`).
- **Conventions**:

  - Narrative prompts ≤ 180 words target; `temperature=0.2`.
  - JSON prompts: start with “Return strict JSON:” and validate.

---

# Implementation Checklist (dev + tests)

- [ ] `pkg/types`: core domain types (AssetID, PriceBar, Fundamentals, MacroSeries, …)
- [ ] `pkg/engines/…`: each engine with `Run` + `GetNext`
- [ ] `pkg/tools/…`: YF, FMP, EDGAR, FRED (normalize to types)
- [ ] `pkg/orch`: orchestrator with guardrails; parallel fan‑out helpers
- [ ] `pkg/next`: tool/engine chain wrapper (cycle detect, max depth)
- [ ] `prompts/…`: templates + tiny renderer
- [ ] `tests/fixtures`: recorded tool payloads → normalization contract tests
- [ ] `tests/golden`: reports markdown snapshots
- [ ] `cmd/analyze`, `cmd/optimize`: wire the flows

---

## Notes on “Engines decide next” (NEXT_PATTERN Analysis)

**Current limitation**: engines can’t drive flow.
**Plan**: each engine returns `NextSpec` proposals; orchestrator merges:

- Validate proposals (authz, cost, SLAs).
- Resolve conflicts; de‑dup; enforce max depth.
- This keeps engines smart but prevents “runaway” chains.

**Example**:

- `Risk` sees `AnnualVol>20%` and `Crypto weight>10%` → proposes `Scenario`, `CryptoRisk`.
- Orchestrator accepts `Scenario`, rejects `CryptoRisk` if CryptoRisk engine disabled in config.

---

## Where to start (practical 2‑week sprint)

1. Types + PortfolioIngest + MarketDataIngest + FXNormalize + TimeAlign (with tests)
2. FundamentalsIngest + MacroIngest + FactorBuild
3. Risk + Scenario
4. Reporting (Analyze core)
5. Optimize (simple mean‑variance) + IC Conclusion
6. Add NewsIngest/Analyze; wire Batch for text (optional)

---

If you want, I can turn this into a starter repo structure with the type files and engine skeletons to copy/paste.
