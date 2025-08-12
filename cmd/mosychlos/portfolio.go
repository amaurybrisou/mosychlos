package mosychlos

import (
	"log"
	"log/slog"
	"os"

	"github.com/amaurybrisou/mosychlos/pkg/cli"
	"github.com/spf13/cobra"
)

func portfolioCommand(cmd *cobra.Command, args []string) {
	// // Load configuration
	// cfg := config.MustLoadConfig()

	// // Create filesystem and shared bag (don't convert to snapshot yet)
	// filesystem := fs.OS{}
	// sharedBag := bag.NewSharedBag()

	// // Set up tools with shared bag for metrics tracking
	// tools.SetSharedBag(sharedBag)
	// if err := tools.NewTools(cfg); err != nil {
	// 	slog.Error("Failed to initialize tools", "error", err)
	// 	os.Exit(1)
	// }

	// // Initialize application health monitoring
	// healthMonitor := health.NewApplicationMonitor(sharedBag)
	// healthMonitor.StartPeriodicHealthCheck(15 * time.Second)

	// // Create portfolio service with shared bag for normalization
	// portfolioService := portfolio.NewService(cfg, filesystem, sharedBag)

	// // Create Binance fetcher adapter
	// binanceProvider := binance.NewPortfolioProvider(&cfg.Binance)

	// // Load portfolio
	// ctx := context.Background()
	// portfolioData, err := portfolioService.GetPortfolio(
	// 	ctx, adapters.NewBinanceFetcher(binanceProvider))
	// if err != nil {
	// 	slog.Error("Failed to load portfolio", "error", err)
	// 	os.Exit(1)
	// }

	// Interactive display mode selection
	displayMode, err := cli.SelectDisplayMode()
	if err != nil {
		slog.Error("Failed to select display mode", "error", err)
		os.Exit(1)
	}

	// Display based on selection
	switch displayMode {
	// case "summary":
	// 	cli.DisplayPortfolioSummary(portfolioData)
	// case "detailed":
	// 	cli.DisplayDetailedHoldings(portfolioData)
	// case "by-account":
	// 	cli.DisplayByAccount(portfolioData)
	// case "compliance":
	// 	jurisdiction := cfg.Localization.Country
	// 	cli.DisplayComplianceCheck(portfolioData, jurisdiction)
	case "ai-analysis":
		// Launch AI analysis - delegate to analyze command
		analyzeCommand(cmd, []string{}) // Call with empty args for interactive mode
		return                          // Don't ask for another view after analysis (analyze has its own loop)
	}

	// Ask if user wants to see another view
	showAnother, err := cli.ConfirmAction("Would you like to see another view?")
	if err != nil {
		log.Printf("Error in confirmation: %v", err)
		return
	}

	if showAnother {
		portfolioCommand(cmd, args) // recursive call for multiple views
	}
}
