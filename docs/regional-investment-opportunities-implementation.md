# Regional Investment Research Engine Implementation Plan

## Executive Summary

This document outlines the implementation plan for a **Regional Investment Research Engine** that leverages OpenAI's web search capabilities to provide localized, structured investment research. The engine will use region/language-driven data architecture to deliver personalized investment research based on user localization context and investment profile.

## Investment Profile Architecture

### Investment Profile vs User Profile

**Current UserProfile Analysis**:

Looking at the existing templates, we currently have `{{.UserProfile}}` with:

- Risk Tolerance
- Investment Goals
- Time Horizon
- Available Capital

**Investment Profile Question: Do We Need a Dedicated Investment Profile?**

**✅ YES - Investment Profile Should Be Separate and Specialized**:

1. **Specialized Investment Context**: Beyond basic user preferences
2. **Research Targeting**: Specific investment themes, asset class preferences, ESG criteria
3. **Market Sophistication**: Professional vs retail investor research depth
4. **Regional Investment Culture**: French preference for ESG, US growth focus, German stability focus

### Investment Profile Structure

```go
type InvestmentProfile struct {
    // Research Preferences
    ResearchDepth       string   `json:"research_depth"`        // "basic", "intermediate", "advanced"
    InvestmentStyle     string   `json:"investment_style"`      // "growth", "value", "balanced", "income"
    SophisticationLevel string   `json:"sophistication_level"`  // "retail", "affluent", "institutional"

    // Thematic Preferences
    PreferredThemes     []string `json:"preferred_themes"`      // ["ai", "clean_energy", "healthcare"]
    AvoidedSectors      []string `json:"avoided_sectors"`       // ["tobacco", "weapons", "fossil_fuels"]
    ESGCriteria         ESGPrefs `json:"esg_criteria"`

    // Asset Class Preferences
    AssetClassFocus     []string `json:"asset_class_focus"`     // ["equities", "bonds", "alternatives", "crypto"]
    GeographicFocus     []string `json:"geographic_focus"`      // ["domestic", "developed", "emerging"]

    // Research Behavior
    InformationSources  []string `json:"information_sources"`   // ["financial_news", "analyst_reports", "social_sentiment"]
    UpdateFrequency     string   `json:"update_frequency"`      // "daily", "weekly", "monthly", "quarterly"

    // Regional Context Integration
    RegionalPreferences RegionalInvestmentPrefs `json:"regional_preferences"`
}

type RegionalInvestmentPrefs struct {
    // French Example
    PEAOptimization     bool     `json:"pea_optimization,omitempty"`
    AssuranceVieUse     bool     `json:"assurance_vie_use,omitempty"`
    EuropeanFocus       bool     `json:"european_focus,omitempty"`

    // US Example
    TaxLossHarvesting   bool     `json:"tax_loss_harvesting,omitempty"`
    RetirementOptimized bool     `json:"retirement_optimized,omitempty"`
    ESGPriority         string   `json:"esg_priority,omitempty"` // "mandatory", "preferred", "neutral"
}
```

### Investment Research Profiles

**Research Profile Integration**:

Based on the various investment research types we've discussed, the `InvestmentProfile` drives different research profiles:

1. **Conservative Research Profile** (French Retiree):

   ```go
   profile := InvestmentProfile{
       ResearchDepth: "basic",
       InvestmentStyle: "income",
       SophisticationLevel: "retail",
       PreferredThemes: []string{"dividend_aristocrats", "european_stability"},
       RegionalPreferences: RegionalInvestmentPrefs{
           PEAOptimization: true,
           AssuranceVieUse: true,
           EuropeanFocus: true,
       },
   }
   ```

2. **Growth Research Profile** (US Tech Professional):

   ```go
   profile := InvestmentProfile{
       ResearchDepth: "advanced",
       InvestmentStyle: "growth",
       SophisticationLevel: "affluent",
       PreferredThemes: []string{"artificial_intelligence", "biotechnology", "clean_energy"},
       RegionalPreferences: RegionalInvestmentPrefs{
           TaxLossHarvesting: true,
           RetirementOptimized: true,
       },
   }
   ```

3. **Institutional Research Profile** (Family Office):
   ```go
   profile := InvestmentProfile{
       ResearchDepth: "comprehensive",
       InvestmentStyle: "balanced",
       SophisticationLevel: "institutional",
       AssetClassFocus: []string{"equities", "bonds", "alternatives", "private_equity"},
       GeographicFocus: []string{"global", "emerging_markets"},
   }
   ```

