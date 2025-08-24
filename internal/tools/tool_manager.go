// Package tools
package tools

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

type ToolManager struct {
	tools     map[string]models.Tool
	sharedBag bag.SharedBag
}

// NewToolManager creates a fresh manager, wires caching/metrics wrappers,
// and initializes tools based on config.
func NewToolManager(cfg *config.Config, sharedBag bag.SharedBag) (*ToolManager, error) {
	m := &ToolManager{
		tools:     make(map[string]models.Tool),
		sharedBag: sharedBag,
	}

	// Example: iterate config to decide which tools to create
	for _, tc := range cfg.Tools.EnabledTools {
		toolFactory, ok := registry[tc]
		if !ok {
			return nil, fmt.Errorf("unknown tool: %s", tc)
		}

		toolConfig := toolFactory(cfg.GetToolConfig(tc), sharedBag)

		tool, err := toolConfig.Constructor()
		if err != nil {
			return nil, fmt.Errorf("failed to build tool: %w", err)
		}

		tool = wrapTool(tool, &toolConfig, cfg.CacheDir, sharedBag)
		m.tools[toolConfig.Key.String()] = tool
	}

	return m, nil
}

// Get returns a tool by name, or nil if not found
func (m *ToolManager) Get(name string) models.Tool {
	return m.tools[name]
}

// List returns all registered tools
func (m *ToolManager) List() []models.Tool {
	out := make([]models.Tool, 0, len(m.tools))
	for _, t := range m.tools {
		out = append(out, t)
	}
	return out
}

func (m *ToolManager) Defs() []models.ToolDef {
	return toolsToToolDefs(m.List())
}

// Bag exposes the manager’s shared bag (if needed by engines)
func (m *ToolManager) Bag() bag.SharedBag {
	return m.sharedBag
}

// Close releases resources (if any tools need it)
func (m *ToolManager) Close() error {
	var errs []error
	for _, t := range m.tools {
		if closer, ok := t.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing tools: %v", errs)
	}
	return nil
}

// wrapTool applies all configured wrappers to a tool
func wrapTool(tool models.Tool, config *models.ToolConfig, cacheDir string, sharedBag bag.SharedBag) models.Tool {
	wrapped := tool

	// Apply rate limiting if configured
	if config.RateLimit != nil {
		wrapped = NewRateLimitedTool(
			wrapped,
			config.RateLimit.RequestsPerSecond,
			config.RateLimit.RequestsPerDay,
			config.RateLimit.Burst,
		)
		slog.Debug("Applied rate limiting",
			"tool", tool.Name(),
			"requests_per_second", config.RateLimit.RequestsPerSecond,
			"requests_per_day", config.RateLimit.RequestsPerDay,
		)
	}

	// Apply caching if enabled
	if config.CacheEnabled {
		wrapped = NewCachedToolWithMonitoring(wrapped, cacheDir, config.CacheTTL, sharedBag)
		slog.Debug("Applied caching",
			"tool", tool.Name(),
			"cache_ttl", config.CacheTTL,
		)
	}

	// Apply metrics tracking if shared bag is available
	if sharedBag != nil {
		wrapped = NewMetricsWrapper(wrapped, sharedBag)
		slog.Debug("Applied metrics wrapper",
			"tool", tool.Name(),
		)
	}

	return wrapped
}

var (
	mu       = &sync.Mutex{}
	registry = make(map[string]ToolFactory)
)

type ToolFactory func(cfg any, sharedBag bag.SharedBag) models.ToolConfig

// Register registers a single tool with all wrappers
func Register(name bag.Key, f ToolFactory) {
	mu.Lock()
	registry[name.String()] = f
	mu.Unlock()

	slog.Info("Tool registered successfully",
		"key", name,
		"name", name.String(),
	)
}

func toolsToToolDefs(tools []models.Tool) []models.ToolDef {
	toolDefs := make([]models.ToolDef, 0, len(tools))
	for _, tool := range tools {
		toolDefs = append(toolDefs, tool.Definition())
	}
	return toolDefs
}
