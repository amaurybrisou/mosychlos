# Prompt System Refactoring for Regional Investment Research

## Executive Summary

This document outlines the refactoring plan for the existing prompt system (`internal/prompt/`) to support regional investment research templates with layered composition architecture. The refactoring **builds on the solid foundation** of `pkg/models/localization.go` and maintains existing functionality while adding regional template support.

**Foundation**: Your `LocalizationConfig` in `pkg/models/localization.go` provides comprehensive localization with ISO compliance, validation, and integration throughout the codebase. All regional extensions are built through **composition** of this foundation.

## Current Prompt System Analysis

### Existing Architecture

```
internal/prompt/
├── interface.go        # Manager interface definitions
├── loader.go          # Template loading logic
├── manager.go         # Core prompt management
├── types.go           # Data structures
└── templates/
    ├── market/
    │   └── outlook.tmpl
    └── portfolio/
        ├── allocation.tmpl
        ├── compliance.tmpl
        ├── performance.tmpl
        ├── reallocation.tmpl
        └── risk.tmpl
```

### Current Template Pattern

**✅ Existing Strengths**:

- **Embedded Templates**: `//go:embed templates/portfolio/*.tmpl`
- **Template Manager**: Full `text/template` execution with data injection
- **Rich Data Binding**: Templates receive normalized portfolio data, localization, market context
- **Consistent Interface**: `Manager.GeneratePrompt(analysisType, data)`

### Current Template Example

From `templates/portfolio/risk.tmpl`:

```gotmpl
You are a portfolio risk analyst. Analyze the following portfolio data...

**User Context:**
- Location: {{.Localization.Country}}, {{.Localization.Region}}
- Currency: {{.Localization.Currency}}
- Language: {{.Localization.Language}}
{{- if .UserProfile}}
- Risk Tolerance: {{.UserProfile.RiskTolerance}}
{{- end}}

**Portfolio Overview:**
{{- if .Portfolio}}
- Total Value: ${{printf "%.2f" .Portfolio.TotalValueUSD}} USD
- Holdings Count: {{.Portfolio.HoldingsCount}}
{{- end}}
```

## Refactoring Plan: Layered Template Architecture

### Goal: Composition over Duplication

**New Architecture**:

```
internal/prompt/
├── interface.go                    # ✅ Existing - no changes needed
├── loader.go                      # 🔄 Extend for regional loading
├── manager.go                     # ✅ Existing manager - preserved unchanged
├── regional_manager.go            # 🆕 Regional template composition (alongside existing)
├── types.go                       # 🔄 Add regional extensions to PromptData
└── templates/
    ├── portfolio/                 # ✅ Existing templates preserved
    │   ├── allocation.tmpl
    │   ├── compliance.tmpl
    │   ├── performance.tmpl
    │   ├── reallocation.tmpl
    │   └── risk.tmpl
    ├── market/                    # ✅ Existing templates preserved
    │   └── outlook.tmpl
    └── investment_research/       # 🆕 New regional templates
        ├── base/
        │   ├── research.tmpl
        │   └── components/
        │       ├── context.tmpl
        │       ├── portfolio_analysis.tmpl
        │       ├── research_framework.tmpl
        │       └── output_format.tmpl
        └── regional/
            ├── overlays/
            │   ├── FR_overlay.tmpl    # ~10 lines each
            │   ├── US_overlay.tmpl
            │   └── CA_overlay.tmpl
            └── localization/
                ├── FR_fr.yaml         # Regional data
                ├── US_en.yaml
                └── CA_en.yaml

pkg/models/
├── regional.go                     # 🆕 Regional types (RegionalConfig, RegionalOverlay)
└── ... (existing model files)
```

### Architectural Decisions

#### ✅ **Localization Foundation Strategy**

- **LocalizationConfig**: Your `pkg/models/localization.go` provides the **solid foundation**
  - **Comprehensive Coverage**: Country, Language, Currency, Timezone, Region, City
  - **ISO Compliance**: Full validation with international standards
  - **Already Integrated**: Used in config, prompt system, and throughout codebase
