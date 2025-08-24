package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Normalize converts a Portfolio to a NormalizedPortfolio for AI analysis.
// This method reuses the existing MarshalJSON logic but returns a structured type.
func (p Portfolio) Normalize() (*NormalizedPortfolio, error) {
	// Parse the AsOf date
	asOfTime, err := p.AsOfTime()
	if err != nil {
		return nil, fmt.Errorf("invalid portfolio date: %w", err)
	}
	if asOfTime.IsZero() {
		asOfTime = time.Now()
	}

	// Calculate total value and collect all holdings
	var allHoldings []Holding
	totalValueUSD := 0.0
	holdingsCount := 0

	for _, account := range p.Accounts {
		for _, holding := range account.Holdings {
			allHoldings = append(allHoldings, holding)
			// For now, assume all values are in USD (we can enhance this later with currency conversion)
			totalValueUSD += holding.Value(0)
			holdingsCount++
		}
	}

	if totalValueUSD == 0 {
		return nil, fmt.Errorf("portfolio has zero total value")
	}

	// Convert holdings to normalized format
	normalizedHoldings := make([]NormalizedHolding, 0, len(allHoldings))
	for _, holding := range allHoldings {
		value := holding.Value(0)
		weight := (value / totalValueUSD) * 100 // Convert to percentage

		normalizedHolding := NormalizedHolding{
			Symbol:          holding.Ticker,
			Name:            holding.Name,
			WeightPercent:   weight,
			ValueUSD:        value,
			Quantity:        holding.Quantity,
			AssetClass:      normalizeAssetClass(holding.Type),
			Region:          normalizeRegion(holding.Region),
			Sector:          holding.Sector,
			Currency:        holding.Currency,
			IsLargePosition: weight > 5.0,
			IsForeign:       holding.Currency != "" && holding.Currency != p.BaseCurrency,
		}
		normalizedHoldings = append(normalizedHoldings, normalizedHolding)
	}

	// Calculate allocations
	assetAllocations := calculateAssetAllocations(normalizedHoldings)
	regionAllocations := calculateRegionAllocations(normalizedHoldings)
	sectorAllocations := calculateSectorAllocations(normalizedHoldings)

	// Calculate risk metrics
	riskMetrics := calculateRiskMetrics(normalizedHoldings, p.BaseCurrency)

	return &NormalizedPortfolio{
		TotalValueUSD:     totalValueUSD,
		BaseCurrency:      p.BaseCurrency,
		AsOfDate:          asOfTime,
		HoldingsCount:     holdingsCount,
		AssetAllocations:  assetAllocations,
		RegionAllocations: regionAllocations,
		SectorAllocations: sectorAllocations,
		Holdings:          normalizedHoldings,
		RiskMetrics:       riskMetrics,
	}, nil
}

// Helper functions for normalization

func normalizeAssetClass(assetType AssetType) string {
	switch assetType {
	case Stock:
		return "stock"
	case ETF:
		return "etf"
	case MutualFund:
		return "mutual_fund"
	case BondIG, BondHY, BondIL, BondGov, BondCorp, BondEM, BondMuni:
		return "bond"
	case Cash, CashEQ:
		return "cash"
	case Crypto, CryptoCore:
		return "crypto"
	default:
		return "other"
	}
}

func normalizeRegion(region string) string {
	if region == "" {
		return "Unknown"
	}
	// Standardize common region names
	region = strings.TrimSpace(strings.ToLower(region))
	switch region {
	case "us", "usa", "united states", "north america":
		return "US"
	case "eu", "europe", "european":
		return "Europe"
	case "asia", "asian":
		return "Asia"
	case "emerging", "emerging markets", "em":
		return "Emerging"
	case "global", "worldwide":
		return "Global"
	default:
		return cases.Title(language.Und).String(region)
	}
}

func calculateAssetAllocations(holdings []NormalizedHolding) map[string]float64 {
	allocations := make(map[string]float64)
	for _, holding := range holdings {
		allocations[holding.AssetClass] += holding.WeightPercent
	}
	return allocations
}

func calculateRegionAllocations(holdings []NormalizedHolding) map[string]float64 {
	allocations := make(map[string]float64)
	for _, holding := range holdings {
		allocations[holding.Region] += holding.WeightPercent
	}
	return allocations
}

func calculateSectorAllocations(holdings []NormalizedHolding) map[string]float64 {
	allocations := make(map[string]float64)
	for _, holding := range holdings {
		if holding.Sector != "" {
			allocations[holding.Sector] += holding.WeightPercent
		}
	}
	// Only return if we have sector data
	if len(allocations) == 0 {
		return nil
	}
	return allocations
}

