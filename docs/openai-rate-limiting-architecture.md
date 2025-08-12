# OpenAI Rate Limiting Architecture Proposal

## Executive Summary

This document proposes implementing a comprehensive rate limiting system for OpenAI API calls in Mosychlos, leveraging the middleware pattern from the OpenAI Go SDK and implementing exponential backoff retry mechanisms based on OpenAI's best practices.

## Current State Analysis

### Existing Implementation

The current OpenAI integration consists of:

1. **Provider Layer** (`internal/ai/openai/provider.go`)

   - Basic OpenAI client initialization
   - No rate limiting mechanisms
   - Simple HTTP client configuration

2. **Session Management** (`internal/ai/openai/session.go`)

   - Two API paths: Chat Completions and Responses API
   - Direct API calls without retry logic
   - Basic error handling

3. **Response API Integration** (`internal/ai/openai/response-api.go`)

   - Web search integration with budget tracking
   - No OpenAI rate limit awareness
   - Complex request/response handling

4. **AI Client** (`internal/ai/client.go`)
   - Tool orchestration and session management
   - Max tool calls protection (infinite loop prevention)
   - No OpenAI-specific rate limiting

### Current Rate Limiting

The codebase has rate limiting for **external tools** (FMP, FRED, NewsAPI, etc.) but **no rate limiting for OpenAI API calls**:

```go
// External tools have rate limiting
type RateLimitedTool struct {
    tool        models.Tool
    rateLimiter *RateLimiter
    toolName    string
}
```

However, OpenAI API calls bypass this system entirely.

## OpenAI Rate Limiting Requirements

### Rate Limit Dimensions

OpenAI enforces rate limits across **five dimensions**:

1. **RPM** - Requests per minute
2. **RPD** - Requests per day
3. **TPM** - Tokens per minute
4. **TPD** - Tokens per day
5. **IPM** - Images per minute (for vision models)

### Rate Limit Headers

OpenAI provides rate limit information in response headers:

```http
x-ratelimit-limit-requests: 60
x-ratelimit-limit-tokens: 150000
x-ratelimit-remaining-requests: 59
x-ratelimit-remaining-tokens: 149984
x-ratelimit-reset-requests: 1s
x-ratelimit-reset-tokens: 6m0s
```

### Usage Tiers

Rate limits vary by organization tier:

| Tier   | Qualification         | Usage Limit    |
| ------ | --------------------- | -------------- |
| Free   | Allowed geography     | $100/month     |
| Tier 1 | $5 paid               | $100/month     |
| Tier 2 | $50 paid + 7 days     | $500/month     |
| Tier 3 | $100 paid + 7 days    | $1,000/month   |
| Tier 4 | $250 paid + 14 days   | $5,000/month   |
| Tier 5 | $1,000 paid + 30 days | $200,000/month |

## Architecture Proposal

### 1. Middleware-Based Rate Limiting

Implement rate limiting using OpenAI's middleware pattern for transparent integration:

```go
// internal/ai/openai/middleware.go
package openaiprov

import (
    "context"
    "fmt"
    "log/slog"
    "net/http"
    "strconv"
    "time"

    "github.com/openai/openai-go/v2/option"
)

type RateLimitMiddleware struct {
    limiter *OpenAIRateLimiter
}

func NewRateLimitMiddleware(limiter *OpenAIRateLimiter) *RateLimitMiddleware {
    return &RateLimitMiddleware{limiter: limiter}
}

func (m *RateLimitMiddleware) Middleware() option.MiddlewareFunc {
    return func(req *http.Request, next option.MiddlewareNext) (*http.Response, error) {
        start := time.Now()

        // Pre-request: Check rate limits
        if err := m.limiter.WaitForCapacity(req.Context()); err != nil {
            return nil, fmt.Errorf("rate limit exceeded: %w", err)
        }

        // Execute request
        res, err := next(req)

        duration := time.Since(start)

        // Post-request: Update rate limit state from headers
        if res != nil {
            m.limiter.UpdateFromHeaders(res.Header)
        }

        // Log rate limit metrics
        m.logRateLimitMetrics(req, res, err, duration)

        return res, err
    }
}
```

### 2. OpenAI-Specific Rate Limiter

Create a specialized rate limiter that understands OpenAI's multi-dimensional limits:

```go
// internal/ai/openai/rate_limiter.go
package openaiprov

import (
    "context"
    "fmt"
    "net/http"
    "strconv"
    "sync"
    "time"
)

type OpenAIRateLimiter struct {
    // Current limits (updated from headers)
    requestsPerMinute    int64
    tokensPerMinute      int64
    remainingRequests    int64
    remainingTokens      int64
    requestResetTime     time.Time
    tokenResetTime       time.Time

    // State management
    mu                   sync.RWMutex
    lastHeaderUpdate     time.Time

    // Configuration
    maxRetries           int
    baseDelay           time.Duration
    maxDelay            time.Duration
    jitterFactor        float64

    // Metrics
    totalRequests       int64
    rateLimitHits       int64
    retriesExecuted     int64
}

func NewOpenAIRateLimiter(config OpenAIRateLimiterConfig) *OpenAIRateLimiter {
    return &OpenAIRateLimiter{
        maxRetries:      config.MaxRetries,
        baseDelay:       config.BaseDelay,
        maxDelay:        config.MaxDelay,
        jitterFactor:    config.JitterFactor,

        // Conservative defaults until we get actual limits from headers
        requestsPerMinute: 60,
        tokensPerMinute:   150000,
        remainingRequests: 60,
        remainingTokens:   150000,
    }
}

// WaitForCapacity blocks until capacity is available
func (rl *OpenAIRateLimiter) WaitForCapacity(ctx context.Context) error {
    rl.mu.RLock()
    hasCapacity := rl.remainingRequests > 0 && rl.remainingTokens > 1000 // Buffer for tokens
    rl.mu.RUnlock()

    if hasCapacity {
        return nil
    }

    // Calculate wait time based on reset times
    rl.mu.RLock()
    waitTime := rl.calculateWaitTime()
    rl.mu.RUnlock()

    slog.Warn("OpenAI rate limit reached, waiting",
        "wait_duration", waitTime,
        "remaining_requests", rl.remainingRequests,
        "remaining_tokens", rl.remainingTokens)

    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-time.After(waitTime):
        return nil
    }
}

// UpdateFromHeaders updates rate limit state from response headers
func (rl *OpenAIRateLimiter) UpdateFromHeaders(headers http.Header) {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    if limit := headers.Get("x-ratelimit-limit-requests"); limit != "" {
        if val, err := strconv.ParseInt(limit, 10, 64); err == nil {
            rl.requestsPerMinute = val
        }
    }

    if limit := headers.Get("x-ratelimit-limit-tokens"); limit != "" {
        if val, err := strconv.ParseInt(limit, 10, 64); err == nil {
            rl.tokensPerMinute = val
        }
    }

    if remaining := headers.Get("x-ratelimit-remaining-requests"); remaining != "" {
        if val, err := strconv.ParseInt(remaining, 10, 64); err == nil {
            rl.remainingRequests = val
        }
    }

    if remaining := headers.Get("x-ratelimit-remaining-tokens"); remaining != "" {
        if val, err := strconv.ParseInt(remaining, 10, 64); err == nil {
            rl.remainingTokens = val
        }
    }

    // Parse reset times
    if reset := headers.Get("x-ratelimit-reset-requests"); reset != "" {
        if duration, err := time.ParseDuration(reset); err == nil {
            rl.requestResetTime = time.Now().Add(duration)
        }
    }

    if reset := headers.Get("x-ratelimit-reset-tokens"); reset != "" {
        if duration, err := time.ParseDuration(reset); err == nil {
            rl.tokenResetTime = time.Now().Add(duration)
        }
    }

    rl.lastHeaderUpdate = time.Now()
}
```

### 3. Exponential Backoff Retry Middleware

Implement retry logic with exponential backoff for rate limit errors:

