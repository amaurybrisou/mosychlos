// Package mosychlos
package mosychlos

import (
	"log"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/spf13/cobra"
)

func RootCmd() {
	var rootCmd = &cobra.Command{
		Use:   "mosychlos",
		Short: "Interactive portfolio management CLI",
		Long:  `An interactive command-line interface for managing and analyzing your portfolio.`,
	}

	cfg := config.MustLoadConfig()
	rootCmd.AddCommand(NewPortfolioCommand(cfg))
	rootCmd.AddCommand(NewAnalyzeCommand(cfg))
	rootCmd.AddCommand(CreateToolsCommand(cfg))
	// Add batch processing command
	rootCmd.AddCommand(CreateBatchCommand(cfg))
	rootCmd.AddCommand(NewReportCommand(cfg))
	rootCmd.AddCommand(NewMarkdownToPDFCommand(cfg))

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