**Template Data Injection**:

```gotmpl
{{.InvestmentProfile.ResearchDepth}} depth analysis for {{.InvestmentProfile.SophisticationLevel}} investor
Focus Areas: {{range .InvestmentProfile.PreferredThemes}}{{.}}, {{end}}
{{if .InvestmentProfile.RegionalPreferences.PEAOptimization}}
Research PEA-eligible opportunities with European focus
{{end}}
```

## Localization Service Analysis

### Current Architecture Assessment

**✅ Strengths - No Dedicated Localization Service Needed**:

1. **Centralized Configuration**: Robust `config.LocalizationConfig` with ISO standards
2. **Automatic Propagation**: `populateComputedFields()` spreads localization across all components
3. **Standard Compliance**: ISO 3166-1 (country), ISO 639-1 (language), ISO 4217 (currency), IANA (timezone)
4. **Existing Integration**: Already connected to jurisdiction, web search, tools, and templates

**🔄 Current Localization Flow**:

```
LocalizationConfig → populateComputedFields() → {
  ├── Jurisdiction.Country
  ├── LLM.Locale
  ├── WebSearch.UserLocation
  ├── NewsAPI.Locale
  ├── FRED.Country
  └── Template.Localization
}
```

### Localization Service Decision: **NOT REQUIRED**

**Rationale**:

- **Configuration-Driven**: Current architecture is already configuration-driven with automatic propagation
- **ISO Standards**: All localization uses international standards (no custom translation needed)
- **Template Integration**: Templates already access localization context seamlessly
- **Tool Integration**: External APIs already receive localized parameters

**What We Have vs. What We Need**:

- ✅ **Geographic Context**: Country, region, city for market relevance
- ✅ **Language Context**: Language code for AI prompts and web search
- ✅ **Currency Context**: Base currency for financial calculations
- ✅ **Regulatory Context**: Jurisdiction rules and compliance
- ❌ **Regional Market Data**: Missing region-specific investment databases
- ❌ **Localized Templates**: Missing hierarchical template structure

## Implementation Architecture

### Phase 1: Regional Data Architecture

#### 1.1 Regional Configuration Structure

```yaml
# config/regions/
├── US/
│   ├── en/
│   │   ├── market_context.yaml
│   │   ├── investment_focus.yaml
│   │   └── web_search_queries.yaml
│   └── es/
│       └── ... (Spanish variants)
├── FR/
│   ├── fr/
│   │   ├── market_context.yaml
│   │   ├── pea_eligible_instruments.yaml
│   │   └── web_search_queries.yaml
│   └── en/
│       └── ... (English variants)
└── DE/
├── de/
└── en/
```

#### 1.2 Regional Market Context Data

**Example: `config/regions/FR/fr/market_context.yaml`**

```yaml
region_config:
  country: 'FR'
  language: 'fr'
  currency: 'EUR'
  regulatory_framework:
    primary_regulator: 'AMF (Autorité des marchés financiers)'
    investment_vehicles:
      - name: "PEA (Plan d'Épargne en Actions)"
        max_amount: '150000 EUR'
        eligible_instruments: ['european_stocks', 'ucits_etfs']
      - name: 'Assurance-vie'
        eligible_instruments: ['ucits_etfs', 'fonds_euros', 'bonds']

  market_structure:
    primary_exchanges:
      - 'Euronext Paris'
      - 'Euronext Brussels'
      - 'Euronext Amsterdam'
    major_indices:
      - 'CAC 40'
      - 'SBF 120'
      - 'CAC Mid 60'

  preferred_instruments:
    equities:
      domicile_preference: ['FR', 'IE', 'LU', 'DE']
      examples: ['OR.PA', 'MC.PA', 'SAN.PA']
    etfs:
      domicile_requirement: ['IE', 'LU', 'DE', 'FR']
      ucits_only: true
      examples: ['CW8.PA', 'EWLD.PA', 'PAEEM.PA']
    bonds:
      types: ['government', 'corporate_investment_grade']
      currency_preference: 'EUR'

  tax_considerations:
    capital_gains_tax: '30%'
    pea_tax_advantage: '0% after 5 years'
    assurance_vie_tax: 'Progressive based on duration'
```

