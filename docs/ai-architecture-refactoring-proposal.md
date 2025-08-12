# AI Package Architecture Analysis and Refactoring Proposal

## Executive Summary

This document provides a comprehensive analysis of the current AI and OpenAI packages in Mosychlos, identifying architectural strengths, weaknesses, and proposing clean refactoring strategies to improve maintainability, extensibility, and integration with the proposed OpenAI rate limiting system.

## Current Architecture Analysis

### Package Structure Overview

```
internal/llm/                    # New LLM package (was internal/ai/)
├── client.go                    # Main LLM client orchestration
├── client_test.go              # Integration tests
├── factory.go                  # Provider factory pattern
├── schema.go                   # JSON schema generation for structured output
├── mock/                       # Generated mocks using mockgen
│   ├── provider_mock.go        # Generated provider mocks
│   ├── strategy_mock.go        # Generated strategy mocks
│   └── middleware_mock.go      # Generated middleware mocks
├── middleware/                 # Middleware system
│   ├── middleware.go           # Core middleware interfaces
│   ├── rate_limit.go          # Rate limiting middleware
│   └── retry.go               # Retry middleware
├── health/                     # Health monitoring
│   └── provider_monitor.go    # Provider health monitoring
├── validation/                 # Configuration validation
│   └── config_validator.go    # Provider config validation
└── openai/                     # OpenAI provider implementation
    ├── provider.go             # OpenAI provider implementation
    ├── provider_test.go        # Provider unit tests
    ├── session.go             # Session management
    ├── session_state.go       # Clean session state management
    ├── api_strategy.go        # API strategy interfaces
    ├── chat_strategy.go       # Chat Completions strategy
    ├── responses_strategy.go  # Responses API strategy
    ├── rate_limiter.go        # OpenAI-specific rate limiter
    ├── stream.go              # Streaming support
    ├── stream_test.go         # Streaming tests
    └── web_search_preview_tracking.go  # Web search metrics
```

### Global Application State: SharedBag

The **SharedBag** (`pkg/bag/shared_bag.go`) is Mosychlos's global application state manager:

```go
// SharedBag interface for mutable shared state
type SharedBag interface {
    Get(k keys.Key) (any, bool)
    GetAs(k keys.Key, out any) bool     // Type-safe extraction
    Set(k keys.Key, v any)
    Update(k keys.Key, fn func(any) any) // Atomic updates
    Has(k keys.Key) bool
    Snapshot() Bag                       // Immutable snapshot
}
```

#### SharedBag Usage in Mosychlos

**Creation and Initialization** (`cmd/mosychlos/analyze.go`, line 43):

```go
// Created once per command execution
sharedBag := bag.NewSharedBag()
sharedBag.Set(keys.KVerboseMode, verbose)

// Passed to all major components
tools.SetSharedBag(sharedBag)
healthMonitor := health.NewApplicationMonitor(sharedBag)
portfolioService := portfolio.NewService(cfg, filesystem, sharedBag)
```

**Key Data Storage**:

- `keys.KPortfolio` - Current portfolio data
- `keys.KPortfolioNormalizedForAI` - AI-ready portfolio format
- `keys.KProfile` - Investment profile settings
- `keys.KInvestmentResearchResult` - Investment research analysis results
- `keys.KToolComputations` - Tool execution metrics
- `keys.KVerboseMode` - Debug output control
- `keys.KAnalysisResults` - Analysis results for report generation

**Cross-Component Communication**:

- **Portfolio Service**: Caches loaded portfolios (`portfolio/service.go`, line 143)
- **Investment Profile Manager**: Caches user profiles (`profile/manager.go`, line 145)
- **Engine Orchestrator**: Stores analysis results between engine executions
- **Tool System**: Tracks tool computations and metrics
- **Report Generation**: Extracts data for different report types

**Thread-Safe Design**:

```go
type sharedBag struct {
    data map[keys.Key]any
    mu   sync.RWMutex  // Thread-safe access
}
```

