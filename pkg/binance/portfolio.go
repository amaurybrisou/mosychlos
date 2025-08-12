package binance

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// portfolioProvider implements the PortfolioProvider interface
type portfolioProvider struct {
	client Client
}

// NewPortfolioProvider creates a new portfolio provider with the given configuration
func NewPortfolioProvider(cfg *config.BinanceConfig) PortfolioProvider {
	return &portfolioProvider{
		client: New(cfg),
	}
}

// GetSpotPortfolio retrieves and processes spot portfolio data
func (p *portfolioProvider) GetSpotPortfolio(ctx context.Context) (*models.BinancePortfolioData, error) {
	// get account information with balances
	accountInfo, err := p.client.GetAccountInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get account info: %w", err)
	}

	// process balances - only include non-zero balances
	balances := make(map[string]float64)
	var assets []string

	for _, balance := range accountInfo.Balances {
		free, err := strconv.ParseFloat(balance.Free, 64)
		if err != nil {
			continue
		}

		locked, err := strconv.ParseFloat(balance.Locked, 64)
		if err != nil {
			continue
		}

		totalBalance := free + locked
		if totalBalance > 0 {
			balances[balance.Asset] = totalBalance
			assets = append(assets, balance.Asset)
		}
	}

	// get current prices for all assets (convert to USDT pairs)
	prices := make(map[string]float64)
	values := make(map[string]float64)
	var totalValue float64

	// handle USDT and other stablecoins separately
	stablecoins := map[string]bool{
		"USDT": true,
		"USDC": true,
		"BUSD": true,
		"TUSD": true,
		"USDP": true,
	}

	for _, asset := range assets {
		var price float64

		if stablecoins[asset] {
			price = 1.0
		} else {
			// try to get price from USDT pair
			symbol := asset + "USDT"
			priceData, err := p.client.GetPrice(ctx, symbol)
			if err != nil {
				// if USDT pair doesn't exist, try BUSD
				symbol = asset + "BUSD"
				priceData, err = p.client.GetPrice(ctx, symbol)
				if err != nil {
					// skip assets we can't price
					continue
				}
			}
			price = priceData.Price
		}

		prices[asset] = price
		assetValue := balances[asset] * price
		values[asset] = assetValue
		totalValue += assetValue
	}

	return &models.BinancePortfolioData{
		AccountType: accountInfo.AccountType,
		UpdateTime:  time.Unix(accountInfo.UpdateTime/1000, 0),
		Balances:    balances,
		Prices:      prices,
		Values:      values,
		TotalValue:  totalValue,
		Permissions: accountInfo.Permissions,
	}, nil
}