- **Regional Extensions**: Build through **composition**, not duplication
- **Validation Authority**: Your `LocalizationConfig.Validate()` is the **gold standard**

#### ✅ **Type Location Strategy**

- **LocalizationConfig**: ✅ Already in `pkg/models/localization.go` (your excellent foundation)
- **PromptData**: Stays in `internal/prompt/types.go` (prompt-specific, uses your LocalizationConfig)
- **Regional Types**: New types go in `pkg/models/regional.go` (shared models, embed your LocalizationConfig)
- **Investment Profile**: Already planned for `pkg/models/investment_profile.go`
- **Existing Types**: No changes to current model locations

#### ✅ **Migration Strategy: Coexistence**

- **RegionalManager**: Lives alongside existing `manager` (no replacement)
- **Existing Manager**: Completely unchanged and unaffected
- **Interface Extension**: Add new methods, don't modify existing
- **Zero Breaking Changes**: All current functionality preserved
- **Your LocalizationConfig**: Remains unchanged, becomes foundation for extensions

### Refactoring Strategy

#### Phase 1: Extend Existing Types (No Breaking Changes)

**Existing Foundation: LocalizationConfig**

✅ **Already Implemented**: `pkg/models/localization.go` provides the solid foundation:

```go
// LocalizationConfig - ✅ EXISTING - Your excellent implementation
type LocalizationConfig struct {
    Country  string `mapstructure:"country" yaml:"country"`     // ISO 3166-1 alpha-2
    Language string `mapstructure:"language" yaml:"language"`   // ISO 639-1
    Timezone string `mapstructure:"timezone" yaml:"timezone"`   // IANA timezone
    Currency string `mapstructure:"currency" yaml:"currency"`   // ISO 4217
    Region   string `mapstructure:"region" yaml:"region"`       // Optional region
    City     string `mapstructure:"city" yaml:"city"`           // Optional city
}

// Validate() provides comprehensive ISO compliance validation
```

**Add Regional Support to PromptData** (internal/prompt/types.go):

```go
// PromptData - Extend existing struct with regional extensions
type PromptData struct {
    // ✅ Existing fields preserved (using your LocalizationConfig)
    Localization models.LocalizationConfig `json:"localization"`  // Your foundation
    UserProfile  *UserProfile              `json:"user_profile,omitempty"`
    Timestamp    time.Time                 `json:"timestamp"`
    Portfolio    *models.NormalizedPortfolio  `json:"portfolio,omitempty"`
    MarketData   *models.NormalizedMarketData `json:"market_data,omitempty"`
    MacroData    *models.NormalizedMacroData  `json:"macro_data,omitempty"`
    AnalysisType models.AnalysisType `json:"analysis_type"`
    Context      map[string]any      `json:"context,omitempty"`

    // 🆕 New regional extensions (optional fields)
    InvestmentProfile *models.InvestmentProfile  `json:"investment_profile,omitempty"`
    RegionalOverlay   *models.RegionalOverlay    `json:"regional_overlay,omitempty"`
    RegionalConfig    *models.RegionalConfig     `json:"regional_config,omitempty"`
}
```

**Add Regional Types** (pkg/models/regional.go - NEW FILE):

```go
// RegionalOverlay holds template additions and localization data
type RegionalOverlay struct {
    TemplateAdditions string                    `json:"template_additions"` // Regional template content
    LocalizationData  map[string]interface{}    `json:"localization_data"`  // YAML data injection
}

// RegionalConfig extends LocalizationConfig with investment-specific regional context
type RegionalConfig struct {
    models.LocalizationConfig                   // ✅ Embed your solid foundation (Country, Language, Currency, etc.)

    // Regional investment-specific extensions only
    Strings         map[string]string      `json:"strings"`         // Localized strings
    MarketContext   RegionalMarketContext  `json:"market_context"`  // Market-specific data
    TaxContext      RegionalTaxContext     `json:"tax_context"`     // Tax optimization context
    Data            map[string]interface{} `json:"data"`            // Raw YAML data
}

// Supporting regional types...
type RegionalMarketContext struct {
    RegulatoryFocus     string   `json:"regulatory_focus"`
    InvestmentCulture   string   `json:"investment_culture"`
    PreferredThemes     []string `json:"preferred_themes"`
    PrimaryExchanges    []string `json:"primary_exchanges"`
    MajorIndices        []string `json:"major_indices"`
}

type RegionalTaxContext struct {
    PrimaryAccounts          []string `json:"primary_accounts"`
    OptimizationStrategies   []string `json:"optimization_strategies"`
}
```

