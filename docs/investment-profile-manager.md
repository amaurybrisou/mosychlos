# Investment Profile Manager Implementation

## Executive Summary

This document outlines the implementation of a **simplified Investment Profile system** that extends beyond basic user preferences to provide specialized investment context for portfolio analysis and research. The Investment Profile complements the existing UserProfile with investment-specific data structures and regional adaptation capabilities.

**Implementation Focus**: Essential types in `pkg/models`, default profiles via YAML configs, and integration with existing `RegionalPromptManager` - **no complex generation or inference logic**.

## Investment Profile vs User Profile

### Current UserProfile Analysis

**Existing UserProfile Capabilities**:

```go
// From current templates: {{.UserProfile}}
type UserProfile struct {
    RiskTolerance        string   // "conservative", "moderate", "aggressive"
    InvestmentGoals      []string // ["retirement", "wealth_building", "income"]
    TimeHorizonYears     int      // 5, 10, 20, etc.
    AvailableCapitalUSD  float64  // Capital available for new investments
}
```

**Template Usage**:

```gotmpl
{{- if .UserProfile}}
- Risk Tolerance: {{.UserProfile.RiskTolerance}}
- Investment Goals: {{range .UserProfile.InvestmentGoals}}{{.}} {{end}}
- Time Horizon: {{.UserProfile.TimeHorizonYears}} years
- Available Capital: ${{printf "%.2f" .UserProfile.AvailableCapitalUSD}} USD
{{- end}}
```

### Why We Need a Dedicated Investment Profile

**✅ Investment Profile Justification**:

1. **Research Specialization**: Beyond basic user preferences into investment methodology
2. **Market Sophistication Levels**: Retail vs Affluent vs Institutional research depth
3. **Thematic Preferences**: Specific investment themes, ESG criteria, sector preferences
4. **Regional Investment Culture**: French ESG focus vs US growth focus vs German stability
5. **Research Behavior**: Information sources, update frequency, analysis depth

**❌ Current UserProfile Limitations**:

- Too generic for sophisticated investment research
- No regional investment culture adaptation
- Missing thematic and sector preferences
- No research depth configuration
- Limited ESG and sustainability criteria

## Investment Profile Architecture - Simplified

### Core Investment Profile Structure (Essential Types Only)

**Location**: `pkg/models/investment_profile.go`

```go
// InvestmentProfile contains essential investment preferences and regional context
type InvestmentProfile struct {
    // Core investment preferences
    InvestmentStyle   string   `json:"investment_style"`     // "growth", "value", "income", "balanced"
    ResearchDepth     string   `json:"research_depth"`       // "basic", "intermediate", "advanced"
    RiskApproach      string   `json:"risk_tolerance"`        // "conservative", "moderate", "aggressive"

    // Asset and sector preferences
    PreferredAssets   []string `json:"preferred_assets"`     // ["equities", "bonds", "alternatives"]
    AvoidedSectors    []string `json:"avoided_sectors"`      // ["tobacco", "weapons", "fossil_fuels"]

    // ESG criteria (simplified)
    ESGCriteria       ESGCriteria `json:"esg_criteria"`

    // Regional context (compose with existing foundation)
    RegionalContext   RegionalInvestmentContext `json:"regional_context"`

    // Metadata
    ProfileVersion    string    `json:"profile_version"`
    Source           string    `json:"source"`           // "default", "user_defined"
}

// ESGCriteria defines Environmental, Social, and Governance preferences (simplified)
type ESGCriteria struct {
    ESGImportance     string   `json:"esg_importance"`       // "mandatory", "important", "moderate", "low"
    ESGFocus         []string  `json:"esg_focus"`            // ["environmental", "social", "governance"]
    ExclusionCriteria []string `json:"exclusion_criteria"`   // ["fossil_fuels", "tobacco", "weapons"]
}

// RegionalInvestmentContext composes with existing LocalizationConfig
type RegionalInvestmentContext struct {
    LocalizationConfig                          // ✅ Your foundation
    TaxOptimizedAccounts []string  `json:"tax_optimized_accounts"` // ["PEA", "401k", "RRSP"]
    LocalPriorities      []string  `json:"local_priorities"`       // ["esg_focus", "dividend_income"]
    RegulatoryFocus      []string  `json:"regulatory_focus"`       // ["mifid_ii", "sec_rules"]
    PreferLocalChampions bool      `json:"prefer_local_champions"` // Regional market leaders preference
}

// Validate delegates to existing LocalizationConfig foundation
func (ip *InvestmentProfile) Validate() error {
    return ip.RegionalContext.LocalizationConfig.Validate()
}
```

