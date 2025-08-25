package summarize

import (
	"time"

	"github.com/amaurybrisou/mosychlos/internal/tools"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// Config struct for this tool (optional). If you don’t have one yet,
// you can wire it in your global tools config map like others.
type Config struct {
	Model       string        `json:"model"`       // e.g., "gpt-5-nano"
	CacheEnable bool          `json:"cacheEnable"` // default false
	CacheTTL    time.Duration `json:"cacheTtl"`
	Persisting  bool          `json:"persisting"`
}

// GetToolConfigs returns tool configuration in your standard format
func GetToolConfigs(_cfg any, sharedBag bag.SharedBag) models.ToolConfig {
	var cfg *Config
	if _cfg == nil {
		cfg = &Config{Model: "gpt-5-nano", CacheEnable: false, CacheTTL: 0, Persisting: false}
	} else {
		cfg = _cfg.(*Config)
		if cfg.Model == "" {
			cfg.Model = "gpt-5-nano"
		}
	}

	return models.ToolConfig{
		Key: bag.SummarizeNews,
		Constructor: func() (models.Tool, error) {
			return new(sharedBag, cfg.Model)
		},
		Config:       cfg,
		CacheEnabled: cfg.CacheEnable, // often false for LLM transforms, but you can enable
		CacheTTL:     cfg.CacheTTL,
		RateLimit: &models.ToolsRateLimit{
			RequestsPerSecond: 3,
			RequestsPerDay:    5000, // logical cap; it’s local LLM usage, not external API
			Burst:             5,
		},
		Persisting: cfg.Persisting,
	}
}

func init() {
	// Register under your tools registry like other tools do.
	tools.Register(bag.SummarizeNews, GetToolConfigs)
}
