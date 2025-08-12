package models

import "time"

// CompanyTickersResponse represents the company tickers API response
type CompanyTickersResponse struct {
	Fields  []string                     `json:"fields"`
	Data    [][]interface{}              `json:"data"`
	Tickers map[string]CompanyTickerInfo `json:",inline"`
}

// CompanyTickerInfo holds company ticker information
type CompanyTickerInfo struct {
	CIKStr int    `json:"cik_str"`
	Ticker string `json:"ticker"`
	Title  string `json:"title"`
}

// CompanyFactsResponse represents company facts API response
type CompanyFactsResponse struct {
	CIK        int                            `json:"cik"`
	EntityName string                         `json:"entityName"`
	Facts      map[string]map[string]FactData `json:"facts"`
}

// FactData represents financial fact data
type FactData struct {
	Label       string                `json:"label"`
	Description string                `json:"description"`
	Units       map[string][]UnitData `json:"units"`
}

// UnitData represents unit-specific fact data
type UnitData struct {
	End   string      `json:"end"`
	Val   interface{} `json:"val"`
	AccN  string      `json:"accn,omitempty"`
	FY    int         `json:"fy,omitempty"`
	FP    string      `json:"fp,omitempty"`
	Form  string      `json:"form,omitempty"`
	Filed string      `json:"filed,omitempty"`
	Start string      `json:"start,omitempty"`
	Frame string      `json:"frame,omitempty"`
}

// CompanyFilingsResponse represents company filings API response
type CompanyFilingsResponse struct {
	CIK                               int          `json:"cik"`
	EntityType                        string       `json:"entityType"`
	SIC                               string       `json:"sic"`
	SICDescription                    string       `json:"sicDescription"`
	InsiderTransactionForOwnerExists  int          `json:"insiderTransactionForOwnerExists"`
	InsiderTransactionForIssuerExists int          `json:"insiderTransactionForIssuerExists"`
	Name                              string       `json:"name"`
	Tickers                           []string     `json:"tickers"`
	Exchanges                         []string     `json:"exchanges"`
	EIN                               string       `json:"ein"`
	Description                       string       `json:"description"`
	Website                           string       `json:"website"`
	InvestorWebsite                   string       `json:"investorWebsite"`
	Category                          string       `json:"category"`
	FiscalYearEnd                     string       `json:"fiscalYearEnd"`
	StateOfIncorporation              string       `json:"stateOfIncorporation"`
	StateOfIncorporationDescription   string       `json:"stateOfIncorporationDescription"`
	Addresses                         AddressInfo  `json:"addresses"`
	Phone                             string       `json:"phone"`
	Flags                             string       `json:"flags"`
	FormerNames                       []FormerName `json:"formerNames"`
	Filings                           FilingsData  `json:"filings"`
}

// AddressInfo represents company address information
type AddressInfo struct {
	Mailing  Address `json:"mailing"`
	Business Address `json:"business"`
}

// Address represents an address
type Address struct {
	Street1                   string `json:"street1"`
	Street2                   string `json:"street2,omitempty"`
	City                      string `json:"city"`
	StateOrCountry            string `json:"stateOrCountry"`
	ZipCode                   string `json:"zipCode"`
	StateOrCountryDescription string `json:"stateOrCountryDescription"`
}

// FormerName represents a former company name
type FormerName struct {
	Name string `json:"name"`
	From string `json:"from"`
	To   string `json:"to"`
}

// FilingsData represents filings data
type FilingsData struct {
	Recent FilingsRecent `json:"recent"`
	Files  []string      `json:"files"`
}

// FilingsRecent represents recent filings
type FilingsRecent struct {
	AccessionNumber       []string `json:"accessionNumber"`
	FilingDate            []string `json:"filingDate"`
	ReportDate            []string `json:"reportDate"`
	AcceptanceDateTime    []string `json:"acceptanceDateTime"`
	Act                   []string `json:"act"`
	Form                  []string `json:"form"`
	FileNumber            []string `json:"fileNumber"`
	FilmNumber            []string `json:"filmNumber"`
	Items                 []string `json:"items"`
	Size                  []int    `json:"size"`
	IsXBRL                []int    `json:"isXBRL"`
	IsInlineXBRL          []int    `json:"isInlineXBRL"`
	PrimaryDocument       []string `json:"primaryDocument"`
	PrimaryDocDescription []string `json:"primaryDocDescription"`
}

// InsiderTransactionsResponse represents insider trading data
type InsiderTransactionsResponse struct {
	CIK     int                `json:"cik"`
	Name    string             `json:"name"`
	Filings InsiderFilingsData `json:"filings"`
}

// InsiderFilingsData represents insider filings data
type InsiderFilingsData struct {
	Recent InsiderFilingsRecent `json:"recent"`
	Files  []string             `json:"files"`
}

// InsiderFilingsRecent represents recent insider filings
type InsiderFilingsRecent struct {
	AccessionNumber       []string `json:"accessionNumber"`
	FilingDate            []string `json:"filingDate"`
	ReportDate            []string `json:"reportDate"`
	AcceptanceDateTime    []string `json:"acceptanceDateTime"`
	Act                   []string `json:"act"`
	Form                  []string `json:"form"`
	FileNumber            []string `json:"fileNumber"`
	FilmNumber            []string `json:"filmNumber"`
	Items                 []string `json:"items"`
	Size                  []int    `json:"size"`
	IsXBRL                []int    `json:"isXBRL"`
	IsInlineXBRL          []int    `json:"isInlineXBRL"`
	PrimaryDocument       []string `json:"primaryDocument"`
	PrimaryDocDescription []string `json:"primaryDocDescription"`
	OwnerCik              []int    `json:"ownerCik"`
	OwnerName             []string `json:"ownerName"`
	IssuerCik             []int    `json:"issuerCik"`
	IssuerName            []string `json:"issuerName"`
	IssuerTradingSymbol   []string `json:"issuerTradingSymbol"`
}

// FilingsParams represents parameters for filings requests
type FilingsParams struct {
	CIK        string `json:"cik,omitempty"`
	Type       string `json:"type,omitempty"`
	DateBefore string `json:"dateb,omitempty"`
	Owner      string `json:"owner,omitempty"`
	Count      int    `json:"count,omitempty"`
	Start      int    `json:"start,omitempty"`
}

// FormType represents common SEC form types
type FormType string

const (
	Form10K    FormType = "10-K"
	Form10Q    FormType = "10-Q"
	Form8K     FormType = "8-K"
	Form4      FormType = "4"
	Form3      FormType = "3"
	Form5      FormType = "5"
	FormDEF14A FormType = "DEF 14A"
	Form11K    FormType = "11-K"
	Form20F    FormType = "20-F"
)

// Common date formats used by SEC
const (
	SECDateFormat     = "2006-01-02"
	SECDateTimeFormat = "2006-01-02T15:04:05.000Z"
)

// ParseSECDate parses SEC date string
func ParseSECDate(dateStr string) (time.Time, error) {
	return time.Parse(SECDateFormat, dateStr)
}

// ParseSECDateTime parses SEC datetime string
func ParseSECDateTime(datetimeStr string) (time.Time, error) {
	return time.Parse(SECDateTimeFormat, datetimeStr)
}
