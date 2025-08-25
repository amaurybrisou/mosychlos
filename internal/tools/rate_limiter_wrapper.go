package tools

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// RateLimiter manages rate limiting for tools
type RateLimiter struct {
	requestsPerSecond int
	requestsPerDay    int
	burst             int

	// Token bucket for per-second limiting
	tokens     int
	maxTokens  int
	lastRefill time.Time
	tokenMux   sync.Mutex

	// Daily request counter
	dailyCount     int
	dailyResetTime time.Time
	dailyMux       sync.Mutex
}

var _ models.Tool = (*RateLimitedTool)(nil)

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond, requestsPerDay, burst int) *RateLimiter {
	if burst == 0 {
		burst = requestsPerSecond
	}

	now := time.Now()
	return &RateLimiter{
		requestsPerSecond: requestsPerSecond,
		requestsPerDay:    requestsPerDay,
		burst:             burst,
		tokens:            burst,
		maxTokens:         burst,
		lastRefill:        now,
		dailyResetTime:    now.Add(24 * time.Hour),
	}
}

// Allow checks if a request is allowed and consumes a token
func (rl *RateLimiter) Allow() bool {
	rl.tokenMux.Lock()
	defer rl.tokenMux.Unlock()

	// Refill tokens based on time passed
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)
	tokensToAdd := int(elapsed.Seconds()) * rl.requestsPerSecond

	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		if rl.tokens > rl.maxTokens {
			rl.tokens = rl.maxTokens
		}
		rl.lastRefill = now
	}

	// Check daily limit
	rl.dailyMux.Lock()
	if now.After(rl.dailyResetTime) {
		rl.dailyCount = 0
		rl.dailyResetTime = now.Add(24 * time.Hour)
	}

	if rl.dailyCount >= rl.requestsPerDay {
		rl.dailyMux.Unlock()
		return false
	}
	rl.dailyMux.Unlock()

	// Check if we have tokens available
	if rl.tokens <= 0 {
		return false
	}

	// Consume token
	rl.tokens--
	rl.dailyCount++

	return true
}

// Wait blocks until a request is allowed
func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		if rl.Allow() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			continue
		}
	}
}

// Stats returns current rate limiter statistics
func (rl *RateLimiter) Stats() map[string]any {
	rl.tokenMux.Lock()
	rl.dailyMux.Lock()
	defer rl.tokenMux.Unlock()
	defer rl.dailyMux.Unlock()

	return map[string]any{
		"tokens_available":     rl.tokens,
		"daily_requests_used":  rl.dailyCount,
		"daily_requests_limit": rl.requestsPerDay,
		"requests_per_second":  rl.requestsPerSecond,
		"next_daily_reset":     rl.dailyResetTime.Format(time.RFC3339),
	}
}

// RateLimitedTool wraps a tool with rate limiting
type RateLimitedTool struct {
	tool        models.Tool
	rateLimiter *RateLimiter
	toolName    string
}

// NewRateLimitedTool creates a rate-limited tool wrapper
func NewRateLimitedTool(tool models.Tool, requestsPerSecond, requestsPerDay, burst int) *RateLimitedTool {
	return &RateLimitedTool{
		tool:        tool,
		rateLimiter: NewRateLimiter(requestsPerSecond, requestsPerDay, burst),
		toolName:    tool.Name(),
	}
}

// Name implements the Tool interface
func (rt *RateLimitedTool) Name() string {
	return rt.tool.Name()
}

// Key implements the Tool interface
func (rt *RateLimitedTool) Key() bag.Key {
	return rt.tool.Key()
}

// Description implements the Tool interface
func (rt *RateLimitedTool) Description() string {
	return rt.tool.Description()
}

func (rt *RateLimitedTool) IsExternal() bool {
	return rt.tool.IsExternal()
}

// Definition implements the Tool interface
func (rt *RateLimitedTool) Definition() models.ToolDef {
	return rt.tool.Definition()
}

// Tags implements the Tool interface
func (rt *RateLimitedTool) Tags() []string {
	return rt.tool.Tags()
}

// Run implements the Tool interface with rate limiting
func (rt *RateLimitedTool) Run(ctx context.Context, args any) (any, error) {
	// Check rate limit before executing
	err := rt.rateLimiter.Wait(ctx)
	if err != nil {
		slog.Warn("Rate limit wait cancelled",
			"tool", rt.toolName,
			"error", err,
		)
		return "", fmt.Errorf("rate limit wait cancelled: %w", err)
	}

	// Log rate limit stats periodically (every 10th request)
	if rt.rateLimiter.dailyCount%10 == 0 {
		stats := rt.rateLimiter.Stats()
		slog.Debug("Rate limiter stats",
			"tool", rt.toolName,
			"stats", stats,
		)
	}

	// Execute the actual tool
	return rt.tool.Run(ctx, args)
}
