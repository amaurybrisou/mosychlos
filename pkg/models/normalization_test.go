package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestPortfolio_Normalization(t *testing.T) {
	portfolio := Portfolio{
		AsOf:         "2025-08-12",
		BaseCurrency: "USD",
		Validated:    true,
		Accounts: []Account{
			{
				Name:     "Test Brokerage",
				Type:     AccountBrokerage,
				Currency: "USD",
				Balance:  1000.0,
				Holdings: []Holding{
					{
						Ticker:    "AAPL",
						Quantity:  10,
						CostBasis: 150.0,
						Currency:  "USD",
						Type:      Stock,
						Sector:    "Technology",
					},
					{
						Ticker:    "BTC",
						Quantity:  0.5,
						CostBasis: 50000.0,
						Currency:  "USD",
						Type:      Crypto,
					},
				},
			},
		},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(portfolio)
	if err != nil {
		t.Fatalf("JSON marshaling failed: %v", err)
	}

	// verify compact structure includes summary fields
	var result map[string]any
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("JSON unmarshaling failed: %v", err)
	}

	// check summary fields
	if result["accounts"] != float64(1) {
		t.Errorf("Expected accounts count 1, got %v", result["accounts"])
	}

	if result["total_holdings"] != float64(2) {
		t.Errorf("Expected total_holdings 2, got %v", result["total_holdings"])
	}

	if !result["validated"].(bool) {
		t.Error("Expected validated to be true")
	}

	// Test String representation
	str := portfolio.String()
	if !strings.Contains(str, "Portfolio[2025-08-12]") {
		t.Errorf("String should contain portfolio date, got: %s", str)
	}

	if !strings.Contains(str, "1 accounts") {
		t.Errorf("String should contain account count, got: %s", str)
	}

	if !strings.Contains(str, "2 holdings") {
		t.Errorf("String should contain holdings count, got: %s", str)
	}
}

func TestMacroData_Normalization(t *testing.T) {
	macro := MacroData{
		Country:     "US",
		LastUpdated: time.Date(2025, 8, 12, 12, 0, 0, 0, time.UTC),
		GDP: &EconomicIndicator{
			Value:  2.1,
			Change: 0.3,
			Trend:  "rising",
			AsOf:   time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC),
		},
		Inflation: &EconomicIndicator{
			Value:  3.2,
			Change: -0.1,
			Trend:  "falling",
			AsOf:   time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(macro)
	if err != nil {
		t.Fatalf("JSON marshaling failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("JSON unmarshaling failed: %v", err)
	}

	if result["country"] != "US" {
		t.Errorf("Expected country US, got %v", result["country"])
	}

	// Test String representation
	str := macro.String()
	if !strings.Contains(str, "US:") {
		t.Errorf("String should start with country, got: %s", str)
	}

	if !strings.Contains(str, "GDP 2.1% (rising)") {
		t.Errorf("String should contain GDP info, got: %s", str)
	}

	if !strings.Contains(str, "Inflation 3.2% (falling)") {
		t.Errorf("String should contain inflation info, got: %s", str)
	}
}

func TestComplianceRules_Normalization(t *testing.T) {
	rules := ComplianceRules{
		AllowedAssetTypes:    []string{"stock", "etf", "bond_gov"},
		DisallowedAssetTypes: []string{"crypto", "derivative_option"},
		TickerBlocklist:      []string{"GME", "AMC"},
		TickerSubstitutes:    map[string]string{"VTI": "VXUS", "SPY": "IVV"},
		MaxLeverage:          2,
		Notes:                "Conservative investment approach",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(rules)
	if err != nil {
		t.Fatalf("JSON marshaling failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("JSON unmarshaling failed: %v", err)
	}

	if result["max_leverage"] != float64(2) {
		t.Errorf("Expected max_leverage 2, got %v", result["max_leverage"])
	}

	if result["allowed_count"] != float64(3) {
		t.Errorf("Expected allowed_count 3, got %v", result["allowed_count"])
	}

	// Test String representation
	str := rules.String()
	if !strings.Contains(str, "max leverage 2x") {
		t.Errorf("String should contain leverage info, got: %s", str)
	}

	if !strings.Contains(str, "3 allowed types") {
		t.Errorf("String should contain allowed types count, got: %s", str)
	}

	if !strings.Contains(str, "2 blocked types") {
		t.Errorf("String should contain blocked types count, got: %s", str)
	}
}
