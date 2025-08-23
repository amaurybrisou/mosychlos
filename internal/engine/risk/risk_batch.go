package risk

import (
	"context"
	"fmt"
	"log/slog"
	"maps"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/internal/engine/base"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// RiskBatchEngine implements risk analysis using the base batch engine with embedding
type RiskBatchEngine struct {
	*base.BatchEngine // Embed the base engine
	promptBuilder     models.PromptBuilder
}

var _ models.Engine = &RiskBatchEngine{}

// NewRiskBatchEngine creates a new risk batch engine using base batch engine
func NewRiskBatchEngine(name string, cfg config.LLMConfig, pb models.PromptBuilder, constraints models.BaseToolConstraints) *RiskBatchEngine {
	if name == "" {
		name = "risk-batch-engine"
	}

	// Create hooks implementation for risk analysis
	hooks := &RiskBatchEngineHooks{
		promptBuilder: pb,
	}

	// Create base batch engine
	baseEngine := base.NewBatchEngine(name, cfg.Model, constraints, hooks)

	return &RiskBatchEngine{
		BatchEngine:   baseEngine,
		promptBuilder: pb,
	}
}

// RiskBatchEngineHooks implements base.BatchEngineHooks for risk analysis
type RiskBatchEngineHooks struct {
	promptBuilder models.PromptBuilder
}

var _ models.BatchEngineHooks = &RiskBatchEngineHooks{}

// GetInitialPrompt returns the initial risk analysis prompt
func (h *RiskBatchEngineHooks) GetInitialPrompt(ctx context.Context) (string, error) {
	prompt, err := h.promptBuilder.BuildPrompt(ctx, models.AnalysisRisk)
	if err != nil {
		return "", fmt.Errorf("build initial risk prompt: %w", err)
	}
	return prompt, nil
}

// GenerateCustomID generates custom IDs for risk analysis jobs
func (h *RiskBatchEngineHooks) GenerateCustomID(iteration, jobIndex int) string {
	if iteration == 0 {
		return "task0"
	}
	return fmt.Sprintf("task_%d_%d", iteration, jobIndex)
}

// PreIteration is called before each batch iteration
func (h *RiskBatchEngineHooks) PreIteration(iteration int, jobs []models.BatchJob) error {
	slog.Debug("Processing risk batch iteration",
		"iteration", iteration,
		"jobs", len(jobs))
	return nil
}

// PostIteration is called after each batch iteration with results
func (h *RiskBatchEngineHooks) PostIteration(iteration int, results *models.BatchResult) error {
	slog.Debug("Risk batch iteration completed",
		"iteration", iteration,
		"successes", results.Successes,
		"failures", results.Failures)
	return nil
}

// ProcessToolResult is called when a tool call result is processed
func (h *RiskBatchEngineHooks) ProcessToolResult(customID, toolName, result string, sharedBag bag.SharedBag) error {
	// Store tool results in the shared bag for risk analysis
	sharedBag.Update(bag.KRiskAnalysisResult, func(a any) any {
		resultMap, ok := a.(map[string]any)
		if !ok {
			resultMap = make(map[string]any)
		}
		// Store both the tool result and metadata
		resultMap[fmt.Sprintf("%s_tool_%s", customID, toolName)] = result
		return resultMap
	})

	slog.Debug("Risk tool result processed",
		"tool", toolName,
		"custom_id", customID,
		"result_length", len(result))

	return nil
}

// ProcessFinalResult is called when a final result (no more tool calls) is processed
func (h *RiskBatchEngineHooks) ProcessFinalResult(customID, content string, sharedBag bag.SharedBag) error {
	// Store final risk analysis result in shared bag
	sharedBag.Update(h.ResultKey(), func(existing any) any {
		existingMap, ok := existing.(map[string]any)
		if !ok {
			existingMap = make(map[string]any)
		}

		// Create a new map to avoid mutation issues
		resultMap := make(map[string]any)
		maps.Copy(resultMap, existingMap)

		// Store only as the consolidated result for orchestrator compatibility
		// The customID storage is not needed as it creates duplicate data
		resultMap["result"] = content

		return resultMap
	})

	slog.Info("Risk analysis final result obtained",
		"custom_id", customID,
		"content_length", len(content))

	return nil
}

// ShouldContinueIteration determines if the batch process should continue for risk analysis
func (h *RiskBatchEngineHooks) ShouldContinueIteration(iteration int, nextJobs []models.BatchJob) bool {
	// Stop if maximum iterations reached
	if iteration >= 20 {
		slog.Info("Risk analysis reached maximum iterations", "iteration", iteration)
		return false
	}

	// Stop if no more jobs to process (natural completion)
	if len(nextJobs) == 0 {
		slog.Info("Risk analysis completed - no more jobs", "iteration", iteration)
		return false
	}

	// Continue processing
	return true
}

// ResultKey returns the key where risk analysis results should be stored
func (h *RiskBatchEngineHooks) ResultKey() bag.Key {
	return bag.KRiskAnalysisResult
}