```go
// internal/ai/openai/retry_middleware.go
package openaiprov

import (
    "context"
    "math"
    "math/rand"
    "net/http"
    "time"

    "github.com/openai/openai-go/v2"
    "github.com/openai/openai-go/v2/option"
)

type RetryMiddleware struct {
    maxRetries      int
    baseDelay       time.Duration
    maxDelay        time.Duration
    exponentialBase float64
    jitterFactor    float64
}

func NewRetryMiddleware(config RetryConfig) *RetryMiddleware {
    return &RetryMiddleware{
        maxRetries:      config.MaxRetries,
        baseDelay:       config.BaseDelay,
        maxDelay:        config.MaxDelay,
        exponentialBase: config.ExponentialBase,
        jitterFactor:    config.JitterFactor,
    }
}

func (r *RetryMiddleware) Middleware() option.MiddlewareFunc {
    return func(req *http.Request, next option.MiddlewareNext) (*http.Response, error) {
        var lastErr error

        for attempt := 0; attempt <= r.maxRetries; attempt++ {
            if attempt > 0 {
                // Calculate delay with exponential backoff and jitter
                delay := r.calculateDelay(attempt)

                slog.Debug("Retrying OpenAI request after rate limit",
                    "attempt", attempt,
                    "delay", delay,
                    "max_retries", r.maxRetries)

                select {
                case <-req.Context().Done():
                    return nil, req.Context().Err()
                case <-time.After(delay):
                    // Continue with retry
                }
            }

            res, err := next(req)

            // Check if this is a rate limit error
            if err != nil {
                if isRateLimitError(err) {
                    lastErr = err
                    continue // Retry
                }
                return res, err // Non-rate-limit error, don't retry
            }

            // Success
            return res, nil
        }

        return nil, fmt.Errorf("max retries (%d) exceeded for rate limit: %w", r.maxRetries, lastErr)
    }
}

func (r *RetryMiddleware) calculateDelay(attempt int) time.Duration {
    // Exponential backoff: baseDelay * (exponentialBase ^ attempt)
    delay := float64(r.baseDelay) * math.Pow(r.exponentialBase, float64(attempt-1))

    // Add jitter
    if r.jitterFactor > 0 {
        jitter := rand.Float64() * r.jitterFactor * delay
        delay += jitter
    }

    // Cap at max delay
    if time.Duration(delay) > r.maxDelay {
        delay = float64(r.maxDelay)
    }

    return time.Duration(delay)
}

func isRateLimitError(err error) bool {
    // Check for OpenAI rate limit errors
    if apiErr, ok := err.(*openai.Error); ok {
        return apiErr.Code == "rate_limit_exceeded"
    }

    // Check HTTP status codes
    if httpErr, ok := err.(*http.Response); ok {
        return httpErr.StatusCode == 429
    }

    return false
}
```

### 4. Configuration Integration

Add rate limiting configuration to the existing config system:

```go
// internal/config/config.go - Add to OpenAIConfig struct

type OpenAIConfig struct {
    // ... existing fields

    RateLimit *OpenAIRateLimitConfig `mapstructure:"rate_limit" yaml:"rate_limit"`
    Retry     *RetryConfig           `mapstructure:"retry" yaml:"retry"`
}

type OpenAIRateLimitConfig struct {
    Enabled        bool          `mapstructure:"enabled" yaml:"enabled"`
    MaxRetries     int           `mapstructure:"max_retries" yaml:"max_retries"`
    BaseDelay      time.Duration `mapstructure:"base_delay" yaml:"base_delay"`
    MaxDelay       time.Duration `mapstructure:"max_delay" yaml:"max_delay"`
    JitterFactor   float64       `mapstructure:"jitter_factor" yaml:"jitter_factor"`
    LogMetrics     bool          `mapstructure:"log_metrics" yaml:"log_metrics"`
}

type RetryConfig struct {
    MaxRetries      int           `mapstructure:"max_retries" yaml:"max_retries"`
    BaseDelay       time.Duration `mapstructure:"base_delay" yaml:"base_delay"`
    MaxDelay        time.Duration `mapstructure:"max_delay" yaml:"max_delay"`
    ExponentialBase float64       `mapstructure:"exponential_base" yaml:"exponential_base"`
    JitterFactor    float64       `mapstructure:"jitter_factor" yaml:"jitter_factor"`
}
```

### 5. Provider Integration

Update the OpenAI provider to use the middleware:

```go
// internal/ai/openai/provider.go - Updated New function

func New(llmConfig config.LLMConfig, sharedBag bag.SharedBag) *Provider {
    slog.Debug("OpenAI.New: initializing provider", "model", llmConfig.Model, "base_url", llmConfig.BaseURL)

    opts := oa.DefaultClientOptions()
    opts = append(opts, option.WithAPIKey(llmConfig.APIKey))

    if llmConfig.BaseURL != "" {
        slog.Debug("OpenAI.New: using custom base URL", "base_url", llmConfig.BaseURL)
        opts = append(opts, option.WithBaseURL(llmConfig.BaseURL))
    }

    // Add rate limiting middleware if enabled
    if llmConfig.OpenAI.RateLimit != nil && llmConfig.OpenAI.RateLimit.Enabled {
        rateLimiter := NewOpenAIRateLimiter(*llmConfig.OpenAI.RateLimit)
        rateLimitMiddleware := NewRateLimitMiddleware(rateLimiter)
        opts = append(opts, option.WithMiddleware(rateLimitMiddleware.Middleware()))

        slog.Info("OpenAI rate limiting enabled",
            "max_retries", llmConfig.OpenAI.RateLimit.MaxRetries,
            "base_delay", llmConfig.OpenAI.RateLimit.BaseDelay,
            "max_delay", llmConfig.OpenAI.RateLimit.MaxDelay)
    }

    // Add retry middleware if enabled
    if llmConfig.OpenAI.Retry != nil {
        retryMiddleware := NewRetryMiddleware(*llmConfig.OpenAI.Retry)
        opts = append(opts, option.WithMiddleware(retryMiddleware.Middleware()))

        slog.Info("OpenAI retry middleware enabled",
            "max_retries", llmConfig.OpenAI.Retry.MaxRetries,
            "exponential_base", llmConfig.OpenAI.Retry.ExponentialBase)
    }

    return &Provider{
        client:    oa.NewClient(opts...),
        config:    llmConfig,
        sharedBag: sharedBag,
    }
}
```

