package report

import (
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// getDataSourcesUsed analyzes the data to determine which data sources were utilized
func (g *Generator) getDataSourcesUsed(customerData *models.CustomerReportData, systemData *models.SystemReportData) []string {
	var sources []string

	// Check customer data sources
	if customerData != nil {
		if customerData.Portfolio != nil {
			sources = append(sources, "Portfolio Data")
		}
		if customerData.RiskMetrics != nil {
			sources = append(sources, "Risk Analysis")
		}
		if customerData.NewsAnalyzed != nil {
			sources = append(sources, "News Analysis")
		}
		if customerData.Fundamentals != nil {
			sources = append(sources, "Fundamental Analysis")
		}
		if customerData.StockAnalysis != nil {
			sources = append(sources, "Stock Analysis")
		}
		if customerData.AllocationData != nil {
			sources = append(sources, "Allocation Analysis")
		}
		if customerData.PerformanceData != nil {
			sources = append(sources, "Performance Analysis")
		}
		if customerData.ComplianceData != nil {
			sources = append(sources, "Compliance Analysis")
		}
	}

	// Check system data sources
	if systemData != nil {
		if systemData.ApplicationHealth.Status == "healthy" {
			sources = append(sources, "System Health")
		}
		if systemData.ToolMetrics != nil {
			sources = append(sources, "Tool Metrics")
		}
		if systemData.CacheStats != nil {
			sources = append(sources, "Cache Statistics")
		}
		if systemData.ExternalDataHealth != nil {
			sources = append(sources, "External Data Health")
		}
		if systemData.MarketDataFreshness != nil {
			sources = append(sources, "Market Data Quality")
		}
	}

	return sources
}
