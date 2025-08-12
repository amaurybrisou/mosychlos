package models

// FMP (Financial Modeling Prep) API Response Types

// FMPCompanyProfile represents company profile from FMP
type FMPCompanyProfile struct {
	Symbol            string  `json:"symbol"`
	Price             float64 `json:"price"`
	Beta              float64 `json:"beta"`
	VolAvg            int64   `json:"volAvg"`
	MktCap            int64   `json:"mktCap"`
	LastDiv           float64 `json:"lastDiv"`
	Range             string  `json:"range"`
	Changes           float64 `json:"changes"`
	CompanyName       string  `json:"companyName"`
	Currency          string  `json:"currency"`
	CIK               string  `json:"cik"`
	ISIN              string  `json:"isin"`
	CUSIP             string  `json:"cusip"`
	Exchange          string  `json:"exchange"`
	ExchangeShortName string  `json:"exchangeShortName"`
	Industry          string  `json:"industry"`
	Website           string  `json:"website"`
	Description       string  `json:"description"`
	CEO               string  `json:"ceo"`
	Sector            string  `json:"sector"`
	Country           string  `json:"country"`
	FullTimeEmployees string  `json:"fullTimeEmployees"`
	Phone             string  `json:"phone"`
	Address           string  `json:"address"`
	City              string  `json:"city"`
	State             string  `json:"state"`
	Zip               string  `json:"zip"`
	DcfDiff           float64 `json:"dcfDiff"`
	Dcf               float64 `json:"dcf"`
	Image             string  `json:"image"`
	IpoDate           string  `json:"ipoDate"`
	DefaultImage      bool    `json:"defaultImage"`
	IsEtf             bool    `json:"isEtf"`
	IsActivelyTrading bool    `json:"isActivelyTrading"`
	IsAdr             bool    `json:"isAdr"`
	IsFund            bool    `json:"isFund"`
}

// FMPFinancialStatements represents a collection of financial statements
type FMPFinancialStatements struct {
	Symbol     string                  `json:"symbol"`
	Type       string                  `json:"type"`
	Statements []FMPFinancialStatement `json:"statements"`
}

// FMPFinancialStatement represents a financial statement
type FMPFinancialStatement struct {
	Date                                    string  `json:"date"`
	Symbol                                  string  `json:"symbol"`
	ReportedCurrency                        string  `json:"reportedCurrency"`
	CIK                                     string  `json:"cik"`
	FillingDate                             string  `json:"fillingDate"`
	AcceptedDate                            string  `json:"acceptedDate"`
	CalendarYear                            string  `json:"calendarYear"`
	Period                                  string  `json:"period"`
	Revenue                                 int64   `json:"revenue"`
	CostOfRevenue                           int64   `json:"costOfRevenue"`
	GrossProfit                             int64   `json:"grossProfit"`
	GrossProfitRatio                        float64 `json:"grossProfitRatio"`
	ResearchAndDevelopmentExpenses          int64   `json:"researchAndDevelopmentExpenses"`
	GeneralAndAdministrativeExpenses        int64   `json:"generalAndAdministrativeExpenses"`
	SellingAndMarketingExpenses             int64   `json:"sellingAndMarketingExpenses"`
	SellingGeneralAndAdministrativeExpenses int64   `json:"sellingGeneralAndAdministrativeExpenses"`
	OtherExpenses                           int64   `json:"otherExpenses"`
	OperatingExpenses                       int64   `json:"operatingExpenses"`
	CostAndExpenses                         int64   `json:"costAndExpenses"`
	InterestIncome                          int64   `json:"interestIncome"`
	InterestExpense                         int64   `json:"interestExpense"`
	DepreciationAndAmortization             int64   `json:"depreciationAndAmortization"`
	Ebitda                                  int64   `json:"ebitda"`
	EbitdaRatio                             float64 `json:"ebitdaratio"`
	OperatingIncome                         int64   `json:"operatingIncome"`
	OperatingIncomeRatio                    float64 `json:"operatingIncomeRatio"`
	TotalOtherIncomeExpensesNet             int64   `json:"totalOtherIncomeExpensesNet"`
	IncomeBeforeTax                         int64   `json:"incomeBeforeTax"`
	IncomeBeforeTaxRatio                    float64 `json:"incomeBeforeTaxRatio"`
	IncomeTaxExpense                        int64   `json:"incomeTaxExpense"`
	NetIncome                               int64   `json:"netIncome"`
	NetIncomeRatio                          float64 `json:"netIncomeRatio"`
	Eps                                     float64 `json:"eps"`
	EpsDiluted                              float64 `json:"epsdiluted"`
	WeightedAverageShsOut                   int64   `json:"weightedAverageShsOut"`
	WeightedAverageShsOutDil                int64   `json:"weightedAverageShsOutDil"`
	Link                                    string  `json:"link"`
	FinalLink                               string  `json:"finalLink"`
}