#### 1.3 Regional Web Search Queries

**Example: `config/regions/FR/fr/web_search_queries.yaml`**

```yaml
search_strategies:
  ai_opportunities:
    - 'meilleures actions intelligence artificielle Europe 2025'
    - 'ETF IA européen éligible PEA'
    - 'investissement technologie France croissance'

  climate_investing:
    - 'ETF transition énergétique Europe UCITS'
    - 'actions énergie renouvelable françaises 2025'
    - 'investissement durable PEA éligible'

  dividend_income:
    - 'actions dividend aristocrats européennes'
    - 'ETF dividendes Europe éligible assurance vie'
    - 'REIT européens rendement élevé'

  emerging_markets:
    - 'ETF marchés émergents UCITS domiciliation européenne'
    - 'opportunités Asie investissement Europe'

market_terminology:
  etf: 'ETF (Exchange Traded Fund)'
  pea: "Plan d'Épargne en Actions"
  ucits: 'OPCVM harmonisé européen'
  assurance_vie: "Contrat d'assurance-vie"
  fonds_euros: 'Fonds en euros'
```

### Phase 2: Investment Research Engine

#### 2.1 Engine Architecture (internal/engine/research.go)

```go
// Regional Investment Research Engine
type InvestmentResearchEngine struct {
    promptManager    RegionalPromptManager
    regionalLoader   RegionalConfigLoader
    fs              fs.FS                  // Filesystem abstraction
    constraints     models.ToolConstraints
}

// Tool Constraints with Web Search as Required Tool
constraints: models.ToolConstraints{
    RequiredTools: []keys.Key{
        keys.WebSearch, // OpenAI web search for real-time research
    },
    PreferredTools: []keys.Key{
        keys.FMP,       // Market data validation
        keys.FRED,      // Economic context
        keys.NewsAPI,   // News correlation
    },
    MaxCallsPerTool: map[keys.Key]int{
        keys.WebSearch: 8,  // Comprehensive research depth
        keys.FMP:       4,  // Supporting data
        keys.NewsAPI:   2,  // News context
    },
    MinCallsPerTool: map[keys.Key]int{
        keys.WebSearch: 3,  // Minimum research quality
    },
}
```

#### 2.2 Regional Template Manager (internal/prompt/regional_prompt_manager.go)

```go
type RegionalPromptManager struct {
    baseTemplates   map[string]*template.Template
    regionalConfigs map[string]RegionalConfig
    localization    config.LocalizationConfig
    fs              fs.FS  // Filesystem abstraction for template/config loading
    configDir       string // Base configuration directory
}

// Template Resolution Strategy using pkg/fs:
// 1. config/regions/{Country}/{Language}/{template}.yaml
// 2. config/regions/{Country}/en/{template}.yaml
// 3. config/regions/default/{template}.yaml

func NewRegionalPromptManager(
    localization config.LocalizationConfig,
    fsys fs.FS,
    configDir string,
) *RegionalPromptManager {
    return &RegionalPromptManager{
        baseTemplates:   make(map[string]*template.Template),
        regionalConfigs: make(map[string]RegionalConfig),
        localization:    localization,
        fs:              fsys,
        configDir:       configDir,
    }
}
```

#### 2.3 Layered Template Architecture (Maintainable & Regionalized)

**Architecture Principle**: Composition over Duplication

```
internal/prompt/templates/investment_research/
├── base/
│   ├── research.tmpl                # Core research logic
│   └── components/
│       ├── context.tmpl             # User/investment profile context
│       ├── portfolio_analysis.tmpl  # Portfolio gap analysis
│       ├── research_framework.tmpl  # Web search instructions
│       └── output_format.tmpl       # Structured response format
└── regional/
    ├── overlays/
    │   ├── FR_overlay.tmpl          # French-specific additions only
    │   ├── US_overlay.tmpl          # US-specific additions only
    │   └── CA_overlay.tmpl          # Canadian-specific additions only
    └── localization/
        ├── FR_fr.yaml               # French language strings/market data
        ├── US_en.yaml               # US English strings/market data
        └── CA_en.yaml               # Canadian English strings/market data
```

**Maintainability Benefits**:

