// internal/llm/tools/runtime.go
// File: internal/llm/tools/runtime.go
package toolsruntime

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// Registry maps a tool key to an implementation.
type Registry map[bag.Key]models.Tool

type Options struct {
	MaxRounds int // safety cap against infinite loops
}

// RunConversation executes LLM steps with tool calls enforced by ToolConsumer budgets.
// It returns the final assistant turn (after tools) or an error.
//
// funcTools: your function tools (models.ToolDef).
// hostedTools: built-in hosted tools like {"type":"web_search"} (marshaled as map[string]any).
func RunConversation(
	ctx context.Context,
	sess models.Session,
	provider models.Provider,
	funcTools []models.ToolDef,
	hostedTools []any,
	reg Registry,
	consumer models.ToolConsumer,
	opts Options,
	rf *models.ResponseFormat,
) (*models.AssistantTurn, error) {
	if opts.MaxRounds <= 0 {
		opts.MaxRounds = 6
	}

	turn, err := sess.Next(ctx, funcTools, rf) // first assistant step (rf forwarded)
	if err != nil {
		slog.Error("initial model step failed", "err", err)
		return nil, err
	}

	for round := 0; round < opts.MaxRounds; round++ {
		slog.Debug("RunConversation loop iteration",
			"round", round,
			"max_rounds", opts.MaxRounds,
			"tool_calls_count", len(turn.ToolCalls),
			"content_length", len(turn.Content))

		if len(turn.ToolCalls) == 0 {
			slog.Debug("No more tool calls, returning final turn",
				"content_length", len(turn.Content),
				"content_preview", turn.Content[:min(100, len(turn.Content))])
			return turn, nil // no tools => done
		}

		for _, call := range turn.ToolCalls {
			// Resolve function tool by name
			var foundKey bag.Key
			var tool models.Tool
			for k, t := range reg {
				if t.Name() == call.Function.Name {
					foundKey, tool = k, t
					break
				}
			}
			if tool == nil {
				// Not a registered function tool — very likely a hosted tool (e.g., web_search),
				// which the platform executes itself. We simply continue the loop.
				slog.Info("hosted tool handled by platform (no local impl)", "tool", call.Function.Name)
				continue
			}

			// Check budget
			if consumer != nil && !consumer.HasCreditsFor(foundKey) {
				slog.Info("budget exhausted for tool", "tool", call.Function.Name, "key", foundKey)
				sess.Add(models.RoleAssistant, fmt.Sprintf("Budget exhausted for tool %q.", call.Function.Name))
				continue
			}

			// Pre-consume (accounting) once per call
			if consumer != nil {
				if err := consumer.ConsumeTools(ctx, foundKey); err != nil {
					return nil, fmt.Errorf("consume budget for %s: %w", call.Function.Name, err)
				}
			}

			// Execute local tool
			out, err := tool.Run(ctx, call.Function.Arguments)
			if err != nil {
				slog.Warn("tool failed", "tool", call.Function.Name, "err", err)
				out = fmt.Sprintf("error: %v", err)
			}

			// Feed result to the model
			sess.AddFunctionCallResult(call, out)
		}

		// Next model turn (rf forwarded)
		slog.Debug("Requesting next model turn after tool execution", "round", round)
		next, err := sess.Next(ctx, funcTools, rf)
		if err != nil {
			return nil, err
		}
		turn = next
		slog.Debug("Received next model turn",
			"round", round,
			"content_length", len(next.Content),
			"tool_calls_count", len(next.ToolCalls),
			"content_preview", next.Content[:min(100, len(next.Content))])
	}
	slog.Warn("RunConversation reached max rounds without completion",
		"max_rounds", opts.MaxRounds,
		"final_turn_content_length", len(turn.Content),
		"final_turn_tool_calls", len(turn.ToolCalls))
	return turn, nil
}
