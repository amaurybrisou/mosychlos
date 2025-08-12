package portfolio

import (
	"context"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// Fetcher defines the interface for fetching portfolios from external sources
type Fetcher interface {
	// Fetch retrieves portfolio data from an external source
	Fetch(ctx context.Context) (*models.Portfolio, error)
}

// Service defines the main interface for portfolio operations
type Service interface {
	// GetPortfolio gets the current portfolio, fetching if needed based on configuration
	GetPortfolio(ctx context.Context, fetcher Fetcher) (*models.Portfolio, error)
}