### Supporting Types (Essential Only)

```go
// AllocationRange for future extension (keep simple)
type AllocationRange struct {
    MinPercent    float64 `json:"min_percent"`
    MaxPercent    float64 `json:"max_percent"`
    TargetPercent float64 `json:"target_percent"`
}

// ThemePreference for future thematic analysis
type ThemePreference struct {
    ThemeName     string `json:"theme_name"`     // "ai", "clean_energy", "healthcare"
    InterestLevel string `json:"interest_level"` // "high", "moderate", "low"
}
```

### Profile Examples by Investor Type (Simplified)

#### Conservative French Retiree Profile

```go
conservativeFrenchProfile := InvestmentProfile{
    InvestmentStyle:   "income",
    ResearchDepth:     "basic",
    RiskApproach:      "conservative",
    PreferredAssets:   []string{"bonds", "dividend_equities", "real_estate"},
    AvoidedSectors:    []string{"tobacco", "weapons", "speculative_tech"},

    ESGCriteria: ESGCriteria{
        ESGImportance:     "important",
        ESGFocus:          []string{"environmental", "social", "governance"},
        ExclusionCriteria: []string{"fossil_fuels", "tobacco", "weapons"},
    },

    RegionalContext: RegionalInvestmentContext{
        LocalizationConfig: config.LocalizationConfig{
            Country:  "FR",
            Language: "fr",
            // ... other localization fields
        },
        TaxOptimizedAccounts: []string{"PEA", "Assurance_Vie"},
        LocalPriorities:      []string{"tax_optimization", "esg_compliance", "capital_preservation"},
        RegulatoryFocus:      []string{"mifid_ii", "amf_compliance"},
        PreferLocalChampions: true,
    },

    ProfileVersion: "1.0",
    Source:         "default_regional",
}
```

#### Aggressive US Tech Professional Profile

```go
aggressiveUSProfile := InvestmentProfile{
    InvestmentStyle:   "growth",
    ResearchDepth:     "advanced",
    RiskApproach:      "aggressive",
    PreferredAssets:   []string{"growth_equities", "tech_stocks", "alternatives"},
    AvoidedSectors:    []string{"tobacco", "utilities"}, // Minimal avoidance

    ESGCriteria: ESGCriteria{
        ESGImportance:     "moderate",
        ESGFocus:          []string{"governance"},
        ExclusionCriteria: []string{}, // No strict exclusions
    },

    RegionalContext: RegionalInvestmentContext{
        LocalizationConfig: config.LocalizationConfig{
            Country:  "US",
            Language: "en",
            // ... other localization fields
        },
        TaxOptimizedAccounts: []string{"401k", "IRA", "Roth_IRA"},
        LocalPriorities:      []string{"growth", "tax_optimization", "innovation"},
        RegulatoryFocus:      []string{"sec_rules"},
        PreferLocalChampions: false, // Global focus
    },

    ProfileVersion: "1.0",
    Source:         "default_regional",
}
```

## Default Profile Storage (Aligned with RegionalPromptManager)

### Profile Storage Structure

**Location**: Extends existing config structure

