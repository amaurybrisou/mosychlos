// internal/llm/batch/cost_optimizer.go
package batch

import (
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// CostOptimizer handles cost estimation for batch processing
type CostOptimizer struct {
	// Pricing per 1K tokens (as of August 2025)
	inputPricing  map[string]float64
	outputPricing map[string]float64
}

// NewCostOptimizer creates a new cost optimizer with current pricing
func NewCostOptimizer() *CostOptimizer {
	return &CostOptimizer{
		inputPricing: map[string]float64{
			"gpt-4o":      0.005,   // $0.005 per 1K input tokens
			"gpt-4o-mini": 0.00015, // $0.00015 per 1K input tokens
			"gpt-5":       0.010,   // $0.010 per 1K input tokens (estimated)
			"gpt-5-mini":  0.0003,  // $0.0003 per 1K input tokens (estimated)
			"default":     0.001,   // Default fallback pricing
		},
		outputPricing: map[string]float64{
			"gpt-4o":      0.015,  // $0.015 per 1K output tokens
			"gpt-4o-mini": 0.0006, // $0.0006 per 1K output tokens
			"gpt-5":       0.030,  // $0.030 per 1K output tokens (estimated)
			"gpt-5-mini":  0.0012, // $0.0012 per 1K output tokens (estimated)
			"default":     0.003,  // Default fallback pricing
		},
	}
}

// EstimateCost calculates the estimated cost for batch requests
func (co *CostOptimizer) EstimateCost(reqs []models.BatchRequest) *models.CostEstimate {
	var totalInputTokens, totalOutputTokens int
	var totalCost float64

	for _, req := range reqs {
		estimate := co.estimateRequestCost(req)
		totalInputTokens += estimate.EstimatedTokensIn
		totalOutputTokens += estimate.EstimatedTokensOut
		totalCost += estimate.EstimatedCost
	}

	// Apply batch discount (50% savings)
	batchCost := totalCost * 0.5
	savingsVsSync := 0.5 // 50% savings

	return &models.CostEstimate{
		EstimatedCost:      batchCost,
		SavingsVsSync:      savingsVsSync,
		EstimatedTokensIn:  totalInputTokens,
		EstimatedTokensOut: totalOutputTokens,
	}
}

// estimateRequestCost estimates cost for a single request
func (co *CostOptimizer) estimateRequestCost(req models.BatchRequest) *models.CostEstimate {
	// Extract model from request body
	model := co.extractModel(req.Body)

	// Estimate input tokens from messages
	inputTokens := co.estimateInputTokens(req.Body)

	// Estimate output tokens based on max_tokens or default
	outputTokens := co.estimateOutputTokens(req.Body)

	// Calculate cost
	inputPrice := co.getInputPrice(model)
	outputPrice := co.getOutputPrice(model)

	inputCost := float64(inputTokens) / 1000.0 * inputPrice
	outputCost := float64(outputTokens) / 1000.0 * outputPrice
	totalCost := inputCost + outputCost

	return &models.CostEstimate{
		EstimatedCost:      totalCost,
		SavingsVsSync:      0.5, // 50% batch savings
		EstimatedTokensIn:  inputTokens,
		EstimatedTokensOut: outputTokens,
	}
}

// extractModel extracts the model name from request body
func (co *CostOptimizer) extractModel(body map[string]any) string {
	if model, ok := body["model"].(string); ok {
		return model
	}
	return "default"
}

// estimateInputTokens estimates input tokens from messages
func (co *CostOptimizer) estimateInputTokens(body map[string]any) int {
	messages, ok := body["messages"].([]any)
	if !ok {
		return 100 // Default estimate
	}

	totalTokens := 0
	for _, msg := range messages {
		if msgMap, ok := msg.(map[string]any); ok {
			if content, ok := msgMap["content"].(string); ok {
				// Rough estimate: 1 token ≈ 4 characters
				tokenCount := utf8.RuneCountInString(content) / 4
				if tokenCount < 1 {
					tokenCount = 1
				}
				totalTokens += tokenCount
			}
		}
	}

	// Add overhead for message formatting, role tokens, etc.
	totalTokens = int(float64(totalTokens) * 1.1)

	if totalTokens < 10 {
		totalTokens = 10 // Minimum token count
	}

	return totalTokens
}

// estimateOutputTokens estimates output tokens from max_tokens or default
func (co *CostOptimizer) estimateOutputTokens(body map[string]any) int {
	if maxTokens, ok := body["max_tokens"]; ok {
		switch v := maxTokens.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case json.Number:
			if intVal, err := v.Int64(); err == nil {
				return int(intVal)
			}
		}
	}

	// Default output token estimate
	return 200
}

// getInputPrice gets input pricing for model
func (co *CostOptimizer) getInputPrice(model string) float64 {
	model = strings.ToLower(model)

	// Check for exact match
	if price, ok := co.inputPricing[model]; ok {
		return price
	}

	// Check for partial matches
	for modelPattern, price := range co.inputPricing {
		if strings.Contains(model, modelPattern) {
			return price
		}
	}

	return co.inputPricing["default"]
}

// getOutputPrice gets output pricing for model
func (co *CostOptimizer) getOutputPrice(model string) float64 {
	model = strings.ToLower(model)

	// Check for exact match
	if price, ok := co.outputPricing[model]; ok {
		return price
	}

	// Check for partial matches
	for modelPattern, price := range co.outputPricing {
		if strings.Contains(model, modelPattern) {
			return price
		}
	}

	return co.outputPricing["default"]
}

// UpdatePricing allows updating pricing information
func (co *CostOptimizer) UpdatePricing(model string, inputPrice, outputPrice float64) {
	co.inputPricing[model] = inputPrice
	co.outputPricing[model] = outputPrice
}

// GetCurrentPricing returns current pricing information
func (co *CostOptimizer) GetCurrentPricing() (map[string]float64, map[string]float64) {
	// Return copies to prevent modification
	inputCopy := make(map[string]float64, len(co.inputPricing))
	outputCopy := make(map[string]float64, len(co.outputPricing))

	for k, v := range co.inputPricing {
		inputCopy[k] = v
	}
	for k, v := range co.outputPricing {
		outputCopy[k] = v
	}

	return inputCopy, outputCopy
}
