// internal/llm/tools/runtime.go
// File: internal/llm/tools/runtime.go
package toolsruntime

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// Registry maps a tool key to an implementation.
type Registry map[keys.Key]models.Tool

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
		if len(turn.ToolCalls) == 0 {
			return turn, nil // no tools => done
		}

		for _, call := range turn.ToolCalls {
			// Resolve function tool by name
			var foundKey keys.Key
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
			sess.AddToolResult(call.CallID, out)
		}

		// Next model turn (rf forwarded)
		next, err := sess.Next(ctx, funcTools, rf)
		if err != nil {
			return nil, err
		}
		turn = next
	}
	return turn, nil
}
