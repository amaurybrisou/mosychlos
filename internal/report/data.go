package report

import (
	"fmt"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// extractCustomerData extracts customer-relevant data from the shared bag
func (g *Generator) extractCustomerData() (*models.CustomerReportData, error) {
	data := &models.CustomerReportData{
		GeneratedAt:  time.Now(),
		CustomerName: g.deps.Config.Report.DefaultCustomerName,
	}

	// Extract portfolio data
	if portfolioData, exists := g.deps.DataBag.Get(keys.KPortfolio); exists {
		if portfolio, ok := portfolioData.(*models.Portfolio); ok {
			data.Portfolio = portfolio
		}
	}

	// Extract risk metrics
	if riskData, exists := g.deps.DataBag.Get(keys.KPortfolioRiskMetrics); exists {
		data.RiskMetrics = riskData
	}

	// Extract allocation data
	if allocationData, exists := g.deps.DataBag.Get(keys.KPortfolioAllocationData); exists {
		data.AllocationData = allocationData
	}

	// Extract performance data
	if performanceData, exists := g.deps.DataBag.Get(keys.KPortfolioPerformanceData); exists {
		data.PerformanceData = performanceData
	}

	// Extract compliance data
	if complianceData, exists := g.deps.DataBag.Get(keys.KPortfolioComplianceData); exists {
		data.ComplianceData = complianceData
	}

	// Extract stock analysis
	if stockAnalysis, exists := g.deps.DataBag.Get(keys.KStockAnalysis); exists {
		data.StockAnalysis = stockAnalysis
	}

	// Extract insights
	if insights, exists := g.deps.DataBag.Get(keys.KAnalysisResults); exists {
		data.Insights = insights
	}

	// Extract news analysis
	if newsData, exists := g.deps.DataBag.Get(keys.KNewsAnalyzed); exists {
		data.NewsAnalyzed = newsData
	}

	// Extract fundamentals
	if fundamentalsData, exists := g.deps.DataBag.Get(keys.KFundamentals); exists {
		data.Fundamentals = fundamentalsData
	}

	// Extract investment research results and populate analysis fields
	if investmentResearch, exists := g.deps.DataBag.Get(keys.KInvestmentResearchResult); exists {
		if research, ok := investmentResearch.(models.InvestmentResearchResult); ok {
			// Populate detailed analysis fields from investment research
			if data.RiskMetrics == nil && len(research.RiskConsiderations) > 0 {
				riskContent := "### Risk Factors Analysis\n\n"
				for _, risk := range research.RiskConsiderations {
					riskContent += fmt.Sprintf("- **%s Risk** (%s severity): %s\n", risk.Type, risk.Severity, risk.Impact)
					if risk.Impact != "" {
						riskContent += fmt.Sprintf("  - Impact: %s\n", risk.Impact)
					}
					if risk.Mitigation != "" {
						riskContent += fmt.Sprintf("  - Mitigation: %s\n", risk.Mitigation)
					}
				}
				data.RiskMetrics = riskContent
			}

			if data.AllocationData == nil && len(research.ResearchFindings) > 0 {
				allocationContent := "### Investment Opportunities\n\n"
				for _, finding := range research.ResearchFindings {
					allocationContent += fmt.Sprintf("- **%s** (%s)\n", finding.Title, finding.AssetClass)
					if len(finding.SpecificInstruments) > 0 {
						allocationContent += "  - Instruments:"
						for _, instrument := range finding.SpecificInstruments {
							if instrument.Ticker != "" {
								allocationContent += fmt.Sprintf(" %s", instrument.Ticker)
							} else {
								allocationContent += fmt.Sprintf(" %s", instrument.Name)
							}
						}
						allocationContent += "\n"
					}
					if finding.ExpectedReturn.BaseCase > 0 {
						allocationContent += fmt.Sprintf("  - Expected Return: %.1f%% (%s)\n", finding.ExpectedReturn.BaseCase*100, finding.ExpectedReturn.Confidence)
					}
				}
				data.AllocationData = allocationContent
			}

			if data.PerformanceData == nil && research.MarketAnalysis.OverallSentiment != "" {
				performanceContent := "### Market Performance Analysis\n\n"
				performanceContent += fmt.Sprintf("**Market Sentiment:** %s\n\n", research.MarketAnalysis.OverallSentiment)
				if research.ExecutiveSummary.TimeHorizon != "" {
					performanceContent += fmt.Sprintf("**Time Horizon:** %s\n\n", research.ExecutiveSummary.TimeHorizon)
				}
				if len(research.ExecutiveSummary.KeyTakeaways) > 0 {
					performanceContent += "**Key Market Drivers:**\n"
					for _, takeaway := range research.ExecutiveSummary.KeyTakeaways {
						performanceContent += fmt.Sprintf("- %s\n", takeaway)
					}
				}
				data.PerformanceData = performanceContent
			}

			if data.ComplianceData == nil && len(research.RegionalContext.TaxOptimizations) > 0 {
				complianceContent := "### Regional Tax & Compliance Considerations\n\n"
				complianceContent += fmt.Sprintf("**Location:** %s (%s)\n\n", research.RegionalContext.Country, research.RegionalContext.CurrencyFocus)
				for _, tax := range research.RegionalContext.TaxOptimizations {
					complianceContent += fmt.Sprintf("- **%s**: %s\n", tax.AccountType, tax.Strategy)
					if tax.Implementation != "" {
						complianceContent += fmt.Sprintf("  - Implementation: %s\n", tax.Implementation)
					}
				}
				data.ComplianceData = complianceContent
			}
		}
	}

	return data, nil
}

// extractSystemData extracts system diagnostic data from the shared bag
func (g *Generator) extractSystemData() (*models.SystemReportData, error) {
	data := &models.SystemReportData{
		GeneratedAt: time.Now(),
	}

	// Get application health from health monitor
	if appHealth, exists := g.deps.DataBag.Get(keys.KApplicationHealth); exists {
		if health, ok := appHealth.(models.ApplicationHealth); ok {
			data.ApplicationHealth = health
		}
	}

	// Extract tool metrics - check for the actual key being used
	if toolMetrics, exists := g.deps.DataBag.Get(keys.KToolMetrics); exists {
		if metrics, ok := toolMetrics.(models.ToolMetrics); ok {
			data.ToolMetrics = &metrics
		}
	}

	// Extract tool computations - for activity log display only
	if toolComputations, exists := g.deps.DataBag.Get(keys.KToolComputations); exists {
		if computations, ok := toolComputations.([]models.ToolComputation); ok {
			data.ToolComputations = computations
			// Note: We don't recalculate metrics here as that would override
			// the session-specific metrics from KToolMetrics with cumulative totals
		}
	}

	// Extract cache stats with enhanced map structure support
	if cacheStats, exists := g.deps.DataBag.Get(keys.KCacheStats); exists {
		if statsMap, ok := cacheStats.(models.CacheStatsMap); ok {
			// Use the aggregated stats for system-level reporting
			data.CacheStats = &statsMap.Aggregated
		} else if stats, ok := cacheStats.(models.CacheHealthStatus); ok {
			// Fallback for old structure
			data.CacheStats = &stats
		} else if stats, ok := cacheStats.(*models.CacheHealthStatus); ok {
			// Fallback for pointer structure
			data.CacheStats = stats
		}
	}

	// Extract external data health
	if externalHealth, exists := g.deps.DataBag.Get(keys.KExternalDataHealth); exists {
		if health, ok := externalHealth.(models.ExternalDataHealth); ok {
			data.ExternalDataHealth = &health
		} else if health, ok := externalHealth.(*models.ExternalDataHealth); ok {
			data.ExternalDataHealth = health
		}
	}

	// Extract market data freshness
	if dataFreshness, exists := g.deps.DataBag.Get(keys.KMarketDataFreshness); exists {
		if freshness, ok := dataFreshness.(models.MarketDataFreshness); ok {
			data.MarketDataFreshness = &freshness
		} else if freshness, ok := dataFreshness.(*models.MarketDataFreshness); ok {
			data.MarketDataFreshness = freshness
		}
	}

	return data, nil
}

// getDataSourcesUsed determines which data sources were used in the report
func (g *Generator) getDataSourcesUsed(customerData *models.CustomerReportData, systemData *models.SystemReportData) []string {
	sources := make(map[string]bool)

	if customerData != nil {
		if customerData.Portfolio != nil {
			sources["portfolio"] = true
		}
		if customerData.RiskMetrics != nil {
			sources["risk_analysis"] = true
		}
		if customerData.NewsAnalyzed != nil {
			sources["news_data"] = true
		}
		if customerData.Fundamentals != nil {
			sources["fundamentals"] = true
		}
	}

	if systemData != nil {
		if systemData.ToolMetrics != nil {
			sources["tool_metrics"] = true
		}
		if systemData.CacheStats != nil {
			sources["cache_metrics"] = true
		}
		if systemData.ExternalDataHealth != nil {
			sources["external_apis"] = true
		}
		if systemData.MarketDataFreshness != nil {
			sources["market_data"] = true
		}
	}

	// Convert map to slice
	result := make([]string, 0, len(sources))
	for source := range sources {
		result = append(result, source)
	}

	return result
}
