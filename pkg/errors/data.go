package errors

import (
	"errors"
	"fmt"
)

var (
	// Data/portfolio related errors
	ErrNoAccounts           = errors.New("portfolio has no accounts")
	ErrInvalidAccount       = errors.New("invalid account")
	ErrInvalidHolding       = errors.New("invalid holding")
	ErrUnsupportedAssetType = errors.New("unsupported asset type")
	ErrEmptyTicker          = errors.New("empty ticker")
	ErrNegativeQuantity     = errors.New("negative quantity")
	ErrNegativeCostBasis    = errors.New("negative cost basis")
	ErrEmptyCurrency        = errors.New("empty currency")
)

func InvalidAccountError(name, reason string) error {
	if name == "" {
		return fmt.Errorf("%w: %s", ErrInvalidAccount, reason)
	}
	return fmt.Errorf("%w %s: %s", ErrInvalidAccount, name, reason)
}

func InvalidHoldingError(ticker, reason string) error {
	if ticker == "" {
		return fmt.Errorf("%w: %s", ErrInvalidHolding, reason)
	}
	return fmt.Errorf("%w %s: %s", ErrInvalidHolding, ticker, reason)
}

func UnsupportedAssetTypeError(t string) error {
	return fmt.Errorf("%w: %s", ErrUnsupportedAssetType, t)
}
