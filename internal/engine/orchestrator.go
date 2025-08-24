package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/internal/llm"
	"github.com/amaurybrisou/mosychlos/internal/tools"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/fs"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/google/uuid"

	_ "github.com/amaurybrisou/mosychlos/internal/toolsimpl" // registers all tools
)

type Orchestrator interface {
	Init(ctx context.Context) error
	ExecutePipeline(ctx context.Context) error
	Bag() bag.SharedBag
	Tools() models.ToolProvider
}

// engineOrchestrator owns shared state (SharedBag), builds shared services, and wires engines via a Builder.
type engineOrchestrator struct {
	ID            uuid.UUID
	StartDate     time.Time
	cfg           *config.Config
	engines       []models.Engine // if provided directly (option 1)
	builder       Builder         // if you prefer DI inside orchestrator (option 2)
	sharedBag     bag.SharedBag
	toolManager   *tools.ToolManager
	promptManager models.PromptBuilder
	filesystem    fs.FS
	aiClient      *llm.Client
	// pluggable initialization steps
	initSteps []InitStep
}

// Option configures the orchestrator at construction time.
type Option func(*engineOrchestrator)

// WithFS overrides the default filesystem implementation.
func WithFS(f fs.FS) Option { return func(o *engineOrchestrator) { o.filesystem = f } }

// WithBag injects an existing shared bag instance.
func WithBag(b bag.SharedBag) Option { return func(o *engineOrchestrator) { o.sharedBag = b } }

// WithEngines sets a pre-built engines slice.
func WithEngines(engs []models.Engine) Option {
	return func(o *engineOrchestrator) { o.engines = engs }
}

// WithBuilder injects a custom engine builder.
func WithBuilder(b Builder) Option { return func(o *engineOrchestrator) { o.builder = b } }

// New constructs an engine orchestrator using functional options.
func New(cfg *config.Config, opts ...Option) Orchestrator {
	o := &engineOrchestrator{
		ID:         uuid.New(),
		StartDate:  time.Now(),
		cfg:        cfg,
		sharedBag:  bag.NewSharedBag(),
		filesystem: fs.OS{},
	}

	for _, opt := range opts {
		opt(o)
	}

	if len(o.initSteps) == 0 {
		o.initSteps = defaultInitSteps()
	}

	return o
}

func (o *engineOrchestrator) Bag() bag.SharedBag {
	return o.sharedBag
}

func (o *engineOrchestrator) Tools() models.ToolProvider {
	return o.toolManager
}

// UseBuilder lets you choose to wire engines inside the orchestrator.
func (o *engineOrchestrator) UseBuilder(b Builder) {
	o.builder = b
}

func (o *engineOrchestrator) BatchManager() models.BatchManager {
	return o.aiClient.BatchManager()
}

func (o *engineOrchestrator) Init(ctx context.Context) error {
	// Execute pluggable init steps (tools, portfolio, profile...)
	for i, step := range o.initSteps {
		if err := step(ctx, o); err != nil {
			return fmt.Errorf("init step %d failed: %w", i, err)
		}
	}

	if len(o.engines) == 0 &&
		o.promptManager != nil &&
		o.aiClient != nil &&
		o.toolManager != nil {

		if o.builder == nil {
			o.builder = DefaultBatchRegistry()
		}

		engs, err := o.builder.Build(ctx, Deps{
			Ctx:       ctx,
			Config:    o.cfg,
			SharedBag: o.sharedBag,
			FS:        o.filesystem,
			AI:        o.aiClient,
			Prompts:   o.promptManager,
			Tools:     o.toolManager,
		})
		if err != nil {
			return fmt.Errorf("failed to build engines: %w", err)
		}
		o.engines = engs
	}

	return nil
}

func (o *engineOrchestrator) ExecutePipeline(ctx context.Context) error {
	if o.aiClient == nil {
		return fmt.Errorf("orchestrator not initialized: aiClient is nil")
	}

	if len(o.engines) == 0 {
		return fmt.Errorf("no engines configured")
	}

	// DumpSharedBag every 10s
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			o.DumpSharedBag()
		}
	}()

	resultKeys := make([]bag.Key, len(o.engines))

	for i, eng := range o.engines {
		slog.Info("engine: start", "name", eng.Name())

		if err := eng.Execute(ctx, o.aiClient, o.sharedBag); err != nil {
			slog.Error("engine: failed", "name", eng.Name(), "err", err)
			return err
		}

		if _, ok := o.sharedBag.Get(eng.ResultKey()); !ok {
			return fmt.Errorf("missing result for engine: %s at key %s", eng.Name(), eng.ResultKey())
		}

		resultKeys[i] = eng.ResultKey()

		slog.Info("engine: done", "name", eng.Name())
	}

	// Final bag dump after all engines complete to capture final results
	slog.Info("Dumping final shared bag state after all engines completed")
	o.DumpSharedBag()

	return nil
}

