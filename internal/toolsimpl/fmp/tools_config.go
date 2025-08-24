package fmp

import (
	"log/slog"
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/internal/tools"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// GetToolConfigs returns all FMP tool configurations
func GetToolConfigs(_cfg any, sharedBag bag.SharedBag) models.ToolConfig {
	cfg := _cfg.(*config.FMPConfig)
	return models.ToolConfig{
		Key: bag.FMP,
		Constructor: func() (models.Tool, error) {
			slog.Debug("Initializing FMP tool",
				"config", func() any {
					safeCfg := *cfg
					safeCfg.APIKey = "<redacted>"
					return safeCfg
				}(),
			)
			return new(cfg.APIKey, sharedBag)
		},
		Config:       cfg,
		CacheEnabled: cfg.CacheEnable,
		CacheTTL:     6 * time.Hour,
		RateLimit: &models.ToolsRateLimit{
			RequestsPerSecond: 3,   // FMP can be strict
			RequestsPerDay:    250, // Free tier limit
			Burst:             5,
		},
	}
}

func init() {
	tools.Register(bag.FMP, GetToolConfigs)
}
