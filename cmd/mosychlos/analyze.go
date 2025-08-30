package mosychlos

import (
	"context"
	"log/slog"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/internal/engine"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/fs"
	"github.com/spf13/cobra"
)

func NewAnalyzeCommand(cfg *config.Config) *cobra.Command {
	analyzeCmd := &cobra.Command{
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
	// Add agents flag to use the new agent-based engines
	analyzeCmd.Flags().Bool("agents", false, "Use agent-based analysis with OpenAI Agents Go SDK (experimental)")
	// Add report generation flags
	analyzeCmd.Flags().Bool("reports", false, "Generate reports after analysis")
	analyzeCmd.Flags().Bool("all-formats", false, "Generate reports in all formats (markdown, PDF, JSON)")
	analyzeCmd.Flags().Bool("markdown", false, "Generate markdown reports")
	analyzeCmd.Flags().Bool("pdf", false, "Generate PDF reports")
	analyzeCmd.Flags().Bool("json", false, "Generate JSON reports")

	return analyzeCmd
}

func runAnalyzeCommand(cmd *cobra.Command, _ []string, cfg *config.Config) error {
	batch, _ := cmd.Flags().GetBool("batch")
	useAgents, _ := cmd.Flags().GetBool("agents")

	var builder *engine.RegistryBuilder
	
	if useAgents {
		slog.Info("Using agent-based analysis with OpenAI Agents Go SDK")
		builder = engine.DefaultAgentsRegistry()
	} else if batch {
		slog.Debug("using batch engine builder", "builder", builder)
		builder = engine.DefaultBatchRegistry()
	} else {
		slog.Debug("using default engine builder", "builder", builder)
		builder = engine.DefaultRegistry()
	}

	// Recreate orchestrator with builder injected
	o := engine.New(
		cfg,
		engine.WithBag(bag.NewSharedBag()),
		engine.WithFS(fs.OS{}),
		engine.WithBuilder(builder),
	)

	ctx := context.Background()
	err := o.Init(ctx)
	if err != nil {
		slog.Error("failed to initialize engine orchestrator", "error", err)
		return err
	}

	err = o.ExecutePipeline(ctx)
	if err != nil {
		slog.Error("failed to execute engine pipeline", "error", err)
		return err
	}

	return nil
}