func calculateRiskMetrics(holdings []NormalizedHolding, baseCurrency string) NormalizedRisk {
	if len(holdings) == 0 {
		return NormalizedRisk{}
	}

	// Calculate Herfindahl Index and other concentration metrics
	herfindahlIndex := 0.0
	largestPosition := 0.0
	foreignCurrencyWeight := 0.0

	// Track top 5 positions
	weights := make([]float64, len(holdings))
	regionConcentration := make(map[string]float64)
	sectorConcentration := make(map[string]float64)

	for i, holding := range holdings {
		weight := holding.WeightPercent / 100.0 // Convert to decimal
		weights[i] = weight

		// Herfindahl Index
		herfindahlIndex += weight * weight

		// Largest position
		if weight > largestPosition {
			largestPosition = weight
		}

		// Foreign currency exposure
		if holding.IsForeign {
			foreignCurrencyWeight += weight
		}

		// Regional concentration
		regionConcentration[holding.Region] += weight

		// Sector concentration
		if holding.Sector != "" {
			sectorConcentration[holding.Sector] += weight
		}
	}

	// Sort weights to find top 5
	// Simple approach: sum the largest weights
	maxRegionConcentration := 0.0
	for _, weight := range regionConcentration {
		if weight > maxRegionConcentration {
			maxRegionConcentration = weight
		}
	}

	maxSectorConcentration := 0.0
	for _, weight := range sectorConcentration {
		if weight > maxSectorConcentration {
			maxSectorConcentration = weight
		}
	}

	// Calculate top 5 positions (simplified)
	top5Weight := 0.0
	count := 0
	for _, weight := range weights {
		if count < 5 {
			top5Weight += weight
			count++
		}
		if count == 5 {
			break
		}
	}

	// Effective holdings = 1 / Herfindahl Index
	effectiveHoldings := 1.0
	if herfindahlIndex > 0 {
		effectiveHoldings = 1.0 / herfindahlIndex
	}

	return NormalizedRisk{
		HerfindahlIndex:     herfindahlIndex,
		LargestPositionPct:  largestPosition * 100,
		Top5PositionsPct:    top5Weight * 100,
		EffectiveHoldings:   effectiveHoldings,
		RegionConcentration: maxRegionConcentration * 100,
		SectorConcentration: maxSectorConcentration * 100,
		ForeignCurrencyPct:  foreignCurrencyWeight * 100,
	}
}

// MarshalJSON provides compact JSON representation optimized for AI agents
func (p Portfolio) MarshalJSON() ([]byte, error) {
	// calculate portfolio summary for compact representation
	totalHoldings := 0
	totalValue := 0.0
	currencies := make(map[string]bool)
	assetTypes := make(map[AssetType]int)

	for _, account := range p.Accounts {
		if account.Currency != "" {
			currencies[account.Currency] = true
		}
		totalHoldings += len(account.Holdings)
		for _, holding := range account.Holdings {
			totalValue += holding.Value(0) // use cost basis
			assetTypes[holding.Type]++
		}
	}

	// build currency list
	currencyList := make([]string, 0, len(currencies))
	for curr := range currencies {
		currencyList = append(currencyList, curr)
	}

	// build asset type summary
	assetSummary := make(map[string]int)
	for assetType, count := range assetTypes {
		assetSummary[string(assetType)] = count
	}

	compact := map[string]any{
		"as_of":           p.AsOf,
		"base_currency":   p.BaseCurrency,
		"accounts":        len(p.Accounts),
		"total_holdings":  totalHoldings,
		"total_value":     totalValue,
		"currencies":      currencyList,
		"asset_types":     assetSummary,
		"validated":       p.Validated,
		"accounts_detail": p.Accounts, // keep full detail for AI analysis
	}

	return json.Marshal(compact)
}

// String provides human-readable summary optimized for AI understanding
func (p Portfolio) String() string {
	totalHoldings := 0
	totalValue := 0.0
	currencies := make(map[string]bool)

	for _, account := range p.Accounts {
		if account.Currency != "" {
			currencies[account.Currency] = true
		}
		totalHoldings += len(account.Holdings)
		for _, holding := range account.Holdings {
			totalValue += holding.Value(0)
		}
	}

	currencyList := make([]string, 0, len(currencies))
	for curr := range currencies {
		currencyList = append(currencyList, curr)
	}

	status := "unvalidated"
	if p.Validated {
		status = "validated"
	}

	return fmt.Sprintf("Portfolio[%s]: %d accounts, %d holdings, %.2f total value (%s), %s",
		p.AsOf, len(p.Accounts), totalHoldings, totalValue,
		strings.Join(currencyList, ","), status)
}

// MarshalJSON provides compact JSON representation for AI agents
func (a Account) MarshalJSON() ([]byte, error) {
	totalValue := 0.0
	assetTypes := make(map[AssetType]int)

	for _, holding := range a.Holdings {
		totalValue += holding.Value(0)
		assetTypes[holding.Type]++
	}

	assetSummary := make(map[string]int)
	for assetType, count := range assetTypes {
		assetSummary[string(assetType)] = count
	}

	compact := map[string]any{
		"name":            a.Name,
		"type":            string(a.Type),
		"currency":        a.Currency,
		"balance":         a.Balance,
		"holdings":        len(a.Holdings),
		"total_value":     totalValue,
		"asset_types":     assetSummary,
		"holdings_detail": a.Holdings, // keep for AI analysis
	}

	return json.Marshal(compact)
}

// String provides human-readable account summary
func (a Account) String() string {
	totalValue := a.Balance
	for _, holding := range a.Holdings {
		totalValue += holding.Value(0)
	}

	return fmt.Sprintf("%s[%s]: %d holdings, %.2f %s total",
		a.Name, string(a.Type), len(a.Holdings), totalValue, a.Currency)
}

// MarshalJSON provides compact JSON representation for AI agents
func (h Holding) MarshalJSON() ([]byte, error) {
	value := h.Value(0)

	compact := map[string]any{
		"ticker":   h.Ticker,
		"quantity": h.Quantity,
		"value":    value,
		"currency": h.Currency,
		"type":     string(h.Type),
	}

	// only include optional fields if they have meaningful values
	if h.Name != "" {
		compact["name"] = h.Name
	}
	if h.Sector != "" {
		compact["sector"] = h.Sector
	}
	if h.Region != "" {
		compact["region"] = h.Region
	}

	return json.Marshal(compact)
}

// String provides human-readable holding summary
func (h Holding) String() string {
	value := h.Value(0)
	parts := []string{
		fmt.Sprintf("%s: %.2f@%.2f=%s%.2f", h.Ticker, h.Quantity, h.CostBasis, h.Currency, value),
		string(h.Type),
	}

	if h.Sector != "" {
		parts = append(parts, h.Sector)
	}

	return strings.Join(parts, " | ")
}
