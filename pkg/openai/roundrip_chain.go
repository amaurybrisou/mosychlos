// pkg/openai/roundtrip_chain.go
package openai

import "net/http"

type RTMiddleware func(next http.RoundTripper) http.RoundTripper

// ChainRoundTripper(base, A, B, C) == A(B(C(base)))
func ChainRoundTripper(base http.RoundTripper, mws ...RTMiddleware) http.RoundTripper {
	rt := base
	for i := len(mws) - 1; i >= 0; i-- {
		rt = mws[i](rt)
	}
	return rt
}
