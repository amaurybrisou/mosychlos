// pkg/openai/middleware.go
package openai

import (
	"context"
	"net/http"

	"github.com/amaurybrisou/mosychlos/internal/config"
)

type MiddlewareChain struct {
	cfg     config.OpenAIConfig
	limiter Limiter
}

func NewDefaultMiddlewareChain(cfg config.OpenAIConfig) *MiddlewareChain {
	return &MiddlewareChain{
		cfg:     cfg,
		limiter: NewRPMLimiter(cfg.RateLimit),
	}
}

func (m *MiddlewareChain) Do(ctx context.Context, do DoFunc) (*http.Response, error) {
	if m.cfg.RateLimit.Enabled {
		if err := m.limiter.Wait(ctx); err != nil {
			return nil, err
		}
	}
	resp, err := WithRetry(ctx, m.cfg.Retry, do)
	if resp != nil {
		m.limiter.Update(parseRateHeaders(resp.Header))
	}
	return resp, err
}
