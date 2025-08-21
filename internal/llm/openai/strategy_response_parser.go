package openai

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/openai/openai-go/v2/responses"
)

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// processResponsesAPIResult processes the response from the Responses API
func (s *session) processResponsesAPIResult(resp *responses.Response, startTime time.Time) (*models.AssistantTurn, error) {
	// Extract the content from the response
	content := map[string]any{}
	toolCalls := []models.ToolCall{}
	var mainResponseText string // Store the main AI response text

	// Process the response items
	slog.Debug("Processing OpenAI response items",
		"total_items", len(resp.Output),
		"response_id", resp.ID)

	for i, item := range resp.Output {
		slog.Debug("Processing response item",
			"index", i,
			"type", item.Type)

		// Any of "message", "file_search_call", "function_call", "web_search_call",
		// "computer_call", "reasoning", "image_generation_call", "code_interpreter_call",
		// "local_shell_call", "mcp_call", "mcp_list_tools", "mcp_approval_request",
		// "custom_tool_call".
		switch item.Type {
		case "message":
			// Extract text content from output message
			outputMsg := item.AsMessage()
			slog.Debug("Processing message content",
				"content_items", len(outputMsg.Content))

			for j, contentItem := range outputMsg.Content {
				slog.Debug("Processing content item",
					"item_index", j,
					"content_type", contentItem.Type)

				switch contentItem.Type {
				case "output_text":
					outputText := contentItem.AsOutputText()
					content[fmt.Sprintf("output_text_%d_%d", i, j)] = outputText.Text
					// Store the main response text - this is the AI's actual analysis
					if mainResponseText == "" {
						mainResponseText = outputText.Text
						slog.Debug("Found main response text",
							"length", len(mainResponseText),
							"preview", mainResponseText[:min(100, len(mainResponseText))])
					}
				case "refusal":
					content[fmt.Sprintf("refusal_%d_%d", i, j)] = contentItem.Refusal
					slog.Debug("Found refusal content", "refusal", contentItem.Refusal)
				default:
					slog.Debug("Unknown content type", "type", contentItem.Type)
				}
			}

		case "web_search_call":
			// Handle web search calls.
			// Here can only track the status.
			searchCall := item.AsWebSearchCall()
			s.processWebSearchCall(searchCall, startTime)
		case "function_call":
			// Handle function tool calls
			funcCall := item.AsFunctionCall()
			slog.Debug("Function call detected",
				"status", funcCall.Status,
				"function_name", funcCall.Name,
				"call_id", funcCall.CallID,
				"arguments", funcCall.Arguments,
			)

			switch funcCall.Name {
			case keys.WebSearch.String():
				slog.Debug("Web search function call - processing separately")
				// Don't add web search calls to toolCalls as they are handled differently
			default:
				// For other custom function calls executed by OpenAI internally
				// Add to toolCalls for tracking
				tc := models.ToolCall{
					ID:     funcCall.ID,
					CallID: funcCall.CallID,
					Type:   "function",
					Status: string(funcCall.Status),
					Function: models.ToolCallFunction{
						Name:      funcCall.Name,
						Arguments: funcCall.Arguments,
					},
				}

				s.AddFunctionCall(tc)
				toolCalls = append(toolCalls, tc)
			}
		case "custom_tool_call":
			// Handle tool calls
			toolCall := item.AsCustomToolCall()
			slog.Debug("Custom tool call detected",
				"call_id", toolCall.CallID,
				"function_name", toolCall.Name,
				"arguments", toolCall.Input,
			)

			tc := models.ToolCall{
				ID:     toolCall.ID,
				CallID: toolCall.CallID,
				Type:   string(toolCall.Type),
				Function: models.ToolCallFunction{
					Name:      toolCall.Name,
					Arguments: toolCall.Input,
				},
			}

			s.AddToolCall(tc)
			toolCalls = append(toolCalls, tc)
		}
	}

	// Extract and track token usage from OpenAI response
	if resp.Usage.TotalTokens > 0 {
		usage := models.Usage{
			PromptTokens:     int(resp.Usage.InputTokens),
			CompletionTokens: int(resp.Usage.OutputTokens),
			InputTokens:      int(resp.Usage.InputTokens),
			OutputTokens:     int(resp.Usage.OutputTokens),
			TotalTokens:      int(resp.Usage.TotalTokens),
		}

		// Track OpenAI API usage with token consumption
		s.trackOpenAITokenUsage(startTime, usage, len(toolCalls), nil)
	}

	// Use the main response text if available, otherwise fall back to JSON content
	var finalContent string
	if mainResponseText != "" {
		finalContent = mainResponseText
		slog.Debug("Using main response text as final content",
			"length", len(finalContent))
	} else {
		contentBytes, err := json.Marshal(content)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal content: %w", err)
		}
		finalContent = string(contentBytes)
		slog.Debug("Using JSON marshaled content as final content",
			"length", len(finalContent),
			"content_keys", len(content))
	}

	slog.Debug("Returning AssistantTurn",
		"final_content_length", len(finalContent),
		"tool_calls_count", len(toolCalls),
		"final_content_preview", finalContent[:min(200, len(finalContent))])

	return &models.AssistantTurn{
		Content:   finalContent,
		ToolCalls: toolCalls,
	}, nil
}

