# Regional Investment Research Engine Implementation

## Executive Summary

This document outlines the implementation of a new **Regional Investment Research Engine** that leverages OpenAI's web search capabilities to provide localized investment research. The engine integrates with the refactored prompt system and uses tool-driven analysis for comprehensive market research.

## Engine Architecture

### Overview

The Investment Research Engine follows Mosychlos' existing engine patterns while introducing regional context and web-based research capabilities. It integrates with the ProfileManager for regional investment preferences and uses structured models for professional-grade analysis.

```
┌─────────────────────────────────────────────────────┐
│         Investment Research Engine                  │
│                                                     │
│ ┌─────────────────┐    ┌─────────────────────────┐ │
│ │ ProfileManager  │────│   Regional Prompt       │ │
│ │                 │    │   Manager               │ │
│ │ • Investment    │    │                         │ │
│ │   Profile       │    │ • Base Templates        │ │
│ │ • Regional      │    │ • Regional Overlays     │ │
│ │   Preferences   │    │ • Localized Content     │ │
│ │ • SharedBag     │    │                         │ │
│ │   Integration   │    │                         │ │
│ └─────────────────┘    └─────────────────────────┘ │
│                                   │                 │
│ ┌─────────────────────────────────────────────────┐ │
│ │            Tool Orchestration                   │ │
│ │                                                 │ │
│ │ • OpenAI Web Search (with regional context)    │ │
│ │ • FMP Market Data (regional markets)           │ │
│ │ • NewsAPI Context (local sources)              │ │
│ │ • FRED Economic Data (regional indicators)     │ │
│ └─────────────────────────────────────────────────┘ │
│                                   │                 │
│ ┌─────────────────────────────────────────────────┐ │
│ │     Structured Output (Regional Context)       │ │
│ │                                                 │ │
│ │ • InvestmentResearchResult with tax optimization│ │
│ │ • Regional market access information           │ │
│ │ • Localized investment recommendations         │ │
│ │ • Country-specific compliance considerations   │ │
│ └─────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────┘
```

## Engine Implementation

### Core Engine Structure

```go
// internal/engine/investment_research/engine.go
type InvestmentResearchEngine struct {
    name                 string
    regionalPromptManager *prompt.RegionalManager  // Existing RegionalManager
    profileManager       *profile.Manager         // ProfileManager integration
    toolConstraints      models.ToolConstraints

    // Analysis configuration
    researchDepth        string  // "basic", "standard", "comprehensive"
    maxWebSearches       int     // Based on research depth
    includeAlternatives  bool    // Include crypto, commodities, etc.
}

var _ models.Engine = &InvestmentResearchEngine{}

// Engine interface implementation
func (e *InvestmentResearchEngine) Name() string {
    return e.name
}

func (e *InvestmentResearchEngine) Execute(ctx context.Context, client models.AiClient, sharedBag bag.SharedBag) error {
    // Store sharedBag reference for use during execution
    e.sharedBag = sharedBag

    // Execute investment research
    return e.executeResearch(ctx, client, sharedBag)
}
```

```go
// executeResearch performs the investment research analysis
func (e *InvestmentResearchEngine) executeResearch(ctx context.Context, client models.AiClient, sharedBag bag.SharedBag) error {
    // 1. Get investment profile from SharedBag (loaded by ProfileManager)
    profileData, ok := sharedBag.Get(keys.KProfile)
    if !ok {
        return fmt.Errorf("investment profile not found in shared bag")
    }

    investmentProfile := profileData.(*models.InvestmentProfile)

    // 2. Extract portfolio from shared bag
    portfolioData, ok := sharedBag.Get(keys.Portfolio)
    if !ok {
        return fmt.Errorf("portfolio not found in shared bag")
    }

    portfolio := portfolioData.(*models.NormalizedPortfolio)

    // 3. Use regional preferences from investment profile
    localization := investmentProfile.RegionalPreferences.LocalizationConfig

    // 4. Generate regional research prompt using existing RegionalManager
    prompt, err := e.regionalPromptManager.GenerateRegionalPrompt(
        ctx,
        models.AnalysisInvestmentResearch,
        prompt.PromptData{
            Portfolio:         portfolio,
            InvestmentProfile: investmentProfile,
            Localization:      localization,
        },
        localization.Country,
        localization.Language,
    )
    if err != nil {
        return fmt.Errorf("failed to generate prompt: %w", err)
    }

    // 5. Execute AI analysis with tool constraints
    result, err := e.executeAnalysisWithTools(ctx, client, prompt)
    if err != nil {
        return fmt.Errorf("analysis execution failed: %w", err)
    }

    // 6. Store structured result in shared bag
    sharedBag.Set(keys.InvestmentResearchResult, result)

    return nil
}

    // 6. Store results in shared bag for potential chaining
    e.storeResultsInBag(result)

    return result, nil
}
```

