package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/internal/report"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/fs"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/manifoldco/promptui"
)

// GenerateReport generates a report based on user selection and saves it
func GenerateReport(ctx context.Context, cfg *config.Config, dataBag bag.Bag, filesystem fs.FS) error {
	// Select report type
	reportType, err := SelectReportType()
	if err != nil {
		return fmt.Errorf("failed to select report type: %w", err)
	}

	// Select output format (or use default from config)
	formats, err := SelectReportFormats(models.FormatMarkdown)
	if err != nil {
		return fmt.Errorf("failed to select report format: %w", err)
	}

	// Ask for customer name if generating customer report and not set in config
	customerName := cfg.Report.DefaultCustomerName
	if reportType == models.TypeCustomer && customerName == "" {
		customerName, err = PromptCustomerName()
		if err != nil {
			customerName = "" // use empty if prompt fails
		}
		// Temporarily override the config for this report
		if customerName != "" {
			cfg.Report.DefaultCustomerName = customerName
		}
	}

	// Create dependencies and generator
	deps := report.Dependencies{
		DataBag:    dataBag,
		Config:     cfg,
		FileSystem: filesystem,
	}
	generator := report.NewGenerator(deps)

	// Generate report
	fmt.Printf("\n🔄 Generating %s report in %v format...\n", reportType, formats)

	var output *models.ReportOutput
	for _, format := range formats {
		switch reportType {
		case models.TypeCustomer:
			output, err = generator.GenerateCustomerReport(ctx, format)
		case models.TypeSystem:
			output, err = generator.GenerateSystemReport(ctx, format)
		case models.TypeFull:
			output, err = generator.GenerateFullReport(ctx, format)
		default:
			return fmt.Errorf("unknown report type: %s", reportType)
		}

		if err != nil {
			return fmt.Errorf("failed to generate report: %w", err)
		}

		// Save report to file using config's output directory
		outputDir := cfg.Report.GetReportOutputDir(cfg.DataDir)
		filename := generateReportFilename(reportType, format)
		outputPath := filepath.Join(outputDir, filename)

		if err := saveReportToFile(filesystem, output, outputPath); err != nil {
			return fmt.Errorf("failed to save report: %w", err)
		}

		// Display success message
		fmt.Printf("\n✅ Report generated successfully!\n")
		fmt.Printf("📁 Saved to: %s\n", outputPath)
		fmt.Printf("⏱️  Generated in: %dms\n", output.Metadata.GenerationTimeMs)
		fmt.Printf("📊 Data sources used: %v\n", output.Metadata.DataSources)
	}

	return nil
}

// GenerateReportWithParams generates a report with specified type and formats (non-interactive)
func GenerateReportWithParams(ctx context.Context, cfg *config.Config, dataBag bag.Bag, filesystem fs.FS, reportType models.ReportType, formats []models.ReportFormat) error {
	// Ask for customer name if generating customer report and not set in config
	customerName := cfg.Report.DefaultCustomerName
	if reportType == models.TypeCustomer && customerName == "" {
		customerName, err := PromptCustomerName()
		if err != nil {
			customerName = "" // use empty if prompt fails
		}
		// Temporarily override the config for this report
		if customerName != "" {
			cfg.Report.DefaultCustomerName = customerName
		}
	}

	// Create dependencies and generator
	deps := report.Dependencies{
		DataBag:    dataBag,
		Config:     cfg,
		FileSystem: filesystem,
	}
	generator := report.NewGenerator(deps)

	// Generate report
	fmt.Printf("\n🔄 Generating %s report in %v format...\n", reportType, formats)

	for _, format := range formats {
		switch reportType {
		case models.TypeCustomer:
			output, err := generator.GenerateCustomerReport(ctx, format)
			if err != nil {
				return fmt.Errorf("failed to generate customer report: %w", err)
			}
			if err := saveReport(filesystem, cfg, output, reportType, format); err != nil {
				return err
			}
		case models.TypeSystem:
			output, err := generator.GenerateSystemReport(ctx, format)
			if err != nil {
				return fmt.Errorf("failed to generate system report: %w", err)
			}
			if err := saveReport(filesystem, cfg, output, reportType, format); err != nil {
				return err
			}
		case models.TypeFull:
			output, err := generator.GenerateFullReport(ctx, format)
			if err != nil {
				return fmt.Errorf("failed to generate full report: %w", err)
			}
			if err := saveReport(filesystem, cfg, output, reportType, format); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown report type: %s", reportType)
		}
	}

	return nil
}

