// Package tools
package tools

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/amaurybrisou/mosychlos/pkg/normalize"
)

type ToolManager struct {
	tools     map[string]models.Tool
	sharedBag bag.SharedBag
	reg       normalize.Registry
}

// NewToolManager creates a fresh manager, wires caching/metrics wrappers,
// and initializes tools based on config.
func NewToolManager(cfg *config.Config, sharedBag bag.SharedBag, reg normalize.Registry) (*ToolManager, error) {
	m := &ToolManager{
		tools:     make(map[string]models.Tool),
		sharedBag: sharedBag,
		reg:       reg,
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

		tool = wrapTool(tool, &toolConfig, cfg.DataDir, cfg.CacheDir, sharedBag, reg)
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
func wrapTool(
	tool models.Tool,
	config *models.ToolConfig,
	dataDir, cacheDir string,
	sharedBag bag.SharedBag,
	reg normalize.Registry,
) models.Tool {
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

	// Apply Input/Output Persiting in config.DataDir (especially useful for debugging & generating test fixtures)
	if runID := sharedBag.MustGet(bag.KEngineRunID).(string); config.Persisting && runID != "" {
		wrapped = NewIOPersistingTool(wrapped,
			filepath.Join(dataDir, "tools_i_o", fmt.Sprintf("run_%s_%s", runID, time.Now().Format("20060102_150405"))))
		slog.Debug("Applied I/O persisting",
			"tool", tool.Name(),
			"data_dir", dataDir,
		)
	}

	// 4) Normalize to a stable envelope (side-channel into SharedBag)
	if sharedBag != nil && reg != nil {
		wrapped = NewNormalizeWrapper(wrapped, reg, sharedBag)
		slog.Debug("Applied normalization wrapper",
			"tool", tool.Name(),
		)
	}

	// 5) Wire-minify for the LLM (return compact JSON)
	if sharedBag != nil && reg != nil {
		wrapped = NewWireMinWrapper(wrapped, reg, sharedBag)
		slog.Debug("Applied wire-minification wrapper",
			"tool", tool.Name(),
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
