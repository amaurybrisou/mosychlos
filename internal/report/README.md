# Report Package

The `internal/report` package provides comprehensive reporting capabilities for the Mosychlos portfolio management system, extracting data from the SharedBag and generating formatted reports in multiple output formats.

## Architecture

### Package Structure

```
internal/report/              # Core report generation logic
├── generator.go              # Main report generator implementation
├── data.go                   # Data extraction from SharedBag
├── renderer.go               # Template rendering and formatting
├── types.go                  # Type definitions and interfaces
├── templates/               # Report templates
│   ├── customer.md          # Customer report template
│   └── system.md            # System report template
└── README.md                # Documentation

pkg/cli/
└── report.go                # CLI interface for interactive report generation
```

### Design Principles

- **Separation of Concerns**: Business logic in `internal/report`, CLI interface in `pkg/cli`
- **Template-based**: Flexible report formatting using Go templates
- **Multiple Formats**: Support for Markdown, PDF, and JSON outputs
- **Data-driven**: Extracts data from SharedBag using well-known keys

## Features

### Report Types

#### 📊 Customer Report (`TypeCustomer`)

Customer-facing portfolio analysis report including:

- Portfolio overview and account summary
- Risk assessment and metrics
- Asset allocation analysis
- Performance data and insights
- Compliance status
- Market context and news analysis
- Individual holdings analysis

#### 🔧 System Report (`TypeSystem`)

System diagnostic and health report including:

- Application health status and uptime
- Tool execution metrics and performance
- Cache performance statistics
- External API health monitoring
- Data freshness indicators
- Recent tool activity logs
- Component health status

#### 📋 Full Report (`TypeFull`)

Comprehensive report combining both customer and system data for complete visibility.

### Output Formats

- **Markdown** (`FormatMarkdown`): Human-readable format with proper formatting
- **PDF** (`FormatPDF`): Professional document format using the existing PDF converter
- **JSON** (`FormatJSON`): Machine-readable format for API integrations

## Usage

### Basic Usage

```go
import (
    "context"
    "github.com/amaurybrisou/mosychlos/internal/config"
    "github.com/amaurybrisou/mosychlos/internal/report"
    "github.com/amaurybrisou/mosychlos/pkg/bag"
    "github.com/amaurybrisou/mosychlos/pkg/fs"
)

// Load configuration (includes report settings)
cfg := config.MustLoadConfig()

// Create dependencies
deps := report.Dependencies{
    DataBag:    sharedBag.Snapshot(),
    Config:     cfg,
    FileSystem: fs.OS{},
}

// Create generator
generator := report.New(deps)

// Generate customer report
ctx := context.Background()
output, err := generator.GenerateCustomerReport(ctx, report.FormatPDF)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Report generated: %s\n", output.FilePath)
```

### CLI Integration

The CLI interface is provided through `pkg/cli/report.go`:

```go
import (
    "context"
    "github.com/amaurybrisou/mosychlos/pkg/cli"
)

// Generate report interactively
err := cli.GenerateReport(ctx, dataBag, "./reports")

// Display available data
cli.DisplayAvailableReportData(dataBag)

// Show report type summary
cli.DisplayReportSummary()
```

## Data Sources

The report package extracts data from the SharedBag using well-known keys:

### Customer Data Keys

- `KPortfolio` - Portfolio data
- `KRiskMetrics` - Risk analysis results
- `KPortfolioAllocationData` - Asset allocation data
- `KPortfolioPerformanceData` - Performance metrics
- `KPortfolioComplianceData` - Compliance information
- `KStockAnalysis` - Individual stock analysis
- `KInsights` - AI-generated insights
- `KNewsAnalyzed` - Analyzed market news
- `KFundamentals` - Fundamental analysis data

### System Data Keys

- `KApplicationHealth` - Overall system health
- `KToolMetrics` - Tool execution statistics
- `KCacheStats` - Cache performance metrics
- `KExternalDataHealth` - External API health
- `KMarketDataFreshness` - Data age and quality
- `KToolComputations` - Recent tool executions

## Template System

The package uses Go's `text/template` system with custom functions:

### Template Functions

- `formatDuration` - Format time.Duration for display
- `formatBytes` - Format byte sizes with appropriate units
- `multiply` - Mathematical operations for percentages
- `toUpper` - String formatting
- `slice` - Array slicing for recent items

### Custom Templates

Templates can be overridden via `ReportConfig.TemplateOverrides`:

```go
config := report.ReportConfig{
    TemplateOverrides: map[string]string{
        "customer_header": "# Custom Portfolio Report for {{.CustomerName}}",
    },
}
```

## Integration with Existing Systems

### PDF Generation

Leverages the existing `pkg/pdf` package for PDF conversion:

- Supports multiple LaTeX engines (xelatex, lualatex)
- Unicode sanitization fallback
- Pandoc integration

### Health Monitoring

Integrates with `internal/health` package:

- Real-time health status
- Performance metrics tracking
- Component health monitoring

### SharedBag Integration

Works seamlessly with the application's SharedBag pattern:

- Thread-safe data access
- Snapshot-based consistency
- Well-known key patterns

## Error Handling

The package provides comprehensive error handling:

- Template parsing errors
- Data extraction failures
- File I/O errors
- PDF conversion errors

## Performance Considerations

- Uses bag snapshots for consistent data views
- Template caching for repeated generations
- Minimal memory footprint for large datasets
- Configurable data source filtering

## Command Line Integration

The report functionality integrates with the main CLI through `pkg/cli/report.go`. You can add it to your main CLI commands:

````go
// Add to cmd/mosychlos/main.go
var reportCmd = &cobra.Command{
    Use:   "report",
    Short: "Generate portfolio and system reports",
    Long:  `Generate comprehensive reports from portfolio and system data.`,
    Run:   reportCommand,
}

func reportCommand(cmd *cobra.Command, args []string) {
    // Load configuration
    cfg := config.MustLoadConfig()

    // Create filesystem and shared bag
    filesystem := fs.OS{}
    sharedBag := bag.NewSharedBag()

    // Initialize system (same as portfolio/analyze commands)
    tools.SetSharedBag(sharedBag)
    if err := tools.NewTools(cfg); err != nil {
        slog.Error("Failed to initialize tools", "error", err)
        os.Exit(1)
    }

    // Initialize health monitoring
    healthMonitor := health.NewApplicationMonitor(sharedBag)
    healthMonitor.StartPeriodicHealthCheck(30 * time.Second)

    // Load portfolio data (optional, reports work without it)
    portfolioService := portfolio.NewService(cfg, filesystem, sharedBag)
    binanceProvider := binance.NewPortfolioProvider(&cfg.Binance)
    ctx := context.Background()
    _, _ = portfolioService.GetPortfolio(ctx, adapters.NewBinanceFetcher(binanceProvider))

    // Generate report interactively
    if err := cli.GenerateReport(context.Background(), cfg, sharedBag.Snapshot(), filesystem); err != nil {
        slog.Error("Failed to generate report", "error", err)
        os.Exit(1)
    }
}

// Don't forget to add the command to root:
rootCmd.AddCommand(reportCmd)
```## Future Enhancements

- Email report delivery
- Scheduled report generation
- Custom template loading from files
- Report comparison and diff generation
- Interactive report viewer
- Report archival and versioning
````
