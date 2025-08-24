// internal/llm/openai/speakeasy_integration_test.go
package openai

import (
	"context"
	"testing"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/speakeasy-sdks/openai-go-sdk/pkg/models/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpeakeasyProvider_ToolIntegration(t *testing.T) {
	cfg := config.LLMConfig{
		Provider: "openai",
		Model:    config.LLMModelGPT4o,
		APIKey:   "test-api-key",
		BaseURL:  "https://api.openai.com",
	}

	sharedBag := bag.NewSharedBag()
	provider := NewSpeakeasyProvider(cfg, sharedBag)
	session := provider.NewSession()

	// Test adding messages
	session.Add(models.RoleUser, "Hello")
	assert.Equal(t, 1, len(session.(*speakeasySession).messages))

	// Test adding tool results
	toolCall := models.ToolCall{
		ID:   "call_123",
		Type: "function",
		Function: models.ToolCallFunction{
			Name:      "get_weather",
			Arguments: `{"location": "San Francisco"}`,
		},
	}

	session.AddToolResult(toolCall, "The weather is sunny")
	speakeasySession := session.(*speakeasySession)
	
	// Should have 2 messages now (user + tool result)
	assert.Equal(t, 2, len(speakeasySession.messages))
	
	// Tool result should be formatted as special user message
	toolMessage := speakeasySession.messages[1]
	assert.Equal(t, "user", string(toolMessage.Role))
	assert.Contains(t, toolMessage.Content, "TOOL_RESULT[get_weather]:")
	assert.Contains(t, toolMessage.Content, "The weather is sunny")
}

func TestSpeakeasyProvider_ToolDefinitionConversion(t *testing.T) {
	// Test FunctionToolDef conversion
	functionTool := &models.FunctionToolDef{
		Type: "function",
		Function: models.FunctionDef{
			Name:        "get_weather",
			Description: "Get current weather",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"location": map[string]any{
						"type":        "string",
						"description": "The location to get weather for",
					},
				},
				"required": []string{"location"},
			},
		},
	}

	apiTool := convertToolDefToAPI(functionTool)
	require.NotNil(t, apiTool)
	
	assert.Equal(t, "function", apiTool["type"])
	function, ok := apiTool["function"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "get_weather", function["name"])
	assert.Equal(t, "Get current weather", function["description"])
	assert.NotNil(t, function["parameters"])

	// Test CustomToolDef conversion
	customTool := &models.CustomToolDef{
		Type: "custom",
		FunctionDef: models.FunctionDef{
			Name:        "search_web",
			Description: "Search the web",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "Search query",
					},
				},
			},
		},
	}

	apiCustomTool := convertToolDefToAPI(customTool)
	require.NotNil(t, apiCustomTool)
	
	assert.Equal(t, "function", apiCustomTool["type"])
	customFunction, ok := apiCustomTool["function"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "search_web", customFunction["name"])
	assert.Equal(t, "Search the web", customFunction["description"])
}

func TestSpeakeasyProvider_StructuredOutputIntegration(t *testing.T) {
	cfg := config.LLMConfig{
		Provider: "openai",
		Model:    config.LLMModelGPT4o,
		APIKey:   "test-api-key",
	}

	sharedBag := bag.NewSharedBag()
	provider := NewSpeakeasyProvider(cfg, sharedBag)
	session := provider.NewSession()

	// Add a user message
	session.Add(models.RoleUser, "Analyze this data")

	// Create a response format for structured output
	responseFormat := &models.ResponseFormat{
		Format: models.Format{
			Type: "json_schema",
			Name: "analysis_result",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"summary": map[string]any{
						"type":        "string",
						"description": "Summary of the analysis",
					},
					"score": map[string]any{
						"type":        "number",
						"description": "Confidence score",
					},
				},
				"required": []string{"summary", "score"},
			},
		},
	}

	// Test that structured output doesn't crash (will fail with network error in test env)
	_, err := session.Next(context.Background(), nil, responseFormat)
	
	// In test environment, we expect network errors, but the important thing is
	// that structured output parameters are processed correctly
	if err != nil {
		// Should fail with HTTP/network error, not parameter processing error
		assert.Contains(t, err.Error(), "direct HTTP call failed")
	}
}

func TestSpeakeasyProvider_MessageConversion(t *testing.T) {
	// Test normal message conversion
	normalMsg := createTestMessage("user", "Hello world")
	apiMsg := convertSpeakeasyMessageToAPI(normalMsg)
	
	assert.Equal(t, "user", apiMsg["role"])
	assert.Equal(t, "Hello world", apiMsg["content"])

	// Test tool result message conversion
	toolResultMsg := createTestMessage("user", "TOOL_RESULT[get_weather]: The weather is sunny")
	apiToolMsg := convertSpeakeasyMessageToAPI(toolResultMsg)
	
	assert.Equal(t, "tool", apiToolMsg["role"])
	assert.Equal(t, "The weather is sunny", apiToolMsg["content"])

	// Test malformed tool result (should fall back to normal conversion)
	malformedMsg := createTestMessage("user", "TOOL_RESULT[incomplete")
	apiMalformedMsg := convertSpeakeasyMessageToAPI(malformedMsg)
	
	assert.Equal(t, "user", apiMalformedMsg["role"])
	assert.Equal(t, "TOOL_RESULT[incomplete", apiMalformedMsg["content"])
}

func TestSpeakeasyProvider_ParseToolCall(t *testing.T) {
	// Test valid tool call parsing
	apiCall := map[string]any{
		"id":   "call_123",
		"type": "function",
		"function": map[string]any{
			"name":      "get_weather",
			"arguments": `{"location": "San Francisco"}`,
		},
	}

	toolCall := parseToolCall(apiCall)
	require.NotNil(t, toolCall)
	
	assert.Equal(t, "call_123", toolCall.ID)
	assert.Equal(t, "function", toolCall.Type)
	assert.Equal(t, "get_weather", toolCall.Function.Name)
	assert.Equal(t, `{"location": "San Francisco"}`, toolCall.Function.Arguments)

	// Test invalid tool call (missing function)
	invalidCall := map[string]any{
		"id":   "call_456",
		"type": "function",
	}

	invalidToolCall := parseToolCall(invalidCall)
	assert.Nil(t, invalidToolCall)
}

func TestSpeakeasyProvider_EmbeddingSupport(t *testing.T) {
	cfg := config.LLMConfig{
		Provider: "openai",
		Model:    config.LLMModelGPT4o,
		APIKey:   "test-api-key",
	}

	sharedBag := bag.NewSharedBag()
	provider := NewSpeakeasyProvider(cfg, sharedBag)

	// Test embedding request (will fail with network error in test env)
	_, err := provider.Embedding(context.Background(), "test text")
	
	// In test environment, we expect network errors
	if err != nil {
		assert.Contains(t, err.Error(), "embedding HTTP request failed")
	}
}

// Helper function to create test messages
func createTestMessage(role, content string) shared.ChatCompletionRequestMessage {
	var roleEnum shared.ChatCompletionRequestMessageRoleEnum
	switch role {
	case "system":
		roleEnum = shared.ChatCompletionRequestMessageRoleEnumSystem
	case "user":
		roleEnum = shared.ChatCompletionRequestMessageRoleEnumUser
	case "assistant":
		roleEnum = shared.ChatCompletionRequestMessageRoleEnumAssistant
	default:
		roleEnum = shared.ChatCompletionRequestMessageRoleEnumUser
	}

	return shared.ChatCompletionRequestMessage{
		Role:    roleEnum,
		Content: content,
	}
}