### Validation Strategy: Building on Your Foundation

#### ✅ **LocalizationConfig as Validation Authority**

Your `LocalizationConfig.Validate()` provides the **gold standard** for localization validation:

- **ISO Compliance**: Country (ISO 3166-1), Language (ISO 639-1), Currency (ISO 4217), Timezone (IANA)
- **Automatic Normalization**: Proper case conversion and string handling
- **Comprehensive Checks**: Required fields, format validation, descriptive errors

#### **Regional Validation Strategy**

Regional types delegate to your validation authority:

```go
// RegionalConfig validation leverages your solid foundation
func (rc *RegionalConfig) Validate() error {
    // First validate core localization using your robust validation
    if err := rc.LocalizationConfig.Validate(); err != nil {
        return fmt.Errorf("localization validation failed: %w", err)
    }

    // Regional-specific validations extend your base
    if len(rc.Strings) == 0 {
        return fmt.Errorf("regional strings cannot be empty")
    }

    if len(rc.MarketContext.PrimaryExchanges) == 0 {
        return fmt.Errorf("primary exchanges required for market context")
    }

    return nil
}
```

#### **Configuration Integration**

Regional configs inherit from your base configuration:

```yaml
# Regional configs reference your validated base
localization: # Your LocalizationConfig drives the foundation
  country: 'FR'
  language: 'fr'
  timezone: 'Europe/Paris'
  currency: 'EUR'
  region: 'Île-de-France'

# Regional extensions build on your validated foundation
regional:
  market_context:
    primary_exchanges: ['Euronext Paris']
    regulatory_focus: 'AMF compliance, ESG integration'
```

```

```

#### Phase 2: Extend Manager Interface (Additive Only)

**Add New Regional Interface** (internal/prompt/interface.go):

```go
// RegionalManager extends template management with regional capabilities
type RegionalManager interface {
    // Regional prompt generation
    GenerateRegionalPrompt(ctx context.Context, analysisType models.AnalysisType, data PromptData, country, language string) (string, error)
    LoadRegionalConfig(country, language string) (*models.RegionalConfig, error)
    LoadRegionalOverlay(analysisType models.AnalysisType, country, language string) (string, error)
}

// Manager interface - UNCHANGED (existing methods preserved)
type Manager interface {
    // ✅ Existing method completely unchanged
    GeneratePrompt(ctx context.Context, analysisType models.AnalysisType, data PromptData) (string, error)
}

// Note: RegionalManager is separate interface - existing Manager unchanged
```

````

#### Phase 3: Template Composition Implementation

**Base Template with Injection Points**:

```gotmpl
<!-- templates/investment_research/base/research.tmpl -->
{{template "context" .}}
{{template "portfolio_analysis" .}}

**Investment Research Framework:**
Use web search to identify investment research that:
1. Addresses portfolio gaps and diversification needs
2. Aligns with {{.InvestmentProfile.InvestmentStyle}} investment style
3. Matches {{.InvestmentProfile.ResearchDepth}} research depth

{{/* Regional customization injection point */}}
{{if .RegionalOverlay}}
{{.RegionalOverlay.TemplateAdditions}}
{{end}}

{{template "research_framework" .}}
{{template "output_format" .}}
````

**Minimal Regional Overlay** (~10 lines):

```gotmpl
<!-- templates/investment_research/regional/overlays/FR_overlay.tmpl -->
**French Investment Context:**
- Tax-advantaged accounts: {{.RegionalConfig.Strings.pea_description}}
- Regulatory focus: {{.RegionalConfig.MarketContext.regulatory_focus}}

