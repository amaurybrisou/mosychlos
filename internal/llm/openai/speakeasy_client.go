// internal/llm/openai/speakeasy_client.go
package openai

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	openaigosdk "github.com/speakeasy-sdks/openai-go-sdk/v4"
	"github.com/speakeasy-sdks/openai-go-sdk/v4/pkg/models/shared"
)

// SpeakeasyProvider implements the Provider interface using the speakeasy SDK v4 entirely
type SpeakeasyProvider struct {
	name      string
	client    *openaigosdk.Gpt
	cfg       config.LLMConfig
	sharedBag bag.SharedBag
}

// NewSpeakeasyProvider creates a new provider using the speakeasy SDK v4 with built-in authentication
func NewSpeakeasyProvider(cfg config.LLMConfig, sharedBag bag.SharedBag) *SpeakeasyProvider {
	// Configure speakeasy client with proper authentication
	opts := []openaigosdk.SDKOption{
		openaigosdk.WithSecurity(cfg.APIKey),
	}
	
	// Set custom base URL if configured
	if cfg.BaseURL != "" {
		opts = append(opts, openaigosdk.WithServerURL(cfg.BaseURL))
	}
	
	client := openaigosdk.New(opts...)
	
	return &SpeakeasyProvider{
		name:      "openai-speakeasy-v4",
		client:    client,
		cfg:       cfg,
		sharedBag: sharedBag,
	}
}

func (p *SpeakeasyProvider) Name() string { return p.name }

func (p *SpeakeasyProvider) Embedding(ctx context.Context, text string) ([]float64, error) {
	// Use Speakeasy SDK v4 for embeddings
	request := shared.CreateEmbeddingRequest{
		Input: shared.CreateInputStr(text),
		Model: shared.CreateCreateEmbeddingRequestModelStr("text-embedding-ada-002"), // Default embedding model
	}

	response, err := p.client.Embeddings.CreateEmbedding(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("speakeasy v4 embedding request failed: %w", err)
	}

	if response == nil || response.CreateEmbeddingResponse == nil {
		return nil, fmt.Errorf("empty response from speakeasy v4 embedding API")
	}

	embeddingResp := response.CreateEmbeddingResponse
	if len(embeddingResp.Data) == 0 {
		return nil, fmt.Errorf("no data in embedding response")
	}

	// Extract the embedding vector from the first result - it's already []float64
	embedding := embeddingResp.Data[0].Embedding
	if len(embedding) == 0 {
		return nil, fmt.Errorf("empty embedding vector")
	}

	return embedding, nil
}

func (p *SpeakeasyProvider) NewSession() models.Session {
	return &speakeasySession{p: p, messages: make([]shared.ChatCompletionRequestMessage, 0, 8)}
}

type speakeasySession struct {
	p          *SpeakeasyProvider
	messages   []shared.ChatCompletionRequestMessage
	toolChoice *models.ToolChoice
}

func (s *speakeasySession) Add(role models.Role, content string) {
	// Convert our Role type to the speakeasy v4 message format
	var message shared.ChatCompletionRequestMessage
	
	switch role {
	case models.RoleUser:
		message = shared.CreateChatCompletionRequestMessageChatCompletionRequestUserMessage(
			shared.ChatCompletionRequestUserMessage{
				Content: shared.CreateContentStr(content),
				Role:    shared.ChatCompletionRequestUserMessageRoleUser,
			},
		)
	case models.RoleAssistant:
		message = shared.CreateChatCompletionRequestMessageChatCompletionRequestAssistantMessage(
			shared.ChatCompletionRequestAssistantMessage{
				Content: openaigosdk.String(content),
				Role:    shared.RoleAssistant,
			},
		)
	case models.RoleSystem:
		message = shared.CreateChatCompletionRequestMessageChatCompletionRequestSystemMessage(
			shared.ChatCompletionRequestSystemMessage{
				Content: content, // System messages use string directly
				Role:    shared.ChatCompletionRequestSystemMessageRoleSystem,
			},
		)
	default:
		// Default to user
		message = shared.CreateChatCompletionRequestMessageChatCompletionRequestUserMessage(
			shared.ChatCompletionRequestUserMessage{
				Content: shared.CreateContentStr(content),
				Role:    shared.ChatCompletionRequestUserMessageRoleUser,
			},
		)
	}
	
	s.messages = append(s.messages, message)
}

func (s *speakeasySession) AddToolResult(toolCall models.ToolCall, content string) {
	// Use proper tool message with Speakeasy SDK v4
	message := shared.CreateChatCompletionRequestMessageChatCompletionRequestToolMessage(
		shared.ChatCompletionRequestToolMessage{
			Content:    content,
			Role:       shared.ChatCompletionRequestToolMessageRoleTool,
			ToolCallID: toolCall.CallID,
		},
	)
	s.messages = append(s.messages, message)
	
	slog.Debug("Added tool result to speakeasy v4 session", 
		"tool_call_id", toolCall.CallID,
		"tool_name", toolCall.Function.Name,
		"content_length", len(content),
	)
}

func (s *speakeasySession) AddFunctionCallResult(toolCall models.ToolCall, content string) {
	// For speakeasy v4 session, use tool message for function call results
	s.AddToolResult(toolCall, content)
	
	slog.Debug("Added function call result to speakeasy v4 session", 
		"function_name", toolCall.Function.Name,
		"content_length", len(content),
	)
}