#### SharedBag Integration in New LLM Architecture

The new LLM package will integrate with SharedBag for:

1. **Provider State**: Store rate limit information and health metrics
2. **Session Context**: Cache conversation state across engine chains
3. **Tool Coordination**: Share tool execution results with budget system
4. **Configuration**: Store dynamic provider configuration updates

```go
// internal/llm/openai/provider.go
func New(llmConfig config.LLMConfig, sharedBag bag.SharedBag) *Provider {
    // Provider stores SharedBag reference for state coordination
    return &Provider{
        client:    createClient(llmConfig),
        config:    llmConfig,
        sharedBag: sharedBag,  // Global state access
        // ...
    }
}
```

### Interface Architecture

#### Core Interfaces (`pkg/models/ai.go`)

```go
type Provider interface {
    Name() string
    NewSession(system string) Session
    Embedding(ctx context.Context, text string) ([]float64, error)
}

type Session interface {
    Add(role Role, content string)
    AddToolResult(toolCallID, content string)
    Next(ctx context.Context, tools []ToolDef, rf *ResponseFormat) (*AssistantTurn, error)
    NextStream(ctx context.Context, tools []ToolDef, rf *ResponseFormat) (<-chan StreamChunk, error)
    SetToolChoice(t *ToolChoice)
}

type AiClient interface {
    RegisterTool(t Tool)
    SetToolConsumer(consumer ToolConsumer)
}
```

## Architectural Strengths

### 1. **Clean Interface Design**

- Well-defined provider abstraction enables multiple AI providers
- Session interface cleanly separates conversation state from provider
- Tool integration through standardized interfaces

### 2. **Dual API Support**

- Supports both Chat Completions and Responses API
- Web search integration through Responses API
- Automatic API selection based on configuration

### 3. **Advanced Tool Integration**

- Tool constraint system through `ToolConsumer`
- Budget management for tool calls
- Support for required vs optional tools
- Web search budget tracking

### 4. **Structured Output Support**

- Generic `Ask[T]` function for type-safe responses
- Automatic JSON schema generation from Go structs
- OpenAI-specific schema compatibility

### 5. **Comprehensive Error Handling**

- Tool call limits to prevent infinite loops
- Graceful handling of unknown tools
- Detailed logging throughout the system

## Architectural Weaknesses

### 1. **Mixed Responsibilities in Session**

- `session` struct handles both conversation state AND OpenAI-specific API calls
- Web search tracking mixed with core session logic
- Configuration application scattered across methods

### 2. **API Path Duplication**

- Similar logic duplicated between Chat Completions and Responses API
- Parameter application repeated in multiple places
- Tool handling logic differs between API paths

### 3. **Tight Coupling to OpenAI**

- OpenAI-specific types leaked into generic interfaces
- Hard-coded OpenAI behavior in supposedly generic client
- Web search functionality assumes OpenAI Responses API

### 4. **Limited Extensibility**

- Hard to add new providers due to OpenAI-specific assumptions
- Factory pattern underutilized
- No middleware or plugin system

### 5. **Complex State Management**

- Session state mixed with provider configuration
- Tool choice state maintained separately
- Message history management complex

### 6. **Testing Challenges**

- Heavy reliance on integration tests with actual API keys
- Limited unit testing due to tight coupling
- Mock implementations are incomplete

## Rate Limiting Integration Challenges

### Current State

- **No rate limiting for OpenAI API calls**
- Rate limiting exists for external tools but bypassed for AI provider
- No awareness of OpenAI's multi-dimensional rate limits
- No retry logic for rate limit errors

### Integration Points

1. **Provider Level**: Rate limiting should be transparent to session, coordinated via SharedBag
2. **Session Level**: Retry logic needs to be API-aware, with state stored in SharedBag
3. **Client Level**: Budget consumption should consider rate limits, tracked in SharedBag
4. **Factory Level**: Rate limiting should be configurable per provider, with SharedBag coordination

