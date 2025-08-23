# Base Batch Engine

The base batch engine provides common infrastructure for all batch processing engines using the template method pattern.

## Architecture

The base engine handles infrastructure concerns while hooks provide business-specific customization:

### Infrastructure (Base Engine Responsibility)

- **Tool execution and external tool handling** - All tool execution is handled uniformly by the base engine
- **Batch job management and iteration control** - Common workflow management
- **Error handling and retry logic** - Consistent error handling patterns
- **Metrics collection and monitoring** - Performance and usage tracking
- **SharedBag integration** - Context management across iterations

### Business Logic (Hooks Responsibility)

- **Prompt generation** for specific analysis types
- **Result processing and storage** strategies
- **Iteration continuation logic** - When to stop processing
- **Custom ID generation** patterns
- **Pre/post iteration** processing

## Key Insight: Tool Execution Separation

**Tool execution is handled entirely by the base engine** - hooks should never handle tool execution directly.

The base engine:

- ✅ Handles all `models.ToolCall` execution uniformly
- ✅ Manages external tool processing (web search, citation handling)
- ✅ Provides consistent error handling and logging
- ✅ Integrates with tool registry and metrics system

The hooks focus purely on:

- ✅ Business-specific prompt generation
- ✅ Result processing and storage strategies
- ✅ Iteration control and continuation logic

This separation ensures:

- **No code duplication** across different engines
- **Consistent tool behavior** regardless of engine type
- **Centralized tool monitoring** and error handling
- **Clean separation of concerns** between infrastructure and business logic

## Usage

```go
// Create hooks for your specific engine
hooks := &YourEngineHooks{
    promptBuilder: pb,
}

// Base engine handles all infrastructure including tool execution
baseEngine := base.NewBatchEngine(name, model, constraints, hooks, sharedBag)

// Your engine embeds base engine
engine := &YourEngine{
    BatchEngine: *baseEngine,
    // your business-specific fields
}
```

## Interface Design

- **`BatchEngineHooks`**: Business logic customization points only
- **`ToolCallExecutor`**: **REMOVED** - tool execution is infrastructure responsibility
- **Base engine handles all tool execution**: External tools, regular tools, error handling

## Hook Methods

### Required Business Logic Hooks

- `GetInitialPrompt(ctx)` - Generate analysis-specific prompts
- `GenerateCustomID(iteration, jobIndex)` - Create unique job identifiers
- `ProcessToolResult(customID, toolName, result, sharedBag)` - Handle tool outputs
- `ProcessFinalResult(customID, content, sharedBag)` - Process final analysis
- `ShouldContinueIteration(iteration, nextJobs)` - Control iteration flow
- `ResultKey()` - Specify where to store final results

### Optional Lifecycle Hooks

- `PreIteration(iteration, jobs)` - Setup before each iteration
- `PostIteration(iteration, results)` - Cleanup after each iteration

These hooks enable powerful analysis workflows:

### **PreIteration Hook Use Cases**

#### **1. Dynamic Prompt Enhancement**

```go
func (h *RiskBatchEngineHooks) PreIteration(iteration int, jobs []base.BatchJob) error {
    if iteration > 0 {
        // Enhance prompts based on previous iteration results
        previousResults := h.getPreviousResults()
        for i, job := range jobs {
            jobs[i].Messages = h.enrichPromptWithContext(job.Messages, previousResults)
        }
    }
    return nil
}
```

#### **2. Risk Context Preparation**

```go
func (h *RiskBatchEngineHooks) PreIteration(iteration int, jobs []base.BatchJob) error {
    // Prepare market context for this iteration
    marketData := h.gatherMarketConditions()
    portfolioState := h.getCurrentPortfolioState()

    // Inject current market conditions into all jobs
    for i := range jobs {
        jobs[i] = h.addMarketContext(jobs[i], marketData, portfolioState)
    }

    slog.Info("Risk iteration prepared",
        "iteration", iteration,
        "market_volatility", marketData.Volatility,
        "jobs_count", len(jobs))

    return nil
}
```

### **PostIteration Hook Use Cases**

#### **1. Results Analysis & Next Step Planning**

```go
func (h *RiskBatchEngineHooks) PostIteration(iteration int, results *models.Aggregated) error {
    // Analyze results to determine next steps
    riskLevel := h.analyzeRiskLevel(results)

    if riskLevel == "HIGH" {
        // Trigger additional stress testing iteration
        return h.scheduleStressTestIteration()
    }

    if results.Successes < len(results.Results)/2 {
        // Too many failures - adjust strategy
        return h.adjustAnalysisStrategy()
    }

    return nil
}
```

#### **2. Incremental Risk Assessment**

```go
func (h *RiskBatchEngineHooks) PostIteration(iteration int, results *models.Aggregated) error {
    // Build cumulative risk picture
    cumulativeRisk := h.buildCumulativeRiskAssessment(results)

    // Store intermediate risk metrics
    sharedBag.Update(bag.KRiskMetrics, func(a any) any {
        metrics := a.(*RiskMetrics)
        metrics.AddIterationResults(iteration, cumulativeRisk)
        return metrics
    })

    // Decide if we have enough data or need more analysis
    if h.riskConfidenceLevel(cumulativeRisk) > 0.9 {
        return h.signalEarlyCompletion()
    }

    return nil
}
```

### **Advanced Patterns**

#### **Multi-Phase Risk Analysis**

```go
func (h *RiskBatchEngineHooks) PreIteration(iteration int, jobs []base.BatchJob) error {
    switch iteration {
    case 0:
        // Phase 1: Market risk assessment
        return h.prepareMarketRiskAnalysis(jobs)
    case 1:
        // Phase 2: Credit risk analysis
        return h.prepareCreditRiskAnalysis(jobs)
    case 2:
        // Phase 3: Liquidity risk analysis
        return h.prepareLiquidityRiskAnalysis(jobs)
    }
    return nil
}

func (h *RiskBatchEngineHooks) PostIteration(iteration int, results *models.Aggregated) error {
    // Combine results from different risk phases
    return h.integrateRiskPhaseResults(iteration, results)
}
```

## Tool Execution Infrastructure

The base engine's tool execution infrastructure handles:

- **External tool processing** (web search with citation handling)
- **Regular tool execution** with consistent error handling
- **Tool registry integration** and metrics collection
- **Logging and monitoring** for all tool operations

This architecture ensures that every engine benefits from consistent, monitored, and properly handled tool execution while focusing their hooks purely on business logic concerns.

### Tool Execution Flow

```
┌─────────────────┐    ┌─────────────────────┐    ┌─────────────────┐
│   Hook Calls    │───▶│  Base Engine       │───▶│  Tool Registry  │
│  (Business)     │    │  (Infrastructure)   │    │  (Execution)    │
│                 │    │                     │    │                 │
│ • Prompt Gen    │    │ • Tool Resolution   │    │ • External APIs │
│ • Result Proc   │    │ • Error Handling    │    │ • Data Sources  │
│ • Continuation  │    │ • Metrics/Logging   │    │ • Caching       │
└─────────────────┘    └─────────────────────┘    └─────────────────┘
```

**Key Benefits:**

- ✅ **No code duplication** - All engines use same tool execution infrastructure
- ✅ **Consistent behavior** - External and regular tools handled uniformly
- ✅ **Centralized monitoring** - All tool execution logged and tracked
- ✅ **Clean architecture** - Clear separation between infrastructure and business logic
- ✅ **Easy maintenance** - Tool execution logic centralized and easier to update
