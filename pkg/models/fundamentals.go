package models

import "time"

// FundamentalMetrics represents key financial metrics for a company
type FundamentalMetrics struct {
	// Valuation metrics
	MarketCap float64 `json:"market_cap"`
	Price     float64 `json:"price"`
	PE        float64 `json:"pe_ratio"`
	PB        float64 `json:"pb_ratio"`
	Beta      float64 `json:"beta"`

	// Dividend metrics
	DividendYield float64 `json:"dividend_yield"`
	LastDividend  float64 `json:"last_dividend"`

	// Performance metrics
	ROE          float64 `json:"roe"`
	DebtToEquity float64 `json:"debt_to_equity"`

	// Price movement
	Change           float64 `json:"change"`
	ChangePercentage float64 `json:"change_percentage"`
	Range            string  `json:"range"` // "164.08-260.1"

	// Volume metrics
	Volume        int64 `json:"volume"`
	AverageVolume int64 `json:"average_volume"`
}

// CompanyInfo represents detailed company information
type CompanyInfo struct {
	CEO               string `json:"ceo"`
	FullTimeEmployees string `json:"full_time_employees"`
	Phone             string `json:"phone"`
	Website           string `json:"website"`
	Description       string `json:"description"`

	// Address information
	Address string `json:"address"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
	Country string `json:"country"`

	// Company identifiers
	CIK   string `json:"cik"`
	ISIN  string `json:"isin"`
	CUSIP string `json:"cusip"`

	// Exchange information
	Exchange         string `json:"exchange"`
	ExchangeFullName string `json:"exchange_full_name"`
	Currency         string `json:"currency"`

	// Company characteristics
	IPODate           string `json:"ipo_date"` // "1980-12-12"
	Image             string `json:"image"`
	DefaultImage      bool   `json:"default_image"`
	IsETF             bool   `json:"is_etf"`
	IsActivelyTrading bool   `json:"is_actively_trading"`
	IsADR             bool   `json:"is_adr"`
	IsFund            bool   `json:"is_fund"`
}

// FundamentalsData represents comprehensive fundamental analysis data for a security
type FundamentalsData struct {
	Ticker      string             `json:"ticker"`
	CompanyName string             `json:"company_name"`
	Sector      string             `json:"sector"`
	Industry    string             `json:"industry"`
	Metrics     FundamentalMetrics `json:"metrics"`
	Company     CompanyInfo        `json:"company"`
	LastUpdated time.Time          `json:"last_updated"`
}
