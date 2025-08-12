# LLM Batch Processing Architecture Proposal

## Executive Summary

This document proposes a comprehensive batch processing architecture for the new `internal/llm` package to support OpenAI's Batch API for asynchronous processing of large-scale AI workloads. The proposed architecture maintains compatibility with existing real-time AI processing while introducing efficient batch capabilities for portfolio analysis, market research, and compliance evaluation scenarios.

## Current Architecture Analysis

### Existing AI Package Structure (Migration Context)

The current `internal/ai/` package will be replaced by the new `internal/llm/` package with batch capabilities:

```
internal/ai/ (TO BE REPLACED)
├── client.go                    # Will migrate to internal/llm/client.go
├── client_test.go              # Will migrate to internal/llm/client_test.go
├── factory.go                  # Will migrate to internal/llm/factory.go
├── schema.go                   # Will migrate to internal/llm/schema.go
└── openai/                     # Will migrate to internal/llm/openai/
    ├── provider.go             # Core provider with dual API support
    ├── session.go              # Session management (Chat/Responses API)
    ├── response-api.go         # Responses API implementation
    ├── stream.go               # Streaming support
    └── web_search_preview_tracking.go
```

### Current Capabilities Analysis

**Strengths from internal/ai/ (to be preserved):**

- Clean provider abstraction supporting multiple AI services
- Dual API support (Chat Completions + Responses API)
- Tool integration with SharedBag state management
- Streaming capabilities for real-time responses
- Web search integration for market intelligence
- Reasoning model support (o1-series, gpt-5) with parameter restrictions

**Limitations for Batch Processing (to be addressed in internal/llm/):**

