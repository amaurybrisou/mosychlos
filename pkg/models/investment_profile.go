package models

// InvestmentProfile contains advanced investment preferences and context
type InvestmentProfile struct {
	// Investment Style and Approach
	InvestmentStyle string   `json:"investment_style,omitempty" yaml:"investment_style,omitempty"` // "growth", "value", "blend", "income", "balanced"
	ResearchDepth   string   `json:"research_depth,omitempty" yaml:"research_depth,omitempty"`     // "basic", "intermediate", "detailed", "comprehensive", "advanced"
	RiskTolerance   string   `json:"risk_tolerance,omitempty" yaml:"risk_tolerance,omitempty"`     // "conservative", "moderate", "aggressive"
	AssetClasses    []string `json:"asset_classes,omitempty" yaml:"asset_classes,omitempty"`       // Preferred asset classes (legacy)
	PreferredAssets []string `json:"preferred_assets,omitempty" yaml:"preferred_assets,omitempty"` // Preferred asset types
	AvoidedSectors  []string `json:"avoided_sectors,omitempty" yaml:"avoided_sectors,omitempty"`   // Sectors to avoid

	// TimeHorizonYears
	TimeHorizonYears int `json:"time_horizon_years,omitempty" yaml:"time_horizon_years,omitempty"`

	// ESG Criteria
	ESGCriteria ESGCriteria `json:"esg_criteria,omitempty" yaml:"esg_criteria,omitempty"`

	// Regional Investment Context
	RegionalContext RegionalInvestmentContext `json:"regional_context,omitempty" yaml:"regional_context,omitempty"`

	// Analysis Preferences
	PreferredAnalysisTypes []AnalysisType `json:"preferred_analysis_types,omitempty" yaml:"preferred_analysis_types,omitempty"` // Preferred analysis types
	CustomRequirements     []string       `json:"custom_requirements,omitempty" yaml:"custom_requirements,omitempty"`           // User-specific requirements

	// Metadata
	ProfileVersion string `json:"profile_version,omitempty" yaml:"profile_version,omitempty"` // Profile format version
	Source         string `json:"source,omitempty" yaml:"source,omitempty"`                   // Source of profile (default_regional, default_global, custom)
}

// RegionalInvestmentContext contains region-specific investment context
type RegionalInvestmentContext struct {
	// Basic localization fields (embedded inline from LocalizationConfig)
	Country  string `json:"country,omitempty" yaml:"country,omitempty"`
	Language string `json:"language,omitempty" yaml:"language,omitempty"`
	Currency string `json:"currency,omitempty" yaml:"currency,omitempty"`
	Timezone string `json:"timezone,omitempty" yaml:"timezone,omitempty"`
	Region   string `json:"region,omitempty" yaml:"region,omitempty"`
	City     string `json:"city,omitempty" yaml:"city,omitempty"`

	// Date formatting preference
	DateFormat string `json:"date_format,omitempty" yaml:"date_format,omitempty"`

	// Investment-specific regional context
	TaxOptimizedAccounts []string `json:"tax_optimized_accounts,omitempty" yaml:"tax_optimized_accounts,omitempty"`
	LocalPriorities      []string `json:"local_priorities,omitempty" yaml:"local_priorities,omitempty"`
	RegulatoryFocus      []string `json:"regulatory_focus,omitempty" yaml:"regulatory_focus,omitempty"`
	PreferLocalChampions bool     `json:"prefer_local_champions,omitempty" yaml:"prefer_local_champions,omitempty"`
}

// RegionalInvestmentPreferences contains region-specific investment context (deprecated, use RegionalInvestmentContext)
type RegionalInvestmentPreferences struct {
	LocalizationConfig    `yaml:",inline"` // Your foundation
	PreferredAssetClasses []string         `json:"preferred_asset_classes,omitempty" yaml:"preferred_asset_classes,omitempty"` // Region-specific preferences
	ExcludedSectors       []string         `json:"excluded_sectors,omitempty" yaml:"excluded_sectors,omitempty"`               // Regional exclusions
	ESGPreferences        ESGCriteria      `json:"esg_preferences,omitempty" yaml:"esg_preferences,omitempty"`                 // Regional ESG standards
	ComplianceRules       []string         `json:"compliance_rules,omitempty" yaml:"compliance_rules,omitempty"`               // Regional compliance requirements
}

// ESGCriteria defines Environmental, Social, and Governance preferences
type ESGCriteria struct {
	ESGImportance     string   `json:"esg_importance,omitempty" yaml:"esg_importance,omitempty"`         // "low", "moderate", "important", "high"
	ESGFocus          []string `json:"esg_focus,omitempty" yaml:"esg_focus,omitempty"`                   // Focus areas: "environmental", "social", "governance"
	ExclusionCriteria []string `json:"exclusion_criteria,omitempty" yaml:"exclusion_criteria,omitempty"` // ESG-based exclusions

	// Legacy fields (deprecated but kept for compatibility)
	Environmental []string `json:"environmental,omitempty" yaml:"environmental,omitempty"` // Environmental criteria
	Social        []string `json:"social,omitempty" yaml:"social,omitempty"`               // Social responsibility criteria
	Governance    []string `json:"governance,omitempty" yaml:"governance,omitempty"`       // Governance standards
	Exclusions    []string `json:"exclusions,omitempty" yaml:"exclusions,omitempty"`       // ESG-based exclusions
}

// Validate validates the investment profile
func (ip *InvestmentProfile) Validate() error {
	// Validate regional context if provided
	if ip.RegionalContext.Country != "" {
		// Create a LocalizationConfig for validation
		lc := LocalizationConfig{
			Country:  ip.RegionalContext.Country,
			Language: ip.RegionalContext.Language,
			Currency: ip.RegionalContext.Currency,
			Timezone: ip.RegionalContext.Timezone,
			Region:   ip.RegionalContext.Region,
			City:     ip.RegionalContext.City,
		}
		return lc.Validate()
	}
	return nil
}
