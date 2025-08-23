package fmp

import (
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// GetToolConfigs returns all FMP tool configurations
func GetToolConfigs(cfg *config.FMPConfig) []models.ToolConfig {
	return []models.ToolConfig{
		{
			Key: bag.FMP,
			Constructor: func(cfg any, sharedBag bag.SharedBag) (models.Tool, error) {
				return NewFromConfig(cfg.(*config.FMPConfig), sharedBag)
			},
			Config:       cfg,
			CacheEnabled: cfg.CacheEnable,
			CacheTTL:     6 * time.Hour,
			RateLimit: &models.ToolsRateLimit{
				RequestsPerSecond: 3,   // FMP can be strict
				RequestsPerDay:    250, // Free tier limit
				Burst:             5,
			},
		},
	}
}