✅ **Single Core Template**: One `research.tmpl` with all discovery logic
✅ **Minimal Regional Files**: Only country-specific additions (5-10 lines)
✅ **Component Reuse**: Shared components across all regions
✅ **Configuration-Driven**: Regional data in YAML, not template code
✅ **Generic Asset Discovery**: AI finds assets via web search, not hardcoded lists

**Template Composition Example**:

```gotmpl
<!-- Base Template: research.tmpl -->
{{template "context" .}}
{{template "portfolio_analysis" .}}

**Investment Research Framework:**
Use web search to identify investment research that:
1. Addresses portfolio gaps and diversification needs
2. Aligns with {{.InvestmentProfile.InvestmentStyle}} investment style
3. Matches {{.InvestmentProfile.ResearchDepth}} research depth
4. Considers current market environment and trends

{{/* Regional customization injection point */}}
{{if .RegionalOverlay}}
{{template "regional_overlay" .}}
{{end}}

{{template "research_framework" .}}
{{template "output_format" .}}
```

**Regional Overlay Example** (FR_overlay.tmpl - ~10 lines):

```gotmpl
**French Investment Context:**
- Tax-advantaged accounts: {{.Localization.Strings.pea_description}}
- Regulatory focus: {{.Localization.MarketContext.regulatory_focus}}
- Cultural preferences: {{.Localization.MarketContext.investment_culture}}

**Research Focus Areas:**
{{range .Localization.MarketContext.preferred_themes}}
- {{.}}
{{end}}
```

#### 2.4 Structured Output Models (pkg/models)

```go
type InvestmentResearchResult struct {
    ExecutiveSummary   ExecutiveSummary      `json:"executive_summary"`
    RegionalContext    RegionalContext       `json:"regional_context"`
    ResearchResults    []InvestmentResearch `json:"investment_research"`
    MarketAnalysis     MarketAnalysis        `json:"market_analysis"`
    RiskConsiderations []RiskFactor          `json:"risk_considerations"`
    ActionPlan         ActionPlan            `json:"action_plan"`
    Sources            []SearchSource        `json:"sources"`
}

type InvestmentResearch struct {
    Ticker              string            `json:"ticker"`
    Name                string            `json:"name"`
    AssetClass          string            `json:"asset_class"`
    Region              string            `json:"region"`
    MarketCap           *string           `json:"market_cap,omitempty"`
    Sector              string            `json:"sector"`
    ThematicExposure    []string          `json:"thematic_exposure"`
    InvestmentThesis    string            `json:"investment_thesis"`
    RegionalEligibility RegionalEligibility `json:"regional_eligibility"`
    RiskProfile         RiskProfile       `json:"risk_profile"`
    ValuationMetrics    ValuationMetrics  `json:"valuation_metrics"`
    RecommendedWeight   AllocationRange   `json:"recommended_weight"`
    TimeHorizon         string            `json:"time_horizon"`
}

type RegionalEligibility struct {
    PEAEligible        bool     `json:"pea_eligible,omitempty"`
    AssuranceVieEligible bool   `json:"assurance_vie_eligible,omitempty"`
    TaxImplications    []string `json:"tax_implications"`
    RegulatoryNotes    []string `json:"regulatory_notes"`
}

type MarketAnalysis struct {
    MarketTrends       []MarketTrend     `json:"market_trends"`
    SectorOutlook      SectorOutlook     `json:"sector_outlook"`
    GeographicInsights GeographicInsights `json:"geographic_insights"`
    TimingConsiderations []TimingFactor  `json:"timing_considerations"`
}
```

### Phase 3: Implementation Strategy

#### 3.1 Template Hierarchy

```
internal/prompt/templates/
├── investment_research/
│   ├── base/
│   │   ├── research.tmpl          # Base research template
│   │   ├── thematic.tmpl          # Base thematic analysis
│   │   └── allocation.tmpl        # Base allocation template
│   └── regional/
│       ├── US/
│       │   ├── en/
│       │   │   ├── market_outlook.tmpl
│       │   │   ├── sector_focus.tmpl
│       │   │   └── tax_optimization.tmpl
│       │   └── es/
│       └── FR/
│           ├── fr/
│           │   ├── pea_research.tmpl
│           │   ├── assurance_vie_strategy.tmpl
│           │   └── european_focus.tmpl
│           └── en/
```

#### 3.2 Regional Configuration Loading (internal/config/regional_config_loader.go)

