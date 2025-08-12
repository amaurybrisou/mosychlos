package tools

import (
	"fmt"
	"log/slog"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/internal/tools/fmp"
	"github.com/amaurybrisou/mosychlos/internal/tools/fred"
	"github.com/amaurybrisou/mosychlos/internal/tools/newsapi"
	"github.com/amaurybrisou/mosychlos/internal/tools/sec_edgar"
	"github.com/amaurybrisou/mosychlos/internal/tools/websearch"
	"github.com/amaurybrisou/mosychlos/internal/tools/yfinance"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// WrapTool applies all configured wrappers to a tool
func WrapTool(tool models.Tool, config *models.ToolConfig, cacheDir string, sharedBag bag.SharedBag) models.Tool {
	wrapped := tool

	// Apply rate limiting if configured
	if config.RateLimit != nil {
		wrapped = NewRateLimitedTool(
			wrapped,
			config.RateLimit.RequestsPerSecond,
			config.RateLimit.RequestsPerDay,
			config.RateLimit.Burst,
		)
		slog.Debug("Applied rate limiting",
			"tool", tool.Name(),
			"requests_per_second", config.RateLimit.RequestsPerSecond,
			"requests_per_day", config.RateLimit.RequestsPerDay,
		)
	}

	// Apply caching if enabled
	if config.CacheEnabled {
		wrapped = NewCachedToolWithMonitoring(wrapped, cacheDir, config.CacheTTL, sharedBag)
		slog.Debug("Applied caching",
			"tool", tool.Name(),
			"cache_ttl", config.CacheTTL,
		)
	}

	// Apply metrics tracking if shared bag is available
	if sharedBag != nil {
		wrapped = NewMetricsWrapper(wrapped, sharedBag)
		slog.Debug("Applied metrics wrapper",
			"tool", tool.Name(),
		)
	}

	return wrapped
}

// GetToolConfigs returns all tool configurations
func GetToolConfigs(cfg *config.Config) []models.ToolConfig {
	var configs []models.ToolConfig

	// NewsAPI Tool
	if cfg.Tools.NewsAPI != nil {
		newsConfigs := newsapi.GetToolConfigs(cfg.Tools.NewsAPI)
		configs = append(configs, newsConfigs...)
	}

	// FRED Tool
	if cfg.Tools.FRED != nil {
		fredConfigs := fred.GetToolConfigs(cfg.Tools.FRED)
		configs = append(configs, fredConfigs...)
	}

	// FMP Tool
	if cfg.Tools.FMP != nil {
		fmpConfigs := fmp.GetToolConfigs(cfg.Tools.FMP)
		configs = append(configs, fmpConfigs...)
	}

	// FMP Analyst Estimates Tool (Premium - disabled)
	// if cfg.Tools.FMPAnalystEstimates != nil {
	// 	fmpEstimatesConfigs := fmpestimates.GetToolConfigs(cfg.Tools.FMPAnalystEstimates)
	// 	configs = append(configs, fmpEstimatesConfigs...)
	// }

	// YFinance Tools
	if cfg.Tools.YFinance != nil {
		yfinanceConfigs := yfinance.GetToolConfigs(cfg.Tools.YFinance)
		configs = append(configs, yfinanceConfigs...)
	}

	// SEC Edgar Tool
	if cfg.Tools.SECEdgar != nil {
		secConfigs := sec_edgar.GetToolConfigs(cfg.Tools.SECEdgar)
		configs = append(configs, secConfigs...)
	}

	// Web Search Tool (internal OpenAI tool)
	// Only add if web search is enabled in OpenAI config
	if cfg.LLM.OpenAI.WebSearch {
		webSearchConfigs := websearch.GetToolConfigs(&cfg.LLM.OpenAI)
		configs = append(configs, webSearchConfigs...)
	}

	return configs
}

// RegisterTool registers a single tool with all wrappers
func RegisterTool(config models.ToolConfig, cacheDir string, sharedBag bag.SharedBag) error {
	// Create the tool instance
	tool, err := config.Constructor(config.Config, sharedBag)
	if err != nil {
		return fmt.Errorf("failed to create tool %s: %w", config.Key, err)
	}

	// Apply all wrappers
	wrappedTool := WrapTool(tool, &config, cacheDir, sharedBag)

	// Register the wrapped tool
	tools[config.Key] = wrappedTool

	slog.Info("Tool registered successfully",
		"key", config.Key,
		"name", tool.Name(),
		"cache_enabled", config.CacheEnabled,
		"rate_limited", config.RateLimit != nil,
	)

	return nil
}
