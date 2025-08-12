package openai

import (
	"log/slog"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/openai/openai-go/v2/responses"
)

// processResponsesAPIResult processes the response from the Responses API
func (s *session) processResponsesAPIResult(resp *responses.Response, startTime time.Time) (*models.AssistantTurn, error) {
	// Extract the content from the response
	var content string
	var toolCalls []models.ToolCall

	// Process the response items
	for _, item := range resp.Output {
		switch item.Type {
		case "message":
			// Extract text content from output message
			outputMsg := item.AsMessage()
			for _, contentItem := range outputMsg.Content {
				if contentItem.Type == "output_text" {
					outputText := contentItem.AsOutputText()
					content += outputText.Text
				}
			}
		case "function_call":
			// Handle function tool calls
			switch item.Name {
			case keys.WebSearch.String():
				// Handle web search - extract citations and track metrics WITHOUT token waste
				searchCall := item.AsWebSearchCall()
				s.processWebSearchCall(searchCall, content, startTime)

			default:
				// For other custom function calls executed by OpenAI internally
				// Log for monitoring but don't add to toolCalls (they were already executed)
				funcCall := item.AsFunctionCall()
				slog.Debug("Custom function call executed by OpenAI",
					"function_name", funcCall.Name,
					"call_id", funcCall.CallID,
				)
			}
		}
	}

	// Add the assistant response to the session state
	if content != "" {
		s.Add(models.RoleAssistant, content)
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
		s.trackOpenAIComputation(startTime, usage, len(toolCalls), nil)
	}

	return &models.AssistantTurn{
		Content:   content,
		ToolCalls: toolCalls, // Empty - OpenAI already executed everything
	}, nil
}

// processWebSearchCall handles web search processing without token waste
func (s *session) processWebSearchCall(searchCall interface{}, content string, startTime time.Time) {
	// Extract query from search call (type-safe extraction would go here)
	query := "web_search_query" // Placeholder - implement proper extraction

	// Process citations from web search content if available
	var citationResult *CitationResult
	if content != "" {
		processor := NewCitationProcessor(s.p.sharedBag)
		var err error
		citationResult, err = processor.ProcessCitations(query, content)
		if err != nil {
			slog.Warn("Citation processing failed", "error", err, "query", query)
		}
	}

	// Track web search usage with citation metrics
	trackingDetails := map[string]any{
		"action": keys.WebSearch,
		"query":  query,
	}

	if citationResult != nil && citationResult.Success {
		trackingDetails["citation_count"] = citationResult.TotalCites
		trackingDetails["success"] = true

		slog.Info("Web search citations processed successfully",
			"query", query,
			"citations_found", citationResult.TotalCites)
	}

	// Track the web search execution
	s.trackWebSearchComputation("used", startTime, trackingDetails, nil)

	// Track completion with results
	if citationResult != nil && citationResult.Success {
		s.trackWebSearchComputation("completed", startTime, map[string]any{
			"query_count":    1,
			"citation_count": citationResult.TotalCites,
			"citations":      len(citationResult.Citations), // Count for tracking
		}, nil)
	}
}
