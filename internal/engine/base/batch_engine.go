package base

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/amaurybrisou/mosychlos/internal/budget"
	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/internal/tools"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// BatchEngine implements models.Engine interface with template method pattern
type BatchEngine struct {
	name        string
	constraints models.BaseToolConstraints
	model       config.LLMModel
	hooks       models.BatchEngineHooks
}

var _ models.Engine = &BatchEngine{}

// NewBatchEngine creates a new base batch engine with hooks for customization
func NewBatchEngine(
	name string,
	model config.LLMModel,
	constraints models.BaseToolConstraints,
	hooks models.BatchEngineHooks,
) *BatchEngine {
	return &BatchEngine{
		name:        name,
		model:       model,
		constraints: constraints,
		hooks:       hooks,
	}
}

// Name returns the engine name
func (b *BatchEngine) Name() string {
	return b.name
}

// ResultKey returns the result key from hooks
func (b *BatchEngine) ResultKey() keys.Key {
	return b.hooks.ResultKey()
}

// Execute implements the template method pattern with hooks for customization
func (b *BatchEngine) Execute(ctx context.Context, aiClient models.AiClient, sharedBag bag.SharedBag) error {
	// Set up tool consumer
	aiClient.SetToolConsumer(budget.NewToolConsumer(&b.constraints))

	// Hook: Get initial prompt
	prompt, err := b.hooks.GetInitialPrompt(ctx)
	if err != nil {
		return fmt.Errorf("get initial prompt: %w", err)
	}

	// Initialize first batch job
	currentJobs := []models.BatchJob{
		{
			Request: models.PromptRequest{
				Model:    b.model.String(),
				Messages: []map[string]any{{"role": "user", "content": prompt}},
				Tools:    b.constraints.Tools,
			},
			CustomID: b.hooks.GenerateCustomID(0, 0),
			Messages: []map[string]any{{"role": "user", "content": prompt}},
		},
	}

	maxIterations := 20
	iteration := 0

	// Main batch processing loop
	for len(currentJobs) > 0 && iteration < maxIterations {
		iteration++

		slog.Info("Starting batch iteration",
			"engine", b.name,
			"iteration", iteration,
			"jobs_count", len(currentJobs))

		// Hook: Pre-iteration processing
		if err := b.hooks.PreIteration(iteration, currentJobs); err != nil {
			return fmt.Errorf("pre-iteration hook failed: %w", err)
		}

		// Submit batch and wait for completion
		batchJob, err := b.submitAndWaitForBatch(ctx, aiClient, currentJobs, iteration)
		if err != nil {
			return err
		}

		// Process batch results
		results, err := aiClient.BatchManager().GetResults(ctx, batchJob.ID)
		if err != nil {
			return fmt.Errorf("get batch results (iteration %d): %w", iteration, err)
		}

		// Process individual job results
		nextJobs, err := b.processJobResults(ctx, currentJobs, results, iteration, sharedBag)
		if err != nil {
			return err
		}

		// Hook: Post-iteration processing
		if err := b.hooks.PostIteration(iteration, results); err != nil {
			return fmt.Errorf("post-iteration hook failed: %w", err)
		}

		// Hook: Check if should continue
		if !b.hooks.ShouldContinueIteration(iteration, nextJobs) {
			slog.Info("Engine decided to stop iteration",
				"engine", b.name,
				"iteration", iteration)
			break
		}

		currentJobs = nextJobs
	}

	if iteration >= maxIterations {
		slog.Warn("Batch processing reached maximum iterations",
			"engine", b.name,
			"max_iterations", maxIterations)
	}

	slog.Info("Batch processing completed",
		"engine", b.name,
		"total_iterations", iteration)

	return nil
}

// submitAndWaitForBatch submits jobs to AI client and waits for completion
func (b *BatchEngine) submitAndWaitForBatch(
	ctx context.Context,
	aiClient models.AiClient,
	jobs []models.BatchJob,
	iteration int,
) (*models.BatchJob, error) {
	// Convert jobs to batch requests
	batchRequests := make([]models.PromptRequest, len(jobs))
	for i, job := range jobs {
		req := job.Request
		req.CustomID = job.CustomID // Preserve the custom ID
		batchRequests[i] = req
	}

	// Submit batch
	batchJob, err := aiClient.DoBatch(ctx, batchRequests)
	if err != nil {
		return nil, fmt.Errorf("submit batch (iteration %d): %w", iteration, err)
	}

	slog.Info("Batch submitted successfully",
		"engine", b.name,
		"iteration", iteration,
		"batch_id", batchJob.ID,
		"jobs_count", len(batchRequests))

	// Wait for completion
	completedJob, err := aiClient.BatchManager().WaitForCompletion(ctx, batchJob.ID)
	if err != nil {
		return nil, fmt.Errorf("wait for batch completion (iteration %d): %w", iteration, err)
	}

	slog.Info("Batch completed",
		"engine", b.name,
		"iteration", iteration,
		"batch_id", completedJob.ID,
		"status", completedJob.Status)

	return completedJob, nil
}