### Tool Constraints Configuration

```go
// Tool constraints optimized for investment research
func (e *InvestmentResearchEngine) getToolConstraints(researchDepth string) models.ToolConstraints {
    baseConstraints := models.ToolConstraints{
        RequiredTools: []keys.Key{
            keys.WebSearch,  // OpenAI web search for research
        },
        PreferredTools: []keys.Key{
            keys.FMP,        // Market data validation
            keys.FRED,       // Economic context
            keys.NewsAPI,    // News correlation
        },
    }

    // Adjust tool usage based on research depth
    switch researchDepth {
    case "comprehensive":
        baseConstraints.MaxCallsPerTool = map[keys.Key]int{
            keys.WebSearch: 8,  // Deep research
            keys.FMP:       4,  // Comprehensive data
            keys.NewsAPI:   2,  // News context
        }
        baseConstraints.MinCallsPerTool = map[keys.Key]int{
            keys.WebSearch: 4,  // Minimum quality
        }

    case "standard":
        baseConstraints.MaxCallsPerTool = map[keys.Key]int{
            keys.WebSearch: 5,  // Balanced research
            keys.FMP:       2,
            keys.NewsAPI:   1,
        }
        baseConstraints.MinCallsPerTool = map[keys.Key]int{
            keys.WebSearch: 3,
        }

    case "basic":
        baseConstraints.MaxCallsPerTool = map[keys.Key]int{
            keys.WebSearch: 3,  // Light research
            keys.FMP:       1,
        }
        baseConstraints.MinCallsPerTool = map[keys.Key]int{
            keys.WebSearch: 2,
        }
    }

    return baseConstraints
}
```

### AI Analysis Execution

```go
func (e *InvestmentResearchEngine) executeAnalysisWithTools(ctx context.Context, client models.AiClient, prompt string) (*models.InvestmentResearchResult, error) {
    // 1. Build structured output schema using ai.BuildSchema from internal/ai/schema.go
    schema := ai.BuildSchema[models.InvestmentResearchResult]()

    // 2. Execute analysis with tool access
    response, err := client.CreateChatCompletion(ctx, models.ChatRequest{
        Messages: []models.ChatMessage{
            {
                Role:    "user",
                Content: prompt,
            },
        },
        ToolConstraints: e.getToolConstraints(e.researchDepth),
        ResponseFormat: &models.ResponseFormat{
            Type:   "json_schema",
            Schema: schema,
        },
    })
    if err != nil {
        return nil, fmt.Errorf("AI analysis failed: %w", err)
    }

    // 3. Parse structured response
    var result models.InvestmentResearchResult
    if err := json.Unmarshal([]byte(response.Content), &result); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }

    // 4. Enhance with metadata
    result.Metadata = models.AnalysisMetadata{
        GeneratedAt:     time.Now().UTC(),
        EngineVersion:   "1.0.0",  // Engine version
        ResearchDepth:   e.researchDepth,
        DataSources:     []string{"web_search", "fmp", "newsapi"},
        RegionalContext: fmt.Sprintf("%s",
            result.RegionalAnalysis.Country),
    }

    return &result, nil
}
```

## Regional Configuration Architecture Decision

**Decision: Use Existing RegionalManager + ProfileManager (No RegionalConfigLoader needed)**

