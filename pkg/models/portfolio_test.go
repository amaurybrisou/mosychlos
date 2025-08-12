package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPortfolio_UserID(t *testing.T) {
	t.Parallel()

	t.Run("generates consistent user ID", func(t *testing.T) {
		portfolio := Portfolio{
			AsOf:         "2024-01-01",
			BaseCurrency: "USD",
			Accounts: []Account{
				{
					Name: "Investment",
					Type: "investment",
					Holdings: []Holding{
						{Ticker: "AAPL", Quantity: 100, Type: Stock},
						{Ticker: "MSFT", Quantity: 50, Type: Stock},
					},
				},
			},
		}

		userID1 := portfolio.UserID()
		userID2 := portfolio.UserID()

		// Should be consistent
		assert.Equal(t, userID1, userID2)

		// Should start with portfolio_
		assert.True(t, strings.HasPrefix(userID1, "portfolio_"))

		// Should be reasonable length (24 characters)
		assert.Equal(t, 24, len(userID1))
	})

	t.Run("different portfolios generate different IDs", func(t *testing.T) {
		portfolio1 := Portfolio{
			AsOf:         "2024-01-01",
			BaseCurrency: "USD",
			Accounts: []Account{
				{
					Name: "Investment",
					Type: "investment",
					Holdings: []Holding{
						{Ticker: "AAPL", Quantity: 100, Type: Stock},
					},
				},
			},
		}

		portfolio2 := Portfolio{
			AsOf:         "2024-01-01",
			BaseCurrency: "USD",
			Accounts: []Account{
				{
					Name: "Investment",
					Type: "investment",
					Holdings: []Holding{
						{Ticker: "MSFT", Quantity: 100, Type: Stock}, // Different ticker
					},
				},
			},
		}

		userID1 := portfolio1.UserID()
		userID2 := portfolio2.UserID()

		// Should be different
		assert.NotEqual(t, userID1, userID2)
	})

	t.Run("same structure generates same ID", func(t *testing.T) {
		createPortfolio := func() Portfolio {
			return Portfolio{
				AsOf:         "2024-01-01",
				BaseCurrency: "USD",
				Accounts: []Account{
					{
						Name: "Investment",
						Type: "investment",
						Holdings: []Holding{
							{Ticker: "AAPL", Quantity: 100, Type: Stock},
							{Ticker: "MSFT", Quantity: 50, Type: Stock},
						},
					},
				},
			}
		}

		portfolio1 := createPortfolio()
		portfolio2 := createPortfolio()

		userID1 := portfolio1.UserID()
		userID2 := portfolio2.UserID()

		// Same structure should generate same ID
		assert.Equal(t, userID1, userID2)
	})

	t.Run("different currency generates different ID", func(t *testing.T) {
		portfolio1 := Portfolio{
			AsOf:         "2024-01-01",
			BaseCurrency: "USD",
			Accounts: []Account{
				{Name: "Investment", Type: "investment", Holdings: []Holding{}},
			},
		}

		portfolio2 := Portfolio{
			AsOf:         "2024-01-01",
			BaseCurrency: "EUR", // Different currency
			Accounts: []Account{
				{Name: "Investment", Type: "investment", Holdings: []Holding{}},
			},
		}

		userID1 := portfolio1.UserID()
		userID2 := portfolio2.UserID()

		assert.NotEqual(t, userID1, userID2)
	})

	t.Run("different date generates different ID", func(t *testing.T) {
		portfolio1 := Portfolio{
			AsOf:         "2024-01-01",
			BaseCurrency: "USD",
			Accounts:     []Account{},
		}

		portfolio2 := Portfolio{
			AsOf:         "2024-01-02", // Different date
			BaseCurrency: "USD",
			Accounts:     []Account{},
		}

		userID1 := portfolio1.UserID()
		userID2 := portfolio2.UserID()

		assert.NotEqual(t, userID1, userID2)
	})

	t.Run("empty portfolio generates valid ID", func(t *testing.T) {
		portfolio := Portfolio{}

		userID := portfolio.UserID()

		assert.True(t, strings.HasPrefix(userID, "portfolio_"))
		assert.Equal(t, 24, len(userID))
	})
}

func TestPortfolio_Tickers(t *testing.T) {
	t.Parallel()

	portfolio := Portfolio{
		Accounts: []Account{
			{
				Holdings: []Holding{
					{Ticker: "AAPL", Type: Stock},
					{Ticker: "MSFT", Type: Stock},
					{Ticker: "", Type: Cash}, // Empty ticker should be ignored
				},
			},
			{
				Holdings: []Holding{
					{Ticker: "AAPL", Type: Stock}, // Duplicate should be deduplicated
					{Ticker: "GOOGL", Type: Stock},
				},
			},
		},
	}

	tickers := portfolio.Tickers()

	// Should contain unique tickers
	require.Len(t, tickers, 3)
	assert.Contains(t, tickers, "AAPL")
	assert.Contains(t, tickers, "MSFT")
	assert.Contains(t, tickers, "GOOGL")
}
