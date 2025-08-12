// Package adapters provides conversion services between external APIs and internal models
package adapters

import (
	"context"
	"time"

	"github.com/amaurybrisou/mosychlos/internal/portfolio"
	"github.com/amaurybrisou/mosychlos/pkg/binance"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// BinanceFetcher adapts Binance portfolio provider to portfolio.Fetcher interface
type BinanceFetcher struct {
	provider binance.PortfolioProvider
}

// NewBinanceFetcher creates a new Binance portfolio fetcher
func NewBinanceFetcher(provider binance.PortfolioProvider) portfolio.Fetcher {
	return &BinanceFetcher{
		provider: provider,
	}
}

// Fetch retrieves portfolio data from Binance and converts it to Portfolio model
func (f *BinanceFetcher) Fetch(ctx context.Context) (*models.Portfolio, error) {
	// Get spot portfolio data from Binance
	spotData, err := f.provider.GetSpotPortfolio(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to Portfolio model
	portfolio := &models.Portfolio{
		AsOf:     time.Now().Format(time.RFC3339),
		Accounts: make([]models.Account, 0),
	}

	// Create a single account for Binance spot data
	account := models.Account{
		Name:     "Binance Spot",
		Type:     "spot",
		Holdings: make([]models.Holding, 0),
	}

	// Convert Binance balances to holdings
	for asset, quantity := range spotData.Balances {
		if quantity > 0 {
			// Get price for this asset (skip if not available/delisted)
			price, hasPriceData := spotData.Prices[asset]
			if !hasPriceData {
				// skip delisted or unpriceable assets
				continue
			}

			holding := models.Holding{
				Ticker:    asset,
				Type:      models.Crypto,
				Quantity:  quantity,
				CostBasis: price,  // Use current price as cost basis since we don't have historical data
				Currency:  "USDT", // Base pricing currency for crypto assets
				Name:      asset,
				Sector:    "Cryptocurrency",
				Region:    "Global",
			}
			account.Holdings = append(account.Holdings, holding)
		}
	}

	portfolio.Accounts = append(portfolio.Accounts, account)

	return portfolio, nil
}
