// Package agents provides tools for working with agent-based architectures.
package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/nlpodyssey/openai-agents-go/agents"
	"github.com/openai/openai-go/v2/packages/param"
)

// FromToolsToAgent converts a slice of Mosychlos tools to a slice of agent tools.
// Optionally, a custom error handling function can be provided to handle tool invocation errors.
// This enable us to take routing decisions based on tool-specific errors.
func FromToolsToAgent(mosychlosTools []models.Tool, failureErrorFunctions ...*agents.ToolErrorFunction) []agents.Tool {
	agentTools := make([]agents.Tool, len(mosychlosTools))

	for i, tool := range mosychlosTools {
		agentTool := agents.FunctionTool{
			Name:             tool.Name(),
			Description:      tool.Description(),
			ParamsJSONSchema: tool.Definition().ToMap(),
			OnInvokeTool: func(ctx context.Context, args string) (any, error) {
				result, err := tool.Run(ctx, args)
				if err != nil {
					slog.Error("failed to run tool", "tool", tool.Name(), "error", err)
					return nil, fmt.Errorf("tool %s invocation error: %w", tool.Name(), err)
				}

				switch v := result.(type) {
				case string:
					return v, nil
				case json.RawMessage:
					return string(v), nil
				case []byte:
					return string(v), nil
				default:
					if b, err := json.Marshal(v); err == nil {
						return string(b), nil
					}
					return fmt.Sprint(v), nil
				}

			},
			IsEnabled:        agents.FunctionToolEnabled(),
			StrictJSONSchema: param.NewOpt(true),
		}

		if len(failureErrorFunctions) > 0 && failureErrorFunctions[0] != nil {
			agentTool.FailureErrorFunction = failureErrorFunctions[0]
		}

		agentTools[i] = agentTool
	}

	return agentTools
}
