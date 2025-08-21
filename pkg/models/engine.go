package models

//go:generate mockgen -source=engine.go -destination=mocks/mock_engine.go -package=mocks

import (
	"context"
	"fmt"
	"slices"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
)

type EngineConfig struct {
	Name        string
	Constraints ToolConstraints
}

// ToolConsumer manages tool execution with credit/limit tracking
type ToolConsumer interface {
	// ConsumeTools executes tools based on constraints until credits are exhausted
	ConsumeTools(ctx context.Context, key keys.Key) error

	// GetRemainingCredits returns remaining calls for each tool
	GetRemainingCredits() map[keys.Key]int

	// HasCreditsFor checks if there are remaining credits for a tool
	HasCreditsFor(toolKey keys.Key) bool

	// Reset resets all tool call counters
	Reset()

	// GetUnusedRequiredTools returns required tools that haven't been called yet
	GetUnusedRequiredTools() []keys.Key

	// HasUnusedRequiredTools checks if there are required tools that haven't been called
	HasUnusedRequiredTools() bool
}

type Engine interface {
	Name() string
	ResultKey() keys.Key // Key to store the result in SharedBag
	Execute(ctx context.Context, client AiClient, sharedBag bag.SharedBag) error
}

type BaseToolConstraints struct {
	Tools           []ToolDef        // All tools that can be used
	PreferredTools  []keys.Key       // Tools to inject first
	RequiredTools   []keys.Key       // Tools that must be used
	MaxCallsPerTool map[keys.Key]int // Maximum calls allowed per tool
	MinCallsPerTool map[keys.Key]int // Minimum calls required per tool
}

func DefaultToolConstraints() ToolConstraints {
	return &BaseToolConstraints{
		Tools:           []ToolDef{},
		PreferredTools:  []keys.Key{},
		RequiredTools:   []keys.Key{},
		MaxCallsPerTool: make(map[keys.Key]int),
		MinCallsPerTool: make(map[keys.Key]int),
	}
}

// IsToolPreferred checks if a tool is in the preferred list
func (tc *BaseToolConstraints) IsToolPreferred(toolKey keys.Key) bool {
	return slices.Contains(tc.PreferredTools, toolKey)
}

// IsToolRequired checks if a tool is in the required list
func (tc *BaseToolConstraints) IsToolRequired(toolKey keys.Key) bool {
	for _, required := range tc.RequiredTools {
		if required == toolKey {
			return true
		}
	}
	return false
}

// GetMaxCalls returns the maximum allowed calls for a tool (0 if unlimited)
func (tc *BaseToolConstraints) GetMaxCalls(toolKey keys.Key) int {
	if tc.MaxCallsPerTool == nil {
		return 0 // unlimited
	}
	return tc.MaxCallsPerTool[toolKey]
}

// GetMinCalls returns the minimum allowed calls for a tool (0 if unlimited)
func (tc *BaseToolConstraints) GetMinCalls(toolKey keys.Key) int {
	if tc.MinCallsPerTool == nil {
		return 0 // unlimited
	}
	return tc.MinCallsPerTool[toolKey]
}

func (tc *BaseToolConstraints) GetAllMaxCalls() []int {
	maxCalls := make([]int, 0, len(tc.MaxCallsPerTool))
	for _, calls := range tc.MaxCallsPerTool {
		maxCalls = append(maxCalls, calls)
	}
	return maxCalls
}

// CanCallTool checks if a tool can still be called based on current call count
func (tc *BaseToolConstraints) CanCallTool(toolKey keys.Key, currentCalls int) bool {
	maxCalls := tc.GetMaxCalls(toolKey)
	if maxCalls == 0 {
		return true // unlimited calls
	}
	return currentCalls < maxCalls
}

// GetAllowedTools returns all tools that are either preferred or required
func (tc *BaseToolConstraints) GetAllowedTools() []keys.Key {
	allowedMap := make(map[keys.Key]bool)

	// Add preferred tools
	for _, tool := range tc.PreferredTools {
		allowedMap[tool] = true
	}

	// Add required tools
	for _, tool := range tc.RequiredTools {
		allowedMap[tool] = true
	}

	// Convert to slice
	var allowed []keys.Key
	for tool := range allowedMap {
		allowed = append(allowed, tool)
	}

	return allowed
}

// GetToolsWithLimits returns all tools that have max call limits configured
func (tc *BaseToolConstraints) GetToolsWithLimits() []keys.Key {
	var tools []keys.Key
	for tool := range tc.MaxCallsPerTool {
		tools = append(tools, tool)
	}
	return tools
}

// GetRequiredTools returns all tools that are required
func (tc *BaseToolConstraints) GetRequiredTools() []keys.Key {
	allowedMap := make(map[keys.Key]bool)

	// Add required tools
	for _, tool := range tc.RequiredTools {
		allowedMap[tool] = true
	}
	// Convert to slice
	var allowed []keys.Key
	for tool := range allowedMap {
		allowed = append(allowed, tool)
	}

	return allowed
}

// HasReachedMaxCalls checks if a tool has reached its maximum call limit
func (tc *BaseToolConstraints) HasReachedMaxCalls(toolKey keys.Key, currentCalls int) bool {
	maxCalls := tc.GetMaxCalls(toolKey)
	if maxCalls == 0 {
		return false // unlimited calls
	}
	return currentCalls >= maxCalls
}

// RemainingCalls returns how many more calls are allowed for a tool
func (tc *BaseToolConstraints) RemainingCalls(toolKey keys.Key, currentCalls int) int {
	maxCalls := tc.GetMaxCalls(toolKey)
	if maxCalls == 0 {
		return -1 // unlimited
	}
	remaining := maxCalls - currentCalls
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Validate checks if the constraints are valid
func (tc *BaseToolConstraints) Validate() error {
	// Check for duplicate tools between preferred and required
	preferredMap := make(map[keys.Key]bool)
	for _, tool := range tc.PreferredTools {
		preferredMap[tool] = true
	}

	for _, tool := range tc.RequiredTools {
		if preferredMap[tool] {
			// This is actually OK - a tool can be both preferred and required
			continue
		}
	}

	// Check for negative max calls
	if tc.MaxCallsPerTool != nil {
		for tool, maxCalls := range tc.MaxCallsPerTool {
			if maxCalls < 0 {
				return fmt.Errorf("tool %s has negative max calls: %d", tool, maxCalls)
			}
		}
	}

	return nil
}

// NewToolConstraints creates a new ToolConstraints with validation
func NewToolConstraints(preferred []keys.Key, required []keys.Key, maxCalls map[keys.Key]int) (*BaseToolConstraints, error) {
	tc := &BaseToolConstraints{
		PreferredTools:  preferred,
		RequiredTools:   required,
		MaxCallsPerTool: maxCalls,
	}

	if err := tc.Validate(); err != nil {
		return nil, err
	}

	return tc, nil
}
