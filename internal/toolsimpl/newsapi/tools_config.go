package newsapi

import (
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/internal/tools"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// GetToolConfigs returns all NewsAPI tool configurations
func GetToolConfigs(_cfg any, sharedBag bag.SharedBag) models.ToolConfig {
	cfg := _cfg.(*config.NewsAPIConfig)
	return models.ToolConfig{
		Key: bag.NewsAPI,
		Constructor: func() (models.Tool, error) {
			return new(cfg.APIKey, cfg.BaseURL, cfg.Locale, sharedBag)
		},
		Config:       cfg,
		CacheEnabled: cfg.CacheEnable,
		CacheTTL:     6 * time.Hour,
		RateLimit: &models.ToolsRateLimit{
			RequestsPerSecond: 1,   // Conservative rate
			RequestsPerDay:    100, // NewsAPI free tier
			Burst:             10,
		},
		Persisting: cfg.Persisting,
	}
}

func init() {
	tools.Register(bag.NewsAPI, GetToolConfigs)
}