// processWebSearchCall handles web search processing without token waste
func (s *session) processWebSearchCall(searchCall responses.ResponseFunctionWebSearch, startTime time.Time) {
	// Extract query from search call (type-safe extraction would go here)
	// Track web search usage with citation metrics
	trackingDetails := map[string]any{
		"action": keys.WebSearch,
		"url":    searchCall.Action.URL,
		"query":  searchCall.Action.Query,
		"type":   searchCall.Action.Type,
	}

	var citationResult *CitationResult
	// Any of "in_progress", "searching", "completed", "failed".
	switch searchCall.Type {
	case "in_progress":
		trackingDetails["status"] = "in_progress"
		trackingDetails["query_count"] = 1 // Track query count
		s.trackWebSearchComputation("in_progress", startTime, trackingDetails, nil)
	case "searching":
		trackingDetails["status"] = "searching"
		trackingDetails["query_count"] = 1 // Track query count
		s.trackWebSearchComputation("searching", startTime, trackingDetails, nil)
	case "completed":
		processor := NewCitationProcessor(s.p.sharedBag)
		var err error
		citationResult, err = processor.ProcessCitations(searchCall.Action.Query, searchCall.RawJSON())
		if err != nil {
			slog.Warn("Citation processing failed", "error", err, "query", searchCall.Action.Query)
		}

		if citationResult == nil {
			citationResult = &CitationResult{}
		}

		slog.Debug("Web search completed successfully",
			"query", searchCall.Action.Query,
			"url", searchCall.Action.URL,
			"type", searchCall.Action.Type,
			slog.Any("citations", citationResult.Citations),
		)

		trackingDetails["status"] = "completed"
		trackingDetails["query_count"] = 1 // Track query count
		trackingDetails["citations"] = citationResult.Citations

		s.trackWebSearchComputation("completed", startTime, trackingDetails, nil)
	case "failed":
		processor := NewCitationProcessor(s.p.sharedBag)
		var err error
		citationResult, err = processor.ProcessCitations(searchCall.Action.Query, searchCall.RawJSON())
		if err != nil {
			slog.Warn("Citation processing failed", "error", err, "query", searchCall.Action.Query)
		}

		if citationResult == nil {
			citationResult = &CitationResult{}
		}

		slog.Debug("Web search citations processed successfully",
			"url", searchCall.Action.URL,
			"query", searchCall.Action.Query,
			"citations", citationResult.Citations,
			"citations_found", citationResult.TotalCites)

		trackingDetails["status"] = "failed"
		trackingDetails["query_count"] = 1 // Track query count
		trackingDetails["citations"] = len(citationResult.Citations)
		s.trackWebSearchComputation("failed", startTime, trackingDetails, nil)
	}

	if citationResult != nil && citationResult.Success {
		trackingDetails["citation_count"] = citationResult.TotalCites
		trackingDetails["success"] = true

		slog.Debug("Web search citations processed successfully",
			"query", searchCall.Action.Query,
			"citations_found", citationResult.TotalCites)
	}
}
