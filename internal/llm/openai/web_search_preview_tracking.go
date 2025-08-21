package openai

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// trackWebSearchComputation tracks web search metrics using the same pattern as MetricsWrapper
func (s *session) trackWebSearchComputation(action string, startTime time.Time, details map[string]any, err error) {
	if s.p.sharedBag == nil {
		return
	}

	duration := time.Since(startTime)
	success := err == nil

	// Create computation record exactly like metrics wrapper
	computation := models.ToolComputation{
		ToolName:  keys.WebSearch.String(),
		Arguments: fmt.Sprintf(`{"action": "%s", "details": %v}`, action, details),
		Result:    s.buildWebSearchResult(action, details, success),
		StartTime: startTime,
		Duration:  duration,
		Success:   success,
	}

	if err != nil {
		computation.Error = err.Error()
	}

	// Use EXACT same pattern as metrics wrapper
	s.p.sharedBag.Update(keys.KToolComputations, func(current any) any {
		var computations []models.ToolComputation
		if current != nil {
			if existing, ok := current.([]models.ToolComputation); ok {
				computations = existing
			}
		}
		return append(computations, computation)
	})

	// Update aggregated metrics (same as metrics wrapper)
	s.updateWebSearchMetrics(computation)

	// Update API health (same as metrics wrapper)
	s.updateWebSearchAPIHealth(computation)

	slog.Debug("Web search computation tracking started",
		"action", action,
		"time", startTime,
		"success", success,
	)
}

// buildWebSearchResult creates a result string for the computation
func (s *session) buildWebSearchResult(action string, details map[string]any, success bool) string {
	result := map[string]any{
		"status":  action,
		"success": success,
		"data":    details,
	}

	// For different actions, include different information
	switch action {
	case "offering":
		result["description"] = "Web search capability offered to AI model"
		if details != nil {
			if contextSize, ok := details["context_size"]; ok {
				result["context_size"] = contextSize
			}
			if userLoc, ok := details["user_location"]; ok {
				result["user_location"] = userLoc
			}
		}
	case "used":
		result["description"] = "Web search executed by OpenAI"
		if details != nil {
			if query, ok := details["query"]; ok {
				result["search_query"] = query
			}
			if callID, ok := details["call_id"]; ok {
				result["call_id"] = callID
			}
		}
	case "completed":
		result["description"] = "Web search session completed"
		if details != nil {
			if queries, ok := details["queries"]; ok {
				result["total_queries"] = queries
			}
			if citations, ok := details["citations"]; ok {
				result["citations_found"] = citations
			}
			if queryCount, ok := details["query_count"]; ok {
				result["query_count"] = queryCount
			}
			if citationCount, ok := details["citation_count"]; ok {
				result["citation_count"] = citationCount
			}
		}
	case "api_error":
		result["description"] = "Web search API error"
		result["success"] = false
	}

	// Convert to JSON string for storage
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Sprintf(`{"status": "%s", "success": %v, "error": "failed to marshal result"}`, action, success)
	}

	return string(resultBytes)
}

// updateWebSearchMetrics updates aggregated metrics using same pattern as MetricsWrapper
func (s *session) updateWebSearchMetrics(comp models.ToolComputation) {
	s.p.sharedBag.Update(keys.KToolMetrics, func(current any) any {
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

// updateWebSearchAPIHealth tracks API health using same pattern as MetricsWrapper
func (s *session) updateWebSearchAPIHealth(comp models.ToolComputation) {
	s.p.sharedBag.Update(keys.KExternalDataHealth, func(current any) any {
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

// trackOpenAITokenUsage tracks OpenAI API usage with token consumption
func (s *session) trackOpenAITokenUsage(startTime time.Time, usage models.Usage, toolCallCount int, err error) {
	if s.p.sharedBag == nil {
		return
	}

	duration := time.Since(startTime)
	success := err == nil

	// Create computation record for OpenAI API usage
	computation := models.ToolComputation{
		ToolName:     "openai_api",
		Arguments:    fmt.Sprintf(`{"model": "%s", "tool_calls": %d}`, s.p.cfg.Model.String(), toolCallCount),
		Result:       fmt.Sprintf(`{"tokens_used": %d, "input_tokens": %d, "output_tokens": %d}`, usage.TotalTokens, usage.InputTokens, usage.OutputTokens),
		StartTime:    startTime,
		Duration:     duration,
		Success:      success,
		TokensUsed:   usage.TotalTokens,
		Cost:         s.calculateCost(usage), // Implement cost calculation
		DataConsumed: []string{"messages", "tools"},
		DataProduced: []string{"response", "tool_calls"},
	}

	if err != nil {
		computation.Error = err.Error()
	}

	// Store in shared bag using same pattern as other tracking
	s.p.sharedBag.Update(keys.KToolComputations, func(current any) any {
		var computations []models.ToolComputation
		if current != nil {
			if existing, ok := current.([]models.ToolComputation); ok {
				computations = existing
			}
		}
		return append(computations, computation)
	})

	// Update OpenAI API metrics
	s.updateOpenAIMetrics(computation)

	// Update OpenAI API health status
	s.updateOpenAIAPIHealth(computation)

	slog.Info("OpenAI API computation tracked",
		"tokens_used", usage.TotalTokens,
		"input_tokens", usage.InputTokens,
		"output_tokens", usage.OutputTokens,
		"duration", duration,
		"success", success,
	)
}

// calculateCost estimates the cost based on token usage and model
func (s *session) calculateCost(usage models.Usage) float64 {
	// Rough cost estimates for gpt-4o-mini (adjust as needed)
	// Input: $0.15 per 1M tokens, Output: $0.6 per 1M tokens
	inputCost := float64(usage.InputTokens) * 0.15 / 1000000
	outputCost := float64(usage.OutputTokens) * 0.6 / 1000000
	return inputCost + outputCost
}

// updateOpenAIMetrics updates OpenAI-specific metrics
func (s *session) updateOpenAIMetrics(comp models.ToolComputation) {
	s.p.sharedBag.Update(keys.KToolMetrics, func(current any) any {
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

		// Update overall metrics for OpenAI API
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

		// Update per-tool statistics for OpenAI API
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

// updateOpenAIAPIHealth tracks OpenAI API health status
func (s *session) updateOpenAIAPIHealth(comp models.ToolComputation) {
	s.p.sharedBag.Update(keys.KExternalDataHealth, func(current any) any {
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

		// Update OpenAI API provider health
		provider := health.Providers[comp.ToolName]
		provider.Name = "OpenAI API"
		provider.AverageLatency = comp.Duration

		if comp.Success {
			provider.Status = "healthy"
			provider.LastSuccess = comp.StartTime
			// Calculate simple success rate (this is a simplified approach)
			if provider.SuccessRate == 0 {
				provider.SuccessRate = 1.0
			} else {
				provider.SuccessRate = (provider.SuccessRate + 1.0) / 2.0 // Simple moving average
			}
		} else {
			provider.Status = "degraded"
			provider.LastFailure = comp.StartTime
			if len(provider.RecentErrors) < 5 {
				provider.RecentErrors = append(provider.RecentErrors, comp.Error)
			} else {
				// Keep only recent 5 errors
				provider.RecentErrors = append(provider.RecentErrors[1:], comp.Error)
			}
			// Reduce success rate
			provider.SuccessRate = provider.SuccessRate * 0.8 // Simple decay
		}

		health.Providers[comp.ToolName] = provider
		health.LastCheck = time.Now()

		return health
	})
}