```go
type RegionalConfigLoader struct {
    fs           fs.FS                    // Filesystem abstraction
    configDir    string                   // Base config directory
    localization config.LocalizationConfig
    cachedConfigs map[string]RegionalConfig

    // Template resolution paths using fs.Join for cross-platform compatibility
    resolutionPaths []string
}

func NewRegionalConfigLoader(cfg *config.Config, fsys fs.FS) *RegionalConfigLoader {
    return &RegionalConfigLoader{
        fs:            fsys,
        configDir:     cfg.DataDir, // Or dedicated config dir
        localization:  cfg.Localization,
        cachedConfigs: make(map[string]RegionalConfig),
    }
}

func (rcl *RegionalConfigLoader) LoadRegionalConfig(
    country, language string,
) (RegionalConfig, error) {
    // Hierarchical resolution using pkg/fs for testability:
    paths := []string{
        fs.Join(rcl.configDir, "regions", country, language),        // Primary
        fs.Join(rcl.configDir, "regions", country, "en"),           // Fallback
        fs.Join(rcl.configDir, "regions", "default"),               // Default
    }

    for _, path := range paths {
        if config, err := rcl.loadFromPath(path); err == nil {
            return config, nil
        }
    }

    return RegionalConfig{}, fmt.Errorf("no regional config found for %s/%s", country, language)
}

func (rcl *RegionalConfigLoader) loadFromPath(configPath string) (RegionalConfig, error) {
    // Use fs.FS for reading config files (enables testing with in-memory fs)
    marketContextPath := fs.Join(configPath, "market_context.yaml")
    searchQueriesPath := fs.Join(configPath, "web_search_queries.yaml")

    marketData, err := rcl.fs.ReadFile(marketContextPath)
    if err != nil {
        return RegionalConfig{}, fmt.Errorf("failed to read market context: %w", err)
    }

    searchData, err := rcl.fs.ReadFile(searchQueriesPath)
    if err != nil {
        return RegionalConfig{}, fmt.Errorf("failed to read search queries: %w", err)
    }

    // Parse YAML and merge configs
    var config RegionalConfig
    // ... YAML unmarshaling logic

    return config, nil
}
```

**Key Advantages of Using `pkg/fs`**:

- **Testability**: Easy to inject in-memory filesystem for unit tests
- **Swappable Backends**: Future support for remote config storage without code changes
- **Atomic Operations**: Reliable file operations with proper error handling
- **Cross-Platform**: `fs.Join` handles path separators correctly
- **Centralized Persistence**: Consistent with other file operations in the system

### Filesystem Abstraction Benefits

**Why Use `pkg/fs` for Regional Configuration Management?**

1. **Testing Excellence**:

   ```go
   // Easy unit testing with in-memory filesystem
   func TestRegionalConfigLoader(t *testing.T) {
       testFS := NewInMemoryFS()
       testFS.WriteFile("config/regions/FR/fr/market_context.yaml", frenchConfig, 0644)

       loader := NewRegionalConfigLoader(cfg, testFS)
       config, err := loader.LoadRegionalConfig("FR", "fr")
       assert.NoError(t, err)
   }
   ```

2. **Production Flexibility**:

   - **Local Files**: Use `fs.OS{}` for standard file operations
   - **Cloud Storage**: Future `CloudFS` implementation for S3/GCS
   - **Embedded Resources**: `EmbedFS` for bundled regional configs
   - **Encryption**: `EncryptedFS` wrapper for sensitive configurations

3. **Consistent Architecture**:

   - Aligns with Mosychlos' existing `pkg/fs` usage for reports and exports
   - Same patterns as other file-based components
   - Centralized persistence layer abstraction

4. **Error Handling**:
   ```go
   // Graceful fallback through configuration hierarchy
   for _, path := range hierarchyPaths {
       if data, err := rcl.fs.ReadFile(path); err == nil {
           return parseConfig(data)
       }
   }
   ```

#### 3.3 Engine Integration

