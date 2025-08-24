// internal/llm/openai/speakeasy_client.go
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

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
	// The old Speakeasy SDK doesn't have embedding support, so we'll make a direct HTTP call
	return p.getEmbeddingViaHTTP(ctx, text)
}

// getEmbeddingViaHTTP gets embeddings using direct HTTP call
func (p *SpeakeasyProvider) getEmbeddingViaHTTP(ctx context.Context, text string) ([]float64, error) {
	payload := map[string]any{
		"model": "text-embedding-ada-002", // Default embedding model
		"input": text,
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	baseURL := p.cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/v1/embeddings", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)

	httpClient := NewAuthenticatedClient(p.cfg.APIKey)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var embeddingResponse map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResponse); err != nil {
		return nil, fmt.Errorf("failed to decode embedding response: %w", err)
	}

	// Extract the embedding vector
	data, ok := embeddingResponse["data"].([]any)
	if !ok || len(data) == 0 {
		return nil, fmt.Errorf("no data in embedding response")
	}

	firstItem, ok := data[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid data format in embedding response")
	}

	embedding, ok := firstItem["embedding"].([]any)
	if !ok {
		return nil, fmt.Errorf("no embedding in response")
	}

	// Convert to float64 slice
	result := make([]float64, len(embedding))
	for i, val := range embedding {
		if floatVal, ok := val.(float64); ok {
			result[i] = floatVal
		} else {
			return nil, fmt.Errorf("invalid embedding value at index %d", i)
		}
	}

	return result, nil
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
	// Since the old Speakeasy SDK doesn't support tool role, we'll store tool results
	// as user messages with a special format that our direct HTTP call handler can understand
	toolMessage := shared.ChatCompletionRequestMessage{
		Role:    shared.ChatCompletionRequestMessageRoleEnumUser,
		Content: fmt.Sprintf("TOOL_RESULT[%s]: %s", toolCall.Function.Name, content),
	}
	s.messages = append(s.messages, toolMessage)
	
	slog.Debug("Added tool result to speakeasy session", 
		"tool_name", toolCall.Function.Name,
		"content_length", len(content),
	)
}

func (s *speakeasySession) AddFunctionCallResult(toolCall models.ToolCall, content string) {
	// For speakeasy session, treat function call results the same as tool results
	s.AddToolResult(toolCall, content)
	
	slog.Debug("Added function call result to speakeasy session", 
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
	slog.Debug("Creating chat completion with speakeasy SDK",
		"message_count", len(s.messages),
		"model", s.p.cfg.Model.String(),
		"tools_count", len(tools),
		"response_format", rf != nil,
	)

	// If tools or structured output are needed, use direct HTTP call
	// since Speakeasy SDK v1.11.0 doesn't support these features
	if len(tools) > 0 || rf != nil {
		return s.nextWithModernFeatures(ctx, tools, rf)
	}

	// Use Speakeasy SDK for basic chat completion without tools
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
		Content:   choice.Message.Content,
		ToolCalls: nil, // No tool calls in basic mode
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

// nextWithModernFeatures handles chat completion with tools and structured output
// using direct HTTP calls since Speakeasy SDK v1.11.0 doesn't support these features
func (s *speakeasySession) nextWithModernFeatures(ctx context.Context, tools []models.ToolDef, rf *models.ResponseFormat) (*models.AssistantTurn, error) {
	slog.Debug("Using direct HTTP calls for modern OpenAI features",
		"tools_count", len(tools),
		"has_response_format", rf != nil,
	)

	// Convert messages to OpenAI API format
	apiMessages := make([]map[string]any, len(s.messages))
	for i, msg := range s.messages {
		apiMessages[i] = convertSpeakeasyMessageToAPI(msg)
	}

	// Build the request payload
	payload := map[string]any{
		"model":    s.p.cfg.Model.String(),
		"messages": apiMessages,
	}

	// Add optional parameters
	if s.p.cfg.OpenAI.MaxCompletionTokens > 0 {
		payload["max_tokens"] = s.p.cfg.OpenAI.MaxCompletionTokens
	}
	if s.p.cfg.OpenAI.Temperature != nil {
		payload["temperature"] = *s.p.cfg.OpenAI.Temperature
	}

	// Add tools if provided
	if len(tools) > 0 {
		apiTools := make([]map[string]any, 0, len(tools))
		for _, tool := range tools {
			apiTool := convertToolDefToAPI(tool)
			if apiTool != nil {
				apiTools = append(apiTools, apiTool)
			}
		}
		if len(apiTools) > 0 {
			payload["tools"] = apiTools
		}
	}

	// Add structured output if provided
	if rf != nil {
		payload["response_format"] = map[string]any{
			"type":   rf.Format.Type,
			"schema": rf.Format.Schema,
		}
		if rf.Format.Name != "" {
			responseFormat := payload["response_format"].(map[string]any)
			responseFormat["name"] = rf.Format.Name
		}
	}

	// Make direct HTTP call using the authenticated client
	result, err := s.makeDirectHTTPCall(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("direct HTTP call failed: %w", err)
	}

	// Add the assistant's response to our message history
	s.Add(models.RoleAssistant, result.Content)

	return result, nil
}

// Helper functions for converting between formats

// convertSpeakeasyMessageToAPI converts a speakeasy message to OpenAI API format
func convertSpeakeasyMessageToAPI(msg shared.ChatCompletionRequestMessage) map[string]any {
	role := string(msg.Role)
	content := msg.Content
	
	// Handle special tool result format for direct API calls
	if role == "user" && strings.HasPrefix(content, "TOOL_RESULT[") {
		// Extract tool name and content for proper API format
		if endBracket := strings.Index(content, "]: "); endBracket != -1 {
			toolContent := content[endBracket+3:]
			return map[string]any{
				"role":    "tool", // Use proper tool role for API
				"content": toolContent,
			}
		}
	}
	
	return map[string]any{
		"role":    role,
		"content": content,
	}
}

// convertToolDefToAPI converts our ToolDef to OpenAI API format
func convertToolDefToAPI(tool models.ToolDef) map[string]any {
	switch t := tool.(type) {
	case *models.FunctionToolDef:
		return map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        t.Function.Name,
				"description": t.Function.Description,
				"parameters":  t.Function.Parameters,
			},
		}
	case *models.CustomToolDef:
		return map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        t.Name,
				"description": t.Description,
				"parameters":  t.Parameters,
			},
		}
	default:
		slog.Debug("Unknown tool type, skipping", "type", fmt.Sprintf("%T", tool))
		return nil
	}
}

