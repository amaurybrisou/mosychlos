package jurisdiction

import (
	"context"
	"testing"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Validate(t *testing.T) {
	t.Parallel()

	cfg := config.JurisdictionConfig{
		Country: "FR",
		Rules: models.ComplianceRules{
			AllowedAssetTypes: []string{"stock", "etf", "bond_gov", "cash"},
			MaxLeverage:       1,
		},
	}
	service, err := New(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	cases := []struct {
		name      string
		portfolio *models.Portfolio
		wantErr   bool
	}{
		{
			name:      "nil portfolio",
			portfolio: nil,
			wantErr:   true,
		},
		{
			name: "valid portfolio with allowed assets",
			portfolio: &models.Portfolio{
				AsOf: "2025-08-12",
				Accounts: []models.Account{
					{
						Name:     "Test",
						Type:     models.AccountBrokerage,
						Currency: "USD",
						Holdings: []models.Holding{
							{
								Ticker:    "AAPL",
								Type:      "stock",
								Quantity:  10,
								CostBasis: 150.0,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "portfolio with disallowed assets",
			portfolio: &models.Portfolio{
				AsOf: "2025-08-12",
				Accounts: []models.Account{
					{
						Name:     "Test",
						Type:     models.AccountBrokerage,
						Currency: "USD",
						Holdings: []models.Holding{
							{
								Ticker:    "BTC",
								Type:      "crypto",
								Quantity:  1,
								CostBasis: 50000.0,
							},
						},
					},
				},
			},
			wantErr: true, // crypto not in allowed list
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := service.Validate(ctx, c.portfolio)
			if c.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	// test with default config
	cfg := config.JurisdictionConfig{
		Country: "FR",
		Rules: models.ComplianceRules{
			AllowedAssetTypes: []string{"stock", "etf"},
			MaxLeverage:       1,
		},
	}
	service, err := New(cfg)
	require.NoError(t, err)
	assert.NotNil(t, service)

	// test with custom config
	cfg2 := config.JurisdictionConfig{
		Country: "US",
		Rules: models.ComplianceRules{
			AllowedAssetTypes: []string{"stock", "bond_gov"},
			MaxLeverage:       2,
		},
	}
	service2, err := New(cfg2)
	require.NoError(t, err)
	assert.NotNil(t, service2)
}
