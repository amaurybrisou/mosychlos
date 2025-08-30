package fmpestimates

import (
	"log/slog"
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/internal/tools"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// GetToolConfigs returns all FMP Analyst Estimates tool configurations
func GetToolConfigs(_cfg any, sharedBag bag.SharedBag) models.ToolConfig {
	cfg := _cfg.(*config.FMPAnalystEstimatesConfig)
	return models.ToolConfig{
		Key: bag.FMPAnalystEstimates,
		Constructor: func() (models.Tool, error) {
			slog.Debug("Initializing FMP Analyst Estimates tool",
				"config", func() any {
					safeCfg := *cfg
					safeCfg.APIKey = "<redacted>"
					return safeCfg
				}(),
			)
			return new(cfg.APIKey, cfg.CacheDir, sharedBag)
		},
		Config:       cfg,
		CacheEnabled: cfg.CacheEnable,
		CacheTTL:     24 * time.Hour, // Estimates change daily
		RateLimit: &models.ToolsRateLimit{
			RequestsPerSecond: 3,   // Same as FMP
			RequestsPerDay:    250, // Shares FMP quota
			Burst:             5,
		},
		Persisting: cfg.Persisting,
	}
}

func init() {
	tools.Register(bag.FMPAnalystEstimates, GetToolConfigs)
}
