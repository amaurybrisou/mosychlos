// internal/llm/openai/speakeasy_client.go
package openai

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/speakeasy-sdks/openai-go-sdk"
	"github.com/speakeasy-sdks/openai-go-sdk/pkg/models/shared"
)

// AuthenticatedClient wraps an HTTP client to add API key authentication
type AuthenticatedClient struct {
	client http.Client
	apiKey string
}

func NewAuthenticatedClient(apiKey string) *AuthenticatedClient {
	return &AuthenticatedClient{
		client: http.Client{},
		apiKey: apiKey,
	}
}

func (c *AuthenticatedClient) Do(req *http.Request) (*http.Response, error) {
	// Add the Authorization header with the API key
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	return c.client.Do(req)
}

// SpeakeasyProvider implements the Provider interface using the speakeasy SDK
type SpeakeasyProvider struct {
	name      string
	client    *gpt.Gpt
	cfg       config.LLMConfig
	sharedBag bag.SharedBag
}

// NewSpeakeasyProvider creates a new provider using the speakeasy SDK
func NewSpeakeasyProvider(cfg config.LLMConfig, sharedBag bag.SharedBag) *SpeakeasyProvider {
	// Create an authenticated HTTP client
	httpClient := NewAuthenticatedClient(cfg.APIKey)
	
	// Configure speakeasy client with authentication and base URL
	opts := []gpt.SDKOption{
		gpt.WithClient(httpClient),
	}
	
	// Set custom base URL if configured
	if cfg.BaseURL != "" {
		opts = append(opts, gpt.WithServerURL(cfg.BaseURL))
	}
	
	client := gpt.New(opts...)
	
	return &SpeakeasyProvider{
		name:      "openai-speakeasy",
		client:    client,
		cfg:       cfg,
		sharedBag: sharedBag,
	}
}

func (p *SpeakeasyProvider) Name() string { return p.name }

func (p *SpeakeasyProvider) Embedding(ctx context.Context, text string) ([]float64, error) {
	// TODO: Implement embedding using speakeasy SDK
	return nil, fmt.Errorf("embedding not implemented for speakeasy provider")
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
	// Convert our Role type to the speakeasy message format
	var speakeasyRole shared.ChatCompletionRequestMessageRoleEnum
	switch role {
	case models.RoleUser:
		speakeasyRole = shared.ChatCompletionRequestMessageRoleEnumUser
	case models.RoleAssistant:
		speakeasyRole = shared.ChatCompletionRequestMessageRoleEnumAssistant
	case models.RoleSystem:
		speakeasyRole = shared.ChatCompletionRequestMessageRoleEnumSystem
	default:
		speakeasyRole = shared.ChatCompletionRequestMessageRoleEnumUser
	}

	message := shared.ChatCompletionRequestMessage{
		Role:    speakeasyRole,
		Content: content,
	}
	
	s.messages = append(s.messages, message)
}

func (s *speakeasySession) AddToolResult(toolCall models.ToolCall, content string) {
	// TODO: Implement tool result handling for speakeasy SDK
	slog.Debug("AddToolResult not fully implemented for speakeasy provider", "tool_name", toolCall.Function.Name)
}

func (s *speakeasySession) AddFunctionCallResult(toolCall models.ToolCall, content string) {
	// TODO: Implement function call result handling for speakeasy SDK
	slog.Debug("AddFunctionCallResult not fully implemented for speakeasy provider", "tool_name", toolCall.Function.Name)
}

func (s *speakeasySession) NextStream(ctx context.Context, tools []models.ToolDef, rf *models.ResponseFormat) (<-chan models.StreamChunk, error) {
	// TODO: Implement streaming for speakeasy SDK
	return nil, fmt.Errorf("streaming not implemented for speakeasy provider")
}

func (s *speakeasySession) SetToolChoice(t *models.ToolChoice) {
	s.toolChoice = t
}

func (s *speakeasySession) Next(ctx context.Context, tools []models.ToolDef, rf *models.ResponseFormat) (*models.AssistantTurn, error) {
	slog.Debug("Creating chat completion with speakeasy SDK",
		"message_count", len(s.messages),
		"model", s.p.cfg.Model.String(),
	)

	// Build the chat completion request
	request := shared.CreateChatCompletionRequest{
		Model:    s.p.cfg.Model.String(),
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

	// TODO: Convert tools to speakeasy format
	// if len(tools) > 0 {
	//     request.Tools = convertToolsToSpeakeasyFormat(tools)
	// }

	// Make the API call
	response, err := s.p.client.OpenAI.CreateChatCompletion(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("speakeasy chat completion failed: %w", err)
	}

	if response == nil || response.CreateChatCompletionResponse == nil {
		return nil, fmt.Errorf("empty response from speakeasy API")
	}

	chatResponse := response.CreateChatCompletionResponse
	if len(chatResponse.Choices) == 0 {
		return nil, fmt.Errorf("no choices in speakeasy API response")
	}

	choice := chatResponse.Choices[0]
	if choice.Message == nil || choice.Message.Content == "" {
		return nil, fmt.Errorf("empty message content in speakeasy API response")
	}

	// Convert the response back to our format
	result := &models.AssistantTurn{
		Content: choice.Message.Content,
		// TODO: Handle tool calls if present
		ToolCalls: nil,
	}

	// Add the assistant's response to our message history
	s.Add(models.RoleAssistant, result.Content)

	finishReason := ""
	if choice.FinishReason != nil {
		finishReason = *choice.FinishReason
	}

	slog.Debug("Speakeasy chat completion successful",
		"response_length", len(result.Content),
		"finish_reason", finishReason,
	)

	return result, nil
}

// TODO: Add helper functions to convert between formats
// func convertToolsToSpeakeasyFormat(tools []models.ToolDef) []shared.Tool {
//     // Implementation to convert our tool definitions to speakeasy format
//     return nil
// }