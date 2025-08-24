// internal/llm/openai/speakeasy_client_test.go
package openai

import (
	"context"
	"testing"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSpeakeasyProvider(t *testing.T) {
	cfg := config.LLMConfig{
		Provider: "openai",
		Model:    config.LLMModelGPT4o,
		APIKey:   "test-api-key",
		BaseURL:  "https://api.openai.com",
	}

	sharedBag := bag.NewSharedBag()
	provider := NewSpeakeasyProvider(cfg, sharedBag)

	assert.NotNil(t, provider)
	assert.Equal(t, "openai-speakeasy", provider.Name())
	assert.NotNil(t, provider.client)
}

func TestSpeakeasyProvider_NewSession(t *testing.T) {
	cfg := config.LLMConfig{
		Provider: "openai",
		Model:    config.LLMModelGPT4o,
		APIKey:   "test-api-key",
	}

	sharedBag := bag.NewSharedBag()
	provider := NewSpeakeasyProvider(cfg, sharedBag)
	session := provider.NewSession()

	assert.NotNil(t, session)

	// Test that we can add messages to the session
	session.Add(models.RoleUser, "Hello")
	session.Add(models.RoleAssistant, "Hi there!")

	// Test the session interface methods (they should not panic)
	session.AddToolResult(models.ToolCall{}, "tool result")
	session.AddFunctionCallResult(models.ToolCall{}, "function result")
	session.SetToolChoice(&models.ToolChoice{})

	// Test that streaming works (even with fallback implementation)
	stream, err := session.NextStream(context.Background(), nil, nil)
	// The new implementation uses Next() fallback, which will try to make an HTTP call
	// In test environment, this will fail with network error, but that's expected
	if err != nil {
		// In test environment, we expect network errors since we don't have real API access
		assert.Contains(t, err.Error(), "speakeasy chat completion failed")
	} else {
		assert.NotNil(t, stream)
	}
}

func TestSpeakeasySession_Add(t *testing.T) {
	cfg := config.LLMConfig{
		Provider: "openai",
		Model:    config.LLMModelGPT4o,
		APIKey:   "test-api-key",
	}

	sharedBag := bag.NewSharedBag()
	provider := NewSpeakeasyProvider(cfg, sharedBag)
	session := provider.NewSession().(*speakeasySession)

	// Test adding different types of messages
	session.Add(models.RoleUser, "Hello")
	session.Add(models.RoleAssistant, "Hi there!")
	session.Add(models.RoleSystem, "You are a helpful assistant")

	require.Len(t, session.messages, 3)

	assert.Equal(t, "Hello", session.messages[0].Content)
	assert.Equal(t, "Hi there!", session.messages[1].Content)
	assert.Equal(t, "You are a helpful assistant", session.messages[2].Content)
}