## Refactoring Proposal

### Phase 1: Interface Cleanup and Separation of Concerns

#### 1.1 Extract API Strategy Pattern

```go
// internal/llm/openai/api_strategy.go
type APIStrategy interface {
    Name() string
    Execute(ctx context.Context, session *sessionState, tools []models.ToolDef, rf *models.ResponseFormat) (*models.AssistantTurn, error)
    SupportsStreaming() bool
    ExecuteStream(ctx context.Context, session *sessionState, tools []models.ToolDef, rf *models.ResponseFormat) (<-chan models.StreamChunk, error)
}

type ChatCompletionsStrategy struct {
    client oa.Client
    config config.OpenAIConfig
}

type ResponsesAPIStrategy struct {
    client oa.Client
    config config.OpenAIConfig
    tracker *WebSearchTracker
}
```

#### 1.2 Clean Session State Management

```go
// internal/llm/openai/session_state.go
type sessionState struct {
    messages   []openai.ChatCompletionMessageParamUnion
    toolChoice *models.ToolChoice
    metadata   map[string]any
}

func (s *sessionState) AddMessage(role models.Role, content string) { /* ... */ }
func (s *sessionState) AddToolResult(toolCallID, content string) { /* ... */ }
func (s *sessionState) Clone() *sessionState { /* ... */ }
```

#### 1.3 Refactored Session Implementation

```go
// internal/llm/openai/session.go
type session struct {
    provider  *Provider
    state     *sessionState
    strategy  APIStrategy
}

func (s *session) Next(ctx context.Context, tools []models.ToolDef, rf *models.ResponseFormat) (*models.AssistantTurn, error) {
    return s.strategy.Execute(ctx, s.state, tools, rf)
}
```

#### 1.2 Clean Session State Management

```go
// internal/llm/openai/session_state.go
type sessionState struct {
    messages   []openai.ChatCompletionMessageParamUnion
    toolChoice *models.ToolChoice
    metadata   map[string]any
}

func (s *sessionState) AddMessage(role models.Role, content string) { /* ... */ }
func (s *sessionState) AddToolResult(toolCallID, content string) { /* ... */ }
func (s *sessionState) Clone() *sessionState { /* ... */ }
```

#### 1.3 Refactored Session Implementation

```go
// internal/llm/openai/session.go
type session struct {
    provider  *Provider
    state     *sessionState
    strategy  APIStrategy
}

func (s *session) Next(ctx context.Context, tools []models.ToolDef, rf *models.ResponseFormat) (*models.AssistantTurn, error) {
    return s.strategy.Execute(ctx, s.state, tools, rf)
}
```

### Phase 2: Middleware Architecture for Rate Limiting

#### 2.1 Provider Middleware System

```go
// internal/llm/middleware/middleware.go
type Middleware interface {
    Name() string
    WrapStrategy(strategy APIStrategy) APIStrategy
}

type MiddlewareChain struct {
    middlewares []Middleware
}

func (mc *MiddlewareChain) Apply(strategy APIStrategy) APIStrategy {
    result := strategy
    for i := len(mc.middlewares) - 1; i >= 0; i-- {
        result = mc.middlewares[i].WrapStrategy(result)
    }
    return result
}
```

#### 2.2 Rate Limiting Middleware

```go
// internal/llm/middleware/rate_limit.go
type RateLimitingMiddleware struct {
    limiter *OpenAIRateLimiter
}

func (r *RateLimitingMiddleware) WrapStrategy(strategy APIStrategy) APIStrategy {
    return &rateLimitedStrategy{
        inner:   strategy,
        limiter: r.limiter,
    }
}

type rateLimitedStrategy struct {
    inner   APIStrategy
    limiter *OpenAIRateLimiter
}

func (r *rateLimitedStrategy) Execute(ctx context.Context, session *sessionState, tools []models.ToolDef, rf *models.ResponseFormat) (*models.AssistantTurn, error) {
    // Wait for rate limit capacity
    if err := r.limiter.WaitForCapacity(ctx); err != nil {
        return nil, err
    }

    // Execute with retry logic
    return r.executeWithRetry(ctx, session, tools, rf)
}
```

