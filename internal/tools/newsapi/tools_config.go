package newsapi

import (
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// GetToolConfigs returns all NewsAPI tool configurations
func GetToolConfigs(cfg *config.NewsAPIConfig) []models.ToolConfig {
	return []models.ToolConfig{
		{
			Key: keys.NewsApi,
			Constructor: func(cfg any, sharedBag bag.SharedBag) (models.Tool, error) {
				return NewFromConfig(cfg.(*config.NewsAPIConfig), sharedBag)
			},
			Config:       cfg,
			CacheEnabled: cfg.CacheEnable,
			CacheTTL:     6 * time.Hour,
			RateLimit: &models.ToolsRateLimit{
				RequestsPerSecond: 5,    // Conservative rate
				RequestsPerDay:    1000, // NewsAPI free tier
				Burst:             10,
			},
		},
	}
}
