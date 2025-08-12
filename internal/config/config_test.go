package config

import (
	"os"
	"testing"

	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/amaurybrisou/mosychlos/pkg/nativeutils"
	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				DataDir:  "/tmp/data",
				CacheDir: "/tmp/cache",
				Localization: models.LocalizationConfig{
					Country:  "FR",
					Language: "fr",
					Timezone: "Europe/Paris",
					Currency: "EUR",
					Region:   "Ile-de-France",
					City:     "Paris",
				},
				LLM: LLMConfig{
					Provider: "openai",
					Model:    "gpt-4o",
					APIKey:   "test-api-key",
					OpenAI: OpenAIConfig{
						Temperature: nativeutils.Ptr(0.1),
					},
				},
				Jurisdiction: JurisdictionConfig{
					Rules: models.ComplianceRules{
						AllowedAssetTypes: []string{"stock", "etf"},
						MaxLeverage:       1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty DataDir",
			config: Config{
				DataDir:  "",
				CacheDir: "/tmp/cache",
			},
			wantErr: true,
		},
		{
			name: "empty CacheDir",
			config: Config{
				DataDir:  "/tmp/data",
				CacheDir: "",
			},
			wantErr: true,
		},
		{
			name: "relative DataDir path",
			config: Config{
				DataDir:  "data",
				CacheDir: "/tmp/cache",
			},
			wantErr: true,
		},
		{
			name: "invalid jurisdiction config",
			config: Config{
				DataDir:  "/tmp/data",
				CacheDir: "/tmp/cache",
				Jurisdiction: JurisdictionConfig{
					Country: "", // invalid empty country
				},
			},
			wantErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := c.config.Validate()
			if c.wantErr {
				assert.Error(t, err)
				assert.False(t, c.config.IsValid())
			} else {
				assert.NoError(t, err)
				assert.True(t, c.config.IsValid())

				// test that subsequent calls return immediately
				err2 := c.config.Validate()
				assert.NoError(t, err2)
			}
		})
	}
}

