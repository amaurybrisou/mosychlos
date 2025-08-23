package batchinvestmentresearch

import (
	"context"
	"fmt"
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/internal/prompt"
	"github.com/amaurybrisou/mosychlos/internal/tools"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/amaurybrisou/mosychlos/pkg/openai"
)

// BatchInvestmentResearchEngine creates batch requests for investment research
// This engine generates OpenAI Batch API requests that include tools and proper prompts
type BatchInvestmentResearchEngine struct {
	regionalPromptManager prompt.RegionalManager
	toolRegistry          map[bag.Key]models.Tool
	researchDepth         string // "basic", "standard", "comprehensive"
}

// NewBatchEngine creates a new batch investment research engine
func NewBatchEngine(
	regionalPromptManager prompt.RegionalManager,
	toolRegistry map[bag.Key]models.Tool,
	researchDepth string,
) *BatchInvestmentResearchEngine {
	if researchDepth == "" {
		researchDepth = "standard"
	}

	return &BatchInvestmentResearchEngine{
		regionalPromptManager: regionalPromptManager,
		toolRegistry:          toolRegistry,
		researchDepth:         researchDepth,
	}
}

// GenerateBatchRequest creates a batch request for investment research analysis
func (e *BatchInvestmentResearchEngine) GenerateBatchRequest(
	ctx context.Context,
	customID string,
	cfg *config.Config,
	sharedBag bag.SharedBag,
) (*models.BatchRequest, error) {
	// 1. Generate regional prompt using the same system as the regular engine
	prompt, err := e.generateRegionalPrompt(ctx, sharedBag)
	if err != nil {
		return nil, fmt.Errorf("failed to generate regional prompt: %w", err)
	}

	// 2. Get tool constraints for investment research
	constraints := e.getToolConstraints()

	// 3. Create chat completion request body with tools
	requestBody := map[string]any{
		"model": cfg.LLM.Model,
		"messages": []map[string]any{
			{
				"role":    "system",
				"content": prompt,
			},
			{
				"role": "user",
				"content": fmt.Sprintf("Conduct comprehensive investment research analysis for the provided portfolio. "+
					"Use all available tools to gather market data, economic indicators, and news. "+
					"Generate a structured InvestmentResearchResult with Sources that include URLs from web search citations. "+
					"Current date: %s", time.Now().Format("2006-01-02T15:04:05Z07:00")),
			},
		},
		"tools":       constraints.Tools,
		"tool_choice": "auto",
		"temperature": 0.2, // Lower temperature for consistent research
		"max_tokens":  cfg.LLM.OpenAI.MaxCompletionTokens,
		// Add response format for structured output
		"response_format": map[string]any{
			"type": "json_schema",
			"json_schema": map[string]any{
				"name":   "investment_research_result",
				"schema": openai.BuildSchema[models.InvestmentResearchResult](),
			},
		},
	}

	// 4. Create batch request
	batchRequest := &models.BatchRequest{
		CustomID: customID,
		Method:   "POST",
		URL:      "/v1/chat/completions",
		Body:     requestBody,
	}

	return batchRequest, nil
}

// generateRegionalPrompt generates the same regional prompt as the regular engine
func (e *BatchInvestmentResearchEngine) generateRegionalPrompt(
	ctx context.Context,
	sharedBag bag.SharedBag,
) (string, error) {

	// 1. Extract data from shared bag
	portfolio := sharedBag.MustGet(bag.KPortfolioNormalizedForAI).(*models.NormalizedPortfolio)

	// 2. Extract investment profile from shared bag
	investmentProfile := sharedBag.MustGet(bag.KProfile).(*models.InvestmentProfile)

	// 3. Extract regional config from shared bag
	regionalConfig := sharedBag.MustGet(bag.KRegionalConfig).(*models.RegionalConfig)

	// Create localization config from investment profile
	localizationConfig := models.LocalizationConfig{
		Country:  investmentProfile.RegionalContext.Country,
		Language: investmentProfile.RegionalContext.Language,
		Currency: investmentProfile.RegionalContext.Currency,
		Timezone: investmentProfile.RegionalContext.Timezone,
		Region:   investmentProfile.RegionalContext.Region,
		City:     investmentProfile.RegionalContext.City,
	}

	// Generate regional prompt using the same method as regular engine
	prompt, err := e.regionalPromptManager.GenerateRegionalPrompt(
		ctx,
		models.AnalysisInvestmentResearch,
		sharedBag,
		prompt.PromptData{
			EngineVersion:     "v2.0",
			AnalysisType:      models.AnalysisInvestmentResearch,
			Timestamp:         time.Now(),
			RegionalConfig:    regionalConfig,
			Portfolio:         portfolio,
			InvestmentProfile: investmentProfile,
			Localization:      localizationConfig,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate regional prompt: %w", err)
	}

	return prompt, nil
}

// getToolConstraints returns tool constraints based on research depth (same as regular engine)
func (e *BatchInvestmentResearchEngine) getToolConstraints() models.BaseToolConstraints {
	baseConstraints := models.BaseToolConstraints{
		Tools: tools.GetToolsDef(),
		RequiredTools: []bag.Key{
			bag.WebSearch,
			bag.FMP,     // Market data for research
			bag.NewsApi, // News and market intelligence
		},
		PreferredTools: []bag.Key{
			bag.Fred,                // Economic context
			bag.YFinanceStockData,   // Additional market data
			bag.FMPAnalystEstimates, // Analyst insights
		},
	}

	// Adjust tool usage based on research depth (same logic as regular engine)
	switch e.researchDepth {
	case "comprehensive":
		baseConstraints.MaxCallsPerTool = map[bag.Key]int{
			bag.WebSearch: 8, // Deep research
			bag.FMP:       4, // Comprehensive data
			bag.NewsApi:   2, // News context
		}
		baseConstraints.MinCallsPerTool = map[bag.Key]int{
			bag.WebSearch: 4, // Minimum quality
		}

	case "standard":
		baseConstraints.MaxCallsPerTool = map[bag.Key]int{
			bag.WebSearch: 5, // Balanced research
			bag.FMP:       2,
			bag.NewsApi:   1,
		}
		baseConstraints.MinCallsPerTool = map[bag.Key]int{
			bag.WebSearch: 3,
		}

	case "basic":
		baseConstraints.MaxCallsPerTool = map[bag.Key]int{
			bag.WebSearch: 3, // Light research
			bag.FMP:       1,
			bag.NewsApi:   1,
		}
		baseConstraints.MinCallsPerTool = map[bag.Key]int{
			bag.WebSearch: 2,
		}
	}

	return baseConstraints
}
