// pkg/openai/headers.go
package openai

import (
	"net/http"
	"strconv"
	"time"
)

type RateHeaders struct {
	RemainingRPM int
	RemainingTPM int
	ResetAt      time.Time // best-effort (from Retry-After or vendor hint)
}

func parseRateHeaders(h http.Header) RateHeaders {
	rh := RateHeaders{}
	if v := h.Get("x-ratelimit-remaining-requests"); v != "" {
		if n, _ := strconv.Atoi(v); n >= 0 {
			rh.RemainingRPM = n
		}
	}
	if v := h.Get("x-ratelimit-remaining-tokens"); v != "" {
		if n, _ := strconv.Atoi(v); n >= 0 {
			rh.RemainingTPM = n
		}
	}
	if ra := h.Get("retry-after"); ra != "" {
		// seconds or HTTP-date—support seconds first
		if s, err := strconv.Atoi(ra); err == nil {
			rh.ResetAt = time.Now().Add(time.Duration(s) * time.Second)
		}
	}
	return rh
}
