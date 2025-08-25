package agents

import (
	"context"
	"fmt"

	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/nlpodyssey/openai-agents-go/agents"
)

// ToolWrapper wraps Mosychlos tools to be compatible with the agents SDK
type ToolWrapper struct {
	mosychlosTool models.Tool
}

// NewToolWrapper creates a new wrapper for a Mosychlos tool
func NewToolWrapper(tool models.Tool) *ToolWrapper {
	return &ToolWrapper{mosychlosTool: tool}
}

// ToAgentTool converts a Mosychlos tool to an agents.Tool
func (tw *ToolWrapper) ToAgentTool() agents.FunctionTool {
	// Create a generic string-based tool that can handle JSON arguments
	tool := agents.NewFunctionTool(
		tw.mosychlosTool.Name(),
		tw.mosychlosTool.Description(),
		tw.executeWrapper,
	)
	
	return tool
}

// executeWrapper is the function that gets called by the agents SDK
// It bridges the call to the underlying Mosychlos tool
func (tw *ToolWrapper) executeWrapper(ctx context.Context, args string) (string, error) {
	// Call the underlying Mosychlos tool
	result, err := tw.mosychlosTool.Run(ctx, args)
	if err != nil {
		return "", fmt.Errorf("tool %s execution failed: %w", tw.mosychlosTool.Name(), err)
	}

	return result, nil
}

// ToolConverter provides utilities for converting Mosychlos tools to agent tools
type ToolConverter struct{}

// NewToolConverter creates a new tool converter
func NewToolConverter() *ToolConverter {
	return &ToolConverter{}
}

// ConvertTools converts a slice of Mosychlos tools to agent tools
func (tc *ToolConverter) ConvertTools(mosychlosTools []models.Tool) []agents.Tool {
	agentTools := make([]agents.Tool, 0, len(mosychlosTools))

	for _, tool := range mosychlosTools {
		wrapper := NewToolWrapper(tool)
		agentTool := wrapper.ToAgentTool()
		agentTools = append(agentTools, agentTool)
	}

	return agentTools
}

// ConvertToolProvider converts a ToolProvider to agent tools
func (tc *ToolConverter) ConvertToolProvider(provider models.ToolProvider) []agents.Tool {
	if provider == nil {
		return []agents.Tool{}
	}

	mosychlosTools := provider.List()
	return tc.ConvertTools(mosychlosTools)
}

// ConvertSelectedTools converts only the specified tools by key
func (tc *ToolConverter) ConvertSelectedTools(provider models.ToolProvider, toolKeys ...string) []agents.Tool {
	if provider == nil {
		return []agents.Tool{}
	}

	var selectedTools []models.Tool
	allTools := provider.List()

	// Build a map for quick lookup
	keySet := make(map[string]bool)
	for _, key := range toolKeys {
		keySet[key] = true
	}

	// Filter tools by the requested keys
	for _, tool := range allTools {
		if keySet[tool.Name()] {
			selectedTools = append(selectedTools, tool)
		}
	}

	return tc.ConvertTools(selectedTools)
}