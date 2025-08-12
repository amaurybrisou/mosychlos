package errors

import (
	"errors"
	"fmt"
)

var (
	ErrContextSourceNotFound = errors.New("context source not found")
	ErrRateLimitExceeded     = errors.New("context rate limit exceeded")
	ErrInvalidAPIKey         = errors.New("invalid API key")
	ErrDataNotAvailable      = errors.New("context data not available")
	ErrQuotaExceeded         = errors.New("API quota exceeded")
	ErrNoContextSources      = errors.New("no context sources registered")
	ErrMissingInputKeys      = errors.New("missing input keys")
)

func ContextSourceNotFoundError(kind string) error {
	return fmt.Errorf("%w: %s", ErrContextSourceNotFound, kind)
}
func RateLimitExceededError(kind string) error {
	return fmt.Errorf("%w: %s", ErrRateLimitExceeded, kind)
}
func InvalidAPIKeyError(provider string) error {
	return fmt.Errorf("%w: %s", ErrInvalidAPIKey, provider)
}
func DataNotAvailableError(kind, ticker string) error {
	return fmt.Errorf("%w: %s for %s", ErrDataNotAvailable, kind, ticker)
}
func QuotaExceededError(provider string) error {
	return fmt.Errorf("%w: %s", ErrQuotaExceeded, provider)
}
func NoContextSourcesError() error { return ErrNoContextSources }

// MissingInputKeysError wraps ErrMissingInputKeys with step name & keys list.
func MissingInputKeysError(step string, keys []any) error {
	return fmt.Errorf("%w: %s needs %v", ErrMissingInputKeys, step, keys)
}
