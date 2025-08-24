package secedgar

import (
	"log/slog"
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/internal/tools"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// GetToolConfigs returns tool configurations for SEC Edgar
func GetToolConfigs(_cfg any, sharedBag bag.SharedBag) models.ToolConfig {
	cfg := _cfg.(*config.SECEdgarConfig)

	return models.ToolConfig{
		Key:    bag.SECFilings,
		Config: cfg,
		Constructor: func() (models.Tool, error) {
			slog.Debug("Initializing SEC Edgar tool",
				"config", func() any {
					return cfg
				}(),
			)

			return new(cfg.UserAgent, cfg.BaseURL, sharedBag)
		},
		CacheEnabled: cfg.CacheEnable,
		CacheTTL:     24 * time.Hour, // SEC data doesn't change frequently
		RateLimit: &models.ToolsRateLimit{
			RequestsPerSecond: 1, // SEC requires polite usage
			RequestsPerDay:    cfg.MaxDaily,
			Burst:             5,
		},
	}
}

func init() {
	tools.Register(bag.SECFilings, GetToolConfigs)
}
