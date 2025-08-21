package mosychlos

import (
	"fmt"
	"log"
	"log/slog"
	"strings"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/spf13/cobra"
)

func NewPortfolioCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "portfolio",
		Short: "Display portfolio information",
		Long:  `Interactively display your portfolio with various view options.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return portfolioCommand(cmd, args, cfg)
		},
	}
}

func portfolioCommand(cmd *cobra.Command, args []string, cfg *config.Config) error {
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
	displayMode, err := selectDisplayMode()
	if err != nil {
		slog.Error("Failed to select display mode", "error", err)
		return err
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

	}

	// Ask if user wants to see another view
	showAnother, err := confirmAction("Would you like to see another view?")
	if err != nil {
		log.Printf("Error in confirmation: %v", err)
		return err
	}

	if showAnother {
		return portfolioCommand(cmd, args, cfg) // recursive call for multiple views
	}

	return nil
}

// selectDisplayMode prompts the user to select a display mode
func selectDisplayMode() (string, error) {
	options := []string{
		"summary - Portfolio overview and summary statistics",
		"detailed - Detailed holdings breakdown",
		"accounts - View by account structure",
		"compliance - Compliance and regulatory check",
	}

	fmt.Println("\n📊 Select Display Mode:")
	for i, option := range options {
		fmt.Printf("  %d. %s\n", i+1, option)
	}

	var choice int
	fmt.Print("\nEnter your choice (1-4): ")
	if _, err := fmt.Scanf("%d", &choice); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	if choice < 1 || choice > len(options) {
		return "", fmt.Errorf("invalid choice: must be between 1 and %d", len(options))
	}

	// Extract the display mode from the option string
	option := options[choice-1]
	displayMode := strings.Split(option, " - ")[0]
	return displayMode, nil
}