// processJobResults processes batch results and generates next jobs if needed
func (b *BatchEngine) processJobResults(
	ctx context.Context,
	jobs []models.BatchJob,
	results *models.BatchResult,
	iteration int,
	sharedBag bag.SharedBag,
) ([]models.BatchJob, error) {
	var nextJobs []models.BatchJob

	// Track token usage from this batch
	for customID, usage := range results.Usage {
		if usage.TotalTokens > 0 {
			// Store token usage in shared bag for metrics tracking
			sharedBag.Set(keys.Key(fmt.Sprintf("token_usage_%s", customID)), usage)
			slog.Debug("Stored batch token usage",
				"custom_id", customID,
				"prompt_tokens", usage.PromptTokens,
				"completion_tokens", usage.CompletionTokens,
				"total_tokens", usage.TotalTokens)
		}
	}

	for _, job := range jobs {
		customID := job.CustomID // Use the original custom ID from the job

		// Check for errors
		if errStr, hasError := results.Errors[customID]; hasError && errStr != "" {
			slog.Error("Batch item failed",
				"engine", b.name,
				"custom_id", customID,
				"error", errStr)
			continue
		}

		// Process based on result type
		toolCalls, hasToolCalls := results.ToolCalls[customID]
		slog.Debug("Processing job result",
			"custom_id", customID,
			"has_tool_calls", hasToolCalls,
			"tool_calls_count", len(toolCalls))

		if !hasToolCalls || len(toolCalls) == 0 {
			// Final result - no more tool calls
			if content, hasContent := results.Content[customID]; hasContent {
				slog.Debug("Processing final result", "custom_id", customID, "content_length", len(content))
				if err := b.hooks.ProcessFinalResult(customID, content, sharedBag); err != nil {
					return nil, fmt.Errorf("process final result: %w", err)
				}
			}
			continue
		}

		// Process tool calls and prepare next iteration
		slog.Debug("Processing tool calls", "custom_id", customID, "tool_calls", len(toolCalls))
		nextJob, err := b.processToolCalls(ctx, job, toolCalls, customID, iteration, sharedBag)
		if err != nil {
			return nil, err
		}

		if nextJob != nil {
			slog.Debug("Created next job", "custom_id", nextJob.CustomID)
			nextJobs = append(nextJobs, *nextJob)
		} else {
			slog.Debug("No next job created", "custom_id", customID)
		}
	}

	slog.Debug("Job processing completed",
		"total_jobs_processed", len(jobs),
		"next_jobs_created", len(nextJobs))

	return nextJobs, nil
}

// processToolCalls processes tool calls and creates next job if needed
func (b *BatchEngine) processToolCalls(
	ctx context.Context, job models.BatchJob, toolCalls []models.ToolCall, customID string, iteration int, sharedBag bag.SharedBag) (*models.BatchJob, error) {
	// Process each tool call and collect results
	messages := append(job.Messages, map[string]any{
		"role":       "assistant",
		"tool_calls": toolCalls,
	})

	for _, toolCall := range toolCalls {
		slog.Debug("Executing tool call",
			"tool", toolCall.Function.Name,
			"custom_id", customID,
			"arguments", toolCall.Function.Arguments)

		// Execute the tool call through hooks
		result, err := b.executeToolCall(ctx, toolCall)
		if err != nil {
			slog.Error("Tool execution failed",
				"tool", toolCall.Function.Name,
				"custom_id", customID,
				"error", err)
			result = "Tool execution failed - nothing found"
		}

		slog.Debug("Tool execution completed",
			"tool", toolCall.Function.Name,
			"custom_id", customID,
			"result_length", len(result))

		// Hook: Process tool result
		if err := b.hooks.ProcessToolResult(customID, toolCall.Function.Name, result, sharedBag); err != nil {
			return nil, fmt.Errorf("process tool result: %w", err)
		}

		// Add tool response to messages
		messages = append(messages, map[string]any{
			"role":         "tool",
			"tool_call_id": toolCall.ID,
			"content":      result,
		})
	}

	// Create next job
	nextJob := &models.BatchJob{
		Request: models.PromptRequest{
			Model:    b.model.String(),
			Messages: messages,
			Tools:    b.constraints.Tools,
		},
		CustomID: b.hooks.GenerateCustomID(iteration+1, 0),
		Messages: messages,
	}

	slog.Debug("Created next batch job after tool execution",
		"custom_id", nextJob.CustomID,
		"messages_count", len(messages))

	return nextJob, nil
}

// executeToolCall is a helper method that can be overridden by embedding engines
func (b *BatchEngine) executeToolCall(ctx context.Context, toolCall models.ToolCall) (string, error) {
	slog.Debug("Looking up tool",
		"tool_name", toolCall.Function.Name,
		"available_tools", len(tools.GetToolsMap()))

	allTools := tools.GetToolsMap()
	tool, exists := allTools[keys.Key(toolCall.Function.Name)]
	if !exists {
		slog.Error("Tool not found",
			"tool_name", toolCall.Function.Name,
			"available_tools", func() []string {
				var names []string
				for k := range allTools {
					names = append(names, string(k))
				}
				return names
			}())
		return "", fmt.Errorf("tool not found: %s", toolCall.Function.Name)
	}

	slog.Debug("Tool found, checking if external",
		"tool_name", toolCall.Function.Name,
		"is_external", tool.IsExternal())

	if tool.IsExternal() {
		// External tools (like web_search_preview) handle citation processing internally
		_, err := tool.Run(ctx, toolCall.Function.Arguments)
		if err != nil {
			return "", fmt.Errorf("external tool execution failed: %w", err)
		}
		// External tools don't return content directly - they process citations
		return "Web search completed - results processed for analysis", nil
	}

	// Execute regular tools
	result, err := tool.Run(ctx, toolCall.Function.Arguments)
	if err != nil {
		slog.Error("Tool execution failed",
			"tool", toolCall.Function.Name,
			"custom_id", toolCall,
			"error", err)
		return "", fmt.Errorf("tool execution failed: %w", err)
	}

	return result, nil
}