#### 2.3 Retry Middleware

```go
// internal/llm/middleware/retry.go
type RetryMiddleware struct {
    config RetryConfig
}

func (r *RetryMiddleware) WrapStrategy(strategy APIStrategy) APIStrategy {
    return &retryStrategy{
        inner:  strategy,
        config: r.config,
    }
}
```

### Phase 3: Configuration and Factory Improvements

#### 3.1 Enhanced Provider Factory

```go
// internal/llm/factory.go
type ProviderFactory interface {
    CreateProvider(name string, config any, sharedBag bag.SharedBag) (models.Provider, error)
    RegisterProviderFactory(name string, factory ProviderFactoryFunc)
    ListAvailableProviders() []string
}

type ProviderFactoryFunc func(config any, sharedBag bag.SharedBag) (models.Provider, error)

// Enhanced factory with middleware support
func NewProviderFactory() ProviderFactory {
    return &providerFactory{
        factories: map[string]ProviderFactoryFunc{
            "openai": createOpenAIProvider,
        },
    }
}
```

#### 3.2 Provider Configuration Enhancement

```go
// internal/llm/models/provider_config.go
type ProviderConfig struct {
    LLMConfig        config.LLMConfig
    SharedBag        bag.SharedBag
    Middlewares      []string          // Enabled middleware names
    RateLimiting     *RateLimitConfig  // Rate limiting configuration
    RetryPolicy      *RetryConfig      // Retry configuration
    CustomStrategies map[string]APIStrategy // Custom API strategies
}

// internal/llm/openai/provider.go
func NewWithConfig(cfg ProviderConfig) *Provider {
    // Build middleware chain based on configuration
    chain := &MiddlewareChain{}

    if cfg.RateLimiting != nil && cfg.RateLimiting.Enabled {
        limiter := NewOpenAIRateLimiter(*cfg.RateLimiting)
        chain.Add(NewRateLimitingMiddleware(limiter))
    }

    if cfg.RetryPolicy != nil {
        chain.Add(NewRetryMiddleware(*cfg.RetryPolicy))
    }

    // Select API strategy based on configuration
    strategy := selectAPIStrategy(cfg.LLMConfig)
    strategy = chain.Apply(strategy)

    return &Provider{
        client:   createClient(cfg.LLMConfig),
        config:   cfg.LLMConfig,
        strategy: strategy,
    }
}
```

### Phase 4: Testing Infrastructure Improvements

#### 4.1 Mock Generation with go:generate

Use `//go:generate` annotations to automatically generate mocks for all interfaces:

```go
// internal/llm/interfaces.go
package llm

//go:generate mockgen -source=../../../pkg/models/ai.go -destination=mock/provider_mock.go -package=mock Provider,Session,AiClient
//go:generate mockgen -source=openai/api_strategy.go -destination=mock/strategy_mock.go -package=mock APIStrategy
//go:generate mockgen -source=middleware/middleware.go -destination=mock/middleware_mock.go -package=mock Middleware
//go:generate mockgen -source=openai/rate_limiter.go -destination=mock/rate_limiter_mock.go -package=mock OpenAIRateLimiter

// Mock generation commands:
// go generate ./internal/llm/...
```

```go
// internal/llm/openai/provider.go
package openai

//go:generate mockgen -source=$GOFILE -destination=../mock/openai_provider_mock.go -package=mock

import (
    // ... imports
)

type Provider struct {
    // ... implementation
}
```

```go
// internal/llm/middleware/middleware.go
package middleware

//go:generate mockgen -source=$GOFILE -destination=../mock/middleware_mock.go -package=mock

type Middleware interface {
    Name() string
    WrapStrategy(strategy APIStrategy) APIStrategy
}
```

