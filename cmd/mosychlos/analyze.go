package mosychlos

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/internal/engine"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/spf13/cobra"
)

func NewAnalyzeCommand(cfg *config.Config) *cobra.Command {

	var analyzeCmd = &cobra.Command{
		Use:   "analyze [analysis-type]",
		Short: "Analyze portfolio with AI insights",
		Long: `Generate AI-powered portfolio analysis. Run without arguments for interactive mode,
		or specify analysis type directly: risk, allocation, performance, compliance, reallocation.

		Examples:
		mosychlos analyze              # Interactive mode - select analysis type
		mosychlos analyze risk         # Direct risk analysis
		mosychlos analyze investment_research # In-depth analysis of investment opportunities`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAnalyzeCommand(cmd, args, cfg)
		},
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

	return analyzeCmd
}

func runAnalyzeCommand(cmd *cobra.Command, args []string, cfg *config.Config) error {
	// Get analysis type
	// analysisType, err := getAnalysisType(args)
	// if err != nil {
	// 	slog.Error("Failed to determine analysis type", "error", err)
	// 	os.Exit(1)
	// }

	o := engine.New(cfg, nil)

	batch, _ := cmd.Flags().GetBool("batch")

	builder := engine.DefaultRegistry() // or build your own registry programmatically
	if batch {
		slog.Debug("Using batch engine builder", "builder", builder)
		builder = engine.DefaultBatchRegistry() // or build your own registry programmatically
	} else {
		slog.Debug("Using default engine builder", "builder", builder)
	}

	o.UseBuilder(builder)

	ctx := context.Background()
	err := o.Init(ctx)
	if err != nil {
		slog.Error("Failed to initialize engine orchestrator", "error", err)
		return err
	}

	err = o.ExecutePipeline(ctx)
	if err != nil {
		slog.Error("Failed to execute engine pipeline", "error", err)
		return err
	}

	return nil
	// Run AI analysis
	// aiResponse, err := o.GetResults(analysisType)
	// if err != nil {
	// 	slog.Error("Failed to get analysis results", "error", err)
	// 	os.Exit(1)
	// }

	// // Display results
	// displayAnalysisResults(*aiResponse)

	// Handle report generation
	// if shouldGenerateReports(cmd) {
	// 	// Determine report type and formats from flags
	// 	reportType, formats := getReportParamsFromFlags(cmd)

	// 	o.GenerateReports(ctx, reportType, formats)
	// }

	// // Only ask for continuation in interactive mode (when no analysis type was specified)
	// if len(args) == 0 {
	// 	// Ask if user wants to run another analysis
	// 	runAnother, err := confirmAction("Would you like to run another analysis?")
	// 	if err != nil {
	// 		slog.Error("Error in confirmation", "error", err)
	// 		return
	// 	}

	// 	if runAnother {
	// 		analyzeCommand(cmd, []string{}) // Recursive call with no args for interactive mode
	// 	}
	// }
}

// getAnalysisType determines analysis type from args or user input
func getAnalysisType(args []string) (models.AnalysisType, error) {
	var analysisType models.AnalysisType
	var err error

	if len(args) == 0 {
		// Interactive mode
		slog.Info("Portfolio Analysis")
		fmt.Println()
		analysisType, err = selectAnalysisType()
		if err != nil {
			return models.AnalysisType(""), fmt.Errorf("failed to select analysis type: %w", err)
		}
	} else {
		// Non-interactive mode
		analysisType = models.AnalysisType(args[0])
	}

	return analysisType, nil
}

// displayAnalysisResults shows the analysis results to the user
func displayAnalysisResults(aiResponse string) {
	slog.Info("Portfolio Analysis Results")
	fmt.Println("═══════════════════════════════")
	fmt.Println(aiResponse)
	fmt.Println("═══════════════════════════════")
	fmt.Println()
}

// shouldGenerateReports checks if reports should be generated based on flags or user input
// func shouldGenerateReports(cmd *cobra.Command) bool {
// 	reports, _ := cmd.Flags().GetBool("reports")
// 	allFormats, _ := cmd.Flags().GetBool("all-formats")
// 	markdown, _ := cmd.Flags().GetBool("markdown")
// 	pdf, _ := cmd.Flags().GetBool("pdf")
// 	jsonFlag, _ := cmd.Flags().GetBool("json")

// 	// If any report flag is set, generate reports
// 	if reports || allFormats || markdown || pdf || jsonFlag {
// 		return true
// 	}

// 	// Ask user interactively
// 	generateReports, err := confirmAction("Generate analysis reports?")
// 	if err != nil {
// 		slog.Warn("Error in report confirmation", "error", err)
// 		return true // Default to true when confirmation fails
// 	}

// 	return generateReports
// }

// getReportParamsFromFlags determines report type and formats from command flags
func getReportParamsFromFlags(cmd *cobra.Command) (models.ReportType, []models.ReportFormat) {
	reports, _ := cmd.Flags().GetBool("reports")
	allFormats, _ := cmd.Flags().GetBool("all-formats")
	markdown, _ := cmd.Flags().GetBool("markdown")
	pdf, _ := cmd.Flags().GetBool("pdf")
	jsonFlag, _ := cmd.Flags().GetBool("json")

	// Default report type (could be configurable in the future)
	reportType := models.TypeFull

	var formats []models.ReportFormat

	// If --reports flag is used, generate all formats
	if reports {
		formats = []models.ReportFormat{models.FormatMarkdown, models.FormatPDF, models.FormatJSON}
		return reportType, formats
	}

	// If --all-formats flag is used, generate all formats
	if allFormats {
		formats = []models.ReportFormat{models.FormatMarkdown, models.FormatPDF, models.FormatJSON}
		return reportType, formats
	}

	// Add specific formats based on individual flags
	if markdown {
		formats = append(formats, models.FormatMarkdown)
	}
	if pdf {
		formats = append(formats, models.FormatPDF)
	}
	if jsonFlag {
		formats = append(formats, models.FormatJSON)
	}

	return reportType, formats
}

// selectAnalysisType prompts the user to select an analysis type
func selectAnalysisType() (models.AnalysisType, error) {
	options := []string{
		"risk - Portfolio risk assessment and concentration analysis",
		"investment_research - In-depth analysis of investment opportunities",
	}

	fmt.Println("\n📊 Select Analysis Type:")
	for i, option := range options {
		fmt.Printf("  %d. %s\n", i+1, option)
	}

	var choice int
	fmt.Print("\nEnter your choice (1-2): ")
	if _, err := fmt.Scanf("%d", &choice); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	if choice < 1 || choice > len(options) {
		return "", fmt.Errorf("invalid choice: must be between 1 and %d", len(options))
	}

	// Extract the analysis type from the option string
	option := options[choice-1]
	analysisType := strings.Split(option, " - ")[0]
	return models.AnalysisType(analysisType), nil
}

// confirmAction prompts the user for a yes/no confirmation
func confirmAction(message string) (bool, error) {
	fmt.Printf("%s (y/N): ", message)
	var input string
	if _, err := fmt.Scanf("%s", &input); err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes", nil
}
