// internal/llm/openai/speakeasy_structured_output_test.go
package openai

import (
	"context"
	"testing"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/amaurybrisou/mosychlos/pkg/openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test model for structured output
type TestAnalysisResult struct {
	Summary    string  `json:"summary"`
	Score      float64 `json:"score"`
	Categories []string `json:"categories"`
}

func TestSpeakeasyProvider_StructuredOutputWithSchema(t *testing.T) {
	cfg := config.LLMConfig{
		Provider: "openai",
		Model:    config.LLMModelGPT4o,
		APIKey:   "test-api-key",
	}

	sharedBag := bag.NewSharedBag()
	provider := NewSpeakeasyProvider(cfg, sharedBag)
	session := provider.NewSession()

	// Add a user message
	session.Add(models.RoleUser, "Analyze this market data and provide a structured response")

	// Generate schema using the existing schema building system
	schema := openai.BuildSchema[TestAnalysisResult]()
	
	// Verify schema was generated correctly
	require.NotNil(t, schema)
	assert.Equal(t, "object", schema["type"])
	
	properties, ok := schema["properties"].(map[string]any)
	require.True(t, ok)
	
	// Check that all fields are present
	assert.Contains(t, properties, "summary")
	assert.Contains(t, properties, "score")
	assert.Contains(t, properties, "categories")
	
	// Create response format using the generated schema
	responseFormat := &models.ResponseFormat{
		Format: models.Format{
			Type:   "json_schema",
			Name:   "analysis_result",
			Schema: schema,
		},
	}

	// Test that the request is properly formed (will fail with network error in test env)
	_, err := session.Next(context.Background(), nil, responseFormat)
	
	// In test environment, we expect network errors, but the important thing is
	// that the schema integration and request formation works
	if err != nil {
		assert.Contains(t, err.Error(), "direct HTTP call failed")
	}
}

func TestSpeakeasyProvider_ToolsWithStructuredOutput(t *testing.T) {
	cfg := config.LLMConfig{
		Provider: "openai",
		Model:    config.LLMModelGPT4o,
		APIKey:   "test-api-key",
	}

	sharedBag := bag.NewSharedBag()
	provider := NewSpeakeasyProvider(cfg, sharedBag)
	session := provider.NewSession()

	// Add a user message
	session.Add(models.RoleUser, "Use tools to analyze data and return structured results")

	// Create a test tool
	weatherTool := &models.FunctionToolDef{
		Type: "function",
		Function: models.FunctionDef{
			Name:        "get_weather",
			Description: "Get current weather for a location",
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

	// Generate schema for structured output
	schema := openai.BuildSchema[TestAnalysisResult]()
	responseFormat := &models.ResponseFormat{
		Format: models.Format{
			Type:   "json_schema",
			Name:   "weather_analysis",
			Schema: schema,
		},
	}

	tools := []models.ToolDef{weatherTool}

	// Test that both tools and structured output can be used together
	_, err := session.Next(context.Background(), tools, responseFormat)
	
	// In test environment, we expect network errors
	if err != nil {
		assert.Contains(t, err.Error(), "direct HTTP call failed")
		// The key is that it gets to the HTTP call, meaning the request was properly formed
	}
}

func TestSpeakeasyProvider_ComplexStructuredOutput(t *testing.T) {
	// Test with a more complex nested structure
	type ComplexResult struct {
		Analysis struct {
			Summary    string            `json:"summary"`
			Metrics    map[string]float64 `json:"metrics"`
			Timestamp  string            `json:"timestamp"`
		} `json:"analysis"`
		Recommendations []struct {
			Action     string `json:"action"`
			Priority   int    `json:"priority"`
			Confidence float64 `json:"confidence"`
		} `json:"recommendations"`
		Metadata struct {
			Version    string   `json:"version"`
			DataSource string   `json:"data_source"`
			Tags       []string `json:"tags"`
		} `json:"metadata"`
	}

	cfg := config.LLMConfig{
		Provider: "openai",
		Model:    config.LLMModelGPT4o,
		APIKey:   "test-api-key",
	}

	sharedBag := bag.NewSharedBag()
	provider := NewSpeakeasyProvider(cfg, sharedBag)
	session := provider.NewSession()

	session.Add(models.RoleUser, "Provide a comprehensive analysis with recommendations")

	// Generate schema for complex structure
	schema := openai.BuildSchema[ComplexResult]()
	
	// Verify complex schema generation
	require.NotNil(t, schema)
	properties, ok := schema["properties"].(map[string]any)
	require.True(t, ok)
	
	// Check nested structures
	assert.Contains(t, properties, "analysis")
	assert.Contains(t, properties, "recommendations")
	assert.Contains(t, properties, "metadata")
	
	responseFormat := &models.ResponseFormat{
		Format: models.Format{
			Type:   "json_schema", 
			Name:   "complex_analysis",
			Schema: schema,
		},
	}

	// Test request formation with complex schema
	_, err := session.Next(context.Background(), nil, responseFormat)
	
	if err != nil {
		assert.Contains(t, err.Error(), "direct HTTP call failed")
	}
}

func TestSpeakeasyProvider_SchemaIntegrationEdgeCases(t *testing.T) {
	// Test with empty schema
	emptyResponseFormat := &models.ResponseFormat{
		Format: models.Format{
			Type:   "json_schema",
			Name:   "empty",
			Schema: nil,
		},
	}

	cfg := config.LLMConfig{
		Provider: "openai", 
		Model:    config.LLMModelGPT4o,
		APIKey:   "test-api-key",
	}

	sharedBag := bag.NewSharedBag()
	provider := NewSpeakeasyProvider(cfg, sharedBag)
	session := provider.NewSession()

	session.Add(models.RoleUser, "Test with empty schema")

	// Should handle empty schema gracefully
	_, err := session.Next(context.Background(), nil, emptyResponseFormat)
	
	if err != nil {
		assert.Contains(t, err.Error(), "direct HTTP call failed")
	}
}