The Investment Research Engine uses a two-tier regional approach:

1. **ProfileManager** - Handles user investment profile and regional preferences
2. **RegionalManager** (existing in `internal/prompt/`) - Handles regional prompt templates and market data

This eliminates the need for a separate RegionalConfigLoader component.

### Architecture Justification

User Profile (ProfileManager)
↓ (Regional preferences: country, language, risk tolerance)
RegionalManager.GenerateRegionalPrompt()
↓ (Loads regional overlays + market context from config/templates/)
Investment Research Analysis
↓ (Regional context + market-specific data)
InvestmentResearchResult with localized recommendations

**Why this approach:**

- **ProfileManager**: User-specific data (risk tolerance, preferences, country)
- **RegionalManager**: Template-based regional prompts and market data
- **No RegionalConfigLoader**: Avoids architectural duplication

### Data Flow Implementation

### Data Flow Implementation

```
ProfileManager.LoadProfile(country, riskTolerance)
↓
InvestmentProfile stored in SharedBag (keys.KProfile)
↓
InvestmentResearchEngine.Execute() reads from SharedBag
↓
Regional context applied to research analysis
↓
InvestmentResearchResult with localized recommendations
```

## Structured Output with AI Schema

The engine uses `ai.BuildSchema[T]()` from `internal/ai/schema.go` for structured output:

```go
import (
    "internal/llm"  // Import for BuildSchema function
)

// Generate JSON Schema for structured AI response
schema := ai.BuildSchema[models.InvestmentResearchResult]()

// Schema automatically includes all struct fields and JSON tags
// Enables reliable structured output from AI models
```

**How BuildSchema works:**

- Uses Go reflection to inspect struct fields
- Generates OpenAI-compatible JSON Schema
- Handles nested structs, slices, and optional fields
- Ensures AI responses match exact struct format

### Regional Context Integration

```go
// Extract regional preferences from ProfileManager data
func (e *InvestmentResearchEngine) buildRegionalContext(profile *models.InvestmentProfile) models.RegionalContext {
    return models.RegionalContext{
        Country:       profile.RegionalPreferences.Country,
        Language:      profile.RegionalPreferences.Language,
        CurrencyFocus: profile.RegionalPreferences.Currency,

        // Tax optimizations based on regional preferences and asset eligibility
        TaxOptimizations: e.buildTaxOptimizations(profile.RegionalPreferences),

        // Local market access based on regional preferences
        LocalMarketAccess: e.buildMarketAccess(profile.RegionalPreferences),
    }
}

// Build tax optimization recommendations based on regional context
func (e *InvestmentResearchEngine) buildTaxOptimizations(prefs models.RegionalInvestmentPreferences) []models.TaxOptimization {
    var optimizations []models.TaxOptimization

    switch prefs.Country {
    case "FR":
        optimizations = append(optimizations, models.TaxOptimization{
            VehicleName: "PEA (Plan d'Épargne en Actions)",
            MaxAmount:   150000,
            Benefits:    []string{"Capital gains tax exemption after 5 years", "Dividend tax optimization"},
        })
        // Add Assurance-vie option...

    case "US":
        optimizations = append(optimizations, models.TaxOptimization{
            VehicleName: "401(k) / IRA",
            Benefits:    []string{"Tax-deferred growth", "Employer matching"},
        })

    case "CA":
        optimizations = append(optimizations, models.TaxOptimization{
            VehicleName: "TFSA (Tax-Free Savings Account)",
            MaxAmount:   88000, // 2025 limit
            Benefits:    []string{"Tax-free growth", "Tax-free withdrawals"},
        })
    }

    return optimizations
}
```

### Investment Profile Usage

The ProfileManager provides comprehensive investment context:

- **Investment Style & Depth**: Guides research focus and analysis depth
- **Regional Preferences**: Provides localization context (country, language, currency)
- **Asset Class Preferences**: Filters research to user's preferred investments
- **ESG Criteria**: Applies environmental, social, and governance filters
- **Tax Context**: Enables tax-optimized recommendations per region

## Structured Output Models

