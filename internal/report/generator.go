// GenerateReport generates a report from FullReportData, outputDir, and format (for CLI use)
package report

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/fs"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/amaurybrisou/mosychlos/pkg/pdf"
)

// Generator implements the ReportGenerator interface
type Generator struct {
	deps      Dependencies
	pdf       pdf.Converter
	bagLoader BagLoader
}

// NewGenerator creates a new report generator
func NewGenerator(deps Dependencies) *Generator {
	// Configure PDF converter based on config
	var pdfOptions []pdf.Option
	if deps.Config.Report.PDFEngine != "" {
		pdfOptions = append(pdfOptions, pdf.WithEngines([]string{deps.Config.Report.PDFEngine}))
	}
	pdfOptions = append(pdfOptions, pdf.WithSanitize(deps.Config.Report.EnablePDFUnicodeSanitization))

	// Create bag loader with file system access
	bagLoader := NewBagLoader(deps.FileSystem)

	return &Generator{
		deps:      deps,
		pdf:       pdf.New(pdfOptions...),
		bagLoader: bagLoader,
	}
}

// GenerateCustomerReport generates a customer-facing portfolio report
func (g *Generator) GenerateCustomerReport(ctx context.Context, format models.ReportFormat) (*models.ReportOutput, error) {
	startTime := time.Now()

	customerData, err := g.bagLoader.LoadCustomerData(ctx, g.deps.DataBag)
	if err != nil {
		return nil, fmt.Errorf("failed to load customer data: %w", err)
	}

	content, dataSources, err := g.renderCustomerReport(customerData)
	if err != nil {
		return nil, fmt.Errorf("failed to render customer report: %w", err)
	}

	outputDir := filepath.Join(g.deps.Config.DataDir, g.deps.Config.Report.OutputDir)

	output := &models.ReportOutput{
		Type:        models.TypeCustomer,
		Format:      format,
		Content:     content,
		FilePath:    g.generateMarkdownFilePath(models.TypeCustomer, outputDir),
		GeneratedAt: time.Now(),
		Metadata: models.ReportMeta{
			Title:            g.getReportTitle("Portfolio Analysis Report"),
			Description:      "Comprehensive portfolio analysis and insights",
			DataSources:      dataSources,
			GenerationTimeMs: time.Since(startTime).Milliseconds(),
			Version:          "1.0.0",
		},
	}

	if err := g.processOutput(ctx, output); err != nil {
		return nil, fmt.Errorf("failed to process output: %w", err)
	}

	return output, nil
}

// GenerateSystemReport generates a system diagnostics report
func (g *Generator) GenerateSystemReport(ctx context.Context, format models.ReportFormat) (*models.ReportOutput, error) {
	startTime := time.Now()

	systemData, err := g.bagLoader.LoadSystemData(ctx, g.deps.DataBag)
	if err != nil {
		return nil, fmt.Errorf("failed to load system data: %w", err)
	}

	content, dataSources, err := g.renderSystemReport(systemData)
	if err != nil {
		return nil, fmt.Errorf("failed to render system report: %w", err)
	}

	outputDir := filepath.Join(g.deps.Config.DataDir, g.deps.Config.Report.OutputDir)

	output := &models.ReportOutput{
		Type:        models.TypeSystem,
		Format:      format,
		Content:     content,
		FilePath:    g.generateMarkdownFilePath(models.TypeSystem, outputDir),
		GeneratedAt: time.Now(),
		Metadata: models.ReportMeta{
			Title:            g.getReportTitle("System Health & Diagnostics Report"),
			Description:      "Comprehensive system health and performance diagnostics",
			DataSources:      dataSources,
			GenerationTimeMs: time.Since(startTime).Milliseconds(),
			Version:          "1.0.0",
		},
	}

	if err := g.processOutput(ctx, output); err != nil {
		return nil, fmt.Errorf("failed to process output: %w", err)
	}

	return output, nil
}

// GenerateFullReport generates a comprehensive report combining customer and system data
func (g *Generator) GenerateFullReport(ctx context.Context, format models.ReportFormat) (*models.ReportOutput, error) {
	startTime := time.Now()

	// Load full data using BagLoader
	fullData, err := g.bagLoader.LoadFullData(ctx, g.deps.DataBag)
	if err != nil {
		return nil, fmt.Errorf("failed to load full data: %w", err)
	}

	content, dataSources, err := g.renderFullReport(fullData.Customer, fullData.System)
	if err != nil {
		return nil, fmt.Errorf("failed to render full report: %w", err)
	}

	outputDir := filepath.Join(g.deps.Config.DataDir, g.deps.Config.Report.OutputDir)

	output := &models.ReportOutput{
		Type:        models.TypeFull,
		Format:      format,
		Content:     content,
		FilePath:    g.generateMarkdownFilePath(models.TypeFull, outputDir),
		GeneratedAt: time.Now(),
		Metadata: models.ReportMeta{
			Title:            g.getReportTitle("Complete Portfolio & System Report"),
			Description:      "Comprehensive analysis combining portfolio insights and system diagnostics",
			DataSources:      dataSources,
			GenerationTimeMs: time.Since(startTime).Milliseconds(),
			Version:          "1.0.0",
		},
	}

	if err := g.processOutput(ctx, output); err != nil {
		return nil, fmt.Errorf("failed to process output: %w", err)
	}

	return output, nil
}

