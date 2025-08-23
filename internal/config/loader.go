// Package config provides configuration loading utilities
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	pkglog "github.com/amaurybrisou/mosychlos/pkg/log"
	"github.com/spf13/viper"
)

// LoadConfig loads configuration with priority: env vars > config.yaml > config.default.yaml
func LoadConfig() (*Config, error) {
	v := viper.New()

	// Set up configuration hierarchy
	// 1. Start with defaults from config.default.yaml
	v.SetConfigName("config.default")
	v.SetConfigType("yaml")
	v.AddConfigPath("config")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read default config: %w", err)
	}

	// 2. Merge with user config from config.yaml (if it exists)
	v.SetConfigName("config")
	if err := v.MergeInConfig(); err != nil {
		// It's OK if config.yaml doesn't exist, we'll use defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to merge config.yaml: %w", err)
		}
	}

	// 3. Override with environment variables (highest priority)
	v.SetEnvPrefix("MOSYCHLOS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Set up specific environment variable mappings for common tools
	// This allows NEWSAPI_API_KEY to map to tools.newsapi.api_key
	v.BindEnv("tools.newsapi.api_key", "NEWSAPI_API_KEY")
	v.BindEnv("tools.fred.api_key", "FRED_API_KEY")
	v.BindEnv("tools.fmp.api_key", "FMP_API_KEY")
	v.BindEnv("llm.api_key", "OPENAI_API_KEY")

	// Unmarshal into our config structure
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

// MustLoadConfig loads configuration and panics on error
func MustLoadConfig() *Config {
	cfg, err := LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load configuration: %v", err))
	}

	// Set up logging using pkg/log and loaded config
	logCfg := pkglog.Config{
		Level:     pkglog.LogLevel(cfg.Logging.Level),
		Format:    pkglog.LogFormat(cfg.Logging.Format),
		AddSource: cfg.Logging.AddSource,
		Output:    os.Stdout,
	}
	pkglog.InitWithConfig(logCfg)

	slog.Debug("Configuration loaded successfully", slog.Any("config", cfg))
	return cfg
}