```go
// Engine Registration in internal/engine/orchestrator.go
func NewInvestmentResearchEngine(
    cfg *config.Config,
    fsys fs.FS, // Filesystem abstraction for testability
) *InvestmentResearchEngine {

    // Create regional prompt manager with fs abstraction
    promptManager := NewRegionalPromptManager(
        cfg.Localization,
        fsys,
        cfg.DataDir, // Base directory for regional configs
    )

    // Create regional config loader
    regionalLoader := NewRegionalConfigLoader(cfg, fsys)

    return &InvestmentResearchEngine{
        promptManager:  promptManager,
        regionalLoader: regionalLoader,
        fs:            fsys, // Store for template loading
        constraints: models.ToolConstraints{
            RequiredTools: []keys.Key{keys.WebSearch},
            PreferredTools: []keys.Key{
                keys.FMP,
                keys.FRED,
                keys.NewsAPI,
            },
            MaxCallsPerTool: map[keys.Key]int{
                keys.WebSearch: 8,
                keys.FMP:       4,
                keys.NewsAPI:   2,
            },
            MinCallsPerTool: map[keys.Key]int{
                keys.WebSearch: 3,
            },
        },
    }
}

// Template loading using fs abstraction
func (rpm *RegionalPromptManager) LoadRegionalTemplate(
    templateType, country, language string,
) (*template.Template, error) {

    // Hierarchical template resolution using fs.Join
    templatePaths := []string{
        fs.Join(rpm.configDir, "templates", "investment_research", "regional", country, language, templateType+".tmpl"),
        fs.Join(rpm.configDir, "templates", "investment_research", "regional", country, "en", templateType+".tmpl"),
        fs.Join(rpm.configDir, "templates", "investment_research", "base", templateType+".tmpl"),
    }

    for _, path := range templatePaths {
        if data, err := rpm.fs.ReadFile(path); err == nil {
            return template.New(templateType).Parse(string(data))
        }
    }

    return nil, fmt.Errorf("template %s not found for %s/%s", templateType, country, language)
}
```

### Phase 4: Regional Data Population

#### 4.1 Priority Regional Configurations

1. **US (English)**: S&P 500, NASDAQ focus, 401k/IRA considerations
2. **France (French)**: PEA/Assurance-vie, UCITS ETFs, European markets
3. **Germany (German)**: DAX focus, German tax optimization
4. **UK (English)**: FTSE focus, ISA eligibility, post-Brexit considerations

#### 4.2 Market Data Integration

```yaml
# Integration with existing tools for validation
market_data_sources:
  price_validation: 'FMP API'
  news_correlation: 'NewsAPI'
  economic_context: 'FRED API'
  fundamental_data: 'FMP Estimates API'
```

### Phase 5: Testing Strategy

#### 5.1 Filesystem Abstraction Benefits

**Using `pkg/fs` provides significant testing advantages**:

1. **In-Memory Testing**: Create ephemeral filesystems for fast unit tests
2. **Isolated Tests**: Each test gets clean filesystem state
3. **Cross-Platform**: Consistent behavior across operating systems
4. **Swappable Backends**: Easy to test different storage scenarios
5. **Atomic Operations**: Reliable file operations with proper error handling

```go
// Example test filesystem setup
type TestFS struct {
    files map[string][]byte
    dirs  map[string]bool
}

func NewTestFS() *TestFS {
    return &TestFS{
        files: make(map[string][]byte),
        dirs:  make(map[string]bool),
    }
}

func (tfs *TestFS) ReadFile(path string) ([]byte, error) {
    if data, exists := tfs.files[path]; exists {
        return data, nil
    }
    return nil, fmt.Errorf("file not found: %s", path)
}

func (tfs *TestFS) WriteFile(path string, data []byte, perm fs.FileMode) error {
    tfs.files[path] = data
    return nil
}
```

#### 5.1 Regional Template Tests

