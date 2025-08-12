package prompt

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"log/slog"
	"path/filepath"
	"text/template"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/fs"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"gopkg.in/yaml.v3"
)

// regionalManager implements RegionalManager interface
type regionalManager struct {
	baseManager  models.PromptBuilder                 // Existing manager (composition)
	fs           fs.FS                                // Filesystem abstraction
	configDir    string                               // Regional config directory
	overlayCache map[string]*template.Template        // Template cache
	configCache  map[string]*models.RegionalConfig    // Config cache
	profileCache map[string]*models.InvestmentProfile // Investment profile cache
}

var _ RegionalManager = (*regionalManager)(nil)

// NewRegionalManager creates regional manager that composes existing manager
func NewRegionalManager(
	baseManager models.PromptBuilder,
	filesystem fs.FS,
	configDir string,
) RegionalManager {
	return &regionalManager{
		baseManager:  baseManager, // Compose, don't replace
		fs:           filesystem,
		configDir:    configDir,
		overlayCache: make(map[string]*template.Template),
		configCache:  make(map[string]*models.RegionalConfig),
		profileCache: make(map[string]*models.InvestmentProfile),
	}
}

// GenerateRegionalPrompt generates a prompt with regional context
func (rm *regionalManager) GenerateRegionalPrompt(
	ctx context.Context,
	analysisType models.AnalysisType,
	sharedBag bag.SharedBag,
	data PromptData,
) (string, error) {
	// key keys.KRegionalConfig
	regionalConfig := &models.RegionalConfig{}
	ok := sharedBag.GetAs(keys.KRegionalConfig, regionalConfig)
	if !ok {
		return "", fmt.Errorf("invalid regional config")
	}

	slog.Debug("Generating regional prompt",
		"analysis_type", analysisType,
		"country", regionalConfig.Country,
		"language", regionalConfig.Language,
	)

	// 1. Load regional overlay template
	overlay, err := rm.loadRegionalOverlay(analysisType, regionalConfig.Country, regionalConfig.Language)
	if err != nil {
		slog.Warn("Failed to load regional overlay, using base template",
			"error", err,
			"country", regionalConfig.Country,
			"language", regionalConfig.Language,
		)
		// Continue without overlay - graceful fallback
		overlay = ""
	}

	if overlay != "" {
		// Execute the overlay template with the regional config to substitute variables
		overlayTemplate, err := template.New("overlay").Parse(overlay)
		if err != nil {
			slog.Warn("Failed to parse overlay template, using as-is", "error", err)
			data.RegionalOverlay = &models.RegionalOverlay{
				TemplateAdditions: overlay,
				LocalizationData:  regionalConfig.Data,
			}
		} else {
			var buf bytes.Buffer
			if err := overlayTemplate.Execute(&buf, data); err != nil {
				slog.Warn("Failed to execute overlay template, using as-is", "error", err)
				data.RegionalOverlay = &models.RegionalOverlay{
					TemplateAdditions: overlay,
					LocalizationData:  regionalConfig.Data,
				}
			} else {
				data.RegionalOverlay = &models.RegionalOverlay{
					TemplateAdditions: buf.String(),
					LocalizationData:  regionalConfig.Data,
				}
			}
		}
	}

	// 4. For now, return regional prompt directly (later delegate to base manager)
	// TODO: Delegate to base manager once it supports GeneratePrompt with PromptData
	return rm.generateRegionalPromptDirect(ctx, data)
}

// generateRegionalPromptDirect generates prompt using actual template files
func (rm *regionalManager) generateRegionalPromptDirect(_ context.Context, data PromptData) (string, error) {
	// Load the base research template from config/templates/
	templatePath := filepath.Join(rm.configDir, "templates", "investment_research", "base", "research.tmpl")
	templateContent, err := rm.fs.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to load base template: %w", err)
	}

	// Load component templates
	componentsDir := filepath.Join(rm.configDir, "templates", "investment_research", "base", "components")
	componentFiles := []string{"context.tmpl", "portfolio_analysis.tmpl", "research_framework.tmpl", "output_format.tmpl"}

	tmpl := template.New("investment_research")

	// Load all component templates first
	for _, componentFile := range componentFiles {
		componentPath := filepath.Join(componentsDir, componentFile)
		componentContent, err := rm.fs.ReadFile(componentPath)
		if err != nil {
			slog.Warn("Failed to load component template", "component", componentFile, "error", err)
			continue
		}

		// Parse each component as a named template
		componentName := strings.TrimSuffix(componentFile, ".tmpl")
		_, err = tmpl.
			New(componentName).
			Funcs(rm.tmplFuncMap()).
			Option("missingkey=zero").
			Parse(string(componentContent))
		if err != nil {
			slog.Warn("Failed to parse component template", "component", componentFile, "error", err)
			continue
		}
	}

	// Parse the main template
	tmpl, err = tmpl.Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse base template: %w", err)
	}

	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// LoadInvestmentProfile loads an investment profile for the given country and risk tolerance
func (rm *regionalManager) LoadInvestmentProfile(
	country, riskTolerance string,
) (*models.InvestmentProfile, error) {
	profileKey := fmt.Sprintf("%s_%s", country, riskTolerance)

	// Check cache first
	if cached, exists := rm.profileCache[profileKey]; exists {
		return cached, nil
	}

	// Load from filesystem using same patterns as regional config
	profilePath := filepath.Join(rm.configDir, "investment_profiles", "defaults", country, riskTolerance+".yaml")

	profileData, err := rm.fs.ReadFile(profilePath)
	if err != nil {
		// Fallback to global default
		return rm.loadGlobalDefaultProfile(riskTolerance)
	}

	var profile models.InvestmentProfile
	if err := yaml.Unmarshal(profileData, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse investment profile: %w", err)
	}

	// Cache the result using existing cache patterns
	rm.profileCache[profileKey] = &profile

	return &profile, nil
}

// loadGlobalDefaultProfile loads a fallback global profile
func (rm *regionalManager) loadGlobalDefaultProfile(riskTolerance string) (*models.InvestmentProfile, error) {
	globalKey := fmt.Sprintf("global_%s", riskTolerance)

	// Check cache first
	if cached, exists := rm.profileCache[globalKey]; exists {
		return cached, nil
	}

	profilePath := filepath.Join(rm.configDir, "investment_profiles", "defaults", "global", riskTolerance+".yaml")

	profileData, err := rm.fs.ReadFile(profilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load global default profile for %s: %w", riskTolerance, err)
	}

	var profile models.InvestmentProfile
	if err := yaml.Unmarshal(profileData, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse global investment profile: %w", err)
	}

	// Cache the result
	rm.profileCache[globalKey] = &profile

	return &profile, nil
}
