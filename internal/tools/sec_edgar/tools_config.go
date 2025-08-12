package sec_edgar

import (
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// GetToolConfigs returns tool configurations for SEC Edgar
func GetToolConfigs(cfg *config.SECEdgarConfig) []models.ToolConfig {
	if cfg == nil {
		return nil
	}

	configs := []models.ToolConfig{
		{
			Key:    keys.SECFilings,
			Config: cfg,
			Constructor: func(cfg any, sharedBag bag.SharedBag) (models.Tool, error) {
				return NewFromConfig(cfg.(*config.SECEdgarConfig), sharedBag)
			},
			CacheEnabled: cfg.CacheEnable,
			CacheTTL:     24 * time.Hour, // SEC data doesn't change frequently
			RateLimit: &models.ToolsRateLimit{
				RequestsPerSecond: 1, // SEC requires polite usage
				RequestsPerDay:    cfg.MaxDaily,
				Burst:             5,
			},
		},
	}

	return configs
}
