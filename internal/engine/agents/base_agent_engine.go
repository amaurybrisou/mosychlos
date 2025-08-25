package agents

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/nlpodyssey/openai-agents-go/agents"
)

// BaseAgentEngine provides the foundation for agent-based engines
// It maintains compatibility with the existing engine interface while using agents internally
type BaseAgentEngine struct {
	name         string
	resultKey    bag.Key
	agent        *agents.Agent
	sharedBag    bag.SharedBag
	instructions string
}

// Ensure BaseAgentEngine implements models.Engine
var _ models.Engine = &BaseAgentEngine{}

// BaseAgentEngineConfig contains configuration for agent-based engines
type BaseAgentEngineConfig struct {
	Name         string
	ResultKey    bag.Key
	Instructions string
	Model        string
	Tools        []agents.Tool
	OutputType   agents.OutputTypeInterface
	MaxTurns     int
}

// NewBaseAgentEngine creates a new agent-based engine
func NewBaseAgentEngine(config BaseAgentEngineConfig) *BaseAgentEngine {
	if config.Name == "" {
		config.Name = "base-agent-engine"
	}

	if config.Model == "" {
		config.Model = "gpt-4o"
	}

	if config.MaxTurns == 0 {
		config.MaxTurns = 10
	}

	// Create the agent using the SDK
	agent := agents.New(config.Name).
		WithInstructions(config.Instructions).
		WithModel(config.Model).
		WithTools(config.Tools...)

	// Set output type if provided
	if config.OutputType != nil {
		agent = agent.WithOutputType(config.OutputType)
	}

	return &BaseAgentEngine{
		name:         config.Name,
		resultKey:    config.ResultKey,
		agent:        agent,
		instructions: config.Instructions,
	}
}

// Name returns the engine name
func (e *BaseAgentEngine) Name() string {
	return e.name
}

// ResultKey returns the key where this engine stores its results
func (e *BaseAgentEngine) ResultKey() bag.Key {
	return e.resultKey
}

// Execute runs the agent and stores results in SharedBag
// This maintains compatibility with the existing Engine interface
func (e *BaseAgentEngine) Execute(ctx context.Context, _ models.AiClient, sharedBag bag.SharedBag) error {
	e.sharedBag = sharedBag

	slog.Info("Agent engine starting", "name", e.name)

	// Build the input prompt from SharedBag context
	input, err := e.buildInput(ctx, sharedBag)
	if err != nil {
		return fmt.Errorf("failed to build agent input: %w", err)
	}

	// Run the agent using the SDK
	result, err := agents.Run(ctx, e.agent, input)
	if err != nil {
		return fmt.Errorf("agent execution failed: %w", err)
	}

	// Store the result in SharedBag
	err = e.storeResult(result, sharedBag)
	if err != nil {
		return fmt.Errorf("failed to store agent result: %w", err)
	}

	slog.Info("Agent engine completed", "name", e.name, "result_key", e.resultKey)
	return nil
}

// buildInput constructs the agent input from SharedBag data
// This can be overridden by specific engine implementations
func (e *BaseAgentEngine) buildInput(ctx context.Context, sharedBag bag.SharedBag) (string, error) {
	// Get portfolio data if available
	if portfolioData, exists := sharedBag.Get(bag.KPortfolio); exists {
		return fmt.Sprintf("Portfolio Analysis Request\n\nPortfolio Data: %+v", portfolioData), nil
	}

	// Fallback to basic analysis request
	return "Portfolio analysis requested", nil
}

// storeResult stores the agent's output in the SharedBag
// This can be overridden by specific engine implementations
func (e *BaseAgentEngine) storeResult(result *agents.RunResult, sharedBag bag.SharedBag) error {
	// Store the final output using the engine's result key
	sharedBag.Set(e.resultKey, result.FinalOutput)

	// Also store metadata about the agent run
	metadata := map[string]any{
		"agent_name":  e.name,
		"new_items":   len(result.NewItems),
		"responses":   len(result.RawResponses),
		"final_agent": result.LastAgent.Name,
	}
	
	metadataKey := bag.Key(string(e.resultKey) + "_metadata")
	sharedBag.Set(metadataKey, metadata)

	return nil
}

// GetAgent returns the underlying agent for advanced use cases
func (e *BaseAgentEngine) GetAgent() *agents.Agent {
	return e.agent
}

// WithTools adds tools to the agent
func (e *BaseAgentEngine) WithTools(tools ...agents.Tool) *BaseAgentEngine {
	e.agent = e.agent.WithTools(tools...)
	return e
}

// WithHandoffs adds handoff capabilities to the agent
func (e *BaseAgentEngine) WithHandoffs(handoffAgents ...*agents.Agent) *BaseAgentEngine {
	e.agent = e.agent.WithAgentHandoffs(handoffAgents...)
	return e
}