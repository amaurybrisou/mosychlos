// pkg/openai/retry.go
package openai

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

type DoFunc func(ctx context.Context) (*http.Response, error)

func WithRetry(ctx context.Context, cfg models.RetryConfig, do DoFunc) (*http.Response, error) {
	var lastErr error
	delay := cfg.BaseDelay
	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		resp, err := do(ctx)
		if err == nil && resp != nil && resp.StatusCode < 500 && resp.StatusCode != 429 {
			return resp, nil
		}
		if resp != nil && resp.StatusCode == 429 {
			lastErr = errors.New("rate limited")
		} else if err != nil {
			lastErr = err
		}
		// backoff
		sleep := jitter(delay, cfg.JitterFactor)
		if sleep > cfg.MaxDelay {
			sleep = cfg.MaxDelay
		}
		t := time.NewTimer(sleep)
		select {
		case <-ctx.Done():
			t.Stop()
			return nil, ctx.Err()
		case <-t.C:
		}
		delay = time.Duration(float64(delay) * cfg.ExponentialBase)
	}
	return nil, lastErr
}