### Model Implementation Status

✅ **Implemented in `pkg/models/investment_research.go`:**

- `InvestmentResearchResult` (main container with excellent professional structure)
- `ExecutiveSummary` (market outlook and key takeaways)
- `RegionalContext` (country-specific context)
- `ResearchFinding` (detailed investment opportunities)
- `InvestmentInstrument` (with regional tax eligibility flags: PEA, ISA, TFSA, IRA)
- `InvestmentTheme` (thematic investment opportunities)
- `ActionableInsight` (professional implementation guidance with risk management)
- `SearchSource` (web search attribution)

❌ **Missing Supporting Models (Need Implementation):**

- `MarketAnalysis` - Market trend analysis and sector performance
- `RiskFactor` - Risk assessment with severity and mitigation
- `AnalysisMetadata` - Analysis timestamp, confidence, and version info
- `ExpectedReturn` - Return projections and scenarios
- `RiskProfile` - Risk assessment structure
- `RegulatoryInfo` - Regulatory framework details
- `TaxOptimization` - Tax strategy recommendations
- `MarketAccess` - Regional market accessibility information

### Professional Investment Features

Your implementation shows excellent understanding of professional investment management:

**Regional Tax Optimization:**

```go
// Regional tax-advantaged account eligibility
type InvestmentInstrument struct {
    PEAEligible  bool `json:"pea_eligible,omitempty"`  // France
    ISAEligible  bool `json:"isa_eligible,omitempty"`  // UK
    TFSAEligible bool `json:"tfsa_eligible,omitempty"` // Canada
    IRAEligible  bool `json:"ira_eligible,omitempty"`  // US
}
```

**Risk Management Integration:**

```go
// Professional risk management in actionable insights
type ActionableInsight struct {
    EntryStrategy    string   `json:"entry_strategy"`    // "dollar_cost_average", "lump_sum"
    StopLoss        *float64 `json:"stop_loss,omitempty"`
    ProfitTarget    *float64 `json:"profit_target,omitempty"`
    PositionSize     string   `json:"position_size"`     // "small", "medium", "large"
    MonitoringPoints []string `json:"monitoring_points"`
    ExitCriteria     []string `json:"exit_criteria"`
}
```

## Tool Constraints Configuration

### Investment Research Result

