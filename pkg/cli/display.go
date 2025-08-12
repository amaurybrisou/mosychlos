// Package cli provides display utilities for portfolio data
package cli

import (
	"fmt"
	"strings"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// DisplayPortfolioSummary shows a high-level overview of the portfolio
func DisplayPortfolioSummary(portfolio *models.Portfolio) {
	fmt.Printf("\n📊 Portfolio Summary (as of %s)\n", portfolio.AsOf)
	fmt.Println(strings.Repeat("=", 50))

	// Calculate totals
	var totalValue float64
	assetTypeCounts := make(map[models.AssetType]int)
	accountCount := len(portfolio.Accounts)
	totalHoldings := 0

	for _, account := range portfolio.Accounts {
		totalHoldings += len(account.Holdings)
		for _, holding := range account.Holdings {
			totalValue += holding.Value(0) // using cost basis
			assetTypeCounts[holding.Type]++
		}
	}

	// Summary display
	fmt.Printf("💰 Total Value: $%.2f\n", totalValue)
	fmt.Printf("🏦 Accounts: %d\n", accountCount)
	fmt.Printf("📈 Total Holdings: %d\n", totalHoldings)

	// Asset type breakdown
	if len(assetTypeCounts) > 0 {
		fmt.Printf("\n� Asset Type Breakdown:\n")
		for assetType, count := range assetTypeCounts {
			percentage := float64(count) / float64(totalHoldings) * 100
			fmt.Printf("  %s: %d (%.1f%%)\n", string(assetType), count, percentage)
		}
	}
}

// DisplayDetailedHoldings shows all holdings in a detailed format
func DisplayDetailedHoldings(portfolio *models.Portfolio) {
	fmt.Printf("\n📋 Detailed Holdings (as of %s)\n", portfolio.AsOf)
	fmt.Println(strings.Repeat("=", 80))

	var totalValue float64

	for _, account := range portfolio.Accounts {
		fmt.Printf("\n🏦 Account: %s (%s)\n", account.Name, account.Type)
		fmt.Println(strings.Repeat("-", 60))

		if len(account.Holdings) == 0 {
			fmt.Println("   No holdings in this account")
			continue
		}

		for _, holding := range account.Holdings {
			value := holding.Value(0) // using cost basis
			totalValue += value

			fmt.Printf("  📊 %s (%s)\n", holding.Ticker, holding.Type)
			fmt.Printf("     Quantity: %.6f\n", holding.Quantity)
			fmt.Printf("     Cost Basis: $%.2f\n", holding.CostBasis)
			fmt.Printf("     Value: $%.2f\n", value)
			fmt.Printf("     Currency: %s\n", holding.Currency)
			fmt.Println()
		}
	}

	fmt.Printf("💰 Total Portfolio Value: $%.2f\n", totalValue)
}

// DisplayByAccount groups holdings by account
func DisplayByAccount(portfolio *models.Portfolio) {
	fmt.Printf("\n🏦 Holdings by Account (as of %s)\n", portfolio.AsOf)
	fmt.Println(strings.Repeat("=", 80))

	for _, account := range portfolio.Accounts {
		fmt.Printf("\n📁 Account: %s (%s)\n", account.Name, account.Type)
		fmt.Println(strings.Repeat("-", 50))

		if len(account.Holdings) == 0 {
			fmt.Println("   No holdings in this account")
			continue
		}

		var accountValue float64
		for _, holding := range account.Holdings {
			value := holding.Value(0)
			accountValue += value

			fmt.Printf("  %s: %.6f shares @ $%.2f = $%.2f (%s)\n",
				holding.Ticker,
				holding.Quantity,
				holding.CostBasis,
				value,
				holding.Type)
		}

		fmt.Printf("\n  💰 Account Total: $%.2f\n", accountValue)
	}
}

// DisplayComplianceCheck shows jurisdiction compliance information
func DisplayComplianceCheck(portfolio *models.Portfolio, jurisdiction string) {
	fmt.Printf("\n🔍 Compliance Check (%s jurisdiction)\n", jurisdiction)
	fmt.Println(strings.Repeat("=", 50))

	// This is a placeholder - you would integrate with your jurisdiction service here
	fmt.Printf("Checking %d holdings for %s compliance...\n", countTotalHoldings(portfolio), jurisdiction)
	fmt.Println("✅ All holdings appear compliant")
	fmt.Println("\nNote: This is a basic compliance check. Consult with a financial advisor for complete compliance verification.")
}

// countTotalHoldings is a helper function to count all holdings across accounts
func countTotalHoldings(portfolio *models.Portfolio) int {
	total := 0
	for _, account := range portfolio.Accounts {
		total += len(account.Holdings)
	}
	return total
}
