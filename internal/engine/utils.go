package engine

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/amaurybrisou/mosychlos/internal/adapters"
	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/internal/llm"
	"github.com/amaurybrisou/mosychlos/internal/localization"
	"github.com/amaurybrisou/mosychlos/internal/portfolio"
	"github.com/amaurybrisou/mosychlos/internal/profile"
	"github.com/amaurybrisou/mosychlos/internal/prompt"
	"github.com/amaurybrisou/mosychlos/internal/tools"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/binance"
	"github.com/amaurybrisou/mosychlos/pkg/fs"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// initializeTools sets up tools with shared bag for metrics tracking
func initializeTools(cfg *config.Config, sharedBag bag.SharedBag) error {
	tools.SetSharedBag(sharedBag)
	if err := tools.NewTools(cfg); err != nil {
		return err
	}

	availableTools := tools.GetTools()
	toolNames := make([]string, 0, len(availableTools))
	for _, tool := range availableTools {
		toolNames = append(toolNames, tool.Name())
	}
	slog.Debug("Initialized tools", "count", len(availableTools), "tools", toolNames)

	return nil
}

// loadPortfolioData creates portfolio service and loads portfolio data
func loadPortfolioData(ctx context.Context, cfg *config.Config, filesystem fs.FS, sharedBag bag.SharedBag) error {
	portfolioService := portfolio.NewService(cfg, filesystem, sharedBag)
	binanceProvider := binance.NewPortfolioProvider(&cfg.Binance)

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

	sharedBag.Set(keys.KPortfolio, portfolioData)
	return nil
}

// loadInvestmentProfile creates profile manager and loads investment profile
func loadInvestmentProfile(ctx context.Context, cfg *config.Config, filesystem fs.FS, sharedBag bag.SharedBag) error {
	// Create profile manager using os.DirFS for the filesystem interface
	profileManager, err := profile.NewProfileManager(filesystem, cfg.ConfigDir, sharedBag)
	if err != nil {
		return fmt.Errorf("failed to create profile manager: %w", err)
	}

	// Use localization from config to determine country and default risk tolerance
	country := cfg.Localization.Country
	riskTolerance := keys.KRiskToleranceAggressive // Default risk tolerance

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
	sharedBag.Set(keys.KProfile, investmentProfile)

	return nil
}

func loadRegionalConfig(_ context.Context, cfg *config.Config, filesystem fs.FS, sharedBag bag.SharedBag) error {
	svc := localization.New(filesystem, cfg.ConfigDir)
	regionalCfg, err := svc.LoadRegionalConfig(sharedBag, cfg.Localization.Country, cfg.Localization.Language)
	slog.Debug("Loaded regional config", slog.Any("config", regionalCfg))
	sharedBag.Set(keys.KRegionalConfig, regionalCfg)
	return err
}

func loadPromptManager(ctx context.Context, cfg *config.Config, filesystem fs.FS, sharedBag bag.SharedBag) (models.PromptBuilder, error) {
	// Create prompt manager
	promptConfig := prompt.Config{
		UserLocalization: cfg.Localization,
		UserProfile:      sharedBag.MustGet(keys.KProfile).(*models.InvestmentProfile),
	}

	promptDeps := prompt.Dependencies{
		Bag:    sharedBag.Snapshot(),
		Config: promptConfig,
	}

	promptManager, err := prompt.NewManager(promptDeps)
	if err != nil {
		return nil, fmt.Errorf("failed to create prompt manager: %w", err)
	}

	return promptManager, nil
}

func loadAiClient(_ context.Context, cfg *config.Config, sharedBag bag.SharedBag) (*llm.Client, error) {
	return llm.NewLLMClient(cfg, sharedBag)
}
