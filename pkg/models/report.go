package models

import (
	"context"
	"time"
)

//go:generate mockgen -source=report.go -destination=mocks/report_mock.go -package=mocks

// ReportFormat represents the output format for reports
type ReportFormat string

const (
	FormatMarkdown ReportFormat = "markdown"
	FormatPDF      ReportFormat = "pdf"
	FormatJSON     ReportFormat = "json"
	FormatAll      ReportFormat = "all"
)

// ReportType represents the type of report to generate
type ReportType string

const (
	TypeCustomer ReportType = "customer"
	TypeSystem   ReportType = "system"
	TypeFull     ReportType = "full"
)

// CustomerReportData contains data for customer-facing reports
type CustomerReportData struct {
	Portfolio       any       `json:"portfolio,omitempty"`
	RiskMetrics     any       `json:"risk_metrics,omitempty"`
	AllocationData  any       `json:"allocation_data,omitempty"`
	PerformanceData any       `json:"performance_data,omitempty"`
	ComplianceData  any       `json:"compliance_data,omitempty"`
	StockAnalysis   any       `json:"stock_analysis,omitempty"`
	Insights        any       `json:"insights,omitempty"`
	NewsAnalyzed    any       `json:"news_analyzed,omitempty"`
	Fundamentals    any       `json:"fundamentals,omitempty"`
	Recommendations any       `json:"recommendations,omitempty"`
	GeneratedAt     time.Time `json:"generated_at"`
	CustomerName    string    `json:"customer_name,omitempty"`
}

// SystemReportData contains data for system diagnostic reports
type SystemReportData struct {
	ApplicationHealth   ApplicationHealth    `json:"application_health"`
	ToolMetrics         *ToolMetrics         `json:"tool_metrics,omitempty"`
	CacheStats          *CacheHealthStatus   `json:"cache_stats,omitempty"`
	ExternalDataHealth  *ExternalDataHealth  `json:"external_data_health,omitempty"`
	MarketDataFreshness *MarketDataFreshness `json:"market_data_freshness,omitempty"`
	ToolComputations    []ToolComputation    `json:"tool_computations,omitempty"`
	GeneratedAt         time.Time            `json:"generated_at"`
	BatchMode           bool                 `json:"batch_mode,omitempty"`
}

// ReportGenerator interface for generating different types of reports
type ReportGenerator interface {
	GenerateCustomerReport(ctx context.Context, format ReportFormat) (*ReportOutput, error)
	GenerateSystemReport(ctx context.Context, format ReportFormat) (*ReportOutput, error)
	GenerateFullReport(ctx context.Context, format ReportFormat) (*ReportOutput, error)
}

// ReportOutput contains the generated report data
type ReportOutput struct {
	Type        ReportType   `json:"type"`
	Format      ReportFormat `json:"format"`
	Content     string       `json:"content"`
	FilePath    string       `json:"file_path,omitempty"`
	GeneratedAt time.Time    `json:"generated_at"`
	Metadata    ReportMeta   `json:"metadata"`
}

// FullReportData combines customer and system data for comprehensive reports
type FullReportData struct {
	Customer *CustomerReportData `json:"customer_data"`
	System   *SystemReportData   `json:"system_data"`
}

// ReportMeta contains metadata about the report
type ReportMeta struct {
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	DataSources      []string          `json:"data_sources"`
	GenerationTimeMs int64             `json:"generation_time_ms"`
	Version          string            `json:"version"`
	CustomFields     map[string]string `json:"custom_fields,omitempty"`
	CustomerName     string            `json:"customer_name,omitempty"`
}

// ReportRequest represents a request to generate a report
type ReportRequest struct {
	Type           ReportType   `json:"type"`
	Format         ReportFormat `json:"format"`
	CustomerName   string       `json:"customer_name,omitempty"`
	OutputPath     string       `json:"output_path,omitempty"`
	IncludeSummary bool         `json:"include_summary"`
	CustomTitle    string       `json:"custom_title,omitempty"`
}
