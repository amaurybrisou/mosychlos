---
applyTo: '**/*.go'
---

---

## applyTo: '\*_/_.go'

# Go Development Instructions for Mosychlos Portfolio Management

## **Project Context**:

- **Primary**: Portfolio management CLI tool built in Go with Cobra command-line interface
- **Integration**: External financial analysis services for institutional-grade financial analysis
- **Architecture**: Go CLI with tool-driven multi-engine analysis pipeline
- **Output**: Professional reports, charts, and comprehensive financial analysis (not basic text)

## **Engine Integration Patterns**:

### **Engine Implementation for Tool-Driven Services**

```go
// Implement tool-driven engines that leverage institutional-grade analysis:
type FinancialAnalysisEngine struct {
    toolClient *tools.Client
    cache      *cache.Cache
    sharedBag  *bag.SharedBag
}

// Methods should return structured professional analysis:
func (e *FinancialAnalysisEngine) AnalyzePortfolio(ctx context.Context, portfolio *models.Portfolio) (*models.InstitutionalAnalysis, error) {
    // Tool-driven analysis pipeline
    // Professional behavior through tool constraints
    // Structured analysis output
    return analysis, nil
}
```

### **Tool Integration with Analysis Services**

```go
// Tools should leverage external capabilities, not basic prompting:
type FinancialAnalysisTool struct {
    analysisService string // External service endpoint
    sharedBag       *bag.SharedBag
}

func (t *FinancialAnalysisTool) Execute(ctx context.Context, input map[string]any) (*models.ToolResult, error) {
    // Get portfolio from shared bag
    portfolio := t.sharedBag.Get("portfolio")

    // Call analysis service (not basic chat)
    analysis, err := t.callAnalysisService(portfolio)
    if err != nil {
        return nil, fmt.Errorf("analysis service call failed: %w", err)
    }

    return &models.ToolResult{
        Status: "completed",
        Data: map[string]any{
            "analysis_result": analysis.Result,
            "risk_assessment": analysis.RiskMetrics,
            "recommendations": analysis.Recommendations,
            "report_path":     analysis.PDFReportPath,
            "charts":          analysis.ChartPaths,
        },
    }, nil
}
```

## **Coding Guidelines**:

- **External Integration**: Use HTTP client to call external analysis service endpoints
- **Professional Output**: Handle PDF reports, charts, and structured analysis (not just text)
- **Error Handling**: Comprehensive error handling for external service failures and timeouts
- **Shared Context**: Use SharedBag for portfolio context access in tools
- **Caching**: Implement caching for expensive analysis calls
- **Logging**: Include service calls, response times, and analysis quality in logs

# **Refer to these various dedicated instructions files**:

- [build](build.md)
- [core](core.rules.md)
- [linting](linting.md)
- [logging](logging.md)
- [run](run.md)
- [testing](testing.md)
- [validation](validation.md)
- [Cobra CLI Guidelines](https://github.com/spf13/cobra/blob/master/docs/_index.md)
- [Go Modules](https://blog.golang.org/using-go-modules)
- [Effective Go](https://golang.org/doc/effective_go.html)
