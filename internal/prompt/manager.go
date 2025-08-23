// Package prompt provides a simple interface for managing AI prompts with dependency injection.
// It focuses on portfolio analysis using normalized data from the shared bag.
package prompt

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"text/template"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

//go:embed templates/portfolio/*.tmpl
var templateFS embed.FS

// manager implements the Manager interface
type manager struct {
	deps      Dependencies
	templates map[models.AnalysisType]*template.Template
}

// NewManager creates a new prompt manager with the given dependencies
func NewManager(deps Dependencies) (models.PromptBuilder, error) {
	m := &manager{
		deps:      deps,
		templates: make(map[models.AnalysisType]*template.Template),
	}

	// Load and parse templates
	if err := m.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return m, nil
}

// BuildPrompt creates a prompt for the specified analysis type
func (m *manager) BuildPrompt(ctx context.Context, analysisType models.AnalysisType) (string, error) {
	tmpl, exists := m.templates[analysisType]
	if !exists {
		return "", fmt.Errorf("unknown analysis type: %s", analysisType)
	}

	// Gather normalized data from shared bag
	data, err := m.gatherPromptData(ctx, analysisType)
	if err != nil {
		return "", fmt.Errorf("failed to gather prompt data: %w", err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// gatherPromptData collects all necessary data from the shared bag and config
func (m *manager) gatherPromptData(_ context.Context, analysisType models.AnalysisType) (*PromptData, error) {
	data := &PromptData{
		Localization:      m.deps.Config.UserLocalization,
		UserProfile:       m.deps.Config.UserProfile,
		InvestmentProfile: m.deps.Config.UserProfile,
		Timestamp:         time.Now(),
		AnalysisType:      analysisType,
		Context:           make(map[string]any),
	}

	// Get portfolio data if available
	if portfolioData, exists := m.deps.Bag.Get(bag.KPortfolioNormalizedForAI); exists {
		if portfolio, ok := portfolioData.(*models.NormalizedPortfolio); ok {
			data.Portfolio = portfolio
		}
	}

	if profileData, exists := m.deps.Bag.Get(bag.KProfile); exists {
		if profile, ok := profileData.(*models.InvestmentProfile); ok {
			data.UserProfile = profile
		}
	}

	// Get market data if available
	if marketData, exists := m.deps.Bag.Get(bag.KMarketDataNormalized); exists {
		if market, ok := marketData.(*models.NormalizedMarketData); ok {
			data.MarketData = market
		}
	}

	// Get macro data if available
	if macroData, exists := m.deps.Bag.Get(bag.KMacroDataNormalized); exists {
		if macro, ok := macroData.(models.NormalizedMacroData); ok {
			data.MacroData = &macro
		}
	}

	// Add analysis-specific context based on type
	switch analysisType {
	case models.AnalysisRisk:
		data.Context["focus_areas"] = []string{"concentration", "diversification", "volatility", "correlation"}
	case models.AnalysisAllocation:
		data.Context["focus_areas"] = []string{"asset_allocation", "geographic_distribution", "sector_balance", "rebalancing"}
	case models.AnalysisPerformance:
		data.Context["focus_areas"] = []string{"returns", "benchmarking", "risk_adjusted_metrics", "attribution"}
	case models.AnalysisCompliance:
		data.Context["focus_areas"] = []string{"position_limits", "concentration_limits", "asset_restrictions", "regulatory_compliance"}
	}

	return data, nil
}
