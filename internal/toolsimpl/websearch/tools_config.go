// Package websearch provides a tool for accessing web search capabilities
package websearch

import (
	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/internal/tools"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// GetToolConfigs returns tool configurations for web search
// Since web_search_preview is an internal OpenAI tool, we only need the basic config
func GetToolConfigs(_cfg any, sharedBag bag.SharedBag) models.ToolConfig {
	cfg := _cfg.(*config.OpenAIConfig)
	return models.ToolConfig{
		Key:          bag.WebSearch,
		CacheEnabled: false, // Internal OpenAI tool - no local caching
		CacheTTL:     0,     // No TTL needed
		RateLimit:    nil,   // OpenAI handles rate limiting internally
		Constructor: func() (models.Tool, error) {
			return new(sharedBag)
		},
		Config: cfg, // Pass the OpenAI config
	}
}

func init() {
	tools.Register(bag.WebSearch, GetToolConfigs)
}
