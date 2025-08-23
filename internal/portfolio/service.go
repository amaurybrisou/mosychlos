package portfolio

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/fs"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// service implements Service interface with fs and shared bag state management
type service struct {
	fs         fs.FS
	validators []models.PortfolioValidator
	bag        bag.SharedBag
	config     *config.Config
}

// NewService creates a new portfolio service with fs, shared bag and validators
func NewService(cfg *config.Config, filesystem fs.FS, sharedBag bag.SharedBag, validators ...models.PortfolioValidator) Service {
	return &service{
		fs:         filesystem,
		validators: validators,
		bag:        sharedBag,
		config:     cfg,
	}
}

// GetPortfolio gets the current portfolio, fetching if needed based on configuration
func (s *service) GetPortfolio(ctx context.Context, fetcher Fetcher) (*models.Portfolio, error) {
	now := time.Now()

	// check if we have a current portfolio in bag
	currentPortfolio := s.getCurrentPortfolio()

	// determine if we need to fetch new data
	needsFetch := s.shouldFetchPortfolio(currentPortfolio, now)

	var portfolio *models.Portfolio
	var err error

	if needsFetch {
		// fetch new portfolio data
		portfolio, err = fetcher.Fetch(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch portfolio: %w", err)
		}

		// update fetch metadata in bag
		s.bag.Set(bag.KPortfolioLastFetched, now)
		s.bag.Set(bag.KPortfolioFetchSource, fmt.Sprintf("%T", fetcher))

		// validate portfolio with all validators
		if err := s.validatePortfolio(ctx, portfolio); err != nil {
			return nil, fmt.Errorf("portfolio validation failed: %w", err)
		}

		// mark as validated and update bag
		portfolio.Validated = true
		s.bag.Set(bag.KPortfolio, portfolio)

		// store normalized portfolio for AI analysis
		if normalizedPortfolio, err := portfolio.Normalize(); err != nil {
			fmt.Printf("Warning: failed to normalize portfolio: %v\n", err)
		} else {
			s.bag.Set(bag.KPortfolioNormalizedForAI, normalizedPortfolio)
		}

		// persist to storage for caching
		if err := s.savePortfolio(portfolio, "current"); err != nil {
			// log warning but don't fail - we have the portfolio in memory
			fmt.Printf("Warning: failed to persist portfolio: %v\n", err)
		}
	} else {
		portfolio = currentPortfolio
	}

	return portfolio, nil
}

// validatePortfolio runs all validators and records results in bag
func (s *service) validatePortfolio(ctx context.Context, portfolio *models.Portfolio) error {
	var validationRecords []models.ValidationRecord

	for _, validator := range s.validators {
		record := models.ValidationRecord{
			ValidatedAt:   time.Now(),
			ValidatorName: fmt.Sprintf("%T", validator),
			Success:       false,
		}

		if err := validator.Validate(ctx, portfolio); err != nil {
			record.ErrorMessage = err.Error()
			validationRecords = append(validationRecords, record)
			return fmt.Errorf("validation failed with %s: %w", record.ValidatorName, err)
		}

		record.Success = true
		validationRecords = append(validationRecords, record)
	}

	// update bag with validation records and time
	s.bag.Set(bag.KPortfolioValidationRecord, validationRecords)
	s.bag.Set(bag.KPortfolioValidationTime, time.Now())

	return nil
}

// shouldFetchPortfolio determines if we need to fetch new portfolio data
func (s *service) shouldFetchPortfolio(current *models.Portfolio, now time.Time) bool {
	if current == nil {
		return true // no current portfolio
	}

	// check if portfolio is from today
	if current.AsOf != "" {
		if asOfTime, err := current.AsOfTime(); err == nil {
			if asOfTime.Truncate(24 * time.Hour).Equal(now.Truncate(24 * time.Hour)) {
				return false // same day, no need to fetch
			}
		}
	}

	// check last fetch time from bag
	if lastFetchValue, ok := s.bag.Get(bag.KPortfolioLastFetched); ok {
		if lastFetch, ok := lastFetchValue.(time.Time); ok {
			if lastFetch.Truncate(24 * time.Hour).Equal(now.Truncate(24 * time.Hour)) {
				return false // already fetched today
			}
		}
	}

	return true // need to fetch
}

// getCurrentPortfolio retrieves current portfolio from bag or loads from disk
func (s *service) getCurrentPortfolio() *models.Portfolio {
	// First check if we have it in memory (bag)
	if portfolioValue, ok := s.bag.Get(bag.KPortfolio); ok {
		if portfolio, ok := portfolioValue.(*models.Portfolio); ok {
			return portfolio
		}
	}

	// If not in memory, try to load from disk
	portfolio, err := s.loadPortfolioFromDisk("current")
	if err != nil {
		// no saved portfolio or error loading - that's fine
		return nil
	}

	// Store in bag for future access
	s.bag.Set(bag.KPortfolio, portfolio)

	// also store normalized version for AI analysis
	if normalizedPortfolio, err := portfolio.Normalize(); err != nil {
		fmt.Printf("Warning: failed to normalize cached portfolio: %v\n", err)
	} else {
		s.bag.Set(bag.KPortfolioNormalizedForAI, normalizedPortfolio)
	}

	return portfolio
}

// loadPortfolioFromDisk loads a portfolio from disk
func (s *service) loadPortfolioFromDisk(name string) (*models.Portfolio, error) {
	filename := s.portfolioPath(name)

	data, err := s.fs.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read portfolio file %s: %w", filename, err)
	}

	var portfolio models.Portfolio
	if err := yaml.Unmarshal(data, &portfolio); err != nil {
		return nil, fmt.Errorf("failed to parse portfolio file %s: %w", filename, err)
	}

	return &portfolio, nil
}

// savePortfolio persists portfolio to storage
func (s *service) savePortfolio(portfolio *models.Portfolio, name string) error {
	if err := s.ensureDataDir(); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	filename := s.portfolioPath(name)

	data, err := yaml.Marshal(portfolio)
	if err != nil {
		return fmt.Errorf("failed to marshal portfolio: %w", err)
	}

	return s.fs.WriteFile(filename, data, 0644)
}

// Helper methods
func (s *service) ensureDataDir() error {
	// Create the base data directory
	if err := s.fs.MkdirAll(s.config.DataDir, 0755); err != nil {
		return err
	}

	// Create the portfolio subdirectory
	portfolioDir := filepath.Join(s.config.DataDir, "portfolio")
	return s.fs.MkdirAll(portfolioDir, 0755)
}

func (s *service) portfolioPath(name string) string {
	return filepath.Join(s.config.DataDir, "portfolio", name+".yaml")
}
