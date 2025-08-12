# Portfolio Analysis Engines

## Overview

The **Engine** package provides a collection of specialized analysis engines that follow **Single Responsibility Principle (SRP)** and support both **standalone operation** and **pipeline orchestration**. Each engine focuses on a specific domain of portfolio analysis while maintaining clean interfaces for composition and chaining.

## Architecture Principles

### Single Responsibility Principle (SRP)

Each engine has **one clear domain expertise**:

- `RiskEngine`: Portfolio risk quantification and assessment
- `NewsEngine`: News intelligence and market sentiment analysis
- `AllocationEngine`: Asset allocation optimization and analysis
- `ComplianceEngine`: Regulatory compliance and jurisdiction analysis
- `ReallocationEngine`: Portfolio rebalancing recommendations

### Dual Output Strategy

All engines support **context-aware output modes**:

- **OutputModeReport**: Human-readable analysis reports
- **OutputModeData**: Structured data for AI agents and other engines

### SharedBag Integration

Engines coordinate through **SharedBag** for:

- **Input Data Access**: Portfolio, market data, configuration
- **State Management**: Intermediate results and metadata
- **Pipeline Coordination**: Cross-engine data sharing

## Common Engine Interface

### Core Interface Pattern

```go
type Engine interface {
    // Execute analysis with context-aware output
    Execute(ctx context.Context, portfolio *models.NormalizedPortfolio, mode models.OutputMode) (models.EngineResult, error)

    // Engine metadata
    Name() string
    Dependencies() []string  // Names of required preceding engines
    Version() string
}
```

### Result Interface

```go
type EngineResult interface {
    GetSummary() string
    GetRecommendations() []string
    GetTimestamp() time.Time
    GetData() interface{}  // Engine-specific structured data
}
```

### Output Modes (models.OutputMode)

```go
const (
    OutputModeReport OutputMode = iota  // Human-readable reports
    OutputModeData                      // Structured data for AI/pipeline
)
```

## Engine Catalog

### 🎯 RiskEngine (`risk/`)

**Domain**: Portfolio risk quantification and assessment

**Responsibilities**:

- Concentration risk analysis (Herfindahl Index, position sizes)
- Geographic and currency risk assessment
- Sector concentration analysis
- Liquidity risk evaluation
- Overall risk scoring and recommendations

**Dependencies**: None (pure mathematical analysis)
**Outputs**: Risk scores, concentration metrics, actionable risk recommendations

### 📰 NewsEngine (`news/`)

**Domain**: News intelligence and market sentiment analysis

**Responsibilities**:

- Multi-phase news processing (fetch → summarize → analyze)
- Portfolio-contextual news analysis
- Market sentiment extraction
- News-based portfolio impact assessment

**Dependencies**: None (data gathering engine)
**Outputs**: News summaries, sentiment analysis, market intelligence

### 🎚️ AllocationEngine (`allocation/`)

**Domain**: Asset allocation optimization and analysis

**Responsibilities**:

- Current allocation analysis vs benchmarks
- Diversification effectiveness assessment
- Allocation optimization recommendations
- Rebalancing threshold analysis

**Dependencies**: RiskEngine (for risk-adjusted allocation)
**Outputs**: Allocation metrics, optimization recommendations, rebalancing triggers

### ✅ ComplianceEngine (`compliance/`)

**Domain**: Regulatory compliance and jurisdiction analysis

**Responsibilities**:

- Jurisdiction-specific compliance checking
- Investment restriction validation
- Regulatory requirement assessment
- Compliance risk identification

**Dependencies**: None (regulatory rule-based analysis)
**Outputs**: Compliance status, violations, regulatory recommendations

### 🔄 ReallocationEngine (`reallocation/`)

**Domain**: Portfolio rebalancing and optimization recommendations

**Responsibilities**:

- Comprehensive portfolio rebalancing recommendations
- AI-powered investment research and suggestions
- Implementation strategy development
- Tax and cost optimization

**Dependencies**: RiskEngine, NewsEngine, AllocationEngine (synthesis engine)
**Outputs**: Rebalancing recommendations, implementation strategies, investment suggestions

## Engine Orchestration

### Pipeline Patterns

#### Sequential Pipeline

