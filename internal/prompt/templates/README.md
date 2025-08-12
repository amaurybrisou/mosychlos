# Prompt Templates

This directory contains embedded Go templates used for AI prompt generation, organized by domain for better maintainability and visibility.

## Directory Structure

```
templates/
└── portfolio/          # Portfolio analysis templates
    ├── risk.tmpl       # Risk assessment analysis
    ├── allocation.tmpl # Asset allocation strategy
    ├── performance.tmpl# Performance evaluation
    └── compliance.tmpl # Regulatory compliance check
```

## Usage

Templates are embedded into the binary using Go's `embed` directive and loaded at runtime through the prompt manager. Each template receives a `PromptData` struct containing:

- **User Context**: Localization (country, language, currency, timezone) and preferences
- **Portfolio Data**: Normalized portfolio structure with holdings, allocations, and risk metrics
- **Market Data**: Current market conditions, indices performance, and sentiment indicators
- **Macro Data**: Economic indicators and trends
- **Analysis Context**: Focus areas specific to each analysis type

## Template Functions

Available template functions:

- `mul`: Multiply two float64 values (e.g., `{{mul .Return1Day 100}}` for percentage conversion)
- Standard Go template functions (printf, range, if, etc.)

## Adding New Templates

1. Create template file in appropriate domain directory (e.g., `templates/portfolio/new_analysis.tmpl`)
2. Add corresponding `AnalysisType` constant in `manager.go`
3. Update `templateFiles` map in `loadTemplates()` method
4. Add analysis-specific context in `gatherPromptData()` method

## Future Domains

Potential future template domains:

- `market/` - Market analysis and research templates
- `macro/` - Economic and macro analysis templates
- `news/` - News analysis and sentiment templates
- `research/` - Investment research templates
