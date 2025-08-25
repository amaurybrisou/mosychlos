package engine

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/amaurybrisou/mosychlos/internal/adapters"
	"github.com/amaurybrisou/mosychlos/internal/health"
	"github.com/amaurybrisou/mosychlos/internal/llm"
	"github.com/amaurybrisou/mosychlos/internal/localization"
	"github.com/amaurybrisou/mosychlos/internal/portfolio"
	"github.com/amaurybrisou/mosychlos/internal/profile"
	"github.com/amaurybrisou/mosychlos/internal/prompt"
	"github.com/amaurybrisou/mosychlos/internal/tools"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/binance"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/amaurybrisou/mosychlos/pkg/normalize"
)

// WithInitSteps replaces the default init steps with custom steps.
func WithInitSteps(steps ...InitStep) Option {
	return func(o *engineOrchestrator) { o.initSteps = append([]InitStep(nil), steps...) }
}

// InitStep is a pluggable initialization step.
type InitStep func(ctx context.Context, o *engineOrchestrator) error

// defaultInitSteps returns the default initialization sequence.
func defaultInitSteps() []InitStep {
	return []InitStep{
		StepStartHealthMonitoring,
		StepInitToolManager,
		StepLoadProfile,
		StepLoadRegionalSettings,
		StepLoadPortfolio,
		StepInitAIClient,
		StepInitPromptManager,
	}
}

// StepInitToolManager sets up tools with shared bag for metrics tracking and caches list to orchestrator
func StepInitToolManager(_ context.Context, o *engineOrchestrator) error {
	reg := normalize.DefaultRegistry()

	tm, err := tools.NewToolManager(o.cfg, o.sharedBag, reg)
	if err != nil {
		return fmt.Errorf("failed to load tool manager: %w", err)
	}

	o.toolManager = tm

	return nil
}

// StepLoadPortfolio loads portfolio data into the shared bag
func StepLoadPortfolio(ctx context.Context, o *engineOrchestrator) error {
	portfolioService := portfolio.NewService(o.cfg, o.filesystem, o.sharedBag)
	binanceProvider := binance.NewPortfolioProvider(&o.cfg.Binance)

	portfolioData, err := portfolioService.GetPortfolio(
		ctx, adapters.NewBinanceFetcher(binanceProvider))
	if err != nil {
		return fmt.Errorf("failed to get portfolio data: %w", err)
	}

	if portfolioData != nil {
		totalHoldings := 0
		for _, account := range portfolioData.Accounts {
			totalHoldings += len(account.Holdings)
		}
		slog.Debug("Portfolio loaded", "accounts", len(portfolioData.Accounts), "total_holdings", totalHoldings)
	} else {
		slog.Debug("No portfolio data available")
	}

	o.sharedBag.Set(bag.KPortfolio, portfolioData)
	return nil
}

// StepLoadProfile loads investment profile and regional configuration
func StepLoadProfile(ctx context.Context, o *engineOrchestrator) error {
	// Create profile manager using os.DirFS for the filesystem interface
	profileManager, err := profile.NewProfileManager(o.filesystem, o.cfg.ConfigDir, o.sharedBag)
	if err != nil {
		return fmt.Errorf("failed to create profile manager: %w", err)
	}

	// Use localization from config to determine country and default risk tolerance
	country := o.cfg.Localization.Country
	riskTolerance := bag.KRiskToleranceAggressive // Default risk tolerance

	// Load the investment profile
	investmentProfile, err := profileManager.LoadProfile(ctx, country, riskTolerance.String())
	if err != nil {
		return fmt.Errorf("failed to load investment profile: %w", err)
	}

	slog.Debug("Investment profile loaded",
		"country", investmentProfile.RegionalContext.Country,
		"language", investmentProfile.RegionalContext.Language,
		"currency", investmentProfile.RegionalContext.Currency,
		"investment_style", investmentProfile.InvestmentStyle,
		"research_depth", investmentProfile.ResearchDepth,
	)

	// Store in shared bag for engine use
	o.sharedBag.Set(bag.KProfile, investmentProfile)

	return nil
}

// StepLoadRegionalSettings loads regional configuration for the orchestrator.
func StepLoadRegionalSettings(ctx context.Context, o *engineOrchestrator) error {
	svc := localization.New(o.filesystem, o.cfg.ConfigDir)
	regionalCfg, err := svc.LoadRegionalConfig(o.sharedBag, o.cfg.Localization.Country, o.cfg.Localization.Language)
	slog.Debug("Loaded regional config", slog.Any("config", regionalCfg))
	o.sharedBag.Set(bag.KRegionalConfig, regionalCfg)
	return err
}

// StepStartHealthMonitoring starts health monitoring for the application.
func StepStartHealthMonitoring(ctx context.Context, o *engineOrchestrator) error {
	healthMonitor := health.NewApplicationMonitor(o.sharedBag)
	healthMonitor.StartPeriodicHealthCheck(15 * time.Second)
	return nil
}

func StepInitPromptManager(ctx context.Context, o *engineOrchestrator) error {
	// Create prompt manager
	promptConfig := prompt.Config{
		UserLocalization: o.cfg.Localization,
		UserProfile:      o.sharedBag.MustGet(bag.KProfile).(*models.InvestmentProfile),
	}

	promptDeps := prompt.Dependencies{
		Bag:    o.sharedBag.Snapshot(),
		Config: promptConfig,
	}

	promptManager, err := prompt.NewManager(promptDeps)
	if err != nil {
		return fmt.Errorf("failed to create prompt manager: %w", err)
	}

	o.promptManager = promptManager
	return nil
}

// StepInitAIClient loads the AI client for the orchestrator.
func StepInitAIClient(ctx context.Context, o *engineOrchestrator) error {
	aiClient, err := llm.NewLLMClient(o.cfg, o.sharedBag)
	if err != nil {
		return fmt.Errorf("failed to load AI client: %w", err)
	}
	o.aiClient = aiClient
	return nil
}