```go
// pkg/models/investment_research.go
type InvestmentResearchResult struct {
    ExecutiveSummary     ExecutiveSummary      `json:"executive_summary"`
    RegionalContext      RegionalContext       `json:"regional_context"`
    ResearchFindings     []ResearchFinding     `json:"research_findings"`
    MarketAnalysis       MarketAnalysis        `json:"market_analysis"`
    InvestmentThemes     []InvestmentTheme     `json:"investment_themes"`
    RiskConsiderations   []RiskFactor          `json:"risk_considerations"`
    ActionableInsights   []ActionableInsight   `json:"actionable_insights"`
    Sources              []SearchSource        `json:"sources"`
    Metadata             AnalysisMetadata      `json:"metadata"`
}

type ExecutiveSummary struct {
    KeyTakeaways         []string              `json:"key_takeaways"`
    MarketOutlook        string                `json:"market_outlook"`        // "bullish", "bearish", "neutral"
    RecommendedActions   []string              `json:"recommended_actions"`
    TimeHorizon          string                `json:"time_horizon"`          // "short_term", "medium_term", "long_term"
}

type RegionalContext struct {
    Country              string                `json:"country"`
    Language             string                `json:"language"`
    CurrencyFocus        string                `json:"currency_focus"`
    RegulatoryFramework  RegulatoryInfo        `json:"regulatory_framework"`
    TaxOptimizations     []TaxOptimization     `json:"tax_optimizations"`
    LocalMarketAccess    []MarketAccess        `json:"local_market_access"`
}

type ResearchFinding struct {
    Title                string                `json:"title"`
    AssetClass          string                `json:"asset_class"`           // "equities", "bonds", "alternatives", "crypto"
    GeographicFocus     string                `json:"geographic_focus"`      // "domestic", "developed", "emerging", "global"
    InvestmentTheme     string                `json:"investment_theme"`      // "ai", "clean_energy", "demographics", etc.

    // Investment details
    SpecificInstruments []InvestmentInstrument `json:"specific_instruments"`
    ExpectedReturn      ExpectedReturn         `json:"expected_return"`
    RiskProfile         RiskProfile            `json:"risk_profile"`
    TimeHorizon         string                 `json:"time_horizon"`

    // Research context
    MarketDrivers       []string               `json:"market_drivers"`
    CompetitivePosition string                 `json:"competitive_position"`
    ValuationMetrics    map[string]interface{} `json:"valuation_metrics"`

    // Regional relevance
    RegionalRelevance   string                 `json:"regional_relevance"`    // Why relevant for this region
    LocalAvailability   bool                   `json:"local_availability"`
    TaxImplications     []string               `json:"tax_implications"`
}

type InvestmentInstrument struct {
    Type                string                 `json:"type"`                  // "stock", "etf", "bond", "fund", "alternative"
    Ticker              string                 `json:"ticker,omitempty"`
    Name                string                 `json:"name"`
    Exchange            string                 `json:"exchange,omitempty"`
    ISIN                string                 `json:"isin,omitempty"`
    Currency            string                 `json:"currency"`

    // Metrics
    CurrentPrice        *float64               `json:"current_price,omitempty"`
    MarketCap           *int64                 `json:"market_cap,omitempty"`
    ExpenseRatio        *float64               `json:"expense_ratio,omitempty"`

    // Regional context
    PEAEligible         bool                   `json:"pea_eligible,omitempty"`          // France
    ISAEligible         bool                   `json:"isa_eligible,omitempty"`          // UK
    TFSAEligible        bool                   `json:"tfsa_eligible,omitempty"`         // Canada
    IRAEligible         bool                   `json:"ira_eligible,omitempty"`          // US

    AccessibilityNotes  []string               `json:"accessibility_notes"`
}

type InvestmentTheme struct {
    Name                string                 `json:"name"`
    Description         string                 `json:"description"`
    MarketSize          *int64                 `json:"market_size_usd,omitempty"`
    GrowthProjection    string                 `json:"growth_projection"`     // "high", "medium", "low"
    TimeHorizon         string                 `json:"time_horizon"`
    KeyDrivers          []string               `json:"key_drivers"`

    // Regional adaptation
    RegionalExposure    map[string]float64     `json:"regional_exposure"`     // % by region
    LocalChampions      []string               `json:"local_champions"`       // Regional leaders
    RegulatorySupport   bool                   `json:"regulatory_support"`

    // Implementation
    AccessMethods       []string               `json:"access_methods"`        // "direct_stocks", "sector_etfs", "thematic_funds"
    RecommendedAllocation string               `json:"recommended_allocation"` // "2-5%", "5-10%", etc.
}

type ActionableInsight struct {
    Priority            string                 `json:"priority"`              // "high", "medium", "low"
    Action              string                 `json:"action"`                // "buy", "sell", "hold", "research_further"
    Instrument          InvestmentInstrument   `json:"instrument"`
    TargetAllocation    string                 `json:"target_allocation"`     // "3-5%", "immediate", etc.
    Rationale           string                 `json:"rationale"`
    Timeline            string                 `json:"timeline"`              // "immediate", "next_quarter", "within_year"

    // Implementation details
    EntryStrategy       string                 `json:"entry_strategy"`        // "dollar_cost_average", "lump_sum", "wait_for_dip"
    StopLoss            *float64               `json:"stop_loss,omitempty"`
    ProfitTarget        *float64               `json:"profit_target,omitempty"`

    // Risk management
    PositionSize        string                 `json:"position_size"`         // "small", "medium", "large"
    MonitoringPoints    []string               `json:"monitoring_points"`
    ExitCriteria        []string               `json:"exit_criteria"`
}

type SearchSource struct {
    URL                 string                 `json:"url"`
    Title               string                 `json:"title"`
    SearchQuery         string                 `json:"search_query"`
    RelevanceScore      float64                `json:"relevance_score"`
    PublishedDate       *time.Time             `json:"published_date,omitempty"`
    Source              string                 `json:"source"`                // "financial_times", "reuters", etc.
}
```