// saveReport saves a single report and displays success message
func saveReport(filesystem fs.FS, cfg *config.Config, output *models.ReportOutput, reportType models.ReportType, format models.ReportFormat) error {
	// Save report to file using config's output directory
	outputDir := cfg.Report.GetReportOutputDir(cfg.DataDir)
	filename := generateReportFilename(reportType, format)
	outputPath := filepath.Join(outputDir, filename)

	if err := saveReportToFile(filesystem, output, outputPath); err != nil {
		return fmt.Errorf("failed to save report: %w", err)
	}

	// Display success message
	fmt.Printf("\n✅ Report generated successfully!\n")
	fmt.Printf("📁 Saved to: %s\n", outputPath)
	fmt.Printf("⏱️  Generated in: %dms\n", output.Metadata.GenerationTimeMs)
	fmt.Printf("📊 Data sources used: %v\n", output.Metadata.DataSources)

	return nil
}

// SelectReportType prompts user to select the type of report
func SelectReportType() (models.ReportType, error) {
	items := []struct {
		Display string
		Value   models.ReportType
	}{
		{"📊 Customer Report - Portfolio analysis and insights", models.TypeCustomer},
		{"🔧 System Report - Health and diagnostics", models.TypeSystem},
		{"📋 Full Report - Complete customer and system data", models.TypeFull},
	}

	prompt := promptui.Select{
		Label: "Select report type",
		Items: items,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}:",
			Active:   "▶ {{ .Display }}",
			Inactive: "  {{ .Display }}",
			Selected: "✓ {{ .Display }}",
		},
	}

	i, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return items[i].Value, nil
}

// SelectReportFormats prompts the user to select which report formats to generate
func SelectReportFormats(defaultFormat models.ReportFormat) ([]models.ReportFormat, error) {
	options := []struct {
		Display string
		Value   models.ReportFormat
	}{
		{"Markdown (.md)", models.FormatMarkdown},
		{"PDF (.pdf)", models.FormatPDF},
		{"JSON (.json)", models.FormatJSON},
		{"All formats", models.FormatAll},
	}

	// Convert to display strings for prompt
	items := make([]string, len(options))
	for i, option := range options {
		items[i] = option.Display
		// Add (default) indicator
		if option.Value == defaultFormat {
			items[i] += " (default)"
		}
	}

	prompt := promptui.SelectWithAdd{
		Label:    "Report format(s)",
		Items:    items,
		AddLabel: "Use default format only",
	}

	index, _, err := prompt.Run()
	if err != nil {
		return nil, fmt.Errorf("report format selection failed: %w", err)
	}

	// Handle "Use default format only" option
	if index == -1 {
		return []models.ReportFormat{defaultFormat}, nil
	}

	// Handle "All formats" option
	if options[index].Value == models.FormatAll {
		return []models.ReportFormat{models.FormatMarkdown, models.FormatPDF, models.FormatJSON}, nil
	}

	// Return single selected format
	return []models.ReportFormat{options[index].Value}, nil
}

// PromptCustomerName prompts for an optional customer name
func PromptCustomerName() (string, error) {
	prompt := promptui.Prompt{
		Label:   "Customer name (optional)",
		Default: "",
	}

	return prompt.Run()
}

