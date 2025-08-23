package investmentresearch

// import (
// 	"context"
// 	"fmt"
// 	"time"

// 	"github.com/amaurybrisou/mosychlos/internal/budget"
// 	"github.com/amaurybrisou/mosychlos/internal/prompt"
// 	"github.com/amaurybrisou/mosychlos/pkg/bag"
// 	"github.com/amaurybrisou/mosychlos/pkg/keys"
// 	"github.com/amaurybrisou/mosychlos/pkg/models"
// )

// // internal/engine/investment_research/engine.go
// type InvestmentResearchEngine struct {
// 	regionalPromptManager prompt.RegionalManager
// 	localizationConfig    models.LocalizationConfig // For regional context
// 	toolConstraints       models.ToolConstraints

// 	// Analysis configuration
// 	researchDepth       string // "basic", "standard", "comprehensive"
// 	maxWebSearches      int    // Based on research depth
// 	includeAlternatives bool   // Include crypto, commodities, etc.
// }

// var _ models.Engine = &InvestmentResearchEngine{}

// func NewEngine(
// 	regionalPromptManager prompt.RegionalManager,
// 	researchDepth string,
// 	maxWebSearches int,
// 	includeAlternatives bool,
// ) *InvestmentResearchEngine {
// 	return &InvestmentResearchEngine{
// 		regionalPromptManager: regionalPromptManager,
// 		researchDepth:         researchDepth,
// 		maxWebSearches:        maxWebSearches,
// 		includeAlternatives:   includeAlternatives,
// 		toolConstraints:       getToolConstraints(researchDepth),
// 	}
// }

// // Engine interface implementation
// func (e *InvestmentResearchEngine) Name() string {
// 	return "investment_research"
// }

// func (e *InvestmentResearchEngine) Description() string {
// 	return "Conducts comprehensive investment research using web search and regional context"
// }

// func (e *InvestmentResearchEngine) RequiredData() []string {
// 	return []string{"portfolio", "investment_profile", "localization"}
// }

// func (e *InvestmentResearchEngine) Execute(ctx context.Context, client models.AiClient, sharedBag bag.SharedBag) error {
// 	e.toolConstraints = getToolConstraints(e.researchDepth)

// 	client.SetToolConsumer(budget.NewToolConsumer(&e.toolConstraints))

// 	// 1. Extract data from shared bag
// 	portfolio := sharedBag.MustGet(bag.KPortfolioNormalizedForAI).(*models.NormalizedPortfolio)

// 	// 2. Extract investment profile from shared bag
// 	investmentProfile := sharedBag.MustGet(bag.KProfile).(*models.InvestmentProfile)

// 	// 2. Generate regional research prompt
// 	prompt, err := e.regionalPromptManager.GenerateRegionalPrompt(
// 		ctx,
// 		models.AnalysisInvestmentResearch,
// 		sharedBag,
// 		prompt.PromptData{
// 			EngineVersion:     "v2.0",
// 			AnalysisType:      models.AnalysisInvestmentResearch,
// 			Timestamp:         time.Now(),
// 			RegionalConfig:    sharedBag.MustGet(bag.KRegionalConfig).(*models.RegionalConfig),
// 			Portfolio:         portfolio,
// 			InvestmentProfile: investmentProfile,
// 			Localization: models.LocalizationConfig{
// 				Country:  investmentProfile.RegionalContext.Country,
// 				Language: investmentProfile.RegionalContext.Language,
// 				Currency: investmentProfile.RegionalContext.Currency,
// 				Timezone: investmentProfile.RegionalContext.Timezone,
// 				Region:   investmentProfile.RegionalContext.Region,
// 				City:     investmentProfile.RegionalContext.City,
// 			},
// 		},
// 	)
// 	if err != nil {
// 		return fmt.Errorf("failed to generate prompt: %w", err)
// 	}

// 	// 2. Execute analysis with tool access
// 	// Add current date instruction to the prompt
// 	prompt = prompt + fmt.Sprintf("\n\n**Important**: When generating the metadata.generated_at field, use the current date and time: %s", time.Now().Format("2006-01-02T15:04:05Z07:00"))

// 	result, err := ai.Ask[models.InvestmentResearchResult](ctx, client, prompt)
// 	if err != nil {
// 		return err
// 	}

// 	// 4. Store results in shared bag for potential chaining
// 	sharedBag.Set(bag.KInvestmentResearchResult, result)

// 	return nil
// }

// // Tool constraints optimized for investment research
// func getToolConstraints(researchDepth string) models.ToolConstraints {
// 	baseConstraints := models.ToolConstraints{
// 		RequiredTools: []bag.Key{
// 			bag.WebSearch, // Web search for market intelligence and real-time data
// 			bag.FMP,       // Market data for research
// 			bag.NewsApi,   // News and market intelligence
// 		},
// 		PreferredTools: []bag.Key{
// 			bag.Fred,                // Economic context
// 			bag.YFinanceStockData,   // Additional market data
// 			bag.FMPAnalystEstimates, // Analyst insights
// 		},
// 	}

// 	// Adjust tool usage based on research depth
// 	switch researchDepth {
// 	case "comprehensive":
// 		baseConstraints.MaxCallsPerTool = map[bag.Key]int{
// 			bag.WebSearch: 8, // Comprehensive web research
// 			bag.FMP:       4, // Comprehensive data
// 			bag.NewsApi:   2, // News context
// 		}
// 		baseConstraints.MinCallsPerTool = map[bag.Key]int{
// 			bag.WebSearch: 4, // Minimum web searches for comprehensive analysis
// 		}

// 	case "standard":
// 		baseConstraints.MaxCallsPerTool = map[bag.Key]int{
// 			bag.WebSearch: 5, // Moderate web research
// 			bag.FMP:       2,
// 			bag.NewsApi:   1,
// 		}
// 		baseConstraints.MinCallsPerTool = map[bag.Key]int{
// 			bag.WebSearch: 3, // Minimum web searches for standard analysis
// 		}

// 	case "basic":
// 		baseConstraints.MaxCallsPerTool = map[bag.Key]int{
// 			bag.WebSearch: 3, // Basic web research
// 			bag.FMP:       1,
// 			bag.NewsApi:   1,
// 		}
// 		baseConstraints.MinCallsPerTool = map[bag.Key]int{
// 			bag.WebSearch: 2, // Minimum web searches for basic analysis
// 		}
// 	}

// 	return baseConstraints
// }
