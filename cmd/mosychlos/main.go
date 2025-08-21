package mosychlos

import (
	"log"

	pkgconfig "github.com/amaurybrisou/mosychlos/pkg/config"
	"github.com/spf13/cobra"
)

func RootCmd() {
	var rootCmd = &cobra.Command{
		Use:   "mosychlos",
		Short: "Interactive portfolio management CLI",
		Long:  `An interactive command-line interface for managing and analyzing your portfolio.`,
	}

	cfg := pkgconfig.MustLoadConfig()
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
