package tools

import (
	"fmt"
	"log/slog"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

var (
	tools     = map[keys.Key]models.Tool{}
	sharedBag bag.SharedBag
)

// SetSharedBag sets the shared bag instance for metrics tracking
func SetSharedBag(sb bag.SharedBag) {
	sharedBag = sb
}

// NewTools initializes all configured tools with systematic registration
func NewTools(cfg *config.Config) error {
	// Ensure shared bag is set
	if sharedBag == nil {
		return fmt.Errorf("shared bag is not initialized")
	}

	// Get all tool configurations
	toolConfigs := GetToolConfigs(cfg)

	// Register each tool
	for _, toolConfig := range toolConfigs {
		if err := RegisterTool(toolConfig, cfg.CacheDir, sharedBag); err != nil {
			return fmt.Errorf("failed to register tool %s: %w", toolConfig.Key, err)
		}
	}

	slog.Info("All tools registered successfully",
		"total_tools", len(toolConfigs),
	)

	return nil
}

// GetTools returns a slice of all registered tools
func GetTools() []models.Tool {
	result := make([]models.Tool, 0, len(tools))
	for _, tool := range tools {
		result = append(result, tool)
	}
	return result
}

// GetToolsMap returns all registered tools
func GetToolsMap() map[keys.Key]models.Tool {
	return tools
}

func GetToolsDef() []models.ToolDef {
	toolDefs := make([]models.ToolDef, 0, len(tools))
	for _, tool := range tools {
		toolDefs = append(toolDefs, tool.Definition())
	}
	return toolDefs
}

// ToolCount returns the number of registered tools
func ToolCount() int {
	return len(tools)
}

func GetTool(key keys.Key) (models.Tool, bool) {
	t, ok := tools[key]
	return t, ok
}

// ClearTools clears all registered tools (useful for testing)
func ClearTools() {
	tools = make(map[keys.Key]models.Tool)
}

func ToolsToToolDefs(tools []models.Tool) []models.ToolDef {
	toolDefs := make([]models.ToolDef, 0, len(tools))
	for _, tool := range tools {
		toolDefs = append(toolDefs, tool.Definition())
	}
	return toolDefs
}
