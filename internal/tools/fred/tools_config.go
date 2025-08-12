package fred

import (
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// GetToolConfigs returns all FRED tool configurations
func GetToolConfigs(cfg *config.FREDConfig) []models.ToolConfig {
	return []models.ToolConfig{
		{
			Key: keys.Fred,
			Constructor: func(cfg any, sharedBag bag.SharedBag) (models.Tool, error) {
				return NewFromConfig(cfg.(*config.FREDConfig), sharedBag)
			},
			Config:       cfg,
			CacheEnabled: cfg.CacheEnable,
			CacheTTL:     24 * time.Hour, // Economic data changes slowly
			RateLimit: &models.ToolsRateLimit{
				RequestsPerSecond: 10,    // FRED is generous
				RequestsPerDay:    10000, // High limit
				Burst:             20,
			},
		},
	}
}
