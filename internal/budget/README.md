# Budget Package

The budget package provides tool consumption management for AI engines, implementing credit-based resource allocation and usage tracking.

## Overview

This package implements the `ToolConsumer` interface defined in `pkg/models/engine.go`, providing a constraint-based approach to managing tool call limits across different analysis engines.

## Key Components

### `defaultToolConsumer`

The main implementation that tracks tool usage against defined constraints:

- **Credit Tracking**: Monitors tool call counts per tool type
- **Limit Enforcement**: Prevents overuse based on `ToolConstraints`
- **Session Management**: Provides reset functionality for new analysis sessions

## Usage

### Creating a Consumer

```go
constraints := &models.ToolConstraints{
    MaxCallsPerTool: map[bag.Key]int{
        bag.FinRobotRiskAssessment: 2,
        bag.FMP: 3,
    },
}

consumer := budget.NewToolConsumer(constraints)
```

### In Engine Implementation

```go
func (e *engine) Execute(ctx context.Context, client models.AiClient, sharedBag bag.SharedBag) error {
    // Create consumer from engine constraints
    consumer := budget.NewToolConsumer(&e.constraints)

    // Configure AI client with consumer
    client.SetToolConsumer(consumer)

    // Execute analysis...
    result, err := ai.Ask[string](ctx, client, prompt)

    return err
}
```

### Manual Credit Management

```go
// Check if tool has remaining credits
if consumer.HasCreditsFor(bag.FinRobotRiskAssessment) {
    // Execute tool...

    // Increment counter after successful execution
    consumer.IncrementCallCount(bag.FinRobotRiskAssessment)
}

// Check remaining credits
remaining := consumer.GetRemainingCredits()
fmt.Printf("FMP calls remaining: %d", remaining[bag.FMP])

// Reset for new session
consumer.Reset()
```

## Design Principles

### Constraint-Based

All behavior is driven by `ToolConstraints`:

- Tools without limits have unlimited credits
- Tools with limits are tracked and enforced
- Limits are per-tool-type, not per-session

### Engine Independence

Each engine creates its own consumer instance:

- No shared state between engines
- Independent credit allocation per analysis type
- Clean session boundaries

### Integration with AI Client

The consumer integrates seamlessly with the AI client's `Ask` function to prevent infinite tool calling loops while respecting business logic constraints.

## Future Extensions

- **Budget Policies**: Conservative vs aggressive tool usage strategies
- **Cost Tracking**: Monitor API costs per tool call
- **Analytics**: Usage reporting and optimization recommendations
- **Dynamic Limits**: Adjust limits based on context or user preferences
