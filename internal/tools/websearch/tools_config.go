package websearch

import (
	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// GetToolConfigs returns tool configurations for web search
// Since web_search_preview is an internal OpenAI tool, we only need the basic config
func GetToolConfigs(cfg *config.OpenAIConfig) []models.ToolConfig {
	// Only include web search if it's enabled in OpenAI config
	if !cfg.WebSearch {
		return []models.ToolConfig{}
	}

	return []models.ToolConfig{
		{
			Key:          keys.WebSearch,
			CacheEnabled: false, // Internal OpenAI tool - no local caching
			CacheTTL:     0,     // No TTL needed
			RateLimit:    nil,   // OpenAI handles rate limiting internally
			Constructor: func(cfg any, sharedBag bag.SharedBag) (models.Tool, error) {
				return New(sharedBag)
			},
			Config: cfg, // Pass the OpenAI config
		},
	}
}
