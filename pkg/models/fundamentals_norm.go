package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

// MarshalJSON provides compact JSON representation optimized for AI agents
func (f FundamentalsData) MarshalJSON() ([]byte, error) {
	compact := map[string]any{
		"ticker":       f.Ticker,
		"company_name": f.CompanyName,
		"sector":       f.Sector,
		"industry":     f.Industry,
		"last_updated": f.LastUpdated.Format("2006-01-02"),
	}

	// key valuation metrics only
	metrics := map[string]any{
		"price":      f.Metrics.Price,
		"market_cap": f.Metrics.MarketCap,
		"change_pct": f.Metrics.ChangePercentage,
	}

	// include key ratios if available
	if f.Metrics.PE > 0 {
		metrics["pe_ratio"] = f.Metrics.PE
	}
	if f.Metrics.PB > 0 {
		metrics["pb_ratio"] = f.Metrics.PB
	}
	if f.Metrics.DividendYield > 0 {
		metrics["dividend_yield"] = f.Metrics.DividendYield
	}
	if f.Metrics.Beta > 0 {
		metrics["beta"] = f.Metrics.Beta
	}
	if f.Metrics.ROE > 0 {
		metrics["roe"] = f.Metrics.ROE
	}

	compact["key_metrics"] = metrics

	// essential company info only
	companyInfo := map[string]any{
		"exchange": f.Company.Exchange,
		"currency": f.Company.Currency,
		"is_etf":   f.Company.IsETF,
		"is_fund":  f.Company.IsFund,
		"country":  f.Company.Country,
	}

	// include market cap category for AI context
	capCategory := "unknown"
	if f.Metrics.MarketCap > 200_000_000_000 {
		capCategory = "mega_cap"
	} else if f.Metrics.MarketCap > 10_000_000_000 {
		capCategory = "large_cap"
	} else if f.Metrics.MarketCap > 2_000_000_000 {
		capCategory = "mid_cap"
	} else if f.Metrics.MarketCap > 300_000_000 {
		capCategory = "small_cap"
	} else if f.Metrics.MarketCap > 50_000_000 {
		capCategory = "micro_cap"
	}
	companyInfo["market_cap_category"] = capCategory

	compact["company"] = companyInfo

	return json.Marshal(compact)
}

// String provides human-readable summary optimized for AI understanding
func (f FundamentalsData) String() string {
	price := fmt.Sprintf("$%.2f", f.Metrics.Price)
	if f.Metrics.ChangePercentage != 0 {
		if f.Metrics.ChangePercentage > 0 {
			price += fmt.Sprintf(" (+%.1f%%)", f.Metrics.ChangePercentage)
		} else {
			price += fmt.Sprintf(" (%.1f%%)", f.Metrics.ChangePercentage)
		}
	}

	var details []string

	if f.Metrics.PE > 0 {
		details = append(details, fmt.Sprintf("PE %.1f", f.Metrics.PE))
	}

	if f.Metrics.MarketCap > 1_000_000_000 {
		details = append(details, fmt.Sprintf("Cap $%.1fB", f.Metrics.MarketCap/1_000_000_000))
	} else if f.Metrics.MarketCap > 1_000_000 {
		details = append(details, fmt.Sprintf("Cap $%.1fM", f.Metrics.MarketCap/1_000_000))
	}

	if f.Metrics.DividendYield > 0 {
		details = append(details, fmt.Sprintf("Yield %.1f%%", f.Metrics.DividendYield))
	}

	detailStr := ""
	if len(details) > 0 {
		detailStr = " | " + strings.Join(details, ", ")
	}

	return fmt.Sprintf("%s (%s): %s, %s%s",
		f.Ticker, f.CompanyName, price, f.Sector, detailStr)
}

// MarshalJSON provides compact JSON representation for AI agents
func (m FundamentalMetrics) MarshalJSON() ([]byte, error) {
	compact := map[string]any{
		"price":      m.Price,
		"change_pct": m.ChangePercentage,
	}

	// only include meaningful ratios
	if m.MarketCap > 0 {
		compact["market_cap"] = m.MarketCap
	}
	if m.PE > 0 && m.PE < 100 { // filter out unrealistic PE ratios
		compact["pe_ratio"] = m.PE
	}
	if m.PB > 0 && m.PB < 20 { // filter out unrealistic PB ratios
		compact["pb_ratio"] = m.PB
	}
	if m.DividendYield > 0 {
		compact["dividend_yield"] = m.DividendYield
	}
	if m.Beta > 0 {
		compact["beta"] = m.Beta
	}
	if m.ROE > 0 {
		compact["roe"] = m.ROE
	}
	if m.DebtToEquity > 0 {
		compact["debt_to_equity"] = m.DebtToEquity
	}

	return json.Marshal(compact)
}

// String provides human-readable metrics summary
func (m FundamentalMetrics) String() string {
	var parts []string

	parts = append(parts, fmt.Sprintf("$%.2f", m.Price))

	if m.PE > 0 && m.PE < 100 {
		parts = append(parts, fmt.Sprintf("PE %.1f", m.PE))
	}

	if m.MarketCap > 1_000_000_000 {
		parts = append(parts, fmt.Sprintf("$%.1fB", m.MarketCap/1_000_000_000))
	} else if m.MarketCap > 1_000_000 {
		parts = append(parts, fmt.Sprintf("$%.1fM", m.MarketCap/1_000_000))
	}

	if m.DividendYield > 0 {
		parts = append(parts, fmt.Sprintf("Yield %.1f%%", m.DividendYield))
	}

	return strings.Join(parts, " | ")
}

// MarshalJSON provides compact JSON representation for AI agents
func (c CompanyInfo) MarshalJSON() ([]byte, error) {
	compact := map[string]any{
		"exchange":            c.Exchange,
		"currency":            c.Currency,
		"country":             c.Country,
		"is_etf":              c.IsETF,
		"is_fund":             c.IsFund,
		"is_adr":              c.IsADR,
		"is_actively_trading": c.IsActivelyTrading,
	}

	// include key identifiers
	if c.ISIN != "" {
		compact["isin"] = c.ISIN
	}
	if c.Website != "" {
		compact["website"] = c.Website
	}

	// company description (truncated for AI)
	if c.Description != "" && len(c.Description) > 200 {
		compact["description"] = c.Description[:200] + "..."
	} else if c.Description != "" {
		compact["description"] = c.Description
	}

	return json.Marshal(compact)
}

// String provides human-readable company summary
func (c CompanyInfo) String() string {
	var details []string

	if c.Exchange != "" {
		details = append(details, c.Exchange)
	}

	if c.Country != "" && c.Country != "US" {
		details = append(details, c.Country)
	}

	if c.IsETF {
		details = append(details, "ETF")
	} else if c.IsFund {
		details = append(details, "Fund")
	} else if c.IsADR {
		details = append(details, "ADR")
	}

	if !c.IsActivelyTrading {
		details = append(details, "Inactive")
	}

	if len(details) > 0 {
		return strings.Join(details, " | ")
	}

	return "Company"
}