#### 4.2 Rate Limiting Test Utilities

```go
// internal/llm/openai/rate_limiter_test.go

//go:generate mockgen -source=rate_limiter.go -destination=../mock/rate_limiter_mock.go -package=mock

func NewTestRateLimiter() *OpenAIRateLimiter {
    return NewOpenAIRateLimiter(OpenAIRateLimiterConfig{
        MaxRetries:   3,
        BaseDelay:    10 * time.Millisecond,
        MaxDelay:     100 * time.Millisecond,
        JitterFactor: 0.1,
    })
}

func SimulateRateLimit(t *testing.T, limiter *OpenAIRateLimiter) {
    // Simulate rate limit headers
    headers := http.Header{}
    headers.Set("x-ratelimit-remaining-requests", "0")
    headers.Set("x-ratelimit-reset-requests", "1s")
    limiter.UpdateFromHeaders(headers)
}
```

#### 4.3 Mock Generation Commands

Generate all mocks with a single command:

```bash
# Generate all mocks in the llm package
go generate ./internal/llm/...

# Or generate specific package mocks
go generate ./internal/llm/openai/
go generate ./internal/llm/middleware/
```

#### 4.4 Test Structure with Generated Mocks

````go
// internal/llm/client_test.go
package llm

import (
    "testing"
    "github.com/amaurybrisou/mosychlos/internal/llm/mock"
    "github.com/golang/mock/gomock"
)

func TestClient_WithMocks(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    // Use generated mocks
    mockProvider := mock.NewMockProvider(ctrl)
    mockSession := mock.NewMockSession(ctrl)
    mockStrategy := mock.NewMockAPIStrategy(ctrl)

    // Configure mock expectations
    mockProvider.EXPECT().NewSession(gomock.Any()).Return(mockSession)
    mockSession.EXPECT().Next(gomock.Any(), gomock.Any(), gomock.Any()).Return(&models.AssistantTurn{}, nil)

    // Test implementation with mocks
    client := NewClient(mockProvider, "test system", nil)
    result, err := Ask[string](ctx, client, "test input")

    assert.NoError(t, err)
    assert.NotEmpty(t, result)
}

### Phase 5: Advanced Features

#### 5.1 Provider Health Monitoring

```go
// internal/llm/health/provider_monitor.go
type ProviderHealthMonitor struct {
    providers map[string]*ProviderHealth
    mutex     sync.RWMutex
}

type ProviderHealth struct {
    Name              string
    Status            string // "healthy", "degraded", "down"
    LastSuccess       time.Time
    LastFailure       time.Time
    SuccessRate       float64
    AverageLatency    time.Duration
    RateLimitHits     int64
    CurrentLimits     map[string]int64
    RecentErrors      []string
}
````

#### 5.2 Configuration Validation

```go
// internal/llm/validation/config_validator.go
type ConfigValidator struct{}

func (v *ConfigValidator) ValidateProviderConfig(name string, config any) error {
    switch name {
    case "openai":
        return v.validateOpenAIConfig(config)
    default:
        return fmt.Errorf("unknown provider: %s", name)
    }
}