// makeDirectHTTPCall makes a direct HTTP call to OpenAI API
func (s *speakeasySession) makeDirectHTTPCall(ctx context.Context, payload map[string]any) (*models.AssistantTurn, error) {
	// Need to import required packages first
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	baseURL := s.p.cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/v1/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.p.cfg.APIKey)

	// Make the request using the authenticated client
	httpClient := NewAuthenticatedClient(s.p.cfg.APIKey)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body) 
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the response
	var apiResponse map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract the response content and tool calls
	return s.parseAPIResponse(apiResponse)
}

// parseAPIResponse parses the OpenAI API response into our AssistantTurn format
func (s *speakeasySession) parseAPIResponse(apiResponse map[string]any) (*models.AssistantTurn, error) {
	choices, ok := apiResponse["choices"].([]any)
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("no choices in API response")
	}

	choice, ok := choices[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid choice format")
	}

	message, ok := choice["message"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid message format")
	}

	// Extract content
	content, _ := message["content"].(string)
	
	// Extract tool calls if present
	var toolCalls []models.ToolCall
	if apiToolCalls, ok := message["tool_calls"].([]any); ok {
		toolCalls = make([]models.ToolCall, 0, len(apiToolCalls))
		for _, tc := range apiToolCalls {
			if toolCall, ok := tc.(map[string]any); ok {
				if parsedCall := parseToolCall(toolCall); parsedCall != nil {
					toolCalls = append(toolCalls, *parsedCall)
				}
			}
		}
	}

	return &models.AssistantTurn{
		Content:   content,
		ToolCalls: toolCalls,
	}, nil
}

// parseToolCall converts an API tool call to our ToolCall format
func parseToolCall(apiCall map[string]any) *models.ToolCall {
	id, _ := apiCall["id"].(string)
	callType, _ := apiCall["type"].(string)
	
	function, ok := apiCall["function"].(map[string]any)
	if !ok {
		return nil
	}
	
	name, _ := function["name"].(string)
	arguments, _ := function["arguments"].(string)
	
	return &models.ToolCall{
		ID:   id,
		Type: callType,
		Function: models.ToolCallFunction{
			Name:      name,
			Arguments: arguments,
		},
	}
}