package errors

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidInputs indicates provided inputs are invalid or inconsistent
	ErrInvalidInputs = errors.New("invalid rebalance inputs")
	// ErrToleranceOutOfRange indicates an invalid tolerance value
	ErrToleranceOutOfRange = errors.New("tolerance out of range")
	// ErrNegativeValue indicates a negative value where not allowed
	ErrNegativeValue = errors.New("negative value not allowed")
	// ErrEmptyUniverse indicates no tickers to rebalance
	ErrEmptyUniverse = errors.New("empty universe")
)

// InvalidInputsError returns a formatted input error
func InvalidInputsError(msg string) error { return fmt.Errorf("%w: %s", ErrInvalidInputs, msg) }

// ToleranceOutOfRangeError returns a formatted tolerance error
func ToleranceOutOfRangeError(v float64) error {
	return fmt.Errorf("%w: %.4f", ErrToleranceOutOfRange, v)
}

// NegativeValueError returns a formatted negative value error
func NegativeValueError(name string, v float64) error {
	return fmt.Errorf("%w: %s=%.4f", ErrNegativeValue, name, v)
}

// EmptyUniverseError returns a formatted empty universe error
func EmptyUniverseError() error { return ErrEmptyUniverse }