func (v *ConfigValidator) validateOpenAIConfig(config any) error {
    cfg, ok := config.(config.LLMConfig)
    if !ok {
        return fmt.Errorf("invalid OpenAI configuration type")
    }

    if cfg.APIKey == "" {
        return fmt.Errorf("OpenAI API key is required")
    }

    if cfg.OpenAI.RateLimit != nil {
        return v.validateRateLimitConfig(cfg.OpenAI.RateLimit)
    }

    return nil
}
```

## Implementation Benefits

### 1. **Separation of Concerns**

- Clean separation between conversation state and API execution
- Middleware pattern allows for cross-cutting concerns
- Strategy pattern enables different API approaches

### 2. **Extensibility**

- Easy to add new providers without changing existing code
- Middleware system supports plugins and extensions
- Configuration-driven behavior

### 3. **Testability**

- Clear interfaces enable comprehensive unit testing
- Mock providers for integration testing
- Isolated testing of rate limiting logic

### 4. **Maintainability**

- Single responsibility principle enforced
- Reduced code duplication between API paths
- Clear configuration and initialization flow

### 5. **Rate Limiting Integration**

- Transparent rate limiting through middleware
- Configurable retry policies
- Multi-dimensional rate limit awareness

### 6. **Performance**

- Reduced memory allocations through object pooling
- Optimized message handling
- Efficient middleware chain execution

## Migration Strategy

### Phase 1: Internal Refactoring (Week 1-2)

1. Extract API strategies without changing public interfaces
2. Implement session state management
3. Add comprehensive unit tests

### Phase 2: Middleware Integration (Week 3-4)

1. Implement middleware architecture
2. Add rate limiting middleware with configuration
3. Add retry middleware with exponential backoff

### Phase 3: Enhanced Factory (Week 5)

1. Upgrade provider factory with new features
2. Add configuration validation
3. Implement health monitoring

### Phase 4: Testing and Documentation (Week 6)

1. Complete test coverage
2. Update documentation
3. Performance benchmarking

### Phase 5: Optional Enhancements (Week 7+)

1. Add new provider support
2. Implement advanced monitoring
3. Custom middleware development

## Backwards Compatibility

**No backwards compatibility maintained** - this is a complete refactoring to the new `llm` package with breaking changes to enable proper architecture.

### Migration Path

1. **New Package Structure**: Move from `internal/ai/` to `internal/llm/`
2. **Updated Imports**: All imports need to be updated to use new package paths
3. **Interface Changes**: Some interfaces may have breaking changes for better design
4. **Configuration Updates**: New configuration structure for rate limiting and middleware

### Justification for Breaking Changes

The current architecture has fundamental design issues that cannot be resolved with backwards-compatible changes:

- Mixed responsibilities in session management
- Tight coupling to OpenAI specifics
- No middleware support for cross-cutting concerns
- Limited extensibility for new providers

A clean break enables a properly designed architecture that will be more maintainable long-term.

## Risk Assessment

### Low Risk Changes

- Internal refactoring with interface preservation
- Additive configuration options
- Optional middleware components

### Medium Risk Changes

- Strategy pattern implementation
- Session state management changes
- Testing infrastructure updates

### High Risk Changes

- Rate limiting integration (requires careful testing)
- Message handling modifications
- Provider factory changes

### Mitigation Strategies

1. **Comprehensive Testing**: Unit tests for all new components
2. **Gradual Rollout**: Feature flags for new functionality
3. **Monitoring**: Health checks and performance metrics
4. **Fallback Options**: Disable middleware if issues occur

## Success Metrics

### Functional Metrics

- Zero regression in existing functionality
- Successful rate limiting without API errors
- Improved retry success rate for transient failures

### Performance Metrics

- No increase in latency for normal operations
- Reduced memory allocations
- Improved throughput under rate limits

### Code Quality Metrics

- Increased test coverage (target: >90%)
- Reduced cyclomatic complexity
- Improved maintainability index

### Reliability Metrics

- Reduced rate limit errors (429 responses)
- Improved success rate for investment research analysis
- Better handling of transient failures

## Conclusion

This refactoring proposal transforms the AI package from a tightly coupled, OpenAI-specific implementation into a flexible, extensible, and well-tested architecture that supports:

1. **Clean rate limiting integration** through middleware
2. **Multiple AI provider support** through improved abstraction
3. **Better testability** through dependency injection and mocking
4. **Enhanced maintainability** through separation of concerns
5. **Production readiness** through comprehensive error handling and monitoring

The proposed architecture maintains full backwards compatibility while enabling the advanced rate limiting features required for production-grade investment research analysis in Mosychlos.

The middleware-based approach ensures that rate limiting, retry logic, and other cross-cutting concerns are handled transparently, allowing the core business logic to remain focused on portfolio analysis and decision-making.
