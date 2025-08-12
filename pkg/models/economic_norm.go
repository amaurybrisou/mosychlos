package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

// MarshalJSON provides compact JSON representation optimized for AI agents
func (m MacroData) MarshalJSON() ([]byte, error) {
	compact := map[string]any{
		"country":      m.Country,
		"last_updated": m.LastUpdated.Format("2006-01-02"),
	}

	// flatten indicators with current values and trends
	if m.GDP != nil {
		compact["gdp"] = map[string]any{
			"value":  m.GDP.Value,
			"change": m.GDP.Change,
			"trend":  m.GDP.Trend,
		}
	}

	if m.Inflation != nil {
		compact["inflation"] = map[string]any{
			"value":  m.Inflation.Value,
			"change": m.Inflation.Change,
			"trend":  m.Inflation.Trend,
		}
	}

	if m.InterestRate != nil {
		compact["interest_rate"] = map[string]any{
			"value":  m.InterestRate.Value,
			"change": m.InterestRate.Change,
			"trend":  m.InterestRate.Trend,
		}
	}

	if m.Unemployment != nil {
		compact["unemployment"] = map[string]any{
			"value":  m.Unemployment.Value,
			"change": m.Unemployment.Change,
			"trend":  m.Unemployment.Trend,
		}
	}

	return json.Marshal(compact)
}

// String provides human-readable summary optimized for AI understanding
func (m MacroData) String() string {
	var parts []string

	if m.GDP != nil {
		parts = append(parts, fmt.Sprintf("GDP %.1f%% (%s)", m.GDP.Value, m.GDP.Trend))
	}

	if m.Inflation != nil {
		parts = append(parts, fmt.Sprintf("Inflation %.1f%% (%s)", m.Inflation.Value, m.Inflation.Trend))
	}

	if m.InterestRate != nil {
		parts = append(parts, fmt.Sprintf("Rate %.2f%% (%s)", m.InterestRate.Value, m.InterestRate.Trend))
	}

	if m.Unemployment != nil {
		parts = append(parts, fmt.Sprintf("Unemployment %.1f%% (%s)", m.Unemployment.Value, m.Unemployment.Trend))
	}

	if len(parts) == 0 {
		return fmt.Sprintf("%s: No economic data", m.Country)
	}

	return fmt.Sprintf("%s: %s", m.Country, strings.Join(parts, ", "))
}

// MarshalJSON provides compact JSON representation for AI agents
func (e EconomicIndicator) MarshalJSON() ([]byte, error) {
	compact := map[string]any{
		"value":  e.Value,
		"change": e.Change,
		"trend":  e.Trend,
		"as_of":  e.AsOf.Format("2006-01-02"),
	}

	return json.Marshal(compact)
}

// String provides human-readable indicator summary
func (e EconomicIndicator) String() string {
	changeStr := ""
	if e.Change > 0 {
		changeStr = fmt.Sprintf(" (+%.1f)", e.Change)
	} else if e.Change < 0 {
		changeStr = fmt.Sprintf(" (%.1f)", e.Change)
	}

	return fmt.Sprintf("%.2f%s [%s]", e.Value, changeStr, e.Trend)
}
