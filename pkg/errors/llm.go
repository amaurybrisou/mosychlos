package errors

import (
	"errors"
	"fmt"
)

// Common error types for LLM operations
var (
	// ErrMissingAPIKey is returned when an API key is required but not provided
	ErrMissingAPIKey = errors.New("missing API key")

	// ErrInvalidProvider is returned when an unsupported LLM provider is specified
	ErrInvalidProvider = errors.New("invalid LLM provider")

	// ErrTemplateNotFound is returned when a requested template cannot be found
	ErrTemplateNotFound = errors.New("template not found")

	// ErrTemplateRender is returned when a template cannot be rendered
	ErrTemplateRender = errors.New("failed to render template")

	// ErrAPIRequest is returned when an API request to the LLM provider fails
	ErrAPIRequest = errors.New("API request failed")

	// ErrInvalidModel is returned when an unsupported model is specified
	ErrInvalidModel = errors.New("invalid model specified")
)

// MissingAPIKeyError returns a formatted error for a specific provider requiring an API key
func MissingAPIKeyError(provider string) error {
	return fmt.Errorf("%w for %s provider", ErrMissingAPIKey, provider)
}

// InvalidProviderError returns a formatted error for an unsupported provider
func InvalidProviderError(provider string) error {
	return fmt.Errorf("%w: %s", ErrInvalidProvider, provider)
}

// TemplateNotFoundError returns a formatted error for a missing template
func TemplateNotFoundError(name string) error {
	return fmt.Errorf("%w: %s", ErrTemplateNotFound, name)
}

// TemplateRenderError returns a formatted error for template rendering failure
func TemplateRenderError(name string, err error) error {
	return fmt.Errorf("%w for template %s: %v", ErrTemplateRender, name, err)
}

// APIRequestError returns a formatted error for API request failures
func APIRequestError(provider string, err error) error {
	return fmt.Errorf("%w to %s: %v", ErrAPIRequest, provider, err)
}

// InvalidModelError returns a formatted error for an unsupported model
func InvalidModelError(provider, model string) error {
	return fmt.Errorf("%w: %s is not supported by %s", ErrInvalidModel, model, provider)
}

// IsMissingAPIKey checks if the error is a missing API key error
func IsMissingAPIKey(err error) bool {
	return errors.Is(err, ErrMissingAPIKey)
}

// IsInvalidProvider checks if the error is an invalid provider error
func IsInvalidProvider(err error) bool {
	return errors.Is(err, ErrInvalidProvider)
}

// IsTemplateNotFound checks if the error is a template not found error
func IsTemplateNotFound(err error) bool {
	return errors.Is(err, ErrTemplateNotFound)
}

// IsTemplateRenderError checks if the error is a template rendering error
func IsTemplateRenderError(err error) bool {
	return errors.Is(err, ErrTemplateRender)
}

// IsAPIRequestError checks if the error is an API request error
func IsAPIRequestError(err error) bool {
	return errors.Is(err, ErrAPIRequest)
}

// IsInvalidModel checks if the error is an invalid model error
func IsInvalidModel(err error) bool {
	return errors.Is(err, ErrInvalidModel)
}