func (s *speakeasySession) NextStream(ctx context.Context, tools []models.ToolDef, rf *models.ResponseFormat) (<-chan models.StreamChunk, error) {
	// The old Speakeasy SDK doesn't support streaming properly
	// We'll provide a basic implementation that returns the complete response as a single chunk
	slog.Debug("Speakeasy streaming support is limited - using non-streaming fallback")
	
	// Get the complete response
	result, err := s.Next(ctx, tools, rf)
	if err != nil {
		return nil, err
	}

	// Create a channel and send the result as a single chunk
	ch := make(chan models.StreamChunk, 1)
	go func() {
		defer close(ch)
		ch <- models.StreamChunk{
			Content:      result.Content,
			ToolCalls:    result.ToolCalls,
			IsComplete:   true,
			Error:        nil,
			FinishReason: stringPtr("stop"),
		}
	}()

	return ch, nil
}

// stringPtr creates a pointer to a string
func stringPtr(s string) *string {
	return &s
}

func (s *speakeasySession) SetToolChoice(t *models.ToolChoice) {
	s.toolChoice = t
}

func (s *speakeasySession) Next(ctx context.Context, tools []models.ToolDef, rf *models.ResponseFormat) (*models.AssistantTurn, error) {
	slog.Debug("Creating chat completion with speakeasy SDK v4",
		"message_count", len(s.messages),
		"model", s.p.cfg.Model.String(),
		"tools_count", len(tools),
		"response_format", rf != nil,
	)

	// Build the request using the SDK entirely
	request := shared.CreateChatCompletionRequest{
		Model:    shared.CreateCreateChatCompletionRequestModelStr(s.p.cfg.Model.String()),
		Messages: s.messages,
	}

	// Add optional parameters from config
	if s.p.cfg.OpenAI.MaxCompletionTokens > 0 {
		maxTokens := int64(s.p.cfg.OpenAI.MaxCompletionTokens)
		request.MaxTokens = &maxTokens
	}

	if s.p.cfg.OpenAI.Temperature != nil {
		request.Temperature = s.p.cfg.OpenAI.Temperature
	}

	// Add tools if provided - now fully supported by v4
	if len(tools) > 0 {
		sdkTools := make([]shared.ChatCompletionTool, 0, len(tools))
		for _, tool := range tools {
			if sdkTool := convertToolDefToSDK(tool); sdkTool != nil {
				sdkTools = append(sdkTools, *sdkTool)
			}
		}
		if len(sdkTools) > 0 {
			request.Tools = sdkTools
		}
	}

	// Add structured output if provided - now fully supported by v4
	if rf != nil {
		// Convert our ResponseFormat to SDK format
		if sdkFormat := convertResponseFormatToSDK(rf); sdkFormat != nil {
			request.ResponseFormat = sdkFormat
		}
	}

	// Make the API call using the SDK
	response, err := s.p.client.Chat.CreateChatCompletion(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("speakeasy v4 chat completion failed: %w", err)
	}

	if response == nil || response.CreateChatCompletionResponse == nil {
		return nil, fmt.Errorf("empty response from speakeasy v4 API")
	}

	chatResponse := response.CreateChatCompletionResponse
	if len(chatResponse.Choices) == 0 {
		return nil, fmt.Errorf("no choices in speakeasy v4 API response")
	}

	choice := chatResponse.Choices[0]
	message := choice.GetMessage()

	// Convert the response back to our format
	result := &models.AssistantTurn{
		Content: "",
	}

	// Handle content - may be empty if there are tool calls
	if message.Content != nil {
		result.Content = *message.Content
	}

	// Handle tool calls if present
	if len(message.ToolCalls) > 0 {
		result.ToolCalls = make([]models.ToolCall, len(message.ToolCalls))
		for i, tc := range message.ToolCalls {
			result.ToolCalls[i] = models.ToolCall{
				ID:     tc.ID,
				CallID: tc.ID, // Use ID as CallID for compatibility
				Type:   string(tc.Type),
				Function: models.ToolCallFunction{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
	}

	// Add the assistant's response to our message history
	s.Add(models.RoleAssistant, result.Content)

	finishReason := string(choice.FinishReason)

	slog.Debug("Speakeasy v4 chat completion successful",
		"response_length", len(result.Content),
		"tool_calls", len(result.ToolCalls),
		"finish_reason", finishReason,
	)

	return result, nil
}

// Helper functions for converting between our models and SDK models

// convertToolDefToSDK converts our ToolDef to SDK ChatCompletionTool format
func convertToolDefToSDK(tool models.ToolDef) *shared.ChatCompletionTool {
	switch t := tool.(type) {
	case *models.FunctionToolDef:
		return &shared.ChatCompletionTool{
			Type: shared.ChatCompletionToolTypeFunction,
			Function: shared.FunctionObject{
				Name:        t.Function.Name,
				Description: openaigosdk.String(t.Function.Description),
				Parameters:  t.Function.Parameters, // Already map[string]any
			},
		}
	case *models.CustomToolDef:
		return &shared.ChatCompletionTool{
			Type: shared.ChatCompletionToolTypeFunction,
			Function: shared.FunctionObject{
				Name:        t.Name,
				Description: openaigosdk.String(t.Description),
				Parameters:  t.Parameters, // Already map[string]any
			},
		}
	default:
		slog.Debug("Unknown tool type, skipping", "type", fmt.Sprintf("%T", tool))
		return nil
	}
}

// convertResponseFormatToSDK converts our ResponseFormat to SDK format
func convertResponseFormatToSDK(rf *models.ResponseFormat) *shared.ResponseFormat {
	if rf == nil {
		return nil
	}

	var formatType shared.CreateChatCompletionRequestType
	switch rf.Format.Type {
	case "json_object":
		formatType = shared.CreateChatCompletionRequestTypeJSONObject
	default:
		formatType = shared.CreateChatCompletionRequestTypeText
	}

	return &shared.ResponseFormat{
		Type: &formatType,
	}
}