### Model Reference

All models are implemented in `pkg/models/investment_research.go`. The main `InvestmentResearchResult` struct contains all the structured output fields with excellent professional investment features including regional tax optimization and risk management.

## Regional Configuration Simplified

## Regional Configuration Simplified

With ProfileManager integration, regional configuration is streamlined and eliminates the need for a separate RegionalConfigLoader.

### Regional Data Sources

1. **InvestmentProfile** (via ProfileManager) - User preferences and regional settings
2. **Prompt Templates** (via RegionalPromptManager) - Regional overlays and templates
3. **Tool Constraints** - Regional tool usage patterns

### ProfileManager Provides All Regional Context

The ProfileManager loads comprehensive regional information through the InvestmentProfile:

- **Regional Preferences**: Country, language, currency, timezone
- **Tax-Advantaged Accounts**: PEA (France), ISA (UK), TFSA (Canada), IRA (US) eligibility
- **ESG Criteria**: Environmental, social, governance preferences relevant to region
- **Asset Class Preferences**: Regional asset class focus
- **Compliance Rules**: Regional regulatory requirements

### No Separate RegionalConfigLoader Needed

```go
// OLD APPROACH (not needed):
type RegionalConfigLoader struct {
    fs        fs.FS
    configDir string
    cache     map[string]*RegionalConfig
}

// NEW APPROACH (via ProfileManager):
// Regional context comes from InvestmentProfile in SharedBag
profile := sharedBag.Get(keys.KProfile).(*models.InvestmentProfile)
regionalContext := models.RegionalContext{
    Country:       profile.RegionalPreferences.Country,
    Language:      profile.RegionalPreferences.Language,
    CurrencyFocus: profile.RegionalPreferences.Currency,
    // Tax optimizations built from regional preferences...
}
```

## Engine Factory Integration

### Simplified Factory

```go
// internal/engine/investment_research/factory.go
func NewInvestmentResearchEngine(
    cfg *config.Config,
    sharedBag bag.SharedBag,  // Fixed: not a pointer
    fsys fs.FS,
) (*InvestmentResearchEngine, error) {

    // Create regional prompt manager
    regionalPromptManager, err := prompt.NewRegionalManager(
        cfg.Localization,
        fsys,
        cfg.ConfigDir,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create regional prompt manager: %w", err)
    }

    // ProfileManager is created separately and used via SharedBag
    // No need for separate regional config loader

    researchDepth := "standard"
    if cfg.Engines.InvestmentResearch != nil {
        researchDepth = cfg.Engines.InvestmentResearch.ResearchDepth
    }

    return &InvestmentResearchEngine{
        regionalPromptManager: regionalPromptManager,
        sharedBag:            sharedBag,
        researchDepth:        researchDepth,
        toolConstraints:      buildToolConstraints(researchDepth),
    }, nil
}
}
```

### Orchestrator Registration

```go
// internal/engine/orchestrator.go - Add to engine registration
func (o *Orchestrator) registerEngines() error {
    // ... existing engine registrations

    // Register investment research engine
    if o.config.Engines.InvestmentResearch != nil && o.config.Engines.InvestmentResearch.Enabled {
        investmentResearchEngine, err := investment_research.NewInvestmentResearchEngine(
            o.config,
            o.sharedBag,
            o.fs,
        )
        if err != nil {
            return fmt.Errorf("failed to create investment research engine: %w", err)
        }

        o.engines[models.EngineTypeInvestmentResearch] = investmentResearchEngine

        slog.Info("Investment Research engine registered",
            "research_depth", o.config.Engines.InvestmentResearch.ResearchDepth,
            "web_search_enabled", o.config.Engines.InvestmentResearch.WebSearch.Enabled,
            "include_alternatives", o.config.Engines.InvestmentResearch.IncludeAlternatives,
        )
    }

    return nil
}
```

## Configuration

### Engine Configuration

