package models

import (
	"fmt"
	"strings"
)

// LocalizationConfig holds centralized locale and location settings
type LocalizationConfig struct {
	// Country is the ISO 3166-1 alpha-2 country code (e.g., "US", "GB", "FR")
	Country string `mapstructure:"country" yaml:"country"`
	// Language is the ISO 639-1 language code (e.g., "en", "fr", "de")
	Language string `mapstructure:"language" yaml:"language"`
	// Timezone is the IANA timezone identifier (e.g., "America/New_York", "Europe/London")
	Timezone string `mapstructure:"timezone" yaml:"timezone"`
	// Currency is the ISO 4217 currency code (e.g., "USD", "EUR", "GBP")
	Currency string `mapstructure:"currency" yaml:"currency"`
	// Region is optional state/province/region (e.g., "California", "England")
	Region string `mapstructure:"region" yaml:"region"`
	// City is optional city name (e.g., "San Francisco", "London")
	City string `mapstructure:"city" yaml:"city"`
}

// Validate validates the localization configuration with strict requirements
func (lc *LocalizationConfig) Validate() error {
	// Country is required and must be valid ISO 3166-1 alpha-2
	if strings.TrimSpace(lc.Country) == "" {
		return fmt.Errorf("country cannot be empty")
	}
	if len(lc.Country) != 2 {
		return fmt.Errorf("country must be ISO 3166-1 alpha-2 code (2 characters), got: %s", lc.Country)
	}

	// Language is required and must be valid ISO 639-1
	if strings.TrimSpace(lc.Language) == "" {
		return fmt.Errorf("language cannot be empty")
	}
	if len(lc.Language) != 2 {
		return fmt.Errorf("language must be ISO 639-1 code (2 characters), got: %s", lc.Language)
	}

	// Timezone is required and must be valid IANA timezone
	if strings.TrimSpace(lc.Timezone) == "" {
		return fmt.Errorf("timezone cannot be empty")
	}
	// Basic format check for IANA timezone (should contain at least one slash)
	if !strings.Contains(lc.Timezone, "/") {
		return fmt.Errorf("timezone must be valid IANA timezone (e.g., 'America/New_York'), got: %s", lc.Timezone)
	}

	// Currency is required and must be valid ISO 4217
	if strings.TrimSpace(lc.Currency) == "" {
		return fmt.Errorf("currency cannot be empty")
	}
	if len(lc.Currency) != 3 {
		return fmt.Errorf("currency must be ISO 4217 code (3 characters), got: %s", lc.Currency)
	}

	// Convert to uppercase for consistency
	lc.Country = strings.ToUpper(lc.Country)
	lc.Language = strings.ToLower(lc.Language)
	lc.Currency = strings.ToUpper(lc.Currency)

	return nil
}