// FMPKeyMetrics represents key financial metrics
type FMPKeyMetrics struct {
	Symbol  string         `json:"symbol"`
	Metrics []FMPKeyMetric `json:"metrics"`
}

// FMPKeyMetric represents individual key metric
type FMPKeyMetric struct {
	Symbol                                 string  `json:"symbol"`
	Date                                   string  `json:"date"`
	Period                                 string  `json:"period"`
	RevenuePerShare                        float64 `json:"revenuePerShare"`
	NetIncomePerShare                      float64 `json:"netIncomePerShare"`
	OperatingCashFlowPerShare              float64 `json:"operatingCashFlowPerShare"`
	FreeCashFlowPerShare                   float64 `json:"freeCashFlowPerShare"`
	CashPerShare                           float64 `json:"cashPerShare"`
	BookValuePerShare                      float64 `json:"bookValuePerShare"`
	TangibleBookValuePerShare              float64 `json:"tangibleBookValuePerShare"`
	ShareholdersEquityPerShare             float64 `json:"shareholdersEquityPerShare"`
	InterestDebtPerShare                   float64 `json:"interestDebtPerShare"`
	MarketCap                              int64   `json:"marketCap"`
	EnterpriseValue                        int64   `json:"enterpriseValue"`
	PeRatio                                float64 `json:"peRatio"`
	PriceToSalesRatio                      float64 `json:"priceToSalesRatio"`
	Pocfratio                              float64 `json:"pocfratio"`
	Pfcffratio                             float64 `json:"pfcffratio"`
	PbRatio                                float64 `json:"pbRatio"`
	PtbRatio                               float64 `json:"ptbRatio"`
	EvToSales                              float64 `json:"evToSales"`
	EnterpriseValueOverEbitda              float64 `json:"enterpriseValueOverEBITDA"`
	EvToOperatingCashFlow                  float64 `json:"evToOperatingCashFlow"`
	EvToFreeCashFlow                       float64 `json:"evToFreeCashFlow"`
	EarningsYield                          float64 `json:"earningsYield"`
	FreeCashFlowYield                      float64 `json:"freeCashFlowYield"`
	DebtToEquity                           float64 `json:"debtToEquity"`
	DebtToAssets                           float64 `json:"debtToAssets"`
	NetDebtToEbitda                        float64 `json:"netDebtToEBITDA"`
	CurrentRatio                           float64 `json:"currentRatio"`
	InterestCoverage                       float64 `json:"interestCoverage"`
	IncomeQuality                          float64 `json:"incomeQuality"`
	DividendYield                          float64 `json:"dividendYield"`
	PayoutRatio                            float64 `json:"payoutRatio"`
	SalesGeneralAndAdministrativeToRevenue float64 `json:"salesGeneralAndAdministrativeToRevenue"`
	ResearchAndDdevelopementToRevenue      float64 `json:"researchAndDdevelopementToRevenue"`
	IntangiblesToTotalAssets               float64 `json:"intangiblesToTotalAssets"`
	CapexToOperatingCashFlow               float64 `json:"capexToOperatingCashFlow"`
	CapexToRevenue                         float64 `json:"capexToRevenue"`
	CapexToDepreciation                    float64 `json:"capexToDepreciation"`
	StockBasedCompensationToRevenue        float64 `json:"stockBasedCompensationToRevenue"`
	GrahamNumber                           float64 `json:"grahamNumber"`
	Roic                                   float64 `json:"roic"`
	ReturnOnTangibleAssets                 float64 `json:"returnOnTangibleAssets"`
	GrahamNetNet                           float64 `json:"grahamNetNet"`
	WorkingCapital                         int64   `json:"workingCapital"`
	TangibleAssetValue                     int64   `json:"tangibleAssetValue"`
	NetCurrentAssetValue                   int64   `json:"netCurrentAssetValue"`
	InvestedCapital                        int64   `json:"investedCapital"`
	AverageReceivables                     int64   `json:"averageReceivables"`
	AveragePayables                        int64   `json:"averagePayables"`
	AverageInventory                       int64   `json:"averageInventory"`
	DaysSalesOutstanding                   float64 `json:"daysSalesOutstanding"`
	DaysPayablesOutstanding                float64 `json:"daysPayablesOutstanding"`
	DaysOfInventoryOnHand                  float64 `json:"daysOfInventoryOnHand"`
	ReceivablesTurnover                    float64 `json:"receivablesTurnover"`
	PayablesTurnover                       float64 `json:"payablesTurnover"`
	InventoryTurnover                      float64 `json:"inventoryTurnover"`
	Roe                                    float64 `json:"roe"`
	CapexPerShare                          float64 `json:"capexPerShare"`
}