- No support for asynchronous job management
- Missing cost optimization for large-scale operations
- No handling of 24-hour processing windows
- Limited support for bulk portfolio analysis
- No batch result correlation and aggregation
- Model-specific parameter handling (gpt-5 class models don't support ToolChoice/Temperature)

## Batch Processing Use Cases in Mosychlos

### 1. Portfolio Risk Assessment at Scale

- Process hundreds of portfolio configurations simultaneously
- Evaluate risk scenarios across different market conditions
- Generate compliance reports for multiple jurisdictions

### 2. Market Research Automation

- Batch analysis of earnings calls, SEC filings, and news sentiment
- Economic indicator correlation across multiple timeframes
- Sector rotation analysis with historical data processing

### 3. Investment Committee Preparation

- Generate analysis reports for all portfolio holdings
- Create comparative analysis across investment strategies
- Produce regulatory compliance documentation in bulk

### 4. Performance Attribution Analysis

- Historical performance analysis across multiple time periods
- Risk-adjusted return calculations for large universes
- Attribution analysis for institutional client reporting

## Proposed Batch Architecture

### Extended LLM Package Structure

```
internal/llm/                   # New LLM package (replacing internal/ai/)
├── client.go                   # Enhanced with batch support
├── batch/                      # New batch processing module
│   ├── manager.go              # Batch job lifecycle management
│   ├── job.go                  # Individual batch job representation
│   ├── queue.go                # Local job queue for status tracking
│   ├── result_aggregator.go    # Batch result processing and correlation
│   └── cost_optimizer.go       # Cost analysis and optimization
├── openai/
│   ├── provider.go             # Enhanced with batch endpoints
│   ├── batch_client.go         # OpenAI Batch API integration
│   ├── batch_formatter.go      # JSONL formatting utilities
│   └── batch_monitor.go        # Batch status monitoring
├── middleware/                 # Rate limiting and retry middleware
│   ├── middleware.go           # Core middleware interfaces
│   ├── rate_limiting.go        # Rate limiting middleware
│   └── retry.go                # Retry middleware
├── health/                     # Health monitoring
│   └── monitor.go              # Provider health monitoring
├── validation/                 # Configuration validation
│   └── config_validator.go     # Provider config validation
└── templates/                  # Batch job templates (leverages internal/prompt)
    ├── portfolio_analysis.go   # Portfolio analysis batch templates
    ├── market_research.go      # Market research batch templates
    └── compliance_reports.go   # Compliance reporting templates
```

### Integration with Existing Prompt System

The new `internal/llm/` package will leverage the existing `internal/prompt/` package for template management:

- **Template Reuse**: Batch templates will utilize existing portfolio analysis templates from `internal/prompt/templates/portfolio/`
- **Regional Support**: Batch processing will support regional prompt variations through the existing `RegionalManager` interface
- **Investment Research Integration**: Batch jobs will use existing investment research templates from `internal/prompt/templates/investment_research/`
- **Prompt Dependencies**: Batch request generation will use the existing `Dependencies` injection pattern for consistent data access

## Core Batch Processing Interfaces

### 1. Enhanced AI Client Interface with Model-Aware Parameter Handling

```go
// Enhanced models.AiClient interface with GPT-5 considerations
type AiClient interface {
    // Existing synchronous methods
    RegisterTool(t Tool)
    SetToolConsumer(consumer ToolConsumer)

    // New batch processing methods with model-aware parameter handling
    SubmitBatch(ctx context.Context, requests []BatchRequest, opts BatchOptions) (*BatchJob, error)
    GetBatchStatus(ctx context.Context, jobID string) (*BatchStatus, error)
    GetBatchResults(ctx context.Context, jobID string) ([]BatchResult, error)
    CancelBatch(ctx context.Context, jobID string) error
    ListBatches(ctx context.Context, filters BatchFilters) ([]BatchJob, error)
}

type BatchRequest struct {
    CustomID    string                 `json:"custom_id"`
    Method      string                 `json:"method"`
    URL         string                 `json:"url"`
    Body        map[string]interface{} `json:"body"`
    ModelClass  ModelClass            `json:"model_class,omitempty"` // Track model capabilities
}

type ModelClass string
const (
    ModelClassStandard  ModelClass = "standard"  // Supports all parameters (gpt-4-turbo, etc.)
    ModelClassReasoning ModelClass = "reasoning" // Limited parameters (gpt-5, gpt-5-mini, o1-series)
)

type BatchOptions struct {
    CompletionWindow string            `json:"completion_window"` // "24h"
    Metadata        map[string]string  `json:"metadata,omitempty"`
    Priority        BatchPriority      `json:"priority,omitempty"`
    CostOptimize    bool              `json:"cost_optimize"`
    ModelClass      ModelClass        `json:"model_class,omitempty"` // Optimize requests per model class
}

// Model capability detection utilities
func DetectModelClass(modelName string) ModelClass {
    switch {
    case strings.HasPrefix(modelName, "gpt-5"):
        return ModelClassReasoning // gpt-5, gpt-5-mini, etc.
    case modelName == "o1-preview" || modelName == "o1-mini":
        return ModelClassReasoning
    default:
        return ModelClassStandard
    }
}

func IsReasoningModel(modelName string) bool {
    return DetectModelClass(modelName) == ModelClassReasoning
}
```

type BatchJob struct {
ID string `json:"id"`
Status BatchStatus `json:"status"`
InputFileID string `json:"input_file_id"`
OutputFileID *string `json:"output_file_id"`
ErrorFileID *string `json:"error_file_id"`
CreatedAt time.Time `json:"created_at"`
CompletedAt *time.Time `json:"completed_at"`
RequestCounts RequestCounts `json:"request_counts"`
Metadata map[string]string `json:"metadata"`
CostEstimate *CostEstimate `json:"cost_estimate"`
}

type BatchStatus string
const (
BatchStatusValidating BatchStatus = "validating"
BatchStatusFailed BatchStatus = "failed"
BatchStatusInProgress BatchStatus = "in_progress"
BatchStatusFinalizing BatchStatus = "finalizing"
BatchStatusCompleted BatchStatus = "completed"
BatchStatusExpired BatchStatus = "expired"
BatchStatusCancelled BatchStatus = "cancelled"
)

````

### 2. Batch Manager Implementation

```go
// internal/ai/batch/manager.go
type Manager struct {
    provider      models.Provider
    storage       BatchStorage
    sharedBag     bag.SharedBag
    costOptimizer *CostOptimizer
    monitor       *BatchMonitor
}

type BatchStorage interface {
    SaveJob(ctx context.Context, job *BatchJob) error
    GetJob(ctx context.Context, jobID string) (*BatchJob, error)
    ListJobs(ctx context.Context, filters BatchFilters) ([]BatchJob, error)
    UpdateJobStatus(ctx context.Context, jobID string, status BatchStatus) error
}

func (m *Manager) SubmitPortfolioAnalysisBatch(
    ctx context.Context,
    portfolios []models.Portfolio,
    analysisType AnalysisType,
) (*BatchJob, error) {
    // Generate batch requests from portfolio templates
    requests := m.buildPortfolioAnalysisRequests(portfolios, analysisType)

    // Optimize for cost efficiency
    if m.costOptimizer != nil {
        requests = m.costOptimizer.OptimizeRequests(requests)
    }

    // Submit batch to OpenAI
    return m.SubmitBatch(ctx, requests, BatchOptions{
        CompletionWindow: "24h",
        Metadata: map[string]string{
            "type": "portfolio_analysis",
            "analysis_type": string(analysisType),
            "portfolio_count": fmt.Sprintf("%d", len(portfolios)),
        },
        CostOptimize: true,
    })
}
````

### 3. OpenAI Batch Client Integration

```go
// internal/ai/openai/batch_client.go
type BatchClient struct {
    client    openai.Client
    config    config.LLMConfig
    formatter *BatchFormatter
    monitor   *BatchMonitor
}

func (bc *BatchClient) SubmitBatch(ctx context.Context, requests []BatchRequest, opts BatchOptions) (*BatchJob, error) {
    // Convert requests to JSONL format
    jsonlFile, err := bc.formatter.RequestsToJSONL(requests)
    if err != nil {
        return nil, fmt.Errorf("failed to format requests: %w", err)
    }

    // Upload input file
    file, err := bc.client.Files.Create(ctx, openai.FileCreateParams{
        File:    jsonlFile,
        Purpose: "batch",
    })
    if err != nil {
        return nil, fmt.Errorf("failed to upload batch file: %w", err)
    }

    // Create batch job
    batch, err := bc.client.Batches.Create(ctx, openai.BatchCreateParams{
        InputFileID:      file.ID,
        Endpoint:         "/v1/chat/completions", // Or "/v1/responses"
        CompletionWindow: opts.CompletionWindow,
        Metadata:         opts.Metadata,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create batch: %w", err)
    }

    // Convert OpenAI batch to our BatchJob format
    job := &BatchJob{
        ID:            batch.ID,
        Status:        BatchStatus(batch.Status),
        InputFileID:   batch.InputFileID,
        OutputFileID:  batch.OutputFileID,
        ErrorFileID:   batch.ErrorFileID,
        CreatedAt:     time.Unix(batch.CreatedAt, 0),
        RequestCounts: RequestCounts{
            Total:     batch.RequestCounts.Total,
            Completed: batch.RequestCounts.Completed,
            Failed:    batch.RequestCounts.Failed,
        },
        Metadata: batch.Metadata,
    }

    // Start monitoring
    bc.monitor.StartMonitoring(ctx, job.ID)

    return job, nil
}
```

### 4. Result Aggregation and Processing

```go
// internal/ai/batch/result_aggregator.go
type ResultAggregator struct {
    sharedBag bag.SharedBag
    storage   BatchStorage
}

type PortfolioAnalysisResults struct {
    JobID           string                    `json:"job_id"`
    AnalysisType    AnalysisType             `json:"analysis_type"`
    CompletedAt     time.Time                `json:"completed_at"`
    Results         []PortfolioAnalysisItem   `json:"results"`
    Summary         *BatchAnalysisSummary     `json:"summary"`
    CostBreakdown   *CostBreakdown           `json:"cost_breakdown"`
    ProcessingStats *ProcessingStatistics     `json:"processing_stats"`
}

type PortfolioAnalysisItem struct {
    CustomID     string                 `json:"custom_id"`
    PortfolioID  string                 `json:"portfolio_id"`
    Analysis     *models.Analysis       `json:"analysis"`
    RiskMetrics  *models.RiskMetrics    `json:"risk_metrics"`
    Compliance   *models.ComplianceCheck `json:"compliance"`
    Error        *string                `json:"error,omitempty"`
}

func (ra *ResultAggregator) ProcessPortfolioAnalysisResults(
    ctx context.Context,
    jobID string,
) (*PortfolioAnalysisResults, error) {
    // Get job details
    job, err := ra.storage.GetJob(ctx, jobID)
    if err != nil {
        return nil, fmt.Errorf("failed to get job: %w", err)
    }

    if job.Status != BatchStatusCompleted {
        return nil, fmt.Errorf("job %s is not completed, status: %s", jobID, job.Status)
    }

    // Download and parse results
    results, err := ra.downloadAndParseResults(ctx, job.OutputFileID)
    if err != nil {
        return nil, fmt.Errorf("failed to process results: %w", err)
    }

    // Aggregate and summarize
    aggregated := &PortfolioAnalysisResults{
        JobID:        jobID,
        CompletedAt:  *job.CompletedAt,
        Results:      results,
        Summary:      ra.generateSummary(results),
        CostBreakdown: ra.calculateCostBreakdown(job),
    }

    // Store aggregated results in SharedBag for engine access
    ra.sharedBag.Set(fmt.Sprintf("batch_analysis_%s", jobID), aggregated)

    return aggregated, nil
}
```

## Batch Processing Templates

### 1. Portfolio Analysis Template Integration with Model-Aware Parameters

```go
// internal/llm/templates/portfolio_analysis.go
type PortfolioAnalysisTemplate struct {
    Type         AnalysisType `json:"type"`
    promptManager prompt.Manager `json:"-"` // Use existing prompt system
    Tools        []string     `json:"tools"`
    Parameters   map[string]interface{} `json:"parameters"`
    ModelName    string       `json:"model_name"`
    ModelClass   ModelClass   `json:"model_class"`
}

func NewRiskAnalysisTemplate(promptManager prompt.Manager, modelName string) *PortfolioAnalysisTemplate {
    modelClass := DetectModelClass(modelName)

    // Adjust parameters based on model capabilities
    params := map[string]interface{}{
        "max_tokens": 2000,
        "risk_metrics_required": true,
        "compliance_check": true,
    }

    // Only add temperature for standard models (not gpt-5 class)
    if modelClass == ModelClassStandard {
        params["temperature"] = 0.1 // Low temperature for analytical consistency
    }

    return &PortfolioAnalysisTemplate{
        Type: AnalysisTypeRisk,
        promptManager: promptManager, // Leverage existing prompt management
        ModelName: modelName,
        ModelClass: modelClass,
        Tools: []string{
            "fmp_stock_data",
            "fred_economic_data",
            "risk_calculator",
            "correlation_analyzer",
        },
        Parameters: params,
    }
}

func (t *PortfolioAnalysisTemplate) BuildBatchRequest(ctx context.Context, portfolio models.Portfolio, customID string) (BatchRequest, error) {
    // Use existing prompt system to generate prompts
    systemPrompt, err := t.promptManager.BuildPrompt(ctx, t.Type)
    if err != nil {
        return BatchRequest{}, fmt.Errorf("failed to build prompt: %w", err)
    }

    // Build request body with model-aware parameters
    body := map[string]interface{}{
        "model": t.ModelName,
        "messages": []map[string]string{
            {"role": "system", "content": systemPrompt},
            {"role": "user", "content": t.buildPortfolioPrompt(portfolio)},
        },
        "max_tokens": t.Parameters["max_tokens"],
    }

    // Only add temperature and tool_choice for standard models
    if t.ModelClass == ModelClassStandard {
        if temp, ok := t.Parameters["temperature"]; ok {
            body["temperature"] = temp
        }

        // Add tools and tool_choice for standard models
        if len(t.Tools) > 0 {
            body["tools"] = t.buildToolDefinitions()
            // GPT-5 class models don't support tool_choice parameter
        }
    } else {
        // For reasoning models (gpt-5, o1-series), tools are handled differently
        // Tools may not be supported or have different syntax
        slog.Warn("Using reasoning model with limited parameter support",
            "model", t.ModelName,
            "model_class", t.ModelClass)
    }

    return BatchRequest{
        CustomID: customID,
        Method:   "POST",
        URL:      "/v1/chat/completions",
        Body:     body,
        ModelClass: t.ModelClass,
    }, nil
}
```

### 2. Regional Batch Processing Support with Model Compatibility

```go
// internal/llm/templates/regional_batch.go
type RegionalBatchTemplate struct {
    regionalManager prompt.RegionalManager
    baseTemplate    *PortfolioAnalysisTemplate
    modelName       string
    modelClass      ModelClass
}

func NewRegionalBatchTemplate(regionalManager prompt.RegionalManager, baseTemplate *PortfolioAnalysisTemplate, modelName string) *RegionalBatchTemplate {
    return &RegionalBatchTemplate{
        regionalManager: regionalManager,
        baseTemplate:    baseTemplate,
        modelName:       modelName,
        modelClass:      DetectModelClass(modelName),
    }
}

func (rbt *RegionalBatchTemplate) BuildRegionalBatchRequests(
    ctx context.Context,
    portfolios []models.Portfolio,
    sharedBag bag.SharedBag,
) ([]BatchRequest, error) {
    requests := make([]BatchRequest, 0, len(portfolios))

    for i, portfolio := range portfolios {
        // Use existing regional prompt system
        regionalPrompt, err := rbt.regionalManager.GenerateRegionalPrompt(
            ctx,
            rbt.baseTemplate.Type,
            sharedBag,
            prompt.PromptData{Portfolio: &portfolio},
        )
        if err != nil {
            return nil, fmt.Errorf("failed to generate regional prompt for portfolio %s: %w", portfolio.ID, err)
        }

        customID := fmt.Sprintf("regional_%s_%s_%d", rbt.baseTemplate.Type, portfolio.ID, i)

        // Build request body with model-aware parameters
        body := map[string]interface{}{
            "model": rbt.modelName,
            "messages": []map[string]string{
                {"role": "system", "content": regionalPrompt},
                {"role": "user", "content": rbt.buildPortfolioContext(portfolio)},
            },
            "max_tokens": rbt.baseTemplate.Parameters["max_tokens"],
        }

        // Model-specific parameter handling
        if rbt.modelClass == ModelClassStandard {
            // Standard models support temperature and tools
            if temp, ok := rbt.baseTemplate.Parameters["temperature"]; ok {
                body["temperature"] = temp
            }

            if len(rbt.baseTemplate.Tools) > 0 {
                body["tools"] = rbt.baseTemplate.buildToolDefinitions()
                // Note: tool_choice not set for regional batch to allow model flexibility
            }
        } else {
            // Reasoning models (gpt-5, o1-series) have parameter restrictions
            slog.Info("Regional batch using reasoning model",
                "model", rbt.modelName,
                "portfolio_id", portfolio.ID,
                "note", "temperature and tool_choice not supported")

            // For reasoning models, we may need to embed tool context in the prompt
            // rather than using the tools parameter
            if len(rbt.baseTemplate.Tools) > 0 {
                toolContext := rbt.buildToolContextForReasoningModel(rbt.baseTemplate.Tools)
                // Append tool context to system message for reasoning models
                body["messages"] = []map[string]string{
                    {"role": "system", "content": regionalPrompt + "\n\n" + toolContext},
                    {"role": "user", "content": rbt.buildPortfolioContext(portfolio)},
                }
            }
        }

        request := BatchRequest{
            CustomID:   customID,
            Method:     "POST",
            URL:        "/v1/chat/completions",
            Body:       body,
            ModelClass: rbt.modelClass,
        }

        requests = append(requests, request)
    }

    return requests, nil
}

// buildToolContextForReasoningModel converts tool definitions to prompt context
// since reasoning models may not support the tools parameter
func (rbt *RegionalBatchTemplate) buildToolContextForReasoningModel(tools []string) string {
    if len(tools) == 0 {
        return ""
    }

    context := "Available analysis capabilities for your use:\n"
    for _, tool := range tools {
        switch tool {
        case "fmp_stock_data":
            context += "- Financial data analysis using comprehensive market data\n"
        case "fred_economic_data":
            context += "- Economic indicator analysis using Federal Reserve economic data\n"
        case "risk_calculator":
            context += "- Quantitative risk metric calculations\n"
        case "correlation_analyzer":
            context += "- Portfolio correlation and diversification analysis\n"
        default:
            context += fmt.Sprintf("- %s analysis capabilities\n", tool)
        }
    }
    context += "\nPlease incorporate these analytical perspectives in your analysis."

    return context
}
```

## Cost Optimization and Monitoring

### 1. Cost Optimizer

```go
// internal/llm/batch/cost_optimizer.go
type CostOptimizer struct {
    historicalCosts map[string]*CostMetrics
    modelPricing    map[string]*ModelPricing
    modelCapabilities map[string]*ModelCapabilities
}

type ModelCapabilities struct {
    SupportsTemperature bool     `json:"supports_temperature"`
    SupportsToolChoice  bool     `json:"supports_tool_choice"`
    SupportsTools       bool     `json:"supports_tools"`
    ModelClass          ModelClass `json:"model_class"`
    OptimalUseCases     []string `json:"optimal_use_cases"`
}

type CostEstimate struct {
    EstimatedCost     float64   `json:"estimated_cost"`
    SavingsVsSync     float64   `json:"savings_vs_sync"`
    TokenBreakdown    *TokenBreakdown `json:"token_breakdown"`
    CompletionTimeEst string    `json:"completion_time_estimate"`
    ModelMix          map[ModelClass]int `json:"model_mix"` // Track model class distribution
    ParameterWarnings []string  `json:"parameter_warnings,omitempty"`
}

func (co *CostOptimizer) OptimizeRequests(requests []BatchRequest) []BatchRequest {
    optimized := make([]BatchRequest, len(requests))

    for i, req := range requests {
        // Detect model class and optimize accordingly
        modelName := req.Body["model"].(string)
        modelClass := DetectModelClass(modelName)
        req.ModelClass = modelClass

        // Optimize token usage
        req.Body = co.optimizeTokenUsage(req.Body)

        // Select optimal model for task considering parameter support
        req.Body["model"] = co.selectOptimalModel(req)

        // Clean unsupported parameters for reasoning models
        req.Body = co.cleanUnsupportedParameters(req.Body, modelClass)

        optimized[i] = req
    }

    return optimized
}

func (co *CostOptimizer) cleanUnsupportedParameters(body map[string]interface{}, modelClass ModelClass) map[string]interface{} {
    if modelClass == ModelClassReasoning {
        // Remove unsupported parameters for gpt-5 class and o1-series models
        cleaned := make(map[string]interface{})
        for k, v := range body {
            switch k {
            case "temperature", "tool_choice":
                // Skip these parameters for reasoning models
                slog.Debug("Removing unsupported parameter for reasoning model", "parameter", k)
            case "tools":
                // Tools may need special handling for reasoning models
                if co.shouldIncludeToolsForReasoningModel(v) {
                    cleaned[k] = v
                } else {
                    slog.Debug("Removing tools parameter for reasoning model compatibility")
                }
            default:
                cleaned[k] = v
            }
        }
        return cleaned
    }
    return body
}

func (co *CostOptimizer) shouldIncludeToolsForReasoningModel(tools interface{}) bool {
    // Future: Determine if reasoning model supports tools
    // For now, assume they don't support traditional tools parameter
    return false
}

func (co *CostOptimizer) EstimateBatchCost(requests []BatchRequest) *CostEstimate {
    totalTokens := 0
    estimatedCost := 0.0
    modelMix := make(map[ModelClass]int)
    var warnings []string

    for _, req := range requests {
        tokens := co.estimateTokens(req)
        totalTokens += tokens

        modelName := req.Body["model"].(string)
        modelClass := DetectModelClass(modelName)
        modelMix[modelClass]++

        modelPrice := co.getModelPrice(modelName)
        estimatedCost += float64(tokens) * modelPrice * 0.5 // 50% batch discount

        // Check for parameter compatibility warnings
        if modelClass == ModelClassReasoning {
            if _, hasTemp := req.Body["temperature"]; hasTemp {
                warnings = append(warnings, fmt.Sprintf("Model %s doesn't support temperature parameter", modelName))
            }
            if _, hasToolChoice := req.Body["tool_choice"]; hasToolChoice {
                warnings = append(warnings, fmt.Sprintf("Model %s doesn't support tool_choice parameter", modelName))
            }
        }
    }

    syncCost := estimatedCost * 2 // Sync would be 2x more expensive

    return &CostEstimate{
        EstimatedCost: estimatedCost,
        SavingsVsSync: syncCost - estimatedCost,
        TokenBreakdown: &TokenBreakdown{
            TotalTokens: totalTokens,
            EstimatedOutputTokens: totalTokens / 4, // Rough estimate
        },
        CompletionTimeEst: co.estimateCompletionTime(modelMix),
        ModelMix:          modelMix,
        ParameterWarnings: warnings,
    }
}

func (co *CostOptimizer) estimateCompletionTime(modelMix map[ModelClass]int) string {
    hasReasoning := modelMix[ModelClassReasoning] > 0
    hasStandard := modelMix[ModelClassStandard] > 0

    switch {
    case hasReasoning && hasStandard:
        return "4-12 hours (mixed model classes)"
    case hasReasoning:
        return "6-16 hours (reasoning models take longer)"
    default:
        return "2-8 hours (standard models)"
    }
}
```

    syncCost := estimatedCost * 2 // Sync would be 2x more expensive

    return &CostEstimate{
        EstimatedCost: estimatedCost,
        SavingsVsSync: syncCost - estimatedCost,
        TokenBreakdown: &TokenBreakdown{
            TotalTokens: totalTokens,
            EstimatedOutputTokens: totalTokens / 4, // Rough estimate
        },
        CompletionTimeEst: "2-8 hours",
    }

}

````

### 2. Batch Monitor

```go
// internal/ai/openai/batch_monitor.go
type BatchMonitor struct {
    client      openai.Client
    storage     BatchStorage
    sharedBag   bag.SharedBag
    pollingInterval time.Duration
}

func (bm *BatchMonitor) StartMonitoring(ctx context.Context, jobID string) {
    go func() {
        ticker := time.NewTicker(bm.pollingInterval)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                if err := bm.checkBatchStatus(ctx, jobID); err != nil {
                    slog.Error("Failed to check batch status",
                        "job_id", jobID,
                        "error", err)
                }
            }
        }
    }()
}

