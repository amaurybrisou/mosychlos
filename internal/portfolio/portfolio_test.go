package portfolio

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/fs"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

func TestNewService(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{DataDir: "/tmp/test"}
	filesystem := fs.OS{}
	validator := NewBasicValidator()
	sharedBag := bag.NewSharedBag()

	service := NewService(cfg, filesystem, sharedBag, validator)
	assert.NotNil(t, service)
}

func TestBasicValidator_Validate(t *testing.T) {
	t.Parallel()

	validator := NewBasicValidator()
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
			name: "empty AsOf",
			portfolio: &models.Portfolio{
				AsOf: "",
			},
			wantErr: true,
		},
		{
			name: "invalid AsOf format",
			portfolio: &models.Portfolio{
				AsOf: "invalid-date",
			},
			wantErr: true,
		},
		{
			name: "no accounts",
			portfolio: &models.Portfolio{
				AsOf:     "2025-01-01",
				Accounts: []models.Account{},
			},
			wantErr: true,
		},
		{
			name: "valid portfolio",
			portfolio: &models.Portfolio{
				AsOf: "2025-01-01",
				Accounts: []models.Account{
					{
						Name:     "Test Account",
						Type:     models.AccountBrokerage,
						Currency: "USD",
						Holdings: []models.Holding{
							{
								Ticker:    "AAPL",
								Quantity:  10,
								CostBasis: 150.0,
								Currency:  "USD",
								Type:      models.Stock,
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := validator.Validate(ctx, c.portfolio)

			if c.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.True(t, c.portfolio.Validated)
			}
		})
	}
}

func TestService_GetPortfolio(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{DataDir: "/tmp/test-portfolio"}
	filesystem := fs.OS{}
	validator := NewBasicValidator()
	sharedBag := bag.NewSharedBag()

	service := NewService(cfg, filesystem, sharedBag, validator)
	ctx := context.Background()

	// create a mock fetcher
	mockFetcher := &mockFetcher{
		portfolio: &models.Portfolio{
			AsOf: "2025-08-12",
			Accounts: []models.Account{
				{
					Name:     "Test",
					Type:     models.AccountBrokerage,
					Currency: "USD",
				},
			},
		},
	}

	portfolio, err := service.GetPortfolio(ctx, mockFetcher)
	require.NoError(t, err)
	assert.NotNil(t, portfolio)
	assert.True(t, portfolio.Validated)
	assert.Equal(t, "2025-08-12", portfolio.AsOf)
}

// mockFetcher implements the Fetcher interface for testing
type mockFetcher struct {
	portfolio *models.Portfolio
	err       error
}

func (m *mockFetcher) Fetch(ctx context.Context) (*models.Portfolio, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.portfolio, nil
}
