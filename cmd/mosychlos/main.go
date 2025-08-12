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

	var portfolioCmd = &cobra.Command{
		Use:   "portfolio",
		Short: "Display portfolio information",
		Long:  `Interactively display your portfolio with various view options.`,
		Run:   portfolioCommand,
	}

	var analyzeCmd = &cobra.Command{
		Use:   "analyze [analysis-type]",
		Short: "Analyze portfolio with AI insights",
		Long: `Generate AI-powered portfolio analysis. Run without arguments for interactive mode,
or specify analysis type directly: risk, allocation, performance, compliance, reallocation.

Examples:
  mosychlos portfolio analyze              # Interactive mode - select analysis type
  mosychlos portfolio analyze risk         # Direct risk analysis
  mosychlos portfolio analyze investment_research # In-depth analysis of investment opportunities`,
		Args: cobra.MaximumNArgs(1),
		Run:  analyzeCommand,
	}

	// Add verbose flag to analyze command
	analyzeCmd.Flags().BoolP("verbose", "v", false, "Show detailed analysis process including prompts and AI conversation")
	// Add batch flag to analyze command
	analyzeCmd.Flags().Bool("batch", false, "Use batch processing for analysis (50% cost savings, longer processing time)")
	// Add report generation flags
	analyzeCmd.Flags().Bool("reports", false, "Generate reports after analysis")
	analyzeCmd.Flags().Bool("all-formats", false, "Generate reports in all formats (markdown, PDF, JSON)")
	analyzeCmd.Flags().Bool("markdown", false, "Generate markdown reports")
	analyzeCmd.Flags().Bool("pdf", false, "Generate PDF reports")
	analyzeCmd.Flags().Bool("json", false, "Generate JSON reports")

	portfolioCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(portfolioCmd)
	rootCmd.AddCommand(CreateToolsCommand())

	// Add batch processing command
	cfg := pkgconfig.MustLoadConfig()
	rootCmd.AddCommand(CreateBatchCommand(cfg))

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
