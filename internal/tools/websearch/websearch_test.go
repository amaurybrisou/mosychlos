package websearch

import (
	"context"
	"testing"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebSearchProvider(t *testing.T) {
	sharedBag := bag.NewSharedBag()
	provider, err := New(sharedBag)
	require.NoError(t, err)

	t.Run("Tool Properties", func(t *testing.T) {
		assert.Equal(t, keys.WebSearch.String(), provider.Name())
		assert.Equal(t, keys.WebSearch, provider.Key())
		assert.NotEmpty(t, provider.Description())
		assert.Contains(t, provider.Tags(), "web")
		assert.Contains(t, provider.Tags(), "openai-internal")
		assert.True(t, provider.IsExternal())
	})

	t.Run("Tool Definition", func(t *testing.T) {
		def := provider.Definition()

		typedDef, ok := def.(*models.CustomToolDef)
		assert.True(t, ok)
		assert.Equal(t, models.CustomToolDefType, typedDef.Type)
		assert.Equal(t, keys.WebSearch.String(), typedDef.FunctionDef.Name)
		assert.NotEmpty(t, typedDef.FunctionDef.Description)

		// Check parameters structure
		params, ok := typedDef.FunctionDef.Parameters["properties"].(map[string]any)
		require.True(t, ok)

		queryParam, ok := params["query"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "string", queryParam["type"])

		required, ok := typedDef.FunctionDef.Parameters["required"].([]string)
		require.True(t, ok)
		assert.Contains(t, required, "query")
	})

	t.Run("Run Should Execute", func(t *testing.T) {
		// This tool should execute and return a success message
		result, err := provider.Run(context.Background(), `{"query": "test"}`)
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "initialized")
	})
}

func TestGetToolConfigs(t *testing.T) {
	sharedBag := bag.NewSharedBag()
	t.Run("Web Search Disabled", func(t *testing.T) {
		cfg := &config.OpenAIConfig{
			WebSearch: false,
		}

		configs := GetToolConfigs(cfg)
		assert.Empty(t, configs)
	})

	t.Run("Web Search Enabled", func(t *testing.T) {
		cfg := &config.OpenAIConfig{
			WebSearch: true,
		}

		configs := GetToolConfigs(cfg)
		require.Len(t, configs, 1)

		config := configs[0]
		assert.Equal(t, keys.WebSearch, config.Key)
		assert.False(t, config.CacheEnabled) // No caching for internal tools
		assert.Nil(t, config.RateLimit)      // OpenAI handles rate limiting
		assert.NotNil(t, config.Constructor)

		// Test constructor
		tool, err := config.Constructor(cfg, sharedBag)
		require.NoError(t, err)
		assert.Equal(t, keys.WebSearch.String(), tool.Name())
	})
}
