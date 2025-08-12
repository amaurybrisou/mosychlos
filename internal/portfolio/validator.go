package portfolio

import (
	"context"
	"fmt"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// basicValidator implements basic portfolio validation
type basicValidator struct{}

// NewBasicValidator creates a new basic portfolio validator
func NewBasicValidator() models.PortfolioValidator {
	return &basicValidator{}
}

// Validate performs basic validation on a portfolio
func (v *basicValidator) Validate(ctx context.Context, portfolio *models.Portfolio) error {
	if portfolio == nil {
		return fmt.Errorf("portfolio cannot be nil")
	}

	// validate AsOf date
	if portfolio.AsOf == "" {
		return fmt.Errorf("portfolio must have an AsOf date")
	}

	// validate that AsOf date can be parsed
	if _, err := portfolio.AsOfTime(); err != nil {
		return fmt.Errorf("invalid AsOf date format: %w", err)
	}

	// validate accounts
	if len(portfolio.Accounts) == 0 {
		return fmt.Errorf("portfolio must have at least one account")
	}

	for i, account := range portfolio.Accounts {
		if err := v.validateAccount(account); err != nil {
			return fmt.Errorf("account %d (%s) validation failed: %w", i, account.Name, err)
		}
	}

	// if validation passes, mark as validated
	portfolio.Validated = true

	return nil
}

// validateAccount performs validation on an individual account
func (v *basicValidator) validateAccount(account models.Account) error {
	if account.Name == "" {
		return fmt.Errorf("account name cannot be empty")
	}

	if account.Type == "" {
		return fmt.Errorf("account type cannot be empty")
	}

	if account.Currency == "" {
		return fmt.Errorf("account currency cannot be empty")
	}

	// validate holdings
	for i, holding := range account.Holdings {
		if err := v.validateHolding(holding); err != nil {
			return fmt.Errorf("holding %d (%s) validation failed: %w", i, holding.Ticker, err)
		}
	}

	return nil
}

// validateHolding performs validation on an individual holding
func (v *basicValidator) validateHolding(holding models.Holding) error {
	if holding.Ticker == "" {
		return fmt.Errorf("holding ticker cannot be empty")
	}

	if holding.Quantity < 0 {
		return fmt.Errorf("holding quantity cannot be negative")
	}

	if holding.CostBasis < 0 {
		return fmt.Errorf("holding cost basis cannot be negative")
	}

	if holding.Currency == "" {
		return fmt.Errorf("holding currency cannot be empty")
	}

	if holding.Type == "" {
		return fmt.Errorf("holding type cannot be empty")
	}

	return nil
}
