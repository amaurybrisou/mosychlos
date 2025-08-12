package binance

import (
	"context"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

//go:generate mockgen -source=interface.go -destination=mocks/interface_mock.go -package=mocks

// Client defines the interface for interacting with Binance API
type Client interface {
	// GetAccountInfo retrieves spot account information including balances
	GetAccountInfo(ctx context.Context) (*models.BinanceAccountInfo, error)

	// GetTicker24hr gets 24hr ticker price change statistics for a symbol
	GetTicker24hr(ctx context.Context, symbol string) (*models.BinanceTicker, error)

	// GetAllTickers24hr gets 24hr ticker price change statistics for all symbols
	GetAllTickers24hr(ctx context.Context) ([]*models.BinanceTicker, error)

	// GetKlines gets candlestick/kline data for a symbol
	GetKlines(ctx context.Context, symbol string, interval string, limit int, startTime, endTime *time.Time) ([]*models.BinanceKline, error)

	// GetPrice gets current price for a symbol
	GetPrice(ctx context.Context, symbol string) (*models.BinancePriceData, error)

	// GetPrices gets current prices for multiple symbols
	GetPrices(ctx context.Context, symbols []string) ([]*models.BinancePriceData, error)
}

// PortfolioProvider defines interface for fetching processed portfolio data
type PortfolioProvider interface {
	// GetSpotPortfolio retrieves and processes spot portfolio data
	GetSpotPortfolio(ctx context.Context) (*models.BinancePortfolioData, error)
}