// DisplayAvailableReportData shows what data is available for reporting
func DisplayAvailableReportData(dataBag bag.Bag) {
	fmt.Println("\n📋 Available data for reporting:")
	fmt.Println("=" + fmt.Sprintf("%50s", "="))

	dataKeys := dataBag.Keys()
	if len(dataKeys) == 0 {
		fmt.Println("  No data available")
		return
	}

	// Group keys by category for better display
	categories := map[string][]string{
		"Portfolio": {},
		"Analysis":  {},
		"Health":    {},
		"Other":     {},
	}

	for _, key := range dataKeys {
		keyStr := key.String()
		if contains(keyStr, "portfolio", "holding") {
			categories["Portfolio"] = append(categories["Portfolio"], keyStr)
		} else if contains(keyStr, "risk", "analysis", "insight", "news") {
			categories["Analysis"] = append(categories["Analysis"], keyStr)
		} else if contains(keyStr, "health", "metrics", "cache", "tool") {
			categories["Health"] = append(categories["Health"], keyStr)
		} else {
			categories["Other"] = append(categories["Other"], keyStr)
		}
	}

	for category, keyList := range categories {
		if len(keyList) > 0 {
			fmt.Printf("\n  %s:\n", category)
			for _, key := range keyList {
				keyObj := keys.Key(key)
				if val, exists := dataBag.Get(keyObj); exists {
					fmt.Printf("    • %s (%T)\n", key, val)
				}
			}
		}
	}
}

// DisplayReportSummary shows what each report type includes
func DisplayReportSummary() {
	fmt.Println("\n📋 Report Types Overview:")
	fmt.Println("=" + fmt.Sprintf("%50s", "="))

	fmt.Println("\n📊 CUSTOMER REPORT")
	fmt.Println("   • Portfolio overview and account summary")
	fmt.Println("   • Risk assessment and metrics")
	fmt.Println("   • Asset allocation analysis")
	fmt.Println("   • Performance data and insights")
	fmt.Println("   • Compliance status")
	fmt.Println("   • Market context and news analysis")
	fmt.Println("   • Individual holdings analysis")

	fmt.Println("\n🔧 SYSTEM REPORT")
	fmt.Println("   • Application health and uptime")
	fmt.Println("   • Tool performance metrics")
	fmt.Println("   • Cache performance statistics")
	fmt.Println("   • External API health monitoring")
	fmt.Println("   • Data freshness indicators")
	fmt.Println("   • Recent tool activity logs")
	fmt.Println("   • Component health status")

	fmt.Println("\n📋 FULL REPORT")
	fmt.Println("   • Complete combination of customer and system reports")
	fmt.Println("   • Comprehensive view for technical stakeholders")

	fmt.Println("\n📁 Output Formats: Markdown, PDF, JSON")
}

// Helper functions

func generateReportFilename(reportType models.ReportType, format models.ReportFormat) string {
	timestamp := time.Now().Format("20060102_150405")

	var extension string
	switch format {
	case models.FormatMarkdown:
		extension = "md"
	case models.FormatPDF:
		extension = "pdf"
	case models.FormatJSON:
		extension = "json"
	default:
		extension = "txt"
	}

	return fmt.Sprintf("%s_report_%s.%s", reportType, timestamp, extension)
}

func saveReportToFile(filesystem fs.FS, output *models.ReportOutput, outputPath string) error {
	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if err := filesystem.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// For PDF, the file may already be created by the PDF converter
	if output.FilePath != "" && output.Format == models.FormatPDF {
		// Move/copy the PDF file to the desired location
		return filesystem.Rename(output.FilePath, outputPath)
	}

	// Write content to file
	return filesystem.WriteFile(outputPath, []byte(output.Content), 0644)
}

func contains(str string, substrings ...string) bool {
	for _, substr := range substrings {
		if len(str) >= len(substr) {
			for i := 0; i <= len(str)-len(substr); i++ {
				if str[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
