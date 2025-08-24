// internal/llm/openai/strategy_response_parser.go
// Completely rewritten: parse Responses output items to collect output_text and tool calls.
package openai

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/openai/openai-go/v2/responses"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Scan a Responses API response and extract:
// - main output text (first output_text item)
// - tool calls (function_call/custom_tool_call)
// - usage (tokens)
func processResponsesAPIResult(resp *responses.Response, start time.Time) (*models.AssistantTurn, error) {
	content := map[string]any{}
	var (
		toolCalls []models.ToolCall
	)

	slog.Debug("Processing OpenAI response",
		"response_id", resp.ID,
		"items", len(resp.Output))

	for i, item := range resp.Output {
		switch item.Type {

		case "message":
			msg := item.AsMessage()
			for j, c := range msg.Content {
				switch c.Type {
				case "output_text":
					ot := c.AsOutputText()
					if len(ot.Text) > 0 {
						slog.Debug("Captured output_text",
							"len", len(ot.Text),
							"preview", preview(ot.Text, 120))
						content[key("output_text", i, j)] = ot.Text
					}

				case "refusal":
					content[key("refusal", i, j)] = c.Refusal

				default:
					raw := c.RawJSON()
					content[key(c.Type, i, j)] = json.RawMessage(raw)
				}
			}

		case "function_call":
			fc := item.AsFunctionCall()
			// Example: ignore built-in web_search if you special-case it elsewhere
			if fc.Name == bag.WebSearch.String() {
				slog.Debug("Ignoring inline web_search; handled by your tool layer")
				break
			}
			toolCalls = append(toolCalls, models.ToolCall{
				ID:     fc.ID,
				CallID: fc.CallID,
				Type:   "function",
				Status: string(fc.Status),
				Function: models.ToolCallFunction{
					Name:      fc.Name,
					Arguments: fc.Arguments,
				},
			})

		case "custom_tool_call":
			tc := item.AsCustomToolCall()
			toolCalls = append(toolCalls, models.ToolCall{
				ID:     tc.ID,
				CallID: tc.CallID,
				Type:   string(tc.Type),
				Function: models.ToolCallFunction{
					Name:      tc.Name,
					Arguments: tc.Input,
				},
			})

		case "reasoning":
			// informational only; do not gate logic on this presence
			slog.Debug("Reasoning item present", "response_id", resp.ID)

		default:
			content[key(item.Type, i, -1)] = json.RawMessage(item.RawJSON())
		}
	}

	usage := models.Usage{}
	if resp.Usage.TotalTokens > 0 {
		usage = models.Usage{
			PromptTokens:     int(resp.Usage.InputTokens),
			CompletionTokens: int(resp.Usage.OutputTokens),
			InputTokens:      int(resp.Usage.InputTokens),
			OutputTokens:     int(resp.Usage.OutputTokens),
			TotalTokens:      int(resp.Usage.TotalTokens),
		}
	}

	buf, err := json.Marshal(content)
	if err != nil {
		return nil, fmt.Errorf("marshal content: %w", err)
	}

	return &models.AssistantTurn{
		Content:   string(buf),
		ToolCalls: toolCalls,
		Usage:     usage,
	}, nil
}

func key(kind string, i, j int) string {
	if j >= 0 {
		return fmt.Sprintf("%s_%d_%d", kind, i, j)
	}
	return fmt.Sprintf("%s_%d", kind, i)
}

func preview(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:min(n, len(s))] + "…"
}