func (bm *BatchMonitor) checkBatchStatus(ctx context.Context, jobID string) error {
    batch, err := bm.client.Batches.Retrieve(ctx, jobID)
    if err != nil {
        return fmt.Errorf("failed to retrieve batch: %w", err)
    }

    // Update job status in storage
    if err := bm.storage.UpdateJobStatus(ctx, jobID, BatchStatus(batch.Status)); err != nil {
        return fmt.Errorf("failed to update job status: %w", err)
    }

    // Update SharedBag with status
    bm.sharedBag.Set(fmt.Sprintf("batch_status_%s", jobID), map[string]interface{}{
        "status": batch.Status,
        "progress": batch.RequestCounts,
        "updated_at": time.Now(),
    })

    // If completed, trigger result processing
    if batch.Status == "completed" {
        if err := bm.triggerResultProcessing(ctx, jobID); err != nil {
            slog.Error("Failed to trigger result processing",
                "job_id", jobID,
                "error", err)
        }
    }

    return nil
}
````

## Integration with Existing Mosychlos Architecture

### 1. Enhanced Engine Integration with Prompt System

```go
// Enhanced engine support for batch processing with existing prompt integration
type BatchCapableEngine interface {
    models.Engine
    SubmitBatchAnalysis(ctx context.Context, portfolios []models.Portfolio) (*BatchJob, error)
    ProcessBatchResults(ctx context.Context, jobID string) error
    GetBatchProgress(ctx context.Context, jobID string) (*BatchProgress, error)
}

// internal/engine/risk/engine.go - Enhanced with batch support and prompt integration
func (e *Engine) SubmitBatchAnalysis(ctx context.Context, portfolios []models.Portfolio) (*BatchJob, error) {
    // Use existing prompt manager for template generation
    template := NewRiskAnalysisTemplate(e.promptManager) // Inject existing prompt manager

    requests := make([]BatchRequest, len(portfolios))
    for i, portfolio := range portfolios {
        customID := fmt.Sprintf("risk_analysis_%s_%d", portfolio.ID, i)

        // Use template with integrated prompt system
        request, err := template.BuildBatchRequest(ctx, portfolio, customID)
        if err != nil {
            return nil, fmt.Errorf("failed to build batch request for portfolio %s: %w", portfolio.ID, err)
        }

        requests[i] = request
    }

    return e.batchManager.SubmitBatch(ctx, requests, BatchOptions{
        CompletionWindow: "24h",
        Metadata: map[string]string{
            "engine": "risk_analysis",
            "analysis_type": "portfolio_risk",
            "prompt_version": e.promptManager.GetVersion(), // Track prompt versions
        },
        CostOptimize: true,
    })
}
```

