package openaiutils

import (
	"encoding/json"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// BatchResponse represents the structure of a batch response from OpenAI
type BatchResponse struct {
	ID       string                 `json:"id"`
	CustomID string                 `json:"custom_id"`
	Response BatchResponseBody      `json:"response,omitempty"`
	Error    map[string]interface{} `json:"error,omitempty"`
}

type BatchResponseBody struct {
	StatusCode int                    `json:"status_code"`
	RequestID  string                 `json:"request_id"`
	Body       ChatCompletionResponse `json:"body"`
}

type ChatCompletionResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Choices []CompletionChoice `json:"choices"`
	Usage   *models.Usage      `json:"usage"`
}

type CompletionChoice struct {
	Index        int                   `json:"index"`
	Message      ChatCompletionMessage `json:"message"`
	FinishReason string                `json:"finish_reason"`
}

type ChatCompletionMessage struct {
	Role      string            `json:"role"`
	Content   *string           `json:"content"`
	ToolCalls []models.ToolCall `json:"tool_calls,omitempty"`
}

// ExtractToolCallsFromBatchResponse parses a batch response and extracts tool calls
func ExtractToolCallsFromBatchResponse(content []byte) ([]models.ToolCall, string, error) {
	var batchResp BatchResponse
	if err := json.Unmarshal(content, &batchResp); err != nil {
		return nil, "", err
	}

	// Check if there's an error in the response
	if batchResp.Error != nil {
		return nil, "", nil // No tool calls if there's an error
	}

	// Check if there are choices
	if len(batchResp.Response.Body.Choices) == 0 {
		return nil, "", nil
	}

	choice := batchResp.Response.Body.Choices[0]
	message := choice.Message

	// Extract content
	content_str := ""
	if message.Content != nil {
		content_str = *message.Content
	}

	// Return tool calls if they exist
	if len(message.ToolCalls) > 0 {
		return message.ToolCalls, content_str, nil
	}

	return nil, content_str, nil
}

// IsToolCall identifies tool calls in the batch response content
// Returns tool name, arguments, and whether it was a tool call
// This is a legacy function for backward compatibility
func IsToolCall(content []byte) (string, string, bool) {
	toolCalls, _, err := ExtractToolCallsFromBatchResponse(content)
	if err != nil || len(toolCalls) == 0 {
		return "", "", false
	}

	// Return the first tool call for backward compatibility
	firstCall := toolCalls[0]
	return firstCall.Function.Name, firstCall.Function.Arguments, true
}

// ParseBatchContent parses batch response content and returns structured data
func ParseBatchContent(content []byte) (*BatchContentResult, error) {
	toolCalls, textContent, err := ExtractToolCallsFromBatchResponse(content)
	if err != nil {
		return nil, err
	}

	return &BatchContentResult{
		Content:   textContent,
		ToolCalls: toolCalls,
		HasTools:  len(toolCalls) > 0,
	}, nil
}

// BatchContentResult represents parsed batch response content
type BatchContentResult struct {
	Content   string            `json:"content"`
	ToolCalls []models.ToolCall `json:"tool_calls"`
	HasTools  bool              `json:"has_tools"`
}
