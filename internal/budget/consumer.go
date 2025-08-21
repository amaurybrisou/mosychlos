package budget

import (
	"context"
	"log/slog"

	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// defaultToolConsumer implements ToolConsumer interface with constraint-based budget management
type defaultToolConsumer struct {
	constraints models.ToolConstraints
	callCounts  map[keys.Key]int
}

// NewToolConsumer creates a new ToolConsumer based on the provided constraints
func NewToolConsumer(constraints models.ToolConstraints) models.ToolConsumer {
	if constraints == nil {
		slog.Warn("budget.NewToolConsumer: nil constraints provided, using empty constraints")
		constraints = models.DefaultToolConstraints()
	}

	return &defaultToolConsumer{
		constraints: constraints,
		callCounts:  make(map[keys.Key]int),
	}
}

// ConsumeTools executes tools based on constraints until credits are exhausted
func (c *defaultToolConsumer) ConsumeTools(ctx context.Context, key keys.Key) error {
	if hasLimit := c.constraints.GetMaxCalls(key); hasLimit == 0 {
		return nil // no limit for this tool, so no consumption needed
	}

	c.callCounts[key]++
	slog.Debug("consuming credits for tool",
		"tool", key,
		"used", c.callCounts[key],
		"max", c.constraints.GetMaxCalls(key),
		"remaining", c.constraints.RemainingCalls(key, c.callCounts[key]),
	)

	return nil
}

// GetRemainingCredits returns remaining calls for each tool
func (c *defaultToolConsumer) GetRemainingCredits() map[keys.Key]int {
	remaining := make(map[keys.Key]int)

	// calculate remaining credits for each tool with a configured limit
	for _, toolName := range c.constraints.GetToolsWithLimits() {
		used := c.callCounts[toolName]
		remaining[toolName] = c.constraints.RemainingCalls(toolName, used)
	}

	return remaining
}

// HasCreditsFor checks if there are remaining credits for a tool
func (c *defaultToolConsumer) HasCreditsFor(toolKey keys.Key) bool {
	limit := c.constraints.GetMaxCalls(toolKey)
	if limit == 0 {
		// no limit means unlimited credits
		return true
	}

	used := c.callCounts[toolKey]
	remaining := limit - used

	slog.Debug("budget.HasCreditsFor",
		"tool", toolKey,
		"used", used,
		"max", limit,
		"remaining", remaining)

	return remaining > 0
}

// IncrementCallCount increments the call count for a tool (called after successful tool execution)
func (c *defaultToolConsumer) IncrementCallCount(toolKey keys.Key) {
	c.callCounts[toolKey]++
	slog.Debug("budget.IncrementCallCount",
		"tool", toolKey,
		"new_count", c.callCounts[toolKey])
}

// Reset resets all tool call counters
func (c *defaultToolConsumer) Reset() {
	slog.Debug("budget.Reset: clearing all tool call counters")
	c.callCounts = make(map[keys.Key]int)
}

// GetCallCount returns the current call count for a tool
func (c *defaultToolConsumer) GetCallCount(toolKey keys.Key) int {
	return c.callCounts[toolKey]
}

// GetConstraints returns a copy of the constraints
func (c *defaultToolConsumer) GetConstraints() models.ToolConstraints {
	return c.constraints
}

// GetUnusedRequiredTools returns required tools that haven't been called yet and still have credits,
// or tools that haven't reached their minimum call requirements
func (c *defaultToolConsumer) GetUnusedRequiredTools() []keys.Key {
	var unused []keys.Key

	for _, requiredTool := range c.constraints.GetRequiredTools() {
		currentCalls := c.callCounts[requiredTool]
		minCalls := c.constraints.GetMinCalls(requiredTool) // defaults to 0 if not set

		// Tool is unused if:
		// 1. It hasn't been called at all (currentCalls == 0), OR
		// 2. It hasn't reached its minimum required calls (currentCalls < minCalls)
		// AND it still has credits available
		if (currentCalls == 0 || currentCalls < minCalls) && c.HasCreditsFor(requiredTool) {
			unused = append(unused, requiredTool)
		}
	}

	return unused
}

// HasUnusedRequiredTools checks if there are required tools that haven't been called and still have credits
func (c *defaultToolConsumer) HasUnusedRequiredTools() bool {
	return len(c.GetUnusedRequiredTools()) > 0
}
