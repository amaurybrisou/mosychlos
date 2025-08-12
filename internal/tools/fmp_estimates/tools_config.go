package fmpestimates

import (
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// GetToolConfigs returns all FMP Analyst Estimates tool configurations
func GetToolConfigs(cfg *config.FMPAnalystEstimatesConfig) []models.ToolConfig {
	return []models.ToolConfig{
		{
			Key: keys.FMPAnalystEstimates,
			Constructor: func(cfg any, sharedBag bag.SharedBag) (models.Tool, error) {
				return NewAnalystEstimatesFromConfig(cfg.(*config.FMPAnalystEstimatesConfig), sharedBag)
			},
			Config:       cfg,
			CacheEnabled: cfg.CacheEnable,
			CacheTTL:     24 * time.Hour, // Estimates change daily
			RateLimit: &models.ToolsRateLimit{
				RequestsPerSecond: 3,   // Same as FMP
				RequestsPerDay:    250, // Shares FMP quota
				Burst:             5,
			},
		},
	}
}