**Research Focus Areas:**
{{range .RegionalConfig.MarketContext.preferred_themes}}
- {{.}}
{{end}}
```

**Regional Configuration Data**:

```yaml
# templates/investment_research/regional/localization/FR_fr.yaml
strings:
  pea_description: "Plan d'Épargne en Actions (PEA)"
  av_description: "Contrats d'Assurance Vie"

market_context:
  regulatory_focus: 'AMF compliance, ESG integration'
  investment_culture: 'Long-term stability, European focus'
  preferred_themes:
    - 'European equities'
    - 'ESG-compliant investments'
    - 'Dividend aristocrats'
    - 'Euro-denominated bonds'

tax_context:
  primary_accounts: ['PEA', 'Assurance Vie', 'Compte Titres']
  optimization_strategies: ['PEA maximization', 'AV unit-linked growth']
```

### Implementation Details

#### Regional Manager Implementation (regional_manager.go - NEW FILE)

```go
// RegionalManager implementation - lives alongside existing manager
type regionalManager struct {
    baseManager     Manager                             // Existing manager (composition)
    fs              fs.FS                              // Filesystem abstraction
    configDir       string                             // Regional config directory
    overlayCache    map[string]*template.Template      // Template cache
    configCache     map[string]*models.RegionalConfig  // Config cache
}

// NewRegionalManager creates regional manager that composes existing manager
func NewRegionalManager(
    baseManager Manager,
    fs fs.FS,
    configDir string,
) RegionalManager {
    return &regionalManager{
        baseManager:  baseManager, // Compose, don't replace
        fs:          fs,
        configDir:   configDir,
        overlayCache: make(map[string]*template.Template),
        configCache:  make(map[string]*models.RegionalConfig),
    }
}

func (rm *regionalManager) GenerateRegionalPrompt(
    ctx context.Context,
    analysisType models.AnalysisType,
    data PromptData,
    country, language string,
) (string, error) {
    // 1. Load regional overlay template
    overlay, err := rm.LoadRegionalOverlay(analysisType, country, language)
    if err != nil {
        return "", fmt.Errorf("failed to load regional overlay: %w", err)
    }

    // 2. Load regional configuration data
    regionalConfig, err := rm.LoadRegionalConfig(country, language)
    if err != nil {
        return "", fmt.Errorf("failed to load regional config: %w", err)
    }

    // 3. Inject regional data into existing PromptData
    data.RegionalOverlay = &models.RegionalOverlay{
        TemplateAdditions: overlay,
        LocalizationData:  regionalConfig.Data,
    }
    data.RegionalConfig = regionalConfig

    // 4. Delegate to existing manager (composition pattern)
    return rm.baseManager.GeneratePrompt(ctx, analysisType, data)
}
```

#### Template Resolution with Graceful Fallback

```go
func (rm *regionalManager) LoadRegionalOverlay(
    analysisType models.AnalysisType,
    country, language string,
) (string, error) {
    // Template resolution paths with fallback strategy
    cacheKey := fmt.Sprintf("%s_%s_%s_overlay", analysisType, country, language)

    // Check cache first
    if cached, exists := rm.overlayCache[cacheKey]; exists {
        var buf bytes.Buffer
        if err := cached.Execute(&buf, nil); err == nil {
            return buf.String(), nil
        }
    }

    // Resolution hierarchy with fallback
    paths := []string{
        fs.Join(rm.configDir, "templates", string(analysisType), "regional", "overlays",
                fmt.Sprintf("%s_%s_overlay.tmpl", country, language)),
        fs.Join(rm.configDir, "templates", string(analysisType), "regional", "overlays",
                fmt.Sprintf("%s_overlay.tmpl", country)),
        "", // No overlay - graceful fallback to base template only
    }

    for _, path := range paths {
        if path == "" {
            return "", nil // No regional overlay - use base template
        }

        if content, err := rm.fs.ReadFile(path); err == nil {
            // Cache the template for future use
            tmpl, parseErr := template.New("overlay").Parse(string(content))
            if parseErr == nil {
                rm.overlayCache[cacheKey] = tmpl
            }
            return string(content), nil
        }
    }

    return "", nil // Graceful fallback - no regional overlay needed
}

