package models

import "time"

// EconomicIndicator represents a single economic metric with trend information
type EconomicIndicator struct {
	Value  float64   `json:"value"`
	Change float64   `json:"change"`
	Trend  string    `json:"trend"`
	AsOf   time.Time `json:"as_of"`
}

// MacroData represents macroeconomic data for a country
type MacroData struct {
	Country      string             `json:"country"`
	GDP          *EconomicIndicator `json:"gdp,omitempty"`
	Inflation    *EconomicIndicator `json:"inflation,omitempty"`
	InterestRate *EconomicIndicator `json:"interest_rate,omitempty"`
	Unemployment *EconomicIndicator `json:"unemployment,omitempty"`
	LastUpdated  time.Time          `json:"last_updated"`
}
