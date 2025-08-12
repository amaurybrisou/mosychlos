package models

import (
	"fmt"
	"strings"
)

// CountryPolicy represents investment policy for a specific country
type CountryPolicy struct {
	Allowed    []string `json:"allowed" yaml:"allowed"`
	Optional   []string `json:"optional" yaml:"optional"`
	Restricted []string `json:"restricted" yaml:"restricted"`
}

// CountryConfig represents complete country configuration with policy
type CountryConfig struct {
	Country string        `json:"country" yaml:"country"`
	Policy  CountryPolicy `json:"policy" yaml:"policy"`
}

// PolicyDocument represents the complete policy document structure
type PolicyDocument struct {
	Countries []CountryConfig `json:"countries" yaml:"countries"`
}

// AssetRestriction represents compliance status for an asset type
type AssetRestriction struct {
	AssetType    string   `json:"asset_type"`
	IsAllowed    bool     `json:"is_allowed"`
	IsRestricted bool     `json:"is_restricted"`
	Notes        []string `json:"notes,omitempty"`
}

// ComplianceRules represents asset compliance rules for a jurisdiction
type ComplianceRules struct {
	AllowedAssetTypes    []string          `yaml:"allowed_asset_types" mapstructure:"allowed_asset_types"`
	DisallowedAssetTypes []string          `yaml:"disallowed_asset_types" mapstructure:"disallowed_asset_types"`
	ETFDomicileAllow     []string          `yaml:"etf_domicile_allow" mapstructure:"etf_domicile_allow"`
	ETFDomicileBlock     []string          `yaml:"etf_domicile_block" mapstructure:"etf_domicile_block"`
	TickerBlocklist      []string          `yaml:"ticker_blocklist" mapstructure:"ticker_blocklist"`
	TickerSubstitutes    map[string]string `yaml:"ticker_substitutes" mapstructure:"ticker_substitutes"`
	MaxLeverage          int               `yaml:"max_leverage" mapstructure:"max_leverage"`
	Notes                string            `yaml:"notes" mapstructure:"notes"`
}

// Validate validates the compliance rules
func (cr *ComplianceRules) Validate() error {
	// validate max leverage
	if cr.MaxLeverage < 0 {
		return fmt.Errorf("MaxLeverage cannot be negative, got: %d", cr.MaxLeverage)
	}

	// validate asset types are not empty strings
	for i, assetType := range cr.AllowedAssetTypes {
		if strings.TrimSpace(assetType) == "" {
			return fmt.Errorf("AllowedAssetTypes[%d] cannot be empty", i)
		}
	}

	for i, assetType := range cr.DisallowedAssetTypes {
		if strings.TrimSpace(assetType) == "" {
			return fmt.Errorf("DisallowedAssetTypes[%d] cannot be empty", i)
		}
	}

	// check for conflicts between allowed and disallowed
	allowedSet := make(map[string]bool)
	for _, assetType := range cr.AllowedAssetTypes {
		allowedSet[strings.ToLower(strings.TrimSpace(assetType))] = true
	}

	for _, assetType := range cr.DisallowedAssetTypes {
		normalized := strings.ToLower(strings.TrimSpace(assetType))
		if allowedSet[normalized] {
			return fmt.Errorf("asset type '%s' cannot be both allowed and disallowed", assetType)
		}
	}

	// validate ticker substitutes
	for originalTicker, substituteTicker := range cr.TickerSubstitutes {
		if strings.TrimSpace(originalTicker) == "" {
			return fmt.Errorf("ticker substitute key cannot be empty")
		}
		if strings.TrimSpace(substituteTicker) == "" {
			return fmt.Errorf("ticker substitute value for '%s' cannot be empty", originalTicker)
		}
	}

	return nil
}

// CountryCode represents ISO 3166-1 alpha-2 country codes
type CountryCode string

const (
	CountryFR CountryCode = "FR"
	CountryUS CountryCode = "US"
)
