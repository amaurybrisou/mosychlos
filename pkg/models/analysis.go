package models

// AnalysisType defines the types of portfolio analysis available
type AnalysisType string

const (
	AnalysisRisk               AnalysisType = "risk"
	AnalysisAllocation         AnalysisType = "allocation"
	AnalysisPerformance        AnalysisType = "performance"
	AnalysisCompliance         AnalysisType = "compliance"
	AnalysisReallocation       AnalysisType = "reallocation"
	AnalysisInvestmentResearch AnalysisType = "investment_research" // New regional analysis type
)