### 2. Prompt System Dependencies Integration

```go
// internal/llm/batch/manager.go - Enhanced with prompt system integration
type Manager struct {
    provider       models.Provider
    storage        BatchStorage
    sharedBag      bag.SharedBag
    costOptimizer  *CostOptimizer
    monitor        *BatchMonitor
    promptManager  prompt.Manager         // Integrate existing prompt system
    regionalManager prompt.RegionalManager // Support regional variations
}

func NewBatchManager(
    provider models.Provider,
    storage BatchStorage,
    sharedBag bag.SharedBag,
    promptManager prompt.Manager,
    regionalManager prompt.RegionalManager,
) *Manager {
    return &Manager{
        provider:        provider,
        storage:         storage,
        sharedBag:       sharedBag,
        promptManager:   promptManager,   // Reuse existing prompt infrastructure
        regionalManager: regionalManager, // Leverage regional capabilities
        costOptimizer:   NewCostOptimizer(),
        monitor:         NewBatchMonitor(provider, storage, sharedBag),
    }
}

func (m *Manager) SubmitPortfolioAnalysisBatch(
    ctx context.Context,
    portfolios []models.Portfolio,
    analysisType models.AnalysisType,
    useRegionalPrompts bool,
) (*BatchJob, error) {
    var requests []BatchRequest
    var err error

    if useRegionalPrompts {
        // Use regional prompt system for international portfolios
        regionalTemplate := NewRegionalBatchTemplate(m.regionalManager, NewPortfolioAnalysisTemplate(m.promptManager, analysisType))
        requests, err = regionalTemplate.BuildRegionalBatchRequests(ctx, portfolios, m.sharedBag)
    } else {
        // Use standard prompt system
        template := NewPortfolioAnalysisTemplate(m.promptManager, analysisType)
        requests = make([]BatchRequest, len(portfolios))
        for i, portfolio := range portfolios {
            customID := fmt.Sprintf("%s_analysis_%s_%d", analysisType, portfolio.ID, i)
            requests[i], err = template.BuildBatchRequest(ctx, portfolio, customID)
            if err != nil {
                break
            }
        }
    }

    if err != nil {
        return nil, fmt.Errorf("failed to build batch requests: %w", err)
    }

    // Optimize for cost efficiency using existing templates
    if m.costOptimizer != nil {
        requests = m.costOptimizer.OptimizeRequests(requests)
    }

    // Submit batch to OpenAI with metadata about prompt usage
    return m.SubmitBatch(ctx, requests, BatchOptions{
        CompletionWindow: "24h",
        Metadata: map[string]string{
            "type": "portfolio_analysis",
            "analysis_type": string(analysisType),
            "portfolio_count": fmt.Sprintf("%d", len(portfolios)),
            "prompt_system": "integrated", // Track that we're using existing prompt system
            "regional_prompts": fmt.Sprintf("%t", useRegionalPrompts),
        },
        CostOptimize: true,
    })
}
```

