package fred

import (
	"log/slog"
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/internal/tools"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// GetToolConfigs returns all FRED tool configurations
func GetToolConfigs(_cfg any, sharedBag bag.SharedBag) models.ToolConfig {
	cfg := _cfg.(*config.FREDConfig)
	return models.ToolConfig{
		Key: bag.Fred,
		Constructor: func() (models.Tool, error) {
			slog.Debug("Initializing FRED tool",
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
		CacheTTL:     24 * time.Hour, // Economic data changes slowly
		RateLimit: &models.ToolsRateLimit{
			RequestsPerSecond: 10,    // FRED is generous
			RequestsPerDay:    10000, // High limit
			Burst:             20,
		},
	}
}

func init() {
	tools.Register(bag.Fred, GetToolConfigs)
}