// func (o *engineOrchestrator) GenerateReports(ctx context.Context, reportType models.ReportType, formats []models.ReportFormat) {
// 	// Create dependencies for report generation
// 	deps := report.Dependencies{
// 		Config:     o.cfg,
// 		DataBag:    o.sharedBag,
// 		FileSystem: o.filesystem,
// 	}

// 	generator := report.NewGenerator(deps)

// 	// If no formats specified, use default markdown
// 	if len(formats) == 0 {
// 		formats = []models.ReportFormat{models.FormatMarkdown}
// 	}

// 	// Generate reports for each format
// 	for _, format := range formats {
// 		var output *models.ReportOutput
// 		var err error

// 		switch reportType {
// 		case models.TypeCustomer:
// 			output, err = generator.GenerateCustomerReport(ctx, format)
// 		case models.TypeSystem:
// 			output, err = generator.GenerateSystemReport(ctx, format)
// 		case models.TypeFull:
// 			output, err = generator.GenerateFullReport(ctx, format)
// 		default:
// 			output, err = generator.GenerateFullReport(ctx, format)
// 		}

// 		if err != nil {
// 			slog.Error("Failed to generate report", "type", reportType, "format", format, "error", err)
// 		} else {
// 			slog.Info("Report generated successfully", "type", output.Type, "format", output.Format, "path", output.FilePath)
// 		}
// 	}
// }

// func (o *engineOrchestrator) GetResults(analysisType models.AnalysisType) (*string, error) {
// 	switch analysisType {
// 	case models.AnalysisRisk:
// 		if riskResult, exists := o.sharedBag.Get(bag.KRiskAnalysisResult); exists {
// 			switch result := riskResult.(type) {
// 			case string:
// 				return &result, nil
// 			case map[string]any:
// 				if jsonResult, err := json.Marshal(result); err == nil {
// 					resultStr := string(jsonResult)
// 					return &resultStr, nil
// 				}
// 			}
// 		}
// 		return nil, fmt.Errorf("no risk analysis results found in shared bag")
// 	case models.AnalysisInvestmentResearch:
// 		if investmentResult, exists := o.sharedBag.Get(bag.KInvestmentResearchResult); exists {
// 			// Convert the structured result to JSON string for display
// 			if result, ok := investmentResult.(models.InvestmentResearchResult); ok {
// 				resultJSON, err := json.Marshal(result)
// 				if err != nil {
// 					return nil, fmt.Errorf("failed to marshal investment research result: %w", err)
// 				}
// 				resultStr := string(resultJSON)
// 				return &resultStr, nil
// 			}
// 		}
// 		return nil, fmt.Errorf("no investment research results found in shared bag")
// 	default:
// 		return nil, fmt.Errorf("result extraction for analysis type %s not implemented", analysisType)
// 	}
// }

func (o *engineOrchestrator) GetID() uuid.UUID {
	return o.ID
}

// DumpSharedBag dumps the bag content in mosychlos-data/bag/<run_id>.json
func (o *engineOrchestrator) DumpSharedBag() {
	path := fmt.Sprintf("%s/bag/%s_%s.json", o.cfg.DataDir, o.StartDate.Format("20060102_150405"), o.ID.String())
	base := filepath.Dir(path)
	err := o.filesystem.MkdirAll(base, os.ModePerm)
	if err != nil {
		slog.Error("Failed to create bag dump directory", "error", err)
		return
	}

	dataBytes, err := json.Marshal(o.sharedBag)
	if err != nil {
		slog.Error("Failed to marshal bag content", "error", err)
		return
	}

	err = o.filesystem.WriteFile(
		filepath.Join(path),
		dataBytes,
		0644,
	)
	if err != nil {
		slog.Error("Failed to open bag dump file", "error", err)
		return
	}
}