```
config/
├── investment_profiles/                    # 🆕 New directory
│   └── defaults/                          # Default profiles by region/risk
│       ├── FR/
│       │   ├── conservative.yaml          # French conservative investor
│       │   ├── moderate.yaml              # French moderate investor
│       │   └── aggressive.yaml            # French aggressive investor
│       ├── US/
│       │   ├── conservative.yaml
│       │   ├── moderate.yaml
│       │   └── aggressive.yaml
│       ├── CA/
│       │   ├── conservative.yaml
│       │   ├── moderate.yaml
│       │   └── aggressive.yaml
│       └── global/                        # Fallback defaults
│           ├── conservative.yaml
│           ├── moderate.yaml
│           └── aggressive.yaml
└── localization/                          # ✅ Existing (your foundation)
    └── countries/
```

### Example Default Profile File (Simplified)

```yaml
# config/investment_profiles/defaults/FR/moderate.yaml
investment_style: 'balanced'
research_depth: 'intermediate'
risk_tolerance: 'moderate'

preferred_assets:
  - 'equities'
  - 'bonds'
  - 'real_estate'

avoided_sectors:
  - 'tobacco'
  - 'weapons'
  - 'fossil_fuels'

esg_criteria:
  esg_importance: 'important'
  esg_focus:
    - 'environmental'
    - 'social'
    - 'governance'
  exclusion_criteria:
    - 'fossil_fuels'
    - 'tobacco'
    - 'weapons'

regional_context:
  # Composes with LocalizationConfig
  country: 'FR'
  language: 'fr'
  currency: 'EUR'
  timezone: 'Europe/Paris'
  date_format: 'DD/MM/YYYY'

  # Investment-specific regional context
  tax_optimized_accounts:
    - 'PEA'
    - 'Assurance_Vie'
  local_priorities:
    - 'tax_optimization'
    - 'esg_compliance'
    - 'long_term_growth'
  regulatory_focus:
    - 'mifid_ii'
    - 'amf_compliance'
  prefer_local_champions: true

profile_version: '1.0'
source: 'default_regional'
```

## RegionalPromptManager Integration

### PromptData Extension

```go
// internal/prompt/types.go (extend existing PromptData)
type PromptData struct {
    // ✅ Existing fields
    Portfolio         *models.NormalizedPortfolio
    UserProfile       *models.UserProfile
    MarketData        *models.MarketData
    Localization      config.LocalizationConfig

    // 🆕 Add investment profile
    InvestmentProfile *models.InvestmentProfile  `json:"investment_profile,omitempty"`
}
```

### RegionalManager Extension (Simple)

```go
// internal/prompt/regional_manager.go (add method)
func (rm *regionalManager) LoadInvestmentProfile(
    country, riskTolerance string,
) (*models.InvestmentProfile, error) {

    profileKey := fmt.Sprintf("%s_%s", country, riskTolerance)

    // Check cache first
    if cached, exists := rm.profileCache[profileKey]; exists {
        return cached, nil
    }

    // Load from filesystem using same patterns as regional config
    profilePath := filepath.Join("investment_profiles", "defaults", country, riskTolerance+".yaml")

    profileData, err := rm.fs.ReadFile(profilePath)
    if err != nil {
        // Fallback to global default
        return rm.loadGlobalDefaultProfile(riskTolerance)
    }

    var profile models.InvestmentProfile
    if err := yaml.Unmarshal(profileData, &profile); err != nil {
        return nil, fmt.Errorf("failed to parse investment profile: %w", err)
    }

    // Cache the result using existing cache patterns
    rm.profileCache[profileKey] = &profile

    return &profile, nil
}
```

### Template Usage Examples (Simplified)