## Configuration Examples

### Default Configuration

```yaml
# config/config.default.yaml
llm:
  provider: openai
  model: gpt-4o
  openai:
    rate_limit:
      enabled: true
      max_retries: 5
      base_delay: 1s
      max_delay: 60s
      jitter_factor: 0.1
      log_metrics: true
    retry:
      max_retries: 3
      base_delay: 1s
      max_delay: 30s
      exponential_base: 2.0
      jitter_factor: 0.1
```

### Production Configuration

```yaml
# Production settings for high-tier usage
llm:
  openai:
    rate_limit:
      enabled: true
      max_retries: 10
      base_delay: 500ms
      max_delay: 30s
      jitter_factor: 0.2
      log_metrics: false # Reduce logging in production
    retry:
      max_retries: 5
      base_delay: 1s
      max_delay: 60s
      exponential_base: 1.5
      jitter_factor: 0.15
```

## Benefits

### 1. Transparent Integration

- Uses OpenAI's official middleware pattern
- No changes to existing session or API call logic
- Automatic rate limit handling across all API calls

### 2. Intelligent Rate Limiting

- Multi-dimensional rate limit awareness (RPM, TPM, RPD, TPD)
- Real-time limit updates from response headers
- Predictive capacity management

### 3. Robust Retry Logic

- Exponential backoff with jitter
- Configurable retry policies
- Rate limit error detection and handling

### 4. Comprehensive Monitoring

- Rate limit metrics and logging
- Request/response timing
- Retry attempt tracking
- Budget consumption awareness

### 5. Configuration Flexibility

- Environment-specific settings
- Fine-tuned retry policies
- Optional middleware components

## Implementation Phases

### Phase 1: Core Middleware

1. Implement `OpenAIRateLimiter` with header parsing
2. Create `RateLimitMiddleware` with basic capacity checking
3. Add configuration structure
4. Update provider initialization

### Phase 2: Retry Logic

1. Implement `RetryMiddleware` with exponential backoff
2. Add rate limit error detection
3. Integrate with existing error handling
4. Test retry scenarios

### Phase 3: Monitoring & Optimization

1. Add comprehensive metrics logging
2. Implement rate limit dashboards
3. Optimize wait time calculations
4. Add usage tier detection

### Phase 4: Advanced Features

1. Token usage prediction
2. Request batching optimization
3. Multi-client rate limit coordination
4. Dynamic configuration updates

## Risk Assessment

### Low Risk

- Middleware pattern is official OpenAI approach
- Existing functionality remains unchanged
- Configuration is optional and backwards-compatible

### Medium Risk

- Additional latency from rate limit checks
- Complexity in multi-dimensional limit tracking
- Potential for false positive rate limit detection

### Mitigation Strategies

- Comprehensive testing with various rate limit scenarios
- Monitoring and alerting for rate limit effectiveness
- Gradual rollout with feature flags
- Fallback to direct API calls if middleware fails

## Success Metrics

### Performance Metrics

- Reduction in rate limit errors (429 responses)
- Average request success rate improvement
- Request latency impact measurement

### Reliability Metrics

- Successful retry completion rate
- Rate limit prediction accuracy
- System stability under high load

### Usage Metrics

- Token usage efficiency
- Cost optimization through better rate management
- Request distribution across time periods

## Conclusion

This architecture proposal provides a comprehensive solution for OpenAI rate limiting that:

1. **Leverages official patterns** from OpenAI's Go SDK
2. **Integrates transparently** with existing code
3. **Provides intelligent rate management** across all limit dimensions
4. **Implements proven retry strategies** with exponential backoff
5. **Offers comprehensive monitoring** and configuration options

The middleware-based approach ensures that rate limiting is applied consistently across all OpenAI API calls (Chat Completions, Responses API, Embeddings) without requiring changes to the existing business logic.

Implementation should proceed in phases to ensure stability and allow for iterative improvement based on real-world usage patterns and rate limit behaviors.