// FMPAnalystEstimates represents analyst estimates
type FMPAnalystEstimates struct {
	Symbol    string               `json:"symbol"`
	Estimates []FMPAnalystEstimate `json:"estimates"`
}

// FMPAnalystEstimate represents individual analyst estimate
type FMPAnalystEstimate struct {
	Symbol                        string  `json:"symbol"`
	Date                          string  `json:"date"`
	EstimatedRevenueLow           int64   `json:"estimatedRevenueLow"`
	EstimatedRevenueHigh          int64   `json:"estimatedRevenueHigh"`
	EstimatedRevenueAvg           int64   `json:"estimatedRevenueAvg"`
	EstimatedEbitdaLow            int64   `json:"estimatedEbitdaLow"`
	EstimatedEbitdaHigh           int64   `json:"estimatedEbitdaHigh"`
	EstimatedEbitdaAvg            int64   `json:"estimatedEbitdaAvg"`
	EstimatedEbitLow              int64   `json:"estimatedEbitLow"`
	EstimatedEbitHigh             int64   `json:"estimatedEbitHigh"`
	EstimatedEbitAvg              int64   `json:"estimatedEbitAvg"`
	EstimatedNetIncomeLow         int64   `json:"estimatedNetIncomeLow"`
	EstimatedNetIncomeHigh        int64   `json:"estimatedNetIncomeHigh"`
	EstimatedNetIncomeAvg         int64   `json:"estimatedNetIncomeAvg"`
	EstimatedSgaExpenseLow        int64   `json:"estimatedSgaExpenseLow"`
	EstimatedSgaExpenseHigh       int64   `json:"estimatedSgaExpenseHigh"`
	EstimatedSgaExpenseAvg        int64   `json:"estimatedSgaExpenseAvg"`
	EstimatedEpsAvg               float64 `json:"estimatedEpsAvg"`
	EstimatedEpsHigh              float64 `json:"estimatedEpsHigh"`
	EstimatedEpsLow               float64 `json:"estimatedEpsLow"`
	NumberAnalystEstimatedRevenue int     `json:"numberAnalystEstimatedRevenue"`
	NumberAnalystEstimatedEps     int     `json:"numberAnalystEstimatedEps"`
}

// FMPStockPrice represents stock price information
type FMPStockPrice struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
	Volume int64   `json:"volume"`
}

// FMPMarketCap represents market capitalization data
type FMPMarketCap struct {
	Symbol          string `json:"symbol"`
	Date            string `json:"date"`
	MarketCap       int64  `json:"marketCap"`
	EnterpriseValue int64  `json:"enterpriseValue"`
}

// FMP Error response
type FMPError struct {
	Error struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error"`
}