### 2. CLI Integration with Prompt System and Model Selection

```go
// cmd/mosychlos/analyze.go - Enhanced with batch support, prompt integration, and model awareness
func init() {
    analyzeCmd.Flags().BoolVar(&batchMode, "batch", false, "Submit analysis as batch job for cost optimization")
    analyzeCmd.Flags().StringVar(&batchPriority, "batch-priority", "normal", "Batch job priority (low, normal, high)")
    analyzeCmd.Flags().BoolVar(&waitForResults, "wait", false, "Wait for batch completion and show results")
    analyzeCmd.Flags().BoolVar(&useRegionalPrompts, "regional", false, "Use regional prompt variations for international portfolios")
    analyzeCmd.Flags().StringVar(&promptTemplate, "template", "default", "Prompt template variant to use (default, aggressive, conservative)")
    analyzeCmd.Flags().StringVar(&modelName, "model", "gpt-4-turbo", "Model to use (gpt-4-turbo, gpt-5, gpt-5-mini, o1-preview)")
}

var analyzeCmd = &cobra.Command{
    Use:   "analyze [analysis-type]",
    Short: "Generate AI-powered portfolio analysis (supports batch mode with model selection)",
    Long: `Generate AI-powered portfolio analysis using the existing prompt management system.
Use --batch for cost-optimized processing of large portfolios or when immediate results are not required.
Regional variations and template customizations from internal/prompt are fully supported.
Model selection automatically handles parameter compatibility (GPT-5 class models don't support temperature/tool_choice).

Examples:
  mosychlos portfolio analyze risk --batch                          # Submit batch job with default model
  mosychlos portfolio analyze allocation --batch --model=gpt-5      # Use GPT-5 (no temperature/tool_choice)
  mosychlos portfolio analyze performance --batch --model=gpt-5-mini --regional
  mosychlos portfolio analyze compliance --batch --model=o1-preview --template=conservative --wait
  mosychlos portfolio batch status <job-id>                         # Check batch status
  mosychlos portfolio batch results <job-id>                        # Get batch results`,
    Run: runAnalyze,
}

