package localization

import (
	"fmt"
	"path/filepath"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/fs"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"gopkg.in/yaml.v3"
)

type LocalizationService interface {
	LoadRegionalConfig(bag bag.SharedBag, country, language string) (*models.RegionalConfig, error)
}

type localizationService struct {
	fs        fs.FS
	configDir string // Directory for localization configs
}

func New(fs fs.FS, configDir string) LocalizationService {
	return &localizationService{
		fs:        fs,
		configDir: configDir,
	}
}

// LoadRegionalConfig loads regional configuration with fallback strategy
func (rm *localizationService) LoadRegionalConfig(sharedBag bag.SharedBag, country, language string) (*models.RegionalConfig, error) {
	// Check cache first
	regionalConfig := &models.RegionalConfig{}
	if ok := sharedBag.GetAs(bag.KRegionalConfig, regionalConfig); ok {
		return regionalConfig, nil
	}

	// Load regional configuration with fallback
	configPath := filepath.Join(rm.configDir, "templates", "investment_research", "regional",
		"localization", fmt.Sprintf("%s_%s.yaml", country, language))

	configData, err := rm.fs.ReadFile(configPath)
	if err != nil {
		// Fallback to country-only config
		fallbackPath := filepath.Join(rm.configDir, "templates", "investment_research", "regional",
			"localization", fmt.Sprintf("%s_en.yaml", country))
		if configData, err = rm.fs.ReadFile(fallbackPath); err != nil {
			// Return minimal default config
			return nil, fmt.Errorf("failed to load fallback regional config: %w", err)
		}
	}

	var yamlData map[string]interface{}
	if err := yaml.Unmarshal(configData, &yamlData); err != nil {
		return nil, fmt.Errorf("failed to parse regional config: %w", err)
	}

	// Convert to RegionalConfig struct
	config := &models.RegionalConfig{
		LocalizationConfig: models.LocalizationConfig{
			Country:  country,
			Language: language,
		},
		Data: yamlData,
	}

	// Parse structured sections
	if strings, ok := yamlData[bag.KStrings.String()].(map[string]interface{}); ok {
		config.Strings = make(map[string]string)
		for k, v := range strings {
			if str, ok := v.(string); ok {
				config.Strings[k] = str
			}
		}
	}

	// Parse market context
	if marketCtx, ok := yamlData[bag.KMarketContext.String()].(map[string]interface{}); ok {
		config.MarketContext = parseMarketContext(marketCtx)
	}

	// Parse tax context
	if taxCtx, ok := yamlData[bag.KTaxContext.String()].(map[string]interface{}); ok {
		config.TaxContext = parseTaxContext(taxCtx)
	}

	// Cache the result
	sharedBag.Set(bag.KRegionalConfig, config)

	return config, nil
}

// parseMarketContext converts YAML data to RegionalMarketContext
func parseMarketContext(data map[string]interface{}) models.RegionalMarketContext {
	ctx := models.RegionalMarketContext{}

	if val, ok := data[bag.KRegulatoryFocus.String()].(string); ok {
		ctx.RegulatoryFocus = val
	}

	if val, ok := data[bag.KInvestmentCulture.String()].(string); ok {
		ctx.InvestmentCulture = val
	}

	if themes, ok := data[bag.KPreferredThemes.String()].([]interface{}); ok {
		for _, theme := range themes {
			if str, ok := theme.(string); ok {
				ctx.PreferredThemes = append(ctx.PreferredThemes, str)
			}
		}
	}

	if exchanges, ok := data[bag.KPrimaryExchanges.String()].([]interface{}); ok {
		for _, exchange := range exchanges {
			if str, ok := exchange.(string); ok {
				ctx.PrimaryExchanges = append(ctx.PrimaryExchanges, str)
			}
		}
	}

	if indices, ok := data[bag.KMajorIndices.String()].([]interface{}); ok {
		for _, index := range indices {
			if str, ok := index.(string); ok {
				ctx.MajorIndices = append(ctx.MajorIndices, str)
			}
		}
	}

	return ctx
}

// parseTaxContext converts YAML data to RegionalTaxContext
func parseTaxContext(data map[string]interface{}) models.RegionalTaxContext {
	ctx := models.RegionalTaxContext{}

	if accounts, ok := data[bag.KPrimaryAccounts.String()].([]interface{}); ok {
		for _, account := range accounts {
			if str, ok := account.(string); ok {
				ctx.PrimaryAccounts = append(ctx.PrimaryAccounts, str)
			}
		}
	}

	if strategies, ok := data[bag.KOptimizationStrategies.String()].([]interface{}); ok {
		for _, strategy := range strategies {
			if str, ok := strategy.(string); ok {
				ctx.OptimizationStrategies = append(ctx.OptimizationStrategies, str)
			}
		}
	}

	return ctx
}