```yaml
# config.default.yaml
engines:
  investment_research:
    enabled: true
    research_depth: 'comprehensive' # "basic", "standard", "comprehensive"
    include_alternatives: true # Include crypto, commodities, REITs

    web_search:
      enabled: true
      max_searches_per_analysis: 8
      geographic_targeting: true
      language_preference: 'auto' # "auto", "force_english", "force_local"
      preferred_sources: [] # Override regional defaults

    structured_output:
      enabled: true
      format: 'detailed' # "minimal", "standard", "detailed"
      include_sources: true
      include_citations: true

    focus_areas:
      - 'thematic_opportunities'
      - 'sector_rotation'
      - 'geographic_diversification'
      - 'tax_optimization'
      - 'regulatory_compliance'

    risk_management:
      include_risk_analysis: true
      position_sizing_guidance: true
      stop_loss_recommendations: true

    regional:
      auto_detect_from_localization: true
      fallback_region: 'US'
      fallback_language: 'en'
```

## CLI Integration

### Command Extension

```bash
# Add investment research to existing portfolio analyze command
mosychlos portfolio analyze investment_research --research-depth comprehensive --include-alternatives

# Regional specification (optional - auto-detected from localization)
mosychlos portfolio analyze investment_research --country FR --language fr

# Output format options
mosychlos portfolio analyze investment_research --pdf --json --markdown
```

## Implementation Status & Next Steps

### ✅ Completed Components

**Model Architecture:**

- Main `InvestmentResearchResult` structure with excellent professional features
- Regional tax optimization flags (PEA, ISA, TFSA, IRA) for localized investment guidance
- Professional investment metrics (market cap, expense ratios, current pricing)
- Actionable insights with comprehensive risk management (stop losses, position sizing)
- Regional context integration with ProfileManager architecture

**Architecture Integration:**

- ProfileManager integration patterns established (SharedBag data flow)
- Simplified regional configuration approach (eliminating duplicate RegionalConfigLoader)
- Tool constraint configuration for research depth optimization
- Structured output design for professional investment analysis

### 🔄 In Progress

**Missing Supporting Models (8 models):**

- `MarketAnalysis` - Market trend analysis and sector performance data
- `RiskFactor` - Risk assessment with severity classification and mitigation strategies
- `AnalysisMetadata` - Analysis timestamp, confidence scoring, and version tracking
- `ExpectedReturn` - Return projections with scenario analysis
- `RiskProfile` - Comprehensive risk assessment structure
- `RegulatoryInfo` - Detailed regulatory framework information
- `TaxOptimization` - Tax strategy recommendations with regional specifics
- `MarketAccess` - Regional market accessibility and trading information

### ⏳ Pending Implementation

**Engine Implementation:**

- Engine factory with ProfileManager integration
- Tool constraint optimization for regional research
- AI analysis execution with structured output parsing
- SharedBag result storage and chaining support

**Integration & Registration:**

- Orchestrator registration with proper configuration
- CLI command integration for investment research analysis
- Configuration file updates for engine parameters
- Testing and validation of ProfileManager data flow

### 📋 Implementation Checklist

**Phase 1: Complete Models**

- [ ] Implement 8 missing supporting model types in `investment_research.go`
- [ ] Add YAML tags for configuration file parsing
- [ ] Validate model relationships and data flow

**Phase 2: Engine Implementation**

- [ ] Create engine factory with ProfileManager injection
- [ ] Implement core engine execution logic
- [ ] Add tool constraint optimization
- [ ] Implement structured output parsing

**Phase 3: Integration**

- [ ] Register engine in orchestrator
- [ ] Add CLI command support
- [ ] Update configuration files
- [ ] Create comprehensive test suite

**Phase 4: Regional Enhancement**

- [ ] Validate tax optimization logic per region
- [ ] Test market access information accuracy
- [ ] Optimize web search queries for regional relevance
- [ ] Validate regulatory compliance features

This implementation provides a comprehensive foundation for regional investment research with professional-grade features. The ProfileManager integration ensures consistent regional context while the structured output models support institutional-level investment analysis.