```go
func TestRegionalTemplateResolution(t *testing.T) {
    // Create in-memory filesystem for testing using pkg/fs abstraction
    memFS := NewMemoryFS() // Hypothetical in-memory implementation

    // Setup test directory structure
    testConfigs := map[string]string{
        "config/regions/FR/fr/market_context.yaml":     frenchMarketConfig,
        "config/regions/FR/en/market_context.yaml":     frenchEnglishFallback,
        "config/regions/default/market_context.yaml":   defaultMarketConfig,
        "config/templates/investment_research/base/research.tmpl": baseTemplate,
    }

    for path, content := range testConfigs {
        memFS.WriteFile(path, []byte(content), 0644)
    }

    // Test cases using fs abstraction
    cases := []struct {
        name     string
        country  string
        language string
        expected string
        fs       fs.FS
    }{
        {
            name:     "french_primary",
            country:  "FR",
            language: "fr",
            expected: "config/regions/FR/fr/",
            fs:       memFS,
        },
        {
            name:     "french_fallback",
            country:  "FR",
            language: "de",
            expected: "config/regions/FR/en/",
            fs:       memFS,
        },
        {
            name:     "default_fallback",
            country:  "XX",
            language: "xx",
            expected: "config/regions/default/",
            fs:       memFS,
        },
    }

    for _, c := range cases {
        t.Run(c.name, func(t *testing.T) {
            loader := NewRegionalConfigLoader(testConfig, c.fs)

            config, err := loader.LoadRegionalConfig(c.country, c.language)
            assert.NoError(t, err)
            assert.NotEmpty(t, config.MarketContext)
        })
    }
}

func TestRegionalPromptManagerWithFS(t *testing.T) {
    memFS := NewMemoryFS()

    // Setup template hierarchy
    templates := map[string]string{
        "templates/investment_research/regional/FR/fr/pea_research.tmpl": frenchPEATemplate,
        "templates/investment_research/base/research.tmpl": baseResearchTemplate,
    }

    for path, content := range templates {
        memFS.WriteFile(path, []byte(content), 0644)
    }

    localization := config.LocalizationConfig{
        Country:  "FR",
        Language: "fr",
        Currency: "EUR",
    }

    manager := NewRegionalPromptManager(localization, memFS, "")

    // Test template resolution
    tmpl, err := manager.LoadRegionalTemplate("pea_research", "FR", "fr")
    assert.NoError(t, err)
    assert.NotNil(t, tmpl)
}
```

#### 5.2 Structured Output Validation

```go
func TestInvestmentOpportunityStructuredOutput(t *testing.T) {
    // Test JSON schema generation
    // Test structured response parsing
    // Test regional data integration
}
```

### Phase 6: Configuration Examples

#### 6.1 Updated config.default.yaml

```yaml
# Investment Research Engine
engines:
  investment_research:
    enabled: true
    research_depth: 'comprehensive' # minimal, standard, comprehensive
    web_search:
      enabled: true
      max_searches_per_analysis: 8
      geographic_targeting: true
      language_preference: 'auto' # auto, force_english, force_local

    regional_data:
      config_dir: 'config/regions' # Base directory for regional configs
      cache_duration: '24h'
      fallback_strategy: 'country_then_default'
      filesystem_backend: 'os' # os, memory, cloud (for different fs.FS implementations)

    structured_output:
      enabled: true
      format: 'detailed' # minimal, standard, detailed
      include_sources: true
      include_citations: true

    focus_areas:
      - 'thematic_opportunities'
      - 'sector_rotation'
      - 'geographic_diversification'
      - 'tax_optimization'
      - 'regulatory_compliance'
```

## Architecture Summary & Key Decisions

### **Addressing Your Key Questions**:

#### **1. "Templates need to be fine-tuned to find more asset types and be less specific"**

**✅ SOLUTION**: **AI-Guided Discovery Architecture**

- **Generic Research Framework**: Templates provide research methodology, not specific assets
- **Web Search Driven**: AI discovers current market opportunities via OpenAI web search
- **Broad Asset Coverage**: No limitations to stocks/bonds - includes alternatives, crypto, commodities, real estate
- **Market Adaptive**: Research adapts to current market conditions and emerging themes

#### **2. "Are we diverging from our documentation previously written?"**

**✅ ALIGNED**: **Template Composition Approach**

- **Maintains Existing Architecture**: Extends current `internal/prompt/templates/` system
- **Tool-Driven Philosophy**: Templates guide AI to use research tools effectively
- **Engine Integration**: Works with existing engine orchestration and structured output
- **pkg/fs Integration**: Leverages existing filesystem abstraction patterns

#### **3. "What architecture do you suggest?"**

**✅ LAYERED TEMPLATE SYSTEM**: **Composition over Duplication**

```
Architecture: Base Template + Regional Overlays + Configuration Data
├── Base Template (research.tmpl): Core research logic - ONE FILE
├── Regional Overlays: Country-specific additions - 5-10 lines each
├── Component Templates: Reusable sections across regions
└── Configuration Data: Market context, strings, preferences in YAML
```

#### **4. "We need to keep everything maintainable and regionalized"**

