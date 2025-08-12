package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvestmentProfile_Validate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		profile InvestmentProfile
		wantErr bool
	}{
		{
			name: "valid profile with regional preferences",
			profile: InvestmentProfile{
				InvestmentStyle: "growth",
				ResearchDepth:   "comprehensive",
				RegionalContext: RegionalInvestmentContext{
					Country:  "US",
					Language: "en",
					Currency: "USD",
					Timezone: "America/New_York",
				},
			},
			wantErr: false,
		},
		{
			name: "valid profile without regional preferences",
			profile: InvestmentProfile{
				InvestmentStyle: "value",
				ResearchDepth:   "basic",
				AssetClasses:    []string{"stocks", "bonds"},
			},
			wantErr: false,
		},
		{
			name:    "empty profile is valid",
			profile: InvestmentProfile{},
			wantErr: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := c.profile.Validate()

			if c.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestESGCriteria_FieldAssignment(t *testing.T) {
	t.Parallel()

	esg := ESGCriteria{
		Environmental: []string{"carbon neutral", "renewable energy"},
		Social:        []string{"diversity", "labor practices"},
		Governance:    []string{"board independence", "executive compensation"},
		Exclusions:    []string{"tobacco", "weapons"},
	}

	assert.Len(t, esg.Environmental, 2)
	assert.Contains(t, esg.Environmental, "carbon neutral")
	assert.Len(t, esg.Social, 2)
	assert.Contains(t, esg.Social, "diversity")
	assert.Len(t, esg.Governance, 2)
	assert.Contains(t, esg.Governance, "board independence")
	assert.Len(t, esg.Exclusions, 2)
	assert.Contains(t, esg.Exclusions, "tobacco")
}

func TestRegionalInvestmentPreferences_FieldAssignment(t *testing.T) {
	t.Parallel()

	prefs := RegionalInvestmentPreferences{
		LocalizationConfig: LocalizationConfig{
			Country:  "FR",
			Language: "fr",
			Currency: "EUR",
			Timezone: "Europe/Paris",
		},
		PreferredAssetClasses: []string{"European equities", "Euro bonds"},
		ExcludedSectors:       []string{"tobacco", "weapons"},
		ESGPreferences: ESGCriteria{
			Environmental: []string{"green energy"},
		},
		ComplianceRules: []string{"UCITS compliance", "MiFID II"},
	}

	assert.Equal(t, "FR", prefs.Country)
	assert.Equal(t, "fr", prefs.Language)
	assert.Len(t, prefs.PreferredAssetClasses, 2)
	assert.Contains(t, prefs.PreferredAssetClasses, "European equities")
	assert.Len(t, prefs.ExcludedSectors, 2)
	assert.Contains(t, prefs.ExcludedSectors, "tobacco")
	assert.Len(t, prefs.ESGPreferences.Environmental, 1)
	assert.Contains(t, prefs.ESGPreferences.Environmental, "green energy")
	assert.Len(t, prefs.ComplianceRules, 2)
	assert.Contains(t, prefs.ComplianceRules, "UCITS compliance")
}
