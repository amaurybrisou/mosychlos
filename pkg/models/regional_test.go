package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegionalConfig_Validate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		config  RegionalConfig
		wantErr bool
	}{
		{
			name: "valid french config",
			config: RegionalConfig{
				LocalizationConfig: LocalizationConfig{
					Country:  "FR",
					Language: "fr",
					Currency: "EUR",
					Timezone: "Europe/Paris",
				},
				Strings: map[string]string{
					"pea_description": "Plan d'Épargne en Actions",
				},
				MarketContext: RegionalMarketContext{
					PrimaryExchanges: []string{"Euronext Paris"},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid localization config",
			config: RegionalConfig{
				LocalizationConfig: LocalizationConfig{
					Country:  "", // Invalid - empty country
					Language: "fr",
					Currency: "EUR",
					Timezone: "Europe/Paris",
				},
				Strings: map[string]string{
					"pea_description": "Plan d'Épargne en Actions",
				},
			},
			wantErr: true,
		},
		{
			name: "empty regional data is valid",
			config: RegionalConfig{
				LocalizationConfig: LocalizationConfig{
					Country:  "US",
					Language: "en",
					Currency: "USD",
					Timezone: "America/New_York",
				},
				Strings:       map[string]string{},     // Empty is OK now
				MarketContext: RegionalMarketContext{}, // Empty is OK now
			},
			wantErr: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := c.config.Validate()

			if c.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRegionalMarketContext_FieldAssignment(t *testing.T) {
	t.Parallel()

	ctx := RegionalMarketContext{
		RegulatoryFocus:   "SEC compliance",
		InvestmentCulture: "Growth-oriented",
		PreferredThemes:   []string{"tech", "healthcare"},
		PrimaryExchanges:  []string{"NYSE", "NASDAQ"},
		MajorIndices:      []string{"S&P 500", "NASDAQ Composite"},
	}

	assert.Equal(t, "SEC compliance", ctx.RegulatoryFocus)
	assert.Equal(t, "Growth-oriented", ctx.InvestmentCulture)
	assert.Len(t, ctx.PreferredThemes, 2)
	assert.Contains(t, ctx.PreferredThemes, "tech")
	assert.Len(t, ctx.PrimaryExchanges, 2)
	assert.Contains(t, ctx.PrimaryExchanges, "NYSE")
	assert.Len(t, ctx.MajorIndices, 2)
	assert.Contains(t, ctx.MajorIndices, "S&P 500")
}

func TestRegionalTaxContext_FieldAssignment(t *testing.T) {
	t.Parallel()

	ctx := RegionalTaxContext{
		PrimaryAccounts:        []string{"401(k)", "IRA"},
		OptimizationStrategies: []string{"tax-loss harvesting", "asset location"},
	}

	assert.Len(t, ctx.PrimaryAccounts, 2)
	assert.Contains(t, ctx.PrimaryAccounts, "401(k)")
	assert.Len(t, ctx.OptimizationStrategies, 2)
	assert.Contains(t, ctx.OptimizationStrategies, "tax-loss harvesting")
}

func TestRegionalOverlay_FieldAssignment(t *testing.T) {
	t.Parallel()

	overlay := RegionalOverlay{
		TemplateAdditions: "French investment context...",
		LocalizationData: map[string]interface{}{
			"currency": "EUR",
			"country":  "FR",
		},
	}

	assert.Equal(t, "French investment context...", overlay.TemplateAdditions)
	assert.Len(t, overlay.LocalizationData, 2)
	assert.Equal(t, "EUR", overlay.LocalizationData["currency"])
	assert.Equal(t, "FR", overlay.LocalizationData["country"])
}
