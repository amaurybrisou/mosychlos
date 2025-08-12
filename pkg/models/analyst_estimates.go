package models

import "time"

// AnalystEstimate represents analyst estimates for a specific period
type AnalystEstimate struct {
	Symbol string `json:"symbol"`
	Date   string `json:"date"` // "2029-09-28"

	// Revenue estimates
	RevenueLow  int64 `json:"revenue_low"`
	RevenueHigh int64 `json:"revenue_high"`
	RevenueAvg  int64 `json:"revenue_avg"`

	// EBITDA estimates
	EbitdaLow  int64 `json:"ebitda_low"`
	EbitdaHigh int64 `json:"ebitda_high"`
	EbitdaAvg  int64 `json:"ebitda_avg"`

	// EBIT estimates
	EbitLow  int64 `json:"ebit_low"`
	EbitHigh int64 `json:"ebit_high"`
	EbitAvg  int64 `json:"ebit_avg"`

	// Net Income estimates
	NetIncomeLow  int64 `json:"net_income_low"`
	NetIncomeHigh int64 `json:"net_income_high"`
	NetIncomeAvg  int64 `json:"net_income_avg"`

	// SG&A expenses
	SgaExpenseLow  int64 `json:"sga_expense_low"`
	SgaExpenseHigh int64 `json:"sga_expense_high"`
	SgaExpenseAvg  int64 `json:"sga_expense_avg"`

	// EPS estimates
	EpsLow  float64 `json:"eps_low"`
	EpsHigh float64 `json:"eps_high"`
	EpsAvg  float64 `json:"eps_avg"`

	// Analyst coverage
	NumAnalystsRevenue int `json:"num_analysts_revenue"`
	NumAnalystsEps     int `json:"num_analysts_eps"`
}

// AnalystEstimatesData represents the complete analyst estimates response
type AnalystEstimatesData struct {
	Ticker      string            `json:"ticker"`
	Period      string            `json:"period"` // "annual" or "quarter"
	Estimates   []AnalystEstimate `json:"estimates"`
	LastUpdated time.Time         `json:"last_updated"`
	Error       string            `json:"error,omitempty"` // Error message if data unavailable
}