**✅ MAINTAINABLE**:

- **Single Source of Truth**: Core research logic in one template
- **Minimal Regional Files**: Only country-specific additions (~10 lines each)
- **Configuration-Driven**: Regional differences in YAML files, not template code
- **Component Reuse**: Shared template components across all regions

**✅ REGIONALIZED**:

- **Cultural Adaptation**: Language, investment culture, risk preferences via overlays
- **Regulatory Context**: Tax implications, account types, compliance via configuration
- **Market Access**: Available investment vehicles and platforms via localization data

#### **5. "Do we need an investment profile?"**

**✅ YES**: **Dedicated Investment Profile Required**

**Why Investment Profile ≠ User Profile**:

| UserProfile        | InvestmentProfile                      |
| ------------------ | -------------------------------------- |
| Basic demographics | Specialized investment context         |
| Risk tolerance     | Research depth & sophistication        |
| Time horizon       | Investment style & themes              |
| General goals      | Asset class preferences & ESG criteria |

**Investment Profile Benefits**:

- **Research Targeting**: Drives different research profiles (conservative, growth, institutional)
- **Regional Integration**: French ESG focus, US growth preference, German stability focus
- **Template Data**: Rich context for AI research direction
- **Maintainable Personalization**: Configuration-driven research customization

### **Final Architecture Decision**:

**Layered Template System** with **Investment Profile-Driven Research** using **AI-Guided Discovery**:

1. **Core Research Template**: Generic discovery methodology
2. **Investment Profile**: Drives research focus and sophistication
3. **Regional Overlays**: Minimal country-specific additions
4. **Configuration Data**: Market context and localization strings
5. **Web Search Tools**: AI discovers current opportunities dynamically

This approach is **maintainable** (minimal duplication), **regionalized** (cultural adaptation), **flexible** (AI-driven discovery), and **aligned** with your existing architecture.

## Implementation Timeline

### Week 1-2: Foundation

- [ ] Create regional config directory structure
- [ ] Implement RegionalConfigLoader
- [ ] Create base structured output models
- [ ] Set up template hierarchy

### Week 3-4: Core Engine

- [ ] Implement InvestmentOpportunitiesEngine
- [ ] Create RegionalPromptManager
- [ ] Integrate with existing tool constraints system
- [ ] Add structured output schema generation

### Week 5-6: Regional Data

- [ ] Populate US market configuration
- [ ] Populate France/EU market configuration
- [ ] Create region-specific web search queries
- [ ] Implement template resolution logic

### Week 7-8: Integration & Testing

- [ ] Integrate with engine orchestrator
- [ ] Add comprehensive test coverage
- [ ] Performance optimization
- [ ] Documentation completion

## Success Metrics

### Technical Metrics

- **Template Resolution**: < 10ms per regional config lookup
- **Structured Output**: 100% schema compliance
- **Web Search Integration**: 90%+ successful query execution
- **Regional Coverage**: 95%+ coverage for supported regions

### Business Metrics

- **Investment Relevance**: Regional compliance > 95%
- **Research Quality**: Structured recommendations with citations
- **Localization Accuracy**: Country-specific market focus
- **User Experience**: Professional-grade investment research output

## Risk Mitigation

### Technical Risks

- **Web Search Rate Limits**: Implement exponential backoff and caching
- **Regional Config Complexity**: Start with 2-3 regions, expand gradually
- **Template Maintenance**: Automated testing for all regional variants

### Business Risks

- **Regulatory Compliance**: Legal review of regional recommendations
- **Market Data Accuracy**: Cross-validation with multiple sources
- **Investment Suitability**: Clear disclaimers and risk warnings

## Conclusion

This implementation plan leverages Mosychlos' existing **robust localization architecture** without requiring a dedicated localization service. The **configuration-driven, hierarchical approach** ensures scalability while maintaining the system's **data-driven, professional-grade** characteristics.

Key advantages:

- **Builds on existing strengths**: Centralized localization config and tool constraints
- **Maintains architectural consistency**: Engine patterns and SharedBag integration
- **Provides professional output**: Structured results with proper citations
- **Scales regionally**: Hierarchical template and configuration system
- **Ensures compliance**: Regional regulatory awareness built-in

The result will be an **institutional-grade investment research engine** that provides localized, structured, and actionable investment opportunities tailored to each user's regional context and regulatory requirements.
