# Investment Profile Manager

The `internal/profile` package provides investment profile management functionality for the mosychlos financial analysis system. It follows established patterns from the tools package, using SharedBag for session caching and filesystem operations for persistent storage.

## Overview

The ProfileManager provides:

- **Loading** of investment profiles from YAML configuration files
- **Caching** via SharedBag for performance and cross-engine coordination
- **Saving** of user profiles to the data directory
- **Country-specific** profiles with global fallbacks

## Architecture

```
ProfileManager
├── SharedBag Integration (session cache)
├── Filesystem Interface (config profiles)
└── DataDir Storage (user profiles)
```

## Usage

### Creating a ProfileManager

```go
import (
    "github.com/amaurybrisou/mosychlos/internal/profile"
    "github.com/amaurybrisou/mosychlos/pkg/bag"
)

// Create with filesystem, data directory, and shared bag
manager, err := profile.NewProfileManager(
    configFS,           // fs.FS for reading config profiles
    "/path/to/datadir", // Data directory for user profiles
    sharedBag,          // SharedBag for caching (can be nil)
)
```

### Loading Profiles

```go
ctx := context.Background()

// Load country-specific profile (falls back to global if not found)
profile, err := manager.LoadProfile(ctx, "US", "aggressive")

// Load global profile directly
profile, err := manager.LoadProfile(ctx, "GLOBAL", "moderate")
```

#### Load Strategy

1. **SharedBag Check**: Returns cached profile if present
2. **Country-Specific**: Attempts `investment_profiles/defaults/{country}/{riskTolerance}.yaml`
3. **Global Fallback**: Falls back to `investment_profiles/defaults/global/{riskTolerance}.yaml`
4. **Cache Result**: Stores successful loads in SharedBag

### Saving Profiles

```go
profile := &models.InvestmentProfile{
    InvestmentStyle: "growth",
    ResearchDepth:   "comprehensive",
    RegionalPreferences: models.RegionalInvestmentPreferences{
        LocalizationConfig: models.LocalizationConfig{
            Country:  "US",
            Currency: "USD",
        },
    },
}

err := manager.SaveProfile(ctx, profile, "user_custom.yaml")
// Saves to: {dataDir}/profiles/user_custom.yaml
```

## Configuration Structure

### Profile Directory Layout

```
investment_profiles/defaults/
├── US/
│   ├── conservative.yaml
│   ├── moderate.yaml
│   └── aggressive.yaml
├── FR/
│   ├── conservative.yaml
│   └── moderate.yaml
└── global/
    ├── conservative.yaml
    ├── moderate.yaml
    └── aggressive.yaml
```

### Profile YAML Format

```yaml
investment_style: 'growth'
research_depth: 'comprehensive'
asset_classes:
  - 'equities'
  - 'bonds'
regional_preferences:
  country: 'US'
  language: 'en'
  currency: 'USD'
  timezone: 'America/New_York'
  preferred_asset_classes:
    - 'US equities'
  esg_preferences:
    environmental:
      - 'clean energy'
custom_requirements:
  - 'ESG focused'
  - 'Tax efficient'
```

## SharedBag Integration

The ProfileManager integrates with the SharedBag system using `keys.KProfile`:

- **Caching**: Successful profile loads are cached in SharedBag
- **Session State**: Profiles persist across engine operations within a session
- **Cross-Engine**: Other engines can access cached profiles via SharedBag

## Error Handling

- **Validation**: Country and risk tolerance parameters are required
- **File Not Found**: Graceful fallback from country-specific to global profiles
- **Parse Errors**: Clear error messages for YAML parsing issues
- **Directory Creation**: Automatic creation of profile directories for saves

## Integration with Models

Uses the comprehensive `models.InvestmentProfile` structure:

- **InvestmentProfile**: Main profile container
- **RegionalInvestmentPreferences**: Country/region-specific settings
- **LocalizationConfig**: Embedded localization settings (country, language, currency, timezone)
- **ESGCriteria**: Environmental, Social, and Governance preferences

## Testing

Comprehensive test suite includes:

- Profile loading (country-specific, global fallback, cached)
- SharedBag integration and caching behavior
- Profile saving to data directory
- Error conditions and validation
- YAML parsing and serialization

Run tests:

```bash
go test ./internal/profile -v
```

## Dependencies

- `gopkg.in/yaml.v3`: YAML parsing and serialization
- `pkg/bag`: SharedBag for session caching
- `pkg/models`: Investment profile data structures
- `pkg/keys`: SharedBag key constants