func runAnalyze(cmd *cobra.Command, args []string) {
    if batchMode {
        // Validate model selection and warn about parameter limitations
        modelClass := DetectModelClass(modelName)
        if modelClass == ModelClassReasoning {
            fmt.Printf("Note: Model %s is a reasoning model with parameter limitations:\n", modelName)
            fmt.Printf("  - Temperature parameter not supported\n")
            fmt.Printf("  - Tool_choice parameter not supported\n")
            fmt.Printf("  - Tools may be embedded in prompts instead of using tools parameter\n\n")
        }

        // Initialize with existing prompt system
        promptManager, err := initializePromptManager()
        if err != nil {
            cobra.CheckErr(fmt.Errorf("failed to initialize prompt manager: %w", err))
        }

        var regionalManager prompt.RegionalManager
        if useRegionalPrompts {
            regionalManager, err = initializeRegionalManager()
            if err != nil {
                cobra.CheckErr(fmt.Errorf("failed to initialize regional manager: %w", err))
            }
        }

        job, err := submitBatchAnalysis(cmd.Context(), analysisType, promptManager, regionalManager, modelName)
        if err != nil {
            cobra.CheckErr(err)
        }

        fmt.Printf("Batch job submitted: %s\n", job.ID)
        fmt.Printf("Model: %s (class: %s)\n", modelName, string(modelClass))
        fmt.Printf("Using prompt system: %s\n", job.Metadata["prompt_system"])
        if useRegionalPrompts {
            fmt.Printf("Regional prompts: enabled\n")
        }
        fmt.Printf("Template variant: %s\n", promptTemplate)

        // Display cost estimate with model-specific information
        if job.CostEstimate != nil {
            fmt.Printf("Estimated completion: %s\n", job.CostEstimate.CompletionTimeEst)
            fmt.Printf("Estimated cost: $%.4f (%.1f%% savings vs sync)\n",
                job.CostEstimate.EstimatedCost,
                job.CostEstimate.SavingsVsSync/job.CostEstimate.EstimatedCost*100)

            // Show model mix information
            if len(job.CostEstimate.ModelMix) > 0 {
                fmt.Printf("Model distribution:\n")
                for class, count := range job.CostEstimate.ModelMix {
                    fmt.Printf("  - %s models: %d requests\n", class, count)
                }
            }

            // Show parameter warnings
            if len(job.CostEstimate.ParameterWarnings) > 0 {
                fmt.Printf("Parameter compatibility warnings:\n")
                for _, warning := range job.CostEstimate.ParameterWarnings {
                    fmt.Printf("  ⚠️  %s\n", warning)
                }
            }
        }

        if waitForResults {
            fmt.Printf("\nWaiting for batch completion...\n")
            results, err := waitForBatchCompletion(cmd.Context(), job.ID)
            if err != nil {
                cobra.CheckErr(err)
            }
            displayBatchResults(results)
        } else {
            fmt.Printf("\nUse 'mosychlos portfolio batch results %s' to get results when ready\n", job.ID)
        }
    } else {
        // Existing synchronous analysis logic (unchanged)
        runSynchronousAnalysis(cmd.Context(), analysisType)
    }
}

