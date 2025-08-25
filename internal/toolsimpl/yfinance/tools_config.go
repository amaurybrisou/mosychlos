// Package yfinance
package yfinance

import (
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/internal/tools"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// GetYFinanceStockDataToolConfig returns the YFinance stock data tool configuration.
func GetYFinanceStockDataToolConfig(_cfg any, sharedBag bag.SharedBag) models.ToolConfig {
	cfg := _cfg.(*config.YFinanceConfig)
	yfinanceRateLimit := &models.ToolsRateLimit{
		RequestsPerSecond: 2,
		RequestsPerDay:    2000,
		Burst:             5,
	}
	return models.ToolConfig{
		Key: bag.YFinanceStockData,
		Constructor: func() (models.Tool, error) {
			return newStockDataFromConfig(cfg, sharedBag)
		},
		Config:       cfg,
		CacheEnabled: cfg.CacheEnable,
		CacheTTL:     6 * time.Hour,
		RateLimit:    yfinanceRateLimit,
		Persisting:   cfg.Persisting,
	}
}

// GetYFinanceStockInfoToolConfig returns the YFinance stock info tool configuration.
func GetYFinanceStockInfoToolConfig(_cfg any, sharedBag bag.SharedBag) models.ToolConfig {
	cfg := _cfg.(*config.YFinanceConfig)
	yfinanceRateLimit := &models.ToolsRateLimit{
		RequestsPerSecond: 2,
		RequestsPerDay:    2000,
		Burst:             5,
	}
	return models.ToolConfig{
		Key: bag.YFinanceStockInfo,
		Constructor: func() (models.Tool, error) {
			return newStockInfoFromConfig(cfg, sharedBag)
		},
		Config:       cfg,
		CacheEnabled: cfg.CacheEnable,
		CacheTTL:     24 * time.Hour,
		RateLimit:    yfinanceRateLimit,
		Persisting:   cfg.Persisting,
	}
}

// GetYFinanceDividendsToolConfig returns the YFinance dividends tool configuration.
func GetYFinanceDividendsToolConfig(_cfg any, sharedBag bag.SharedBag) models.ToolConfig {
	cfg := _cfg.(*config.YFinanceConfig)
	yfinanceRateLimit := &models.ToolsRateLimit{
		RequestsPerSecond: 2,
		RequestsPerDay:    2000,
		Burst:             5,
	}
	return models.ToolConfig{
		Key: bag.YFinanceDividends,
		Constructor: func() (models.Tool, error) {
			return newDividendsFromConfig(cfg, sharedBag)
		},
		Config:       cfg,
		CacheEnabled: cfg.CacheEnable,
		CacheTTL:     24 * time.Hour,
		RateLimit:    yfinanceRateLimit,
		Persisting:   cfg.Persisting,
	}
}

// GetYFinanceFinancialsToolConfig returns the YFinance financials tool configuration.
func GetYFinanceFinancialsToolConfig(_cfg any, sharedBag bag.SharedBag) models.ToolConfig {
	cfg := _cfg.(*config.YFinanceConfig)
	yfinanceRateLimit := &models.ToolsRateLimit{
		RequestsPerSecond: 2,
		RequestsPerDay:    2000,
		Burst:             5,
	}
	return models.ToolConfig{
		Key: bag.YFinanceFinancials,
		Constructor: func() (models.Tool, error) {
			return newFinancialsFromConfig(cfg, sharedBag)
		},
		Config:       cfg,
		CacheEnabled: cfg.CacheEnable,
		CacheTTL:     24 * time.Hour,
		RateLimit:    yfinanceRateLimit,
		Persisting:   cfg.Persisting,
	}
}

// GetYFinanceMarketDataToolConfig returns the YFinance market data tool configuration.
func GetYFinanceMarketDataToolConfig(_cfg any, sharedBag bag.SharedBag) models.ToolConfig {
	cfg := _cfg.(*config.YFinanceConfig)
	yfinanceRateLimit := &models.ToolsRateLimit{
		RequestsPerSecond: 2,
		RequestsPerDay:    2000,
		Burst:             5,
	}
	return models.ToolConfig{
		Key: bag.YFinanceMarketData,
		Constructor: func() (models.Tool, error) {
			return newMarketDataFromConfig(cfg, sharedBag)
		},
		Config:       cfg,
		CacheEnabled: cfg.CacheEnable,
		CacheTTL:     6 * time.Hour,
		RateLimit:    yfinanceRateLimit,
		Persisting:   cfg.Persisting,
	}
}

func init() {
	tools.Register(bag.YFinanceDividends, GetYFinanceDividendsToolConfig)
	tools.Register(bag.YFinanceFinancials, GetYFinanceFinancialsToolConfig)
	tools.Register(bag.YFinanceMarketData, GetYFinanceMarketDataToolConfig)
	tools.Register(bag.YFinanceStockData, GetYFinanceStockDataToolConfig)
	tools.Register(bag.YFinanceStockInfo, GetYFinanceStockInfoToolConfig)
}
