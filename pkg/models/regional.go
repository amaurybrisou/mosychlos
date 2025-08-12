package models

// RegionalOverlay holds template additions and localization data
type RegionalOverlay struct {
	TemplateAdditions string         `json:"template_additions"` // Regional template content
	LocalizationData  map[string]any `json:"localization_data"`  // YAML data injection
}

// RegionalConfig extends LocalizationConfig with investment-specific regional context
type RegionalConfig struct {
	LocalizationConfig // Embed your solid foundation (Country, Language, Currency, etc.)

	// Regional investment-specific extensions only
	Strings       map[string]string     `json:"strings"`        // Localized strings
	MarketContext RegionalMarketContext `json:"market_context"` // Market-specific data
	TaxContext    RegionalTaxContext    `json:"tax_context"`    // Tax optimization context
	Data          map[string]any        `json:"data"`           // Raw YAML data
}

// RegionalMarketContext holds region-specific market and investment information
type RegionalMarketContext struct {
	RegulatoryFocus   string   `json:"regulatory_focus"`   // Primary regulatory framework
	InvestmentCulture string   `json:"investment_culture"` // Regional investment approach
	PreferredThemes   []string `json:"preferred_themes"`   // Popular investment themes
	PrimaryExchanges  []string `json:"primary_exchanges"`  // Main stock exchanges
	MajorIndices      []string `json:"major_indices"`      // Key market indices
}

// RegionalTaxContext contains region-specific tax optimization information
type RegionalTaxContext struct {
	PrimaryAccounts        []string `json:"primary_accounts"`        // Tax-advantaged account types
	OptimizationStrategies []string `json:"optimization_strategies"` // Regional tax optimization approaches
}

// Validate validates the regional configuration using LocalizationConfig as foundation
func (rc *RegionalConfig) Validate() error {
	// Delegate core validation to your robust LocalizationConfig
	return rc.LocalizationConfig.Validate()
}
