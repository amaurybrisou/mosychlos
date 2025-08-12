package tools

import (
	"context"
	"sync"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// MetricsWrapper wraps an models.Tool with execution metrics tracking
type MetricsWrapper struct {
	tool      models.Tool
	sharedBag bag.SharedBag
	mu        sync.Mutex
}

// NewMetricsWrapper creates a wrapper that tracks tool execution metrics
func NewMetricsWrapper(tool models.Tool, sharedBag bag.SharedBag) *MetricsWrapper {
	return &MetricsWrapper{
		tool:      tool,
		sharedBag: sharedBag,
	}
}

// Implement models.Tool interface by delegating to wrapped tool

func (w *MetricsWrapper) Name() string {
	return w.tool.Name()
}

func (w *MetricsWrapper) Key() keys.Key {
	return w.tool.Key()
}

func (w *MetricsWrapper) Description() string {
	return w.tool.Description()
}

func (ct *MetricsWrapper) IsExternal() bool {
	return ct.tool.IsExternal()
}

func (w *MetricsWrapper) Tags() []string {
	return w.tool.Tags()
}

func (w *MetricsWrapper) Definition() models.ToolDef {
	return w.tool.Definition()
}

func (w *MetricsWrapper) Run(ctx context.Context, args string) (string, error) {
	startTime := time.Now()

	// Execute the wrapped tool
	result, err := w.tool.Run(ctx, args)

	duration := time.Since(startTime)
	success := err == nil

	// Create computation record
	computation := models.ToolComputation{
		ToolName:  w.tool.Name(),
		Arguments: args,
		Result:    result,
		StartTime: startTime,
		Duration:  duration,
		Success:   success,
	}

	if err != nil {
		computation.Error = err.Error()
	}

	// Record the computation
	w.recordComputation(computation)

	// Update aggregated metrics
	w.updateMetrics(computation)

	// Track API call health status
	w.updateAPIHealth(computation)

	return result, err
}

// recordComputation adds the computation to the shared bag
func (w *MetricsWrapper) recordComputation(comp models.ToolComputation) {
	w.sharedBag.Update(keys.KToolComputations, func(current any) any {
		var computations []models.ToolComputation
		if current != nil {
			if existing, ok := current.([]models.ToolComputation); ok {
				computations = existing
			}
		}
		return append(computations, comp)
	})
}

// updateMetrics updates the aggregated metrics in the shared bag
func (w *MetricsWrapper) updateMetrics(comp models.ToolComputation) {
	w.sharedBag.Update(keys.KToolMetrics, func(current any) any {
		var metrics models.ToolMetrics
		if current != nil {
			if existing, ok := current.(models.ToolMetrics); ok {
				metrics = existing
			}
		}

		// Initialize ByTool map if needed
		if metrics.ByTool == nil {
			metrics.ByTool = make(map[string]models.ToolStats)
		}

		// Update overall metrics
		metrics.TotalCalls++
		metrics.TotalDuration += comp.Duration
		metrics.TotalCost += comp.Cost
		metrics.TotalTokens += comp.TokensUsed

		if comp.Success {
			metrics.SuccessCount++
		} else {
			metrics.ErrorCount++
			if comp.Error != "" {
				metrics.Errors = append(metrics.Errors, comp.Error)
			}
		}

		// Calculate success rate
		if metrics.TotalCalls > 0 {
			metrics.SuccessRate = float64(metrics.SuccessCount) / float64(metrics.TotalCalls)
		}

		// Calculate average duration
		if metrics.TotalCalls > 0 {
			metrics.AverageDuration = time.Duration(int64(metrics.TotalDuration) / int64(metrics.TotalCalls))
		}

		// Update per-tool statistics
		toolStats := metrics.ByTool[comp.ToolName]
		toolStats.Calls++
		toolStats.Duration += comp.Duration
		toolStats.Cost += comp.Cost
		toolStats.Tokens += comp.TokensUsed

		if comp.Success {
			toolStats.Successes++
		} else {
			toolStats.Errors++
		}

		// Calculate per-tool average duration
		if toolStats.Calls > 0 {
			toolStats.AverageDuration = time.Duration(int64(toolStats.Duration) / int64(toolStats.Calls))
		}

		metrics.ByTool[comp.ToolName] = toolStats
		metrics.LastUpdated = time.Now()

		return metrics
	})
}

// updateAPIHealth tracks API call health status for external data providers
func (w *MetricsWrapper) updateAPIHealth(comp models.ToolComputation) {
	// Update overall external data health (contains per-tool status)
	w.sharedBag.Update(keys.KExternalDataHealth, func(current any) any {
		var health models.ExternalDataHealth
		if current != nil {
			if existing, ok := current.(models.ExternalDataHealth); ok {
				health = existing
			}
		}

		// Initialize providers map if needed
		if health.Providers == nil {
			health.Providers = make(map[string]models.DataProviderHealth)
		}

		// Update or create provider health
		provider := health.Providers[comp.ToolName]
		provider.Name = comp.ToolName

		if comp.Success {
			provider.LastSuccess = comp.StartTime
			provider.Status = "healthy"
		} else {
			provider.LastFailure = comp.StartTime
			if provider.Status != "down" {
				provider.Status = "degraded"
			}
			// Add recent error
			if provider.RecentErrors == nil {
				provider.RecentErrors = make([]string, 0)
			}
			provider.RecentErrors = append(provider.RecentErrors, comp.Error)
			// Keep only last 5 errors
			if len(provider.RecentErrors) > 5 {
				provider.RecentErrors = provider.RecentErrors[1:]
			}
		}

		// Update average latency (simple moving average approximation)
		if provider.AverageLatency == 0 {
			provider.AverageLatency = comp.Duration
		} else {
			provider.AverageLatency = (provider.AverageLatency + comp.Duration) / 2
		}

		health.Providers[comp.ToolName] = provider
		health.LastCheck = time.Now()

		return health
	})
}
