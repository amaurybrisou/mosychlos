package mosychlos

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/amaurybrisou/mosychlos/internal/engine"
	"github.com/amaurybrisou/mosychlos/pkg/cli"
	pkgconfig "github.com/amaurybrisou/mosychlos/pkg/config"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/spf13/cobra"
)

func analyzeCommand(cmd *cobra.Command, args []string) {
	cfg := pkgconfig.MustLoadConfig()

	// Get analysis type
	analysisType, err := getAnalysisType(args)
	if err != nil {
		slog.Error("Failed to determine analysis type", "error", err)
		os.Exit(1)
	}

	o := engine.New(cfg, nil)
	builder := engine.DefaultRegistryWithOrder() // or build your own registry programmatically
	o.UseBuilder(builder)

	ctx := context.Background()
	err = o.Init(ctx)
	if err != nil {
		slog.Error("Failed to initialize engine orchestrator", "error", err)
		os.Exit(1)
	}

	err = o.ExecutePipeline(ctx)
	if err != nil {
		slog.Error("Failed to execute engine pipeline", "error", err)
		os.Exit(1)
	}

	// Run AI analysis
	aiResponse, err := o.GetResults(analysisType)
	if err != nil {
		slog.Error("Failed to get analysis results", "error", err)
		os.Exit(1)
	}

	// Display results
	displayAnalysisResults(*aiResponse)

	// Handle report generation
	if shouldGenerateReports(cmd) {
		// Determine report type and formats from flags
		reportType, formats := getReportParamsFromFlags(cmd)

		o.GenerateReports(ctx, reportType, formats)
	}

	// Only ask for continuation in interactive mode (when no analysis type was specified)
	if len(args) == 0 {
		// Ask if user wants to run another analysis
		runAnother, err := cli.ConfirmAction("Would you like to run another analysis?")
		if err != nil {
			slog.Error("Error in confirmation", "error", err)
			return
		}

		if runAnother {
			analyzeCommand(cmd, []string{}) // Recursive call with no args for interactive mode
		}
	}
}

// getAnalysisType determines analysis type from args or user input
func getAnalysisType(args []string) (models.AnalysisType, error) {
	var analysisType models.AnalysisType
	var err error

	if len(args) == 0 {
		// Interactive mode
		slog.Info("Portfolio Analysis")
		fmt.Println()
		analysisType, err = cli.SelectAnalysisType()
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
func shouldGenerateReports(cmd *cobra.Command) bool {
	reports, _ := cmd.Flags().GetBool("reports")
	allFormats, _ := cmd.Flags().GetBool("all-formats")
	markdown, _ := cmd.Flags().GetBool("markdown")
	pdf, _ := cmd.Flags().GetBool("pdf")
	jsonFlag, _ := cmd.Flags().GetBool("json")

	// If any report flag is set, generate reports
	if reports || allFormats || markdown || pdf || jsonFlag {
		return true
	}

	// Ask user interactively
	generateReports, err := cli.ConfirmAction("Generate analysis reports?")
	if err != nil {
		slog.Warn("Error in report confirmation", "error", err)
		return true // Default to true when confirmation fails
	}

	return generateReports
}

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