```gotmpl
<!-- Regional investment research template with investment profile -->

**Investor Profile Analysis:**
- Investment Style: {{.InvestmentProfile.InvestmentStyle}}
- Research Depth: {{.InvestmentProfile.ResearchDepth}}
- Risk Approach: {{.InvestmentProfile.RiskApproach}}

**Asset Preferences:**
{{range .InvestmentProfile.PreferredAssets}}
- {{.}} (preferred)
{{end}}

**Sector Exclusions:**
{{range .InvestmentProfile.AvoidedSectors}}
- {{.}} (excluded)
{{end}}

**ESG Requirements:**
{{if ne .InvestmentProfile.ESGCriteria.ESGImportance "low"}}
ESG Importance: {{.InvestmentProfile.ESGCriteria.ESGImportance}}
ESG Focus: {{range .InvestmentProfile.ESGCriteria.ESGFocus}}{{.}}, {{end}}
ESG Exclusions: {{range .InvestmentProfile.ESGCriteria.ExclusionCriteria}}{{.}}, {{end}}
{{end}}

**Regional Investment Context:**
Tax-Optimized Accounts: {{range .InvestmentProfile.RegionalContext.TaxOptimizedAccounts}}{{.}} {{end}}
Local Priorities: {{range .InvestmentProfile.RegionalContext.LocalPriorities}}{{.}}, {{end}}
{{if .InvestmentProfile.RegionalContext.PreferLocalChampions}}
Focus on domestic market leaders and regional champions.
{{end}}

**Research Instructions:**
Based on this investor profile, conduct {{.InvestmentProfile.ResearchDepth}} research focusing on:
1. {{.InvestmentProfile.InvestmentStyle}} investment opportunities
2. Preferred assets: {{range .InvestmentProfile.PreferredAssets}}{{.}}, {{end}}
3. Regional priorities: {{range .InvestmentProfile.RegionalContext.LocalPriorities}}{{.}}, {{end}}

{{if ne .InvestmentProfile.ESGCriteria.ESGImportance "low"}}
IMPORTANT: All recommendations must meet {{.InvestmentProfile.ESGCriteria.ESGImportance}} ESG standards.
Exclude: {{range .InvestmentProfile.ESGCriteria.ExclusionCriteria}}{{.}}, {{end}}
{{end}}
```

## Implementation Roadmap (Simplified)

### Phase 1: Essential Types ✅

- [x] Document simplified InvestmentProfile types for `pkg/models`
- [ ] Create basic InvestmentProfile, ESGCriteria, RegionalInvestmentContext
- [ ] Add simple validation delegating to LocalizationConfig
- [ ] Skip complex profile builders, managers, and inference logic

### Phase 2: Default Profiles ✅

- [x] Document config/investment_profiles/defaults/ structure
- [ ] Create directory structure aligned with RegionalPromptManager
- [ ] Write 3-4 example YAML files (FR conservative, US aggressive, etc.)
- [ ] Focus on realistic, useful profiles that match regional investment cultures

### Phase 3: RegionalManager Integration ✅

- [x] Document PromptData extension with InvestmentProfile field
- [ ] Add LoadInvestmentProfile method to existing RegionalManager
- [ ] Use existing fs.FS, caching, and error handling patterns
- [ ] Simple loading - no complex generation logic

### Phase 4: Template Integration

- [x] Document template usage examples with {{.InvestmentProfile}}
- [ ] Update 1-2 existing templates to demonstrate InvestmentProfile usage
- [ ] Test with current RegionalPromptManager integration
- [ ] Validate end-to-end workflow

## Key Decisions Made

**✅ Scope Reduction:**

- **No ProfileBuilder, ProfileManager, or complex inference logic**
- **No portfolio-based profile generation**
- **No sophisticated behavioral modeling**

**✅ Alignment with Existing Architecture:**

- **Types in pkg/models** (shared domain models)
- **Config-based defaults** (YAML files in config/)
- **RegionalPromptManager integration** (extend existing patterns)
- **LocalizationConfig foundation** (compose, don't replace)

**✅ Immediate Value Focus:**

- **Essential investment context for templates**
- **Regional investment culture awareness**
- **Simple ESG and sector preference modeling**
- **Tax-optimized account awareness**

This simplified approach provides immediate value for template-driven investment research while maintaining architectural consistency and avoiding unnecessary complexity.