```go
// Data gathering engines first
await newsEngine.Execute(ctx, portfolio, models.OutputModeData)
await marketDataEngine.Execute(ctx, portfolio, models.OutputModeData)

// Analysis engines (can run in parallel)
await riskEngine.Execute(ctx, portfolio, models.OutputModeData)
await allocationEngine.Execute(ctx, portfolio, models.OutputModeData)
await complianceEngine.Execute(ctx, portfolio, models.OutputModeData)

// Synthesis engine
await reallocationEngine.Execute(ctx, portfolio, models.OutputModeReport)
```

#### Parallel Execution

```go
// Independent engines can run concurrently
go riskEngine.Execute(ctx, portfolio, models.OutputModeData)
go complianceEngine.Execute(ctx, portfolio, models.OutputModeData)
go newsEngine.Execute(ctx, portfolio, models.OutputModeData)
```

### Orchestrator Interface

```go
type EngineOrchestrator interface {
    // Add engines to execution pipeline
    AddEngine(engine Engine) EngineOrchestrator

    // Execute pipeline with dependency resolution
    Execute(ctx context.Context, portfolio *models.NormalizedPortfolio) (*OrchestratedResult, error)

    // Execute specific engines only
    ExecuteEngines(ctx context.Context, portfolio *models.NormalizedPortfolio, engines []string) (*OrchestratedResult, error)
}
```

## SharedBag Integration

### Standard Key Patterns

```go
const (
    // Engine-specific result keys
    KNewsAnalysis       keys.Key = "news_analysis"
    KRiskAnalysis       keys.Key = "risk_analysis"
    KAllocationAnalysis keys.Key = "allocation_analysis"
    KComplianceAnalysis keys.Key = "compliance_analysis"

    // Engine metadata keys
    KEngineExecutionOrder keys.Key = "engine_execution_order"
    KEngineTimestamps     keys.Key = "engine_timestamps"
    KEngineErrors         keys.Key = "engine_errors"
)
```

### State Coordination Pattern

```go
// Engine reads shared state
if newsData, ok := sharedBag.Get(keys.KNewsAnalysis); ok {
    // Use news analysis from NewsEngine
}

// Engine stores results for others
sharedBag.Set(keys.KRiskAnalysis, riskData)
sharedBag.Set(keys.KRiskAnalysisTime, time.Now())
```

## Implementation Guidelines

### Engine Structure Template

```go
package enginename

type Engine struct {
    sharedBag  bag.SharedBag
    config     *config.Config
    aiClient   ai.Client      // For AI-powered engines

    // Engine-specific components
    analyzer   SpecificAnalyzer
}

func NewEngine(cfg *config.Config, sharedBag bag.SharedBag, aiClient ai.Client) *Engine {
    return &Engine{
        sharedBag: sharedBag,
        config:    cfg,
        aiClient:  aiClient,
        analyzer:  NewSpecificAnalyzer(),
    }
}

func (e *Engine) Execute(ctx context.Context, portfolio *models.NormalizedPortfolio, mode models.OutputMode) (models.EngineResult, error) {
    // Implementation with mode-specific output
}
```

### Testing Standards

- **Unit Tests**: Each engine tested independently
- **Integration Tests**: Engine pipeline workflows
- **Mock SharedBag**: For isolated testing
- **Deterministic Results**: Reproducible analysis outputs

## Configuration

### Engine-Specific Config

```yaml
engines:
  risk:
    enabled: true
    calculation_method: 'herfindahl'
    risk_tolerance: 'moderate'

  news:
    enabled: true
    max_articles: 20
    analysis_depth: 'detailed'

  allocation:
    enabled: true
    target_allocations:
      stocks: 70
      bonds: 20
      cash: 10
```

## Future Enhancements

### Planned Engines

- **EconomicEngine**: Macroeconomic analysis and indicators
- **MarketDataEngine**: Real-time market data and technical analysis
- **TechnicalEngine**: Technical analysis and chart patterns
- **ESGEngine**: Environmental, Social, and Governance analysis

### Advanced Features

- **Dynamic Dependencies**: Runtime dependency resolution
- **Conditional Execution**: Skip engines based on conditions
- **Result Caching**: Cache expensive engine computations
- **Streaming Results**: Real-time engine result streaming

## Data Flow

Portfolio Data → Engine Pipeline → SharedBag State → Orchestrated Analysis → Final Reports

The engine architecture provides **modular, testable, and composable** portfolio analysis capabilities while maintaining clean separation of concerns and supporting complex analytical workflows.