func TestJurisdictionConfig_Validate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		config  JurisdictionConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: JurisdictionConfig{
				Country: "FR",
				Rules: models.ComplianceRules{
					AllowedAssetTypes: []string{"stock"},
					MaxLeverage:       1,
				},
			},
			wantErr: false,
		},
		{
			name: "empty country",
			config: JurisdictionConfig{
				Country: "",
			},
			wantErr: true,
		},
		{
			name: "lowercase country",
			config: JurisdictionConfig{
				Country: "fr",
			},
			wantErr: true,
		},
		{
			name: "invalid country length",
			config: JurisdictionConfig{
				Country: "FRA",
			},
			wantErr: true,
		},
		{
			name: "invalid custom schema path",
			config: JurisdictionConfig{
				Country:          "FR",
				CustomSchemaPath: "/nonexistent/schema.json",
				Rules: models.ComplianceRules{
					AllowedAssetTypes: []string{"stock"},
					MaxLeverage:       1,
				},
			},
			wantErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := c.config.Validate()
			if c.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBinanceConfig_Validate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		config  BinanceConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: BinanceConfig{
				APIKey:    "test-key",
				APISecret: "test-secret",
				BaseURL:   "https://api.binance.com",
			},
			wantErr: false,
		},
		{
			name:    "empty config (not configured)",
			config:  BinanceConfig{},
			wantErr: false, // should be valid when not configured
		},
		{
			name: "partial config - missing secret",
			config: BinanceConfig{
				APIKey:  "test-key",
				BaseURL: "https://api.binance.com",
			},
			wantErr: true,
		},
		{
			name: "invalid URL",
			config: BinanceConfig{
				APIKey:    "test-key",
				APISecret: "test-secret",
				BaseURL:   "invalid-url",
			},
			wantErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := c.config.Validate()
			if c.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_MustValidate(t *testing.T) {
	t.Parallel()

	// test successful validation
	validConfig := Config{
		DataDir:  "/tmp/data-test",
		CacheDir: "/tmp/cache-test",
		Localization: models.LocalizationConfig{
			Country:  "FR",
			Language: "fr",
			Timezone: "Europe/Paris",
			Currency: "EUR",
			Region:   "Ile-de-France",
			City:     "Paris",
		},
		LLM: LLMConfig{
			Provider: "openai",
			Model:    "gpt-4o",
			APIKey:   "test-api-key",
		},
		Jurisdiction: JurisdictionConfig{
			Rules: models.ComplianceRules{
				AllowedAssetTypes: []string{"stock"},
				MaxLeverage:       1,
			},
		},
	}

	// should not panic
	validConfig.MustValidate()
	assert.True(t, validConfig.IsValid())

	// test panic on invalid config
	invalidConfig := Config{
		DataDir:  "", // invalid empty
		CacheDir: "/tmp/cache",
	}

	assert.Panics(t, func() {
		invalidConfig.MustValidate()
	})
}

func TestComplianceRules_Validate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		rules   models.ComplianceRules
		wantErr bool
	}{
		{
			name: "valid rules",
			rules: models.ComplianceRules{
				AllowedAssetTypes: []string{"stock", "etf"},
				MaxLeverage:       1,
			},
			wantErr: false,
		},
		{
			name: "negative max leverage",
			rules: models.ComplianceRules{
				MaxLeverage: -1,
			},
			wantErr: true,
		},
		{
			name: "empty asset type in allowed",
			rules: models.ComplianceRules{
				AllowedAssetTypes: []string{"stock", ""},
				MaxLeverage:       1,
			},
			wantErr: true,
		},
		{
			name: "conflicting allowed and disallowed",
			rules: models.ComplianceRules{
				AllowedAssetTypes:    []string{"stock"},
				DisallowedAssetTypes: []string{"STOCK"}, // case insensitive conflict
				MaxLeverage:          1,
			},
			wantErr: true,
		},
		{
			name: "empty ticker substitute key",
			rules: models.ComplianceRules{
				TickerSubstitutes: map[string]string{"": "REPLACEMENT"},
				MaxLeverage:       1,
			},
			wantErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := c.rules.Validate()
			if c.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLLMConfig_Validate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		config  LLMConfig
		wantErr bool
	}{
		{
			name: "valid OpenAI config",
			config: LLMConfig{
				Provider: "openai",
				Model:    "gpt-4o",
				APIKey:   "test-api-key",
				OpenAI: OpenAIConfig{
					Temperature:         nativeutils.Ptr(0.1),
					MaxCompletionTokens: 4096,
					ReasoningEffort:     nativeutils.Ptr("medium"),
					Verbosity:           nativeutils.Ptr("medium"),
					ParallelToolCalls:   true,
				},
			},
			wantErr: false,
		},
		{
			name: "empty provider",
			config: LLMConfig{
				Provider: "",
				Model:    "gpt-4o",
				APIKey:   "test-api-key",
			},
			wantErr: true,
		},
		{
			name: "empty model",
			config: LLMConfig{
				Provider: "openai",
				Model:    "",
				APIKey:   "test-api-key",
			},
			wantErr: true,
		},
		{
			name: "empty API key",
			config: LLMConfig{
				Provider: "openai",
				Model:    "gpt-4o",
				APIKey:   "",
			},
			wantErr: true,
		},
		{
			name: "invalid base URL",
			config: LLMConfig{
				Provider: "openai",
				Model:    "gpt-4o",
				APIKey:   "test-api-key",
				BaseURL:  "invalid-url",
			},
			wantErr: true,
		},
		{
			name: "valid base URL",
			config: LLMConfig{
				Provider: "openai",
				Model:    "gpt-4o",
				APIKey:   "test-api-key",
				BaseURL:  "https://api.custom.ai/v1",
			},
			wantErr: false,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := c.config.Validate()
			if c.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOpenAIConfig_Validate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		config  OpenAIConfig
		wantErr bool
	}{
		{
			name:    "empty config is valid",
			config:  OpenAIConfig{},
			wantErr: false,
		},
		{
			name: "valid temperature",
			config: OpenAIConfig{
				Temperature: nativeutils.Ptr(0.5),
			},
			wantErr: false,
		},
		{
			name: "temperature too low",
			config: OpenAIConfig{
				Temperature: nativeutils.Ptr(-1.0),
			},
			wantErr: true,
		},
		{
			name: "temperature too high",
			config: OpenAIConfig{
				Temperature: nativeutils.Ptr(3.0),
			},
			wantErr: true,
		},
		{
			name: "valid top_p",
			config: OpenAIConfig{
				TopP: nativeutils.Ptr(0.5),
			},
			wantErr: false,
		},
		{
			name: "top_p too low",
			config: OpenAIConfig{
				TopP: nativeutils.Ptr(-0.1),
			},
			wantErr: true,
		},
		{
			name: "top_p too high",
			config: OpenAIConfig{
				TopP: nativeutils.Ptr(1.5),
			},
			wantErr: true,
		},
		{
			name: "valid reasoning effort",
			config: OpenAIConfig{
				ReasoningEffort: nativeutils.Ptr("high"),
			},
			wantErr: false,
		},
		{
			name: "invalid reasoning effort",
			config: OpenAIConfig{
				ReasoningEffort: nativeutils.Ptr("extreme"),
			},
			wantErr: true,
		},
		{
			name: "valid verbosity",
			config: OpenAIConfig{
				Verbosity: nativeutils.Ptr("low"),
			},
			wantErr: false,
		},
		{
			name: "invalid verbosity",
			config: OpenAIConfig{
				Verbosity: nativeutils.Ptr("none"),
			},
			wantErr: true,
		},
		{
			name: "valid service tier",
			config: OpenAIConfig{
				ServiceTier: nativeutils.Ptr("priority"),
			},
			wantErr: false,
		},
		{
			name: "invalid service tier",
			config: OpenAIConfig{
				ServiceTier: nativeutils.Ptr("premium"),
			},
			wantErr: true,
		},
		{
			name: "valid presence penalty",
			config: OpenAIConfig{
				PresencePenalty: nativeutils.Ptr(1.0),
			},
			wantErr: false,
		},
		{
			name: "presence penalty too low",
			config: OpenAIConfig{
				PresencePenalty: nativeutils.Ptr(-3.0),
			},
			wantErr: true,
		},
		{
			name: "presence penalty too high",
			config: OpenAIConfig{
				PresencePenalty: nativeutils.Ptr(3.0),
			},
			wantErr: true,
		},
		{
			name: "valid frequency penalty",
			config: OpenAIConfig{
				FrequencyPenalty: nativeutils.Ptr(-1.5),
			},
			wantErr: false,
		},
		{
			name: "frequency penalty too low",
			config: OpenAIConfig{
				FrequencyPenalty: nativeutils.Ptr(-2.5),
			},
			wantErr: true,
		},
		{
			name: "frequency penalty too high",
			config: OpenAIConfig{
				FrequencyPenalty: nativeutils.Ptr(2.5),
			},
			wantErr: true,
		},
		{
			name: "valid max completion tokens",
			config: OpenAIConfig{
				MaxCompletionTokens: 4096,
			},
			wantErr: false,
		},
		{
			name: "invalid max completion tokens",
			config: OpenAIConfig{
				MaxCompletionTokens: -1,
			},
			wantErr: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			err := c.config.Validate()
			if c.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// cleanup test directories
func init() {
	os.RemoveAll("/tmp/data")
	os.RemoveAll("/tmp/cache")
	os.RemoveAll("/tmp/data-test")
	os.RemoveAll("/tmp/cache-test")
}