// Enhanced submission function with model awareness
func submitBatchAnalysis(ctx context.Context, analysisType string, promptManager prompt.Manager, regionalManager prompt.RegionalManager, modelName string) (*BatchJob, error) {
    // Detect model capabilities
    modelClass := DetectModelClass(modelName)

    // Create model-aware template
    template := NewPortfolioAnalysisTemplate(promptManager, analysisType, modelName)

    // Build requests based on model capabilities
    var requests []BatchRequest
    var err error

    portfolios, err := loadPortfoliosForAnalysis(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to load portfolios: %w", err)
    }

    if regionalManager != nil {
        // Use regional template with model awareness
        regionalTemplate := NewRegionalBatchTemplate(regionalManager, template, modelName)
        requests, err = regionalTemplate.BuildRegionalBatchRequests(ctx, portfolios, getSharedBag())
    } else {
        // Use standard template with model awareness
        requests = make([]BatchRequest, len(portfolios))
        for i, portfolio := range portfolios {
            customID := fmt.Sprintf("%s_analysis_%s_%d", analysisType, portfolio.ID, i)
            requests[i], err = template.BuildBatchRequest(ctx, portfolio, customID)
            if err != nil {
                break
            }
        }
    }

    if err != nil {
        return nil, fmt.Errorf("failed to build batch requests: %w", err)
    }

    // Submit with model class information
    return getBatchManager().SubmitBatch(ctx, requests, BatchOptions{
        CompletionWindow: "24h",
        ModelClass:       modelClass,
        Metadata: map[string]string{
            "type": "portfolio_analysis",
            "analysis_type": analysisType,
            "model": modelName,
            "model_class": string(modelClass),
            "portfolio_count": fmt.Sprintf("%d", len(portfolios)),
            "prompt_system": "integrated",
            "regional_prompts": fmt.Sprintf("%t", regionalManager != nil),
        },
        CostOptimize: true,
    })
}
```

## Performance and Scalability Considerations

### 1. Rate Limit Management with Model-Aware Optimization

- **Separate Rate Pool**: Batch API has separate rate limits from sync APIs
- **Higher Throughput**: 50,000 requests per batch with 200MB file size limit
- **Token Management**: Separate enqueued prompt token limits per model
- **Model-Specific Limits**: GPT-5 class models may have different rate limits and processing times

### 2. Cost Optimization Strategy with Model Selection

- **50% Cost Savings**: Automatic cost reduction vs synchronous API calls
- **Token Usage Optimization**: Smart request batching and prompt optimization
- **Model Selection**: Automated optimal model selection based on task requirements and parameter support
- **Parameter Efficiency**: Avoid unsupported parameters to prevent API errors and optimize processing
- **Processing Time Variance**: Reasoning models (GPT-5, o1-series) typically take 2-4x longer than standard models

### 3. Storage and Caching with Model Metadata

- **Result Persistence**: 30-day automatic retention of batch results
- **Local Caching**: Cache batch job metadata and status in SharedBag with model information
- **Database Integration**: Store job metadata in portfolio database with model class tracking
- **Model Compatibility Tracking**: Track which parameters worked with which models for future optimization

### 4. Error Handling and Reliability with Model Compatibility

- **Parameter Validation**: Pre-validate requests against model capabilities to prevent batch failures
- **Partial Completion**: Handle expired batches with partial results, especially for mixed model batches
- **Error File Processing**: Parse and report failed requests with model-specific error details
- **Retry Logic**: Automatic retry for failed batch submissions with parameter cleanup
- **Model Fallback**: Automatic fallback to compatible models when parameter conflicts are detected
- **Monitoring and Alerting**: Proactive monitoring of batch job health with model-specific metrics

## Migration Strategy

### Phase 1: Core LLM Infrastructure (Week 1-2)

1. Create new `internal/llm/` package structure
2. Implement batch interfaces and basic manager with prompt system integration
3. Add OpenAI Batch API client integration
4. Create basic monitoring and storage components
5. Migrate core functionality from `internal/ai/` to `internal/llm/`
6. Add CLI commands for batch job management

### Phase 2: Prompt System Integration (Week 3-4)

1. Integrate existing `internal/prompt/` system with batch templates
2. Add regional batch processing support using existing `RegionalManager`
3. Implement template variant support for different analysis approaches
4. Create cost optimization and estimation system leveraging prompt efficiency
5. Add result aggregation and processing with prompt versioning

### Phase 3: Engine Integration and Migration (Week 5-6)

1. Enhance existing engines with batch capabilities using integrated prompt system
2. Update SharedBag integration for batch results
3. Add batch-aware report generation with existing templates
4. Complete migration from `internal/ai/` to `internal/llm/`
5. Implement comprehensive testing suite with prompt validation

### Phase 4: Advanced Features (Week 7-8)

1. Add intelligent job scheduling and prioritization
2. Implement advanced cost optimization strategies with prompt optimization
3. Add batch result analytics and insights
4. Create monitoring dashboard and alerts
5. Remove deprecated `internal/ai/` package

## Integration Benefits with Existing Prompt System

### 1. **Template Consistency**

- Batch processing uses same prompt templates as synchronous processing
- Regional variations automatically available for batch jobs
- Investment research templates seamlessly integrated
- Template versioning and evolution tracking

### 2. **Operational Efficiency**

- No duplication of prompt management logic
- Consistent prompt behavior across sync and batch modes
- Regional compliance automatically handled through existing system
- Template dependency injection maintains clean architecture

### 3. **Cost and Quality Optimization**

- Existing prompt optimization applies to batch processing
- Regional prompt variations reduce unnecessary context
- Template efficiency improvements benefit both processing modes
- Consistent quality metrics across all processing types

### 4. **Maintenance and Evolution**

- Single source of truth for prompt management
- Template updates automatically benefit batch processing
- Regional compliance changes propagate to batch jobs
- Prompt system improvements enhance both sync and async processing

## Security and Compliance Considerations

### 1. Data Privacy

- **Secure File Upload**: Encrypt batch input files during upload
- **Data Retention**: Automatic cleanup of batch files after processing
- **Access Control**: Role-based access to batch jobs and results

### 2. Audit Trail

- **Job Tracking**: Complete audit trail of batch job lifecycle
- **Cost Tracking**: Detailed cost attribution and reporting
- **Compliance Reporting**: Batch processing audit logs for regulatory compliance

### 3. Rate Limiting and Quotas

- **Usage Monitoring**: Track batch API usage against quotas
- **Cost Controls**: Configurable cost limits and alerts
- **Resource Management**: Intelligent job scheduling to optimize resource usage

## Benefits Summary

### 1. **Cost Efficiency with Prompt Optimization**

- 50% reduction in AI processing costs for large-scale analysis
- Optimized model selection and token usage through existing prompt system
- Template efficiency improvements apply to both sync and batch processing
- Predictable cost structure for budgeting with prompt versioning

### 2. **Scalability with Consistent Quality**

- Process hundreds of portfolios simultaneously using proven prompt templates
- Higher rate limits compared to synchronous APIs
- Regional variations automatically supported through existing system
- Efficient resource utilization with prompt-based optimization

### 3. **Operational Excellence with Unified Management**

- Automated job lifecycle management with prompt tracking
- Comprehensive monitoring and alerting
- Robust error handling and recovery
- Single source of truth for prompt management across sync and batch modes

### 4. **Enhanced Capabilities with Existing Infrastructure**

- Support for complex, multi-step analysis workflows using existing templates
- Integration with existing Mosychlos architecture (SharedBag, engines, reports)
- Flexible template system leveraging existing investment research prompts
- Regional compliance and localization through existing prompt system

This batch processing architecture enables Mosychlos to scale AI-powered portfolio analysis while maintaining cost efficiency and operational reliability, leveraging the existing sophisticated prompt management system to ensure consistent quality and compliance across all processing modes. The integration with `internal/prompt/` provides a seamless transition from the current `internal/ai/` package to the new `internal/llm/` package with enhanced batch capabilities.