func (rm *regionalManager) LoadRegionalConfig(country, language string) (*models.RegionalConfig, error) {
    cacheKey := fmt.Sprintf("%s_%s_config", country, language)

    // Check cache first
    if cached, exists := rm.configCache[cacheKey]; exists {
        return cached, nil
    }

    // Load regional configuration with fallback
    configPath := fs.Join(rm.configDir, "templates", "investment_research", "regional",
                         "localization", fmt.Sprintf("%s_%s.yaml", country, language))

    configData, err := rm.fs.ReadFile(configPath)
    if err != nil {
        // Fallback to country-only config
        fallbackPath := fs.Join(rm.configDir, "templates", "investment_research", "regional",
                               "localization", fmt.Sprintf("%s_en.yaml", country))
        if configData, err = rm.fs.ReadFile(fallbackPath); err != nil {
            // Return minimal default config
            return &models.RegionalConfig{
                Country:  country,
                Language: language,
                Currency: "USD", // Default fallback
                Strings:  make(map[string]string),
                Data:     make(map[string]interface{}),
            }, nil
        }
    }

    var yamlData map[string]interface{}
    if err := yaml.Unmarshal(configData, &yamlData); err != nil {
        return nil, fmt.Errorf("failed to parse regional config: %w", err)
    }

    // Convert to RegionalConfig struct
    config := &models.RegionalConfig{
        Country:  country,
        Language: language,
        Data:     yamlData,
    }

    // Parse structured sections
    if strings, ok := yamlData["strings"].(map[string]interface{}); ok {
        config.Strings = make(map[string]string)
        for k, v := range strings {
            if str, ok := v.(string); ok {
                config.Strings[k] = str
            }
        }
    }

    // Cache the result
    rm.configCache[cacheKey] = config

    return config, nil
}
```

````

### Backward Compatibility & Coexistence

#### ✅ **Existing Manager Completely Unchanged**
- **No modifications** to existing `internal/prompt/manager.go`
- **No changes** to `internal/prompt/interface.go` Manager interface
- **All current templates** work identically (`templates/portfolio/*.tmpl`)
- **All existing engines** continue using existing manager without changes
- **Zero breaking changes** to any current functionality

#### ✅ **Migration Strategy: Side-by-Side Coexistence**

**Phase 1**: Add RegionalManager alongside existing manager
```go
// Both managers coexist - no replacement
type promptSystem struct {
    standardManager Manager         // ✅ Existing manager (unchanged)
    regionalManager RegionalManager // 🆕 New regional manager
}

// Existing engines continue using standard manager
riskEngine.UseManager(promptSystem.standardManager)

// New investment research engine uses regional manager
investmentEngine.UseRegionalManager(promptSystem.regionalManager)
````

**Phase 2**: New investment research engine uses RegionalManager
**Phase 3**: Optional migration of other engines (if desired, not required)

### Testing Strategy

#### Unit Tests for Regional Manager

```go
func TestRegionalManager_FallbackStrategy(t *testing.T) {
    // Setup: existing manager and in-memory filesystem
    existingManager := NewMockManager()
    testFS := NewInMemoryFS()

    // Only create base template, no regional overlays
    testFS.WriteFile("templates/investment_research/base/research.tmpl", baseTemplate, 0644)

    regionalManager := NewRegionalManager(existingManager, testFS, ".")

    // Should gracefully fall back to base template
    data := PromptData{AnalysisType: models.AnalysisInvestmentResearch}
    result, err := regionalManager.GenerateRegionalPrompt(ctx, models.AnalysisInvestmentResearch, data, "XX", "xx")

    assert.NoError(t, err)
    assert.Contains(t, result, "base template content")
    // Verify existing manager was called (composition)
    assert.True(t, existingManager.WasCalled())
}

func TestRegionalManager_RegionalOverlayComposition(t *testing.T) {
    existingManager := NewMockManager()
    testFS := NewInMemoryFS()

    // Setup base template and regional files
    testFS.WriteFile("templates/investment_research/base/research.tmpl", baseTemplate, 0644)
    testFS.WriteFile("templates/investment_research/regional/overlays/FR_overlay.tmpl", frenchOverlay, 0644)
    testFS.WriteFile("templates/investment_research/regional/localization/FR_fr.yaml", frenchConfig, 0644)

    regionalManager := NewRegionalManager(existingManager, testFS, ".")

    data := PromptData{AnalysisType: models.AnalysisInvestmentResearch}
    result, err := regionalManager.GenerateRegionalPrompt(ctx, models.AnalysisInvestmentResearch, data, "FR", "fr")

    assert.NoError(t, err)
    assert.Contains(t, result, "Plan d'Épargne en Actions")
    // Verify regional data was injected into PromptData
    assert.NotNil(t, data.RegionalConfig)
    assert.Equal(t, "FR", data.RegionalConfig.Country)
}

func TestRegionalManager_CoexistenceWithExisting(t *testing.T) {
    // Test that existing manager functionality is unaffected
    existingManager := NewManager() // Real existing manager
    regionalManager := NewRegionalManager(existingManager, fs.OS{}, ".")

    // Standard prompt generation should work unchanged
    standardData := PromptData{AnalysisType: models.AnalysisRisk}
    result, err := existingManager.GeneratePrompt(ctx, models.AnalysisRisk, standardData)

    assert.NoError(t, err)
    assert.Contains(t, result, "portfolio risk analyst") // Existing template content
    // Verify no regional data was added to standard flow
    assert.Nil(t, standardData.RegionalConfig)
    assert.Nil(t, standardData.RegionalOverlay)
}
```

#### Integration Tests

```go
func TestPromptSystemCoexistence_Integration(t *testing.T) {
    // Test both managers working side by side
    config := &config.Config{/* standard config */}

    // Create both manager types
    existingManager, err := prompt.NewManager(prompt.Dependencies{Config: config})
    require.NoError(t, err)

    regionalManager := prompt.NewRegionalManager(existingManager, fs.OS{}, config.ConfigDir)

    // Test existing functionality unchanged
    standardResult, err := existingManager.GeneratePrompt(ctx, models.AnalysisRisk, standardData)
    require.NoError(t, err)
    assert.Contains(t, standardResult, "risk analyst")

    // Test new regional functionality
    regionalResult, err := regionalManager.GenerateRegionalPrompt(ctx, models.AnalysisInvestmentResearch, regionalData, "FR", "fr")
    require.NoError(t, err)
    assert.Contains(t, regionalResult, "Plan d'Épargne")

    // Results should be different (regional context added)
    assert.NotEqual(t, standardResult, regionalResult)
}
```

```

## Benefits of This Refactoring

### ✅ Maintainable

- **Single Source of Truth**: Core logic in base templates
- **Minimal Duplication**: Regional overlays are 5-10 lines
- **Component Reuse**: Shared components across regions
- **Backward Compatible**: No breaking changes

### ✅ Regionalized

- **Cultural Adaptation**: Language and investment culture
- **Regulatory Context**: Tax implications and compliance
- **Market Focus**: Regional investment preferences
- **Flexible Fallback**: Graceful degradation to base templates

### ✅ Scalable

- **Easy Region Addition**: New country = overlay + config file
- **Asset Discovery**: AI-driven vs hardcoded recommendations
- **Template Evolution**: Base improvements benefit all regions
- **Configuration-Driven**: Changes via YAML, not code

### ✅ Testable

- **Filesystem Abstraction**: Easy unit testing with in-memory FS
- **Isolated Components**: Test overlays and base templates separately
- **Fallback Testing**: Verify graceful degradation
- **Integration Testing**: Full composition pipeline

This refactoring approach extends your existing prompt system without disruption while adding powerful regional capabilities for investment research.
```