// processOutput handles the formatting and file writing for the report output
func (g *Generator) processOutput(_ context.Context, output *models.ReportOutput) error {
	// Ensure the output directory exists
	if output.FilePath != "" {
		outputDir := filepath.Dir(output.FilePath)
		if err := g.deps.FileSystem.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Handle different output formats
	switch output.Format {
	case models.FormatJSON:
		// Ensure the output file has .json extension before marshaling
		if !strings.HasSuffix(output.FilePath, ".json") {
			output.FilePath = strings.TrimSuffix(output.FilePath, filepath.Ext(output.FilePath)) + ".json"
		}

		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal report to JSON: %w", err)
		}
		output.Content = string(jsonData)

	case models.FormatPDF:
		// First write the markdown content to a temporary file
		tempMarkdownFile := strings.TrimSuffix(output.FilePath, ".pdf") + "_temp.md"
		if err := g.deps.FileSystem.WriteFile(tempMarkdownFile, []byte(output.Content), 0644); err != nil {
			return fmt.Errorf("failed to write temporary markdown file: %w", err)
		}
		defer g.deps.FileSystem.Remove(tempMarkdownFile) // Clean up temp file

		// Convert markdown to PDF using the PDF converter (returns path to generated PDF)
		generatedPDFPath, err := g.pdf.Convert(tempMarkdownFile)
		if err != nil {
			return fmt.Errorf("failed to convert report to PDF: %w", err)
		}
		defer g.deps.FileSystem.Remove(generatedPDFPath) // Clean up generated PDF after reading

		// Read the generated PDF file content
		pdfContent, err := g.deps.FileSystem.ReadFile(generatedPDFPath)
		if err != nil {
			return fmt.Errorf("failed to read generated PDF: %w", err)
		}
		output.Content = string(pdfContent)

		// Ensure the output file has .pdf extension
		if !strings.HasSuffix(output.FilePath, ".pdf") {
			output.FilePath = strings.TrimSuffix(output.FilePath, filepath.Ext(output.FilePath)) + ".pdf"
		}

	case models.FormatMarkdown:
		// Content is already in markdown format, no conversion needed
		break

	default:
		return fmt.Errorf("unsupported report format: %s", output.Format)
	}

	return g.deps.FileSystem.WriteFile(output.FilePath, []byte(output.Content), 0644)
}

// getReportTitle returns the report title with custom overrides
func (g *Generator) getReportTitle(defaultTitle string) string {
	// For now, just return the default title
	// Future enhancement: add custom title support
	return defaultTitle
}

// generateMarkdownFilePath generates the output file path for markdown reports
func (g *Generator) generateMarkdownFilePath(reportType models.ReportType, outputDir string) string {
	timestamp := ""
	if g.deps.Config.Report.IncludeTimestamp {
		timestamp = "_" + time.Now().Format("20060102_150405")
	}

	filename := fmt.Sprintf("%s_report%s.md", reportType, timestamp)
	return filepath.Join(outputDir, filename)
}

func GenerateReport(fullData *models.FullReportData, outputDir, format string) error {
	// Minimal dependencies for file writing
	fsys := fs.New()
	deps := Dependencies{
		Config:     nil, // Not needed for direct CLI
		DataBag:    nil, // Not used here
		FileSystem: fsys,
	}
	gen := &Generator{deps: deps}

	// Render report content (markdown only for now)
	content, _, err := gen.renderFullReport(fullData.Customer, fullData.System)
	if err != nil {
		return err
	}

	// Determine file extension
	ext := ".md"
	switch format {
	case "pdf":
		ext = ".pdf"
	case "json":
		ext = ".json"
	}

	filename := "full_report_" + time.Now().Format("20060102_150405") + ext
	filePath := filepath.Join(outputDir, filename)

	// Write file
	if format == "json" {
		// Marshal as JSON
		out := map[string]any{
			"customer": fullData.Customer,
			"system":   fullData.System,
		}
		data, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return err
		}
		return fsys.WriteFile(filePath, data, 0644)
	}

	return fsys.WriteFile(filePath, []byte(content), 0644)
}
