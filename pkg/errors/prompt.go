package errors

import (
	"errors"
	"fmt"
)

var (
	// ErrPromptCanceled indicates that a prompt was canceled by the user
	ErrPromptCanceled = errors.New("prompt canceled by user")

	// ErrPromptInterrupted indicates that a prompt was interrupted (e.g., Ctrl+C)
	ErrPromptInterrupted = errors.New("prompt interrupted")

	// ErrPromptInvalidInput indicates that invalid input was provided to a prompt
	ErrPromptInvalidInput = errors.New("invalid prompt input")

	// ErrPromptEmpty indicates that empty input was provided to a prompt that requires non-empty input
	ErrPromptEmpty = errors.New("prompt input cannot be empty")

	// ErrPromptValidation indicates that input validation failed for a prompt
	ErrPromptValidation = errors.New("prompt validation failed")

	// ErrPromptTimeout indicates that a prompt timed out waiting for input
	ErrPromptTimeout = errors.New("prompt timed out")

	// ErrNoItemsAvailable indicates that a select prompt has no items to select from
	ErrNoItemsAvailable = errors.New("no items available for selection")
)

// PromptCanceledError returns a formatted error with optional context about where the prompt was canceled
func PromptCanceledError(context string) error {
	if context == "" {
		return ErrPromptCanceled
	}
	return fmt.Errorf("%w: %s", ErrPromptCanceled, context)
}

// PromptInterruptedError returns a formatted error with optional context about where the prompt was interrupted
func PromptInterruptedError(context string) error {
	if context == "" {
		return ErrPromptInterrupted
	}
	return fmt.Errorf("%w: %s", ErrPromptInterrupted, context)
}

// PromptValidationError returns a formatted error with details about the validation failure
func PromptValidationError(validationMsg string) error {
	return fmt.Errorf("%w: %s", ErrPromptValidation, validationMsg)
}

// PromptInvalidInputError returns a formatted error about invalid input with optional context
func PromptInvalidInputError(input string, expected string) error {
	return fmt.Errorf("%w: got '%s', expected %s", ErrPromptInvalidInput, input, expected)
}

// IsPromptCanceled checks if the error indicates a canceled prompt
func IsPromptCanceled(err error) bool {
	return errors.Is(err, ErrPromptCanceled)
}

// IsPromptInterrupted checks if the error indicates an interrupted prompt
func IsPromptInterrupted(err error) bool {
	return errors.Is(err, ErrPromptInterrupted)
}

// IsPromptValidationError checks if the error indicates a prompt validation failure
func IsPromptValidationError(err error) bool {
	return errors.Is(err, ErrPromptValidation)
}

// IsPromptEmpty checks if the error indicates an empty prompt input
func IsPromptEmpty(err error) bool {
	return errors.Is(err, ErrPromptEmpty)
}

// IsNoItemsAvailable checks if the error indicates no items are available for selection
func IsNoItemsAvailable(err error) bool {
	return errors.Is(err, ErrNoItemsAvailable)
}
