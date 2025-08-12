// pkg/openai/client.go
package openai

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
)

//go:generate mockgen -source=client.go -destination=mocks/client_mock.go -package=mocks

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	doer       HTTPDoer
	middleware *MiddlewareChain
}

func NewClient(doer HTTPDoer, cfg config.OpenAIConfig) *Client {
	return &Client{
		doer:       doer,
		middleware: NewDefaultMiddlewareChain(cfg),
	}
}

func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return c.middleware.Do(ctx, func(ctx context.Context) (*http.Response, error) {
		req = req.WithContext(ctx)
		return c.doer.Do(req)
	})
}

// pkg/openai/client.go  (replace NewHTTPClient)
func NewHTTPClient(mw ...RTMiddleware) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = (&net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 60 * time.Second,
	}).DialContext
	transport.TLSHandshakeTimeout = 10 * time.Second
	transport.IdleConnTimeout = 90 * time.Second
	transport.MaxIdleConns = 100
	transport.MaxConnsPerHost = 0
	transport.MaxIdleConnsPerHost = 100

	rt := http.RoundTripper(transport)
	if len(mw) > 0 {
		rt = ChainRoundTripper(rt, mw...)
	}
	return &http.Client{Transport: rt, Timeout: 30 * time.Second}
}
