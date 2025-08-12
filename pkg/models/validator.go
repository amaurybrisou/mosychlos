package models

import (
	"context"
	"time"
)

//go:generate mockgen -source=validator.go -destination=mocks/validator_mock.go -package=mocks

// PortfolioValidator defines a common interface for validating portfolios
type PortfolioValidator interface {
	// Validate validates a portfolio and returns an error if invalid
	Validate(ctx context.Context, portfolio *Portfolio) error
}

// ValidationRecord tracks validation history for a portfolio
type ValidationRecord struct {
	ValidatedAt   time.Time `yaml:"validated_at"`
	ValidatorName string    `yaml:"validator_name"`
	Success       bool      `yaml:"success"`
	ErrorMessage  string    `yaml:"error_message,omitempty"`
}
