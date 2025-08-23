package tools

import (
	"testing"
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/internal/tools/yfinance"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

func TestModularToolConfigs(t *testing.T) {
	cfg := &config.Config{
		Tools: config.ToolsConfig{
			NewsAPI: &config.NewsAPIConfig{
				APIKey:      "test-key",
				CacheEnable: true,
			},
			FRED: &config.FREDConfig{
				APIKey:      "test-fred-key",
				CacheEnable: true,
			},
			FMP: &config.FMPConfig{
				APIKey:      "test-fmp-key",
				CacheEnable: true,
			},
			FMPAnalystEstimates: &config.FMPAnalystEstimatesConfig{
				APIKey:      "test-estimates-key",
				CacheEnable: true,
			},
			YFinance: &config.YFinanceConfig{
				CacheEnable: true,
			},
		},
	}

	configs := GetToolConfigs(cfg)

	// Should have 8 tools total (1+1+1+5) - FMP Analyst Estimates disabled (premium)
	expectedCount := 8
	if len(configs) != expectedCount {
		t.Errorf("Expected %d tool configs, got %d", expectedCount, len(configs))
	}

	// Verify we have all expected keys
	expectedKeys := map[bag.Key]bool{
		bag.NewsApi: false,
		bag.Fred:    false,
		bag.FMP:     false,
		// bag.FMPAnalystEstimates: false, // Premium - disabled
		bag.YFinanceStockData:  false,
		bag.YFinanceStockInfo:  false,
		bag.YFinanceDividends:  false,
		bag.YFinanceFinancials: false,
		bag.YFinanceMarketData: false,
	}

	// Check each config
	for _, cfg := range configs {
		// Verify key exists
		if _, exists := expectedKeys[cfg.Key]; !exists {
			t.Errorf("Unexpected tool key: %s", cfg.Key)
		}
		expectedKeys[cfg.Key] = true

		// Verify all configs have rate limiting
		if cfg.RateLimit == nil {
			t.Errorf("Tool %s missing rate limit config", cfg.Key)
		}

		// Verify all configs have constructors
		if cfg.Constructor == nil {
			t.Errorf("Tool %s missing constructor", cfg.Key)
		}

		// Verify cache settings
		if !cfg.CacheEnabled {
			t.Errorf("Tool %s should have caching enabled", cfg.Key)
		}

		// Verify cache TTL is set
		if cfg.CacheTTL == 0 {
			t.Errorf("Tool %s missing cache TTL", cfg.Key)
		}
	}

	// Verify all expected keys were found
	for key, found := range expectedKeys {
		if !found {
			t.Errorf("Missing expected tool key: %s", key)
		}
	}
}

func TestYFinanceToolConfigs_CacheTTL(t *testing.T) {
	cfg := &config.YFinanceConfig{
		CacheEnable: true,
	}

	configs := GetYFinanceConfigs(cfg)

	// Verify different cache TTLs for different tool types
	expectedTTLs := map[bag.Key]time.Duration{
		bag.YFinanceStockData:  6 * time.Hour,  // Real-time data
		bag.YFinanceStockInfo:  24 * time.Hour, // Company info changes slowly
		bag.YFinanceDividends:  24 * time.Hour, // Dividends change rarely
		bag.YFinanceFinancials: 24 * time.Hour, // Financial statements change quarterly
		bag.YFinanceMarketData: 6 * time.Hour,  // Market data updates frequently
	}

	for _, config := range configs {
		expectedTTL, exists := expectedTTLs[config.Key]
		if !exists {
			t.Errorf("Unexpected YFinance tool key: %s", config.Key)
			continue
		}

		if config.CacheTTL != expectedTTL {
			t.Errorf("Tool %s: expected cache TTL %v, got %v",
				config.Key, expectedTTL, config.CacheTTL)
		}
	}
}

// Helper function to get YFinance configs specifically
func GetYFinanceConfigs(cfg *config.YFinanceConfig) []models.ToolConfig {
	// Use the yfinance package's GetToolConfigs function
	return yfinance.GetToolConfigs(cfg)
}
