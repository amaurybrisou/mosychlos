// pkg/openai/limiter.go
package openai

import (
	"context"
	"sync"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

//go:generate mockgen -source=limiter.go -destination=mocks/limiter_mock.go -package=mocks

type Limiter interface {
	Wait(ctx context.Context) error
	Update(rh RateHeaders)
}

type rpmLimiter struct {
	cfg   models.RateLimitConfig
	mu    sync.Mutex
	until time.Time
}

func NewRPMLimiter(cfg models.RateLimitConfig) Limiter {
	return &rpmLimiter{cfg: cfg}
}

func (l *rpmLimiter) Wait(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	if now.Before(l.until) {
		wait := time.Until(l.until)
		t := time.NewTimer(wait)
		defer t.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}
	}
	return nil
}

func (l *rpmLimiter) Update(rh RateHeaders) {
	if !l.cfg.Enabled {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	// If server says to back off (via ResetAt), respect it.
	if !rh.ResetAt.IsZero() && rh.ResetAt.After(time.Now()) {
		l.until = rh.ResetAt
		return
	}

	// Soft backoff when we’re close to empty.
	if rh.RemainingRPM == 0 {
		delay := jitter(l.cfg.BaseDelay, l.cfg.JitterFactor)
		if delay > l.cfg.MaxDelay {
			delay = l.cfg.MaxDelay
		}
		l.until = time.Now().Add(delay)
	}
}

func jitter(base time.Duration, factor float64) time.Duration {
	if factor <= 0 {
		return base
	}
	j := base.Seconds() * factor
	return time.Duration((base.Seconds() + (randFloat()*2-1)*j) * float64(time.Second))
}

func randFloat() float64 { return float64(time.Now().UnixNano()%1000) / 1000.0 }
