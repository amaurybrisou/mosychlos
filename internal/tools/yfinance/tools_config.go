package yfinance

import (
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// GetToolConfigs returns all YFinance tool configurations
func GetToolConfigs(cfg *config.YFinanceConfig) []models.ToolConfig {
	// Common rate limits for YFinance (no API key required, but be respectful)
	yfinanceRateLimit := &models.ToolsRateLimit{
		RequestsPerSecond: 2,    // Conservative for Yahoo
		RequestsPerDay:    2000, // Reasonable daily limit
		Burst:             5,
	}

	return []models.ToolConfig{
		{
			Key: keys.YFinanceStockData,
			Constructor: func(cfg any, sharedBag bag.SharedBag) (models.Tool, error) {
				return NewStockDataFromConfig(cfg.(*config.YFinanceConfig), sharedBag)
			},
			Config:       cfg,
			CacheEnabled: cfg.CacheEnable,
			CacheTTL:     6 * time.Hour,
			RateLimit:    yfinanceRateLimit,
		},
		{
			Key: keys.YFinanceStockInfo,
			Constructor: func(cfg any, sharedBag bag.SharedBag) (models.Tool, error) {
				return NewStockInfoFromConfig(cfg.(*config.YFinanceConfig), sharedBag)
			},
			Config:       cfg,
			CacheEnabled: cfg.CacheEnable,
			CacheTTL:     24 * time.Hour, // Company info changes slowly
			RateLimit:    yfinanceRateLimit,
		},
		{
			Key: keys.YFinanceDividends,
			Constructor: func(cfg any, sharedBag bag.SharedBag) (models.Tool, error) {
				return NewDividendsFromConfig(cfg.(*config.YFinanceConfig), sharedBag)
			},
			Config:       cfg,
			CacheEnabled: cfg.CacheEnable,
			CacheTTL:     24 * time.Hour, // Dividends change rarely
			RateLimit:    yfinanceRateLimit,
		},
		{
			Key: keys.YFinanceFinancials,
			Constructor: func(cfg any, sharedBag bag.SharedBag) (models.Tool, error) {
				return NewFinancialsFromConfig(cfg.(*config.YFinanceConfig), sharedBag)
			},
			Config:       cfg,
			CacheEnabled: cfg.CacheEnable,
			CacheTTL:     24 * time.Hour, // Financial statements change quarterly
			RateLimit:    yfinanceRateLimit,
		},
		{
			Key: keys.YFinanceMarketData,
			Constructor: func(cfg any, sharedBag bag.SharedBag) (models.Tool, error) {
				return NewMarketDataFromConfig(cfg.(*config.YFinanceConfig), sharedBag)
			},
			Config:       cfg,
			CacheEnabled: cfg.CacheEnable,
			CacheTTL:     6 * time.Hour, // Market data updates frequently
			RateLimit:    yfinanceRateLimit,
		},
	}
}
