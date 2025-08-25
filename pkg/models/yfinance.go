package models

import "time"

// YFinance API Response Types

// StockDataResponse represents Yahoo Finance chart API response
type StockDataResponse struct {
	Chart ChartResponse `json:"chart"`
}

// ChartResponse represents the chart data response
type ChartResponse struct {
	Result []ChartResult `json:"result"`
	Error  *APIError     `json:"error,omitempty"`
}

// ChartResult represents individual chart result
type ChartResult struct {
	Meta       ChartMeta       `json:"meta"`
	Timestamp  []int64         `json:"timestamp"`
	Indicators ChartIndicators `json:"indicators"`
}

// ChartMeta represents chart metadata
type ChartMeta struct {
	Currency             string         `json:"currency"`
	Symbol               string         `json:"symbol"`
	ExchangeName         string         `json:"exchangeName"`
	InstrumentType       string         `json:"instrumentType"`
	FirstTradeDate       int64          `json:"firstTradeDate"`
	RegularMarketTime    int64          `json:"regularMarketTime"`
	Gmtoffset            int            `json:"gmtoffset"`
	Timezone             string         `json:"timezone"`
	ExchangeTimezoneName string         `json:"exchangeTimezoneName"`
	RegularMarketPrice   float64        `json:"regularMarketPrice"`
	ChartPreviousClose   float64        `json:"chartPreviousClose"`
	PreviousClose        float64        `json:"previousClose"`
	Scale                int            `json:"scale"`
	PriceHint            int            `json:"priceHint"`
	CurrentTradingPeriod TradingPeriods `json:"currentTradingPeriod"`
	DataGranularity      string         `json:"dataGranularity"`
	Range                string         `json:"range"`
	ValidRanges          []string       `json:"validRanges"`
}

// TradingPeriods represents trading periods
type TradingPeriods struct {
	Pre     TradingPeriod `json:"pre"`
	Regular TradingPeriod `json:"regular"`
	Post    TradingPeriod `json:"post"`
}

// TradingPeriod represents a trading period
type TradingPeriod struct {
	Timezone  string `json:"timezone"`
	Start     int64  `json:"start"`
	End       int64  `json:"end"`
	Gmtoffset int    `json:"gmtoffset"`
}

// ChartIndicators represents chart indicators
type ChartIndicators struct {
	Quote    []QuoteData    `json:"quote"`
	Adjclose []AdjCloseData `json:"adjclose,omitempty"`
}

// QuoteData represents quote data
type QuoteData struct {
	Open   []float64 `json:"open"`
	Low    []float64 `json:"low"`
	High   []float64 `json:"high"`
	Close  []float64 `json:"close"`
	Volume []int64   `json:"volume"`
}

// AdjCloseData represents adjusted close data
type AdjCloseData struct {
	Adjclose []float64 `json:"adjclose"`
}

// StockInfoResponse represents stock information response
type StockInfoResponse struct {
	Chart ChartResponse `json:"chart"`
}

// DividendsResponse represents dividends response
type DividendsResponse struct {
	Chart ChartResponse `json:"chart"`
}

// FinancialsResponse represents financial data response
type FinancialsResponse struct {
	QuoteSummary QuoteSummaryResponse `json:"quoteSummary"`
}

// QuoteSummaryResponse represents quote summary response
type QuoteSummaryResponse struct {
	Result []QuoteSummaryResult `json:"result"`
	Error  *APIError            `json:"error,omitempty"`
}

// QuoteSummaryResult represents quote summary result
type QuoteSummaryResult struct {
	FinancialData            *FinancialData            `json:"financialData,omitempty"`
	DefaultKeyStatistics     *DefaultKeyStatistics     `json:"defaultKeyStatistics,omitempty"`
	IncomeStatementHistory   *IncomeStatementHistory   `json:"incomeStatementHistory,omitempty"`
	BalanceSheetHistory      *BalanceSheetHistory      `json:"balanceSheetHistory,omitempty"`
	CashflowStatementHistory *CashflowStatementHistory `json:"cashflowStatementHistory,omitempty"`
}

// FinancialData represents financial data
type FinancialData struct {
	CurrentPrice            YFinanceValue `json:"currentPrice,omitempty"`
	TargetHighPrice         YFinanceValue `json:"targetHighPrice,omitempty"`
	TargetLowPrice          YFinanceValue `json:"targetLowPrice,omitempty"`
	TargetMeanPrice         YFinanceValue `json:"targetMeanPrice,omitempty"`
	RecommendationMean      YFinanceValue `json:"recommendationMean,omitempty"`
	RecommendationKey       string        `json:"recommendationKey,omitempty"`
	NumberOfAnalystOpinions YFinanceValue `json:"numberOfAnalystOpinions,omitempty"`
	TotalCash               YFinanceValue `json:"totalCash,omitempty"`
	TotalCashPerShare       YFinanceValue `json:"totalCashPerShare,omitempty"`
	Ebitda                  YFinanceValue `json:"ebitda,omitempty"`
	TotalDebt               YFinanceValue `json:"totalDebt,omitempty"`
	QuickRatio              YFinanceValue `json:"quickRatio,omitempty"`
	CurrentRatio            YFinanceValue `json:"currentRatio,omitempty"`
	TotalRevenue            YFinanceValue `json:"totalRevenue,omitempty"`
	DebtToEquity            YFinanceValue `json:"debtToEquity,omitempty"`
	RevenuePerShare         YFinanceValue `json:"revenuePerShare,omitempty"`
	ReturnOnAssets          YFinanceValue `json:"returnOnAssets,omitempty"`
	ReturnOnEquity          YFinanceValue `json:"returnOnEquity,omitempty"`
	GrossProfits            YFinanceValue `json:"grossProfits,omitempty"`
	FreeCashflow            YFinanceValue `json:"freeCashflow,omitempty"`
	OperatingCashflow       YFinanceValue `json:"operatingCashflow,omitempty"`
	EarningsGrowth          YFinanceValue `json:"earningsGrowth,omitempty"`
	RevenueGrowth           YFinanceValue `json:"revenueGrowth,omitempty"`
	GrossMargins            YFinanceValue `json:"grossMargins,omitempty"`
	EbitdaMargins           YFinanceValue `json:"ebitdaMargins,omitempty"`
	OperatingMargins        YFinanceValue `json:"operatingMargins,omitempty"`
	ProfitMargins           YFinanceValue `json:"profitMargins,omitempty"`
}

// DefaultKeyStatistics represents key statistics
type DefaultKeyStatistics struct {
	MaxAge                       int           `json:"maxAge,omitempty"`
	PriceHint                    YFinanceValue `json:"priceHint,omitempty"`
	EnterpriseValue              YFinanceValue `json:"enterpriseValue,omitempty"`
	ForwardPE                    YFinanceValue `json:"forwardPE,omitempty"`
	ProfitMargins                YFinanceValue `json:"profitMargins,omitempty"`
	FloatShares                  YFinanceValue `json:"floatShares,omitempty"`
	SharesOutstanding            YFinanceValue `json:"sharesOutstanding,omitempty"`
	SharesShort                  YFinanceValue `json:"sharesShort,omitempty"`
	SharesShortPriorMonth        YFinanceValue `json:"sharesShortPriorMonth,omitempty"`
	SharesShortPreviousMonthDate YFinanceValue `json:"sharesShortPreviousMonthDate,omitempty"`
	DateShortInterest            YFinanceValue `json:"dateShortInterest,omitempty"`
	SharesPercentSharesOut       YFinanceValue `json:"sharesPercentSharesOut,omitempty"`
	HeldPercentInsiders          YFinanceValue `json:"heldPercentInsiders,omitempty"`
	HeldPercentInstitutions      YFinanceValue `json:"heldPercentInstitutions,omitempty"`
	ShortRatio                   YFinanceValue `json:"shortRatio,omitempty"`
	ShortPercentOfFloat          YFinanceValue `json:"shortPercentOfFloat,omitempty"`
	Beta                         YFinanceValue `json:"beta,omitempty"`
	ImpliedSharesOutstanding     YFinanceValue `json:"impliedSharesOutstanding,omitempty"`
	MorningStarOverallRating     YFinanceValue `json:"morningStarOverallRating,omitempty"`
	MorningStarRiskRating        YFinanceValue `json:"morningStarRiskRating,omitempty"`
	Category                     string        `json:"category,omitempty"`
	BookValue                    YFinanceValue `json:"bookValue,omitempty"`
	PriceToBook                  YFinanceValue `json:"priceToBook,omitempty"`
	AnnualReportExpenseRatio     YFinanceValue `json:"annualReportExpenseRatio,omitempty"`
	YtdReturn                    YFinanceValue `json:"ytdReturn,omitempty"`
	Beta3Year                    YFinanceValue `json:"beta3Year,omitempty"`
	TotalAssets                  YFinanceValue `json:"totalAssets,omitempty"`
	Yield                        YFinanceValue `json:"yield,omitempty"`
	FundFamily                   string        `json:"fundFamily,omitempty"`
	FundInceptionDate            YFinanceValue `json:"fundInceptionDate,omitempty"`
	LegalType                    string        `json:"legalType,omitempty"`
	ThreeYearAverageReturn       YFinanceValue `json:"threeYearAverageReturn,omitempty"`
	FiveYearAverageReturn        YFinanceValue `json:"fiveYearAverageReturn,omitempty"`
	PriceToSalesTrailing12Months YFinanceValue `json:"priceToSalesTrailing12Months,omitempty"`
	LastFiscalYearEnd            YFinanceValue `json:"lastFiscalYearEnd,omitempty"`
	NextFiscalYearEnd            YFinanceValue `json:"nextFiscalYearEnd,omitempty"`
	MostRecentQuarter            YFinanceValue `json:"mostRecentQuarter,omitempty"`
	EarningsQuarterlyGrowth      YFinanceValue `json:"earningsQuarterlyGrowth,omitempty"`
	RevenueQuarterlyGrowth       YFinanceValue `json:"revenueQuarterlyGrowth,omitempty"`
	NetIncomeToCommon            YFinanceValue `json:"netIncomeToCommon,omitempty"`
	TrailingEps                  YFinanceValue `json:"trailingEps,omitempty"`
	ForwardEps                   YFinanceValue `json:"forwardEps,omitempty"`
	PegRatio                     YFinanceValue `json:"pegRatio,omitempty"`
	LastSplitFactor              string        `json:"lastSplitFactor,omitempty"`
	LastSplitDate                YFinanceValue `json:"lastSplitDate,omitempty"`
	EnterpriseToRevenue          YFinanceValue `json:"enterpriseToRevenue,omitempty"`
	EnterpriseToEbitda           YFinanceValue `json:"enterpriseToEbitda,omitempty"`
	Yield52WeekHigh              YFinanceValue `json:"52WeekChange,omitempty"`
	SandP52WeekChange            YFinanceValue `json:"SandP52WeekChange,omitempty"`
}

// IncomeStatementHistory represents income statement history
type IncomeStatementHistory struct {
	IncomeStatementHistory []IncomeStatementData `json:"incomeStatementHistory"`
	MaxAge                 int                   `json:"maxAge,omitempty"`
}

// IncomeStatementData represents income statement data
type IncomeStatementData struct {
	MaxAge                            int           `json:"maxAge,omitempty"`
	EndDate                           YFinanceValue `json:"endDate,omitempty"`
	TotalRevenue                      YFinanceValue `json:"totalRevenue,omitempty"`
	CostOfRevenue                     YFinanceValue `json:"costOfRevenue,omitempty"`
	GrossProfit                       YFinanceValue `json:"grossProfit,omitempty"`
	ResearchDevelopment               YFinanceValue `json:"researchDevelopment,omitempty"`
	SellingGeneralAdministrative      YFinanceValue `json:"sellingGeneralAdministrative,omitempty"`
	NonRecurring                      YFinanceValue `json:"nonRecurring,omitempty"`
	OtherOperatingExpenses            YFinanceValue `json:"otherOperatingExpenses,omitempty"`
	TotalOperatingExpenses            YFinanceValue `json:"totalOperatingExpenses,omitempty"`
	OperatingIncome                   YFinanceValue `json:"operatingIncome,omitempty"`
	TotalOtherIncomeExpenseNet        YFinanceValue `json:"totalOtherIncomeExpenseNet,omitempty"`
	Ebit                              YFinanceValue `json:"ebit,omitempty"`
	InterestExpense                   YFinanceValue `json:"interestExpense,omitempty"`
	IncomeBeforeTax                   YFinanceValue `json:"incomeBeforeTax,omitempty"`
	IncomeTaxExpense                  YFinanceValue `json:"incomeTaxExpense,omitempty"`
	MinorityInterest                  YFinanceValue `json:"minorityInterest,omitempty"`
	NetIncomeFromContinuingOps        YFinanceValue `json:"netIncomeFromContinuingOps,omitempty"`
	DiscontinuedOperations            YFinanceValue `json:"discontinuedOperations,omitempty"`
	ExtraordinaryItems                YFinanceValue `json:"extraordinaryItems,omitempty"`
	EffectOfAccountingCharges         YFinanceValue `json:"effectOfAccountingCharges,omitempty"`
	OtherItems                        YFinanceValue `json:"otherItems,omitempty"`
	NetIncome                         YFinanceValue `json:"netIncome,omitempty"`
	NetIncomeApplicableToCommonShares YFinanceValue `json:"netIncomeApplicableToCommonShares,omitempty"`
}

// BalanceSheetHistory represents balance sheet history
type BalanceSheetHistory struct {
	BalanceSheetStatements []BalanceSheetData `json:"balanceSheetStatements"`
	MaxAge                 int                `json:"maxAge,omitempty"`
}

// BalanceSheetData represents balance sheet data
type BalanceSheetData struct {
	MaxAge                  int           `json:"maxAge,omitempty"`
	EndDate                 YFinanceValue `json:"endDate,omitempty"`
	Cash                    YFinanceValue `json:"cash,omitempty"`
	ShortTermInvestments    YFinanceValue `json:"shortTermInvestments,omitempty"`
	NetReceivables          YFinanceValue `json:"netReceivables,omitempty"`
	Inventory               YFinanceValue `json:"inventory,omitempty"`
	OtherCurrentAssets      YFinanceValue `json:"otherCurrentAssets,omitempty"`
	TotalCurrentAssets      YFinanceValue `json:"totalCurrentAssets,omitempty"`
	LongTermInvestments     YFinanceValue `json:"longTermInvestments,omitempty"`
	PropertyPlantEquipment  YFinanceValue `json:"propertyPlantEquipment,omitempty"`
	OtherAssets             YFinanceValue `json:"otherAssets,omitempty"`
	TotalAssets             YFinanceValue `json:"totalAssets,omitempty"`
	AccountsPayable         YFinanceValue `json:"accountsPayable,omitempty"`
	ShortLongTermDebt       YFinanceValue `json:"shortLongTermDebt,omitempty"`
	OtherCurrentLiab        YFinanceValue `json:"otherCurrentLiab,omitempty"`
	LongTermDebt            YFinanceValue `json:"longTermDebt,omitempty"`
	OtherLiab               YFinanceValue `json:"otherLiab,omitempty"`
	MinorityInterest        YFinanceValue `json:"minorityInterest,omitempty"`
	TotalCurrentLiabilities YFinanceValue `json:"totalCurrentLiabilities,omitempty"`
	TotalLiab               YFinanceValue `json:"totalLiab,omitempty"`
	CommonStock             YFinanceValue `json:"commonStock,omitempty"`
	RetainedEarnings        YFinanceValue `json:"retainedEarnings,omitempty"`
	TreasuryStock           YFinanceValue `json:"treasuryStock,omitempty"`
	OtherStockholderEquity  YFinanceValue `json:"otherStockholderEquity,omitempty"`
	TotalStockholderEquity  YFinanceValue `json:"totalStockholderEquity,omitempty"`
	NetTangibleAssets       YFinanceValue `json:"netTangibleAssets,omitempty"`
}

// CashflowStatementHistory represents cashflow statement history
type CashflowStatementHistory struct {
	CashflowStatements []CashflowStatementData `json:"cashflowStatements"`
	MaxAge             int                     `json:"maxAge,omitempty"`
}

// CashflowStatementData represents cashflow statement data
type CashflowStatementData struct {
	MaxAge                                int           `json:"maxAge,omitempty"`
	EndDate                               YFinanceValue `json:"endDate,omitempty"`
	NetIncome                             YFinanceValue `json:"netIncome,omitempty"`
	Depreciation                          YFinanceValue `json:"depreciation,omitempty"`
	ChangeToNetincome                     YFinanceValue `json:"changeToNetincome,omitempty"`
	ChangeToAccountReceivables            YFinanceValue `json:"changeToAccountReceivables,omitempty"`
	ChangeToLiabilities                   YFinanceValue `json:"changeToLiabilities,omitempty"`
	ChangeToInventory                     YFinanceValue `json:"changeToInventory,omitempty"`
	ChangeToOperatingActivities           YFinanceValue `json:"changeToOperatingActivities,omitempty"`
	TotalCashFromOperatingActivities      YFinanceValue `json:"totalCashFromOperatingActivities,omitempty"`
	CapitalExpenditures                   YFinanceValue `json:"capitalExpenditures,omitempty"`
	Investments                           YFinanceValue `json:"investments,omitempty"`
	OtherCashflowsFromInvestingActivities YFinanceValue `json:"otherCashflowsFromInvestingActivities,omitempty"`
	TotalCashflowsFromInvestingActivities YFinanceValue `json:"totalCashflowsFromInvestingActivities,omitempty"`
	DividendsPaid                         YFinanceValue `json:"dividendsPaid,omitempty"`
	NetBorrowings                         YFinanceValue `json:"netBorrowings,omitempty"`
	OtherCashflowsFromFinancingActivities YFinanceValue `json:"otherCashflowsFromFinancingActivities,omitempty"`
	TotalCashFromFinancingActivities      YFinanceValue `json:"totalCashFromFinancingActivities,omitempty"`
	ChangeInCash                          YFinanceValue `json:"changeInCash,omitempty"`
	RepurchaseOfStock                     YFinanceValue `json:"repurchaseOfStock,omitempty"`
	IssuanceOfStock                       YFinanceValue `json:"issuanceOfStock,omitempty"`
}

// MarketDataResponse represents market data response
type MarketDataResponse struct {
	QuoteResponse QuoteResponse `json:"quoteResponse"`
}

// QuoteResponse represents quote response
type QuoteResponse struct {
	Result []QuoteResult `json:"result"`
	Error  *APIError     `json:"error,omitempty"`
}

// QuoteResult represents individual quote result
type QuoteResult struct {
	Language                          string  `json:"language,omitempty"`
	Region                            string  `json:"region,omitempty"`
	QuoteType                         string  `json:"quoteType,omitempty"`
	TypeDisp                          string  `json:"typeDisp,omitempty"`
	QuoteSourceName                   string  `json:"quoteSourceName,omitempty"`
	Triggerable                       bool    `json:"triggerable,omitempty"`
	CustomPriceAlertConfidence        string  `json:"customPriceAlertConfidence,omitempty"`
	Currency                          string  `json:"currency,omitempty"`
	MarketState                       string  `json:"marketState,omitempty"`
	RegularMarketChangePercent        float64 `json:"regularMarketChangePercent,omitempty"`
	RegularMarketPrice                float64 `json:"regularMarketPrice,omitempty"`
	Exchange                          string  `json:"exchange,omitempty"`
	ShortName                         string  `json:"shortName,omitempty"`
	LongName                          string  `json:"longName,omitempty"`
	MessageBoardID                    string  `json:"messageBoardId,omitempty"`
	ExchangeTimezoneName              string  `json:"exchangeTimezoneName,omitempty"`
	ExchangeTimezoneShortName         string  `json:"exchangeTimezoneShortName,omitempty"`
	GmtOffSetMilliseconds             int64   `json:"gmtOffSetMilliseconds,omitempty"`
	Market                            string  `json:"market,omitempty"`
	EsgPopulated                      bool    `json:"esgPopulated,omitempty"`
	FirstTradeDateMilliseconds        int64   `json:"firstTradeDateMilliseconds,omitempty"`
	PriceHint                         int     `json:"priceHint,omitempty"`
	RegularMarketChange               float64 `json:"regularMarketChange,omitempty"`
	RegularMarketTime                 int64   `json:"regularMarketTime,omitempty"`
	RegularMarketDayHigh              float64 `json:"regularMarketDayHigh,omitempty"`
	RegularMarketDayRange             string  `json:"regularMarketDayRange,omitempty"`
	RegularMarketDayLow               float64 `json:"regularMarketDayLow,omitempty"`
	RegularMarketVolume               int64   `json:"regularMarketVolume,omitempty"`
	RegularMarketPreviousClose        float64 `json:"regularMarketPreviousClose,omitempty"`
	Bid                               float64 `json:"bid,omitempty"`
	Ask                               float64 `json:"ask,omitempty"`
	BidSize                           int     `json:"bidSize,omitempty"`
	AskSize                           int     `json:"askSize,omitempty"`
	FullExchangeName                  string  `json:"fullExchangeName,omitempty"`
	FinancialCurrency                 string  `json:"financialCurrency,omitempty"`
	RegularMarketOpen                 float64 `json:"regularMarketOpen,omitempty"`
	AverageDailyVolume3Month          int64   `json:"averageDailyVolume3Month,omitempty"`
	AverageDailyVolume10Day           int64   `json:"averageDailyVolume10Day,omitempty"`
	FiftyTwoWeekLowChange             float64 `json:"fiftyTwoWeekLowChange,omitempty"`
	FiftyTwoWeekLowChangePercent      float64 `json:"fiftyTwoWeekLowChangePercent,omitempty"`
	FiftyTwoWeekRange                 string  `json:"fiftyTwoWeekRange,omitempty"`
	FiftyTwoWeekHighChange            float64 `json:"fiftyTwoWeekHighChange,omitempty"`
	FiftyTwoWeekHighChangePercent     float64 `json:"fiftyTwoWeekHighChangePercent,omitempty"`
	FiftyTwoWeekLow                   float64 `json:"fiftyTwoWeekLow,omitempty"`
	FiftyTwoWeekHigh                  float64 `json:"fiftyTwoWeekHigh,omitempty"`
	DividendDate                      int64   `json:"dividendDate,omitempty"`
	EarningsTimestamp                 int64   `json:"earningsTimestamp,omitempty"`
	EarningsTimestampStart            int64   `json:"earningsTimestampStart,omitempty"`
	EarningsTimestampEnd              int64   `json:"earningsTimestampEnd,omitempty"`
	TrailingAnnualDividendRate        float64 `json:"trailingAnnualDividendRate,omitempty"`
	TrailingPE                        float64 `json:"trailingPE,omitempty"`
	TrailingAnnualDividendYield       float64 `json:"trailingAnnualDividendYield,omitempty"`
	EpsTrailingTwelveMonths           float64 `json:"epsTrailingTwelveMonths,omitempty"`
	EpsForward                        float64 `json:"epsForward,omitempty"`
	EpsCurrentYear                    float64 `json:"epsCurrentYear,omitempty"`
	PriceEpsCurrentYear               float64 `json:"priceEpsCurrentYear,omitempty"`
	SharesOutstanding                 int64   `json:"sharesOutstanding,omitempty"`
	BookValue                         float64 `json:"bookValue,omitempty"`
	FiftyDayAverage                   float64 `json:"fiftyDayAverage,omitempty"`
	FiftyDayAverageChange             float64 `json:"fiftyDayAverageChange,omitempty"`
	FiftyDayAverageChangePercent      float64 `json:"fiftyDayAverageChangePercent,omitempty"`
	TwoHundredDayAverage              float64 `json:"twoHundredDayAverage,omitempty"`
	TwoHundredDayAverageChange        float64 `json:"twoHundredDayAverageChange,omitempty"`
	TwoHundredDayAverageChangePercent float64 `json:"twoHundredDayAverageChangePercent,omitempty"`
	MarketCap                         int64   `json:"marketCap,omitempty"`
	ForwardPE                         float64 `json:"forwardPE,omitempty"`
	PriceToBook                       float64 `json:"priceToBook,omitempty"`
	SourceInterval                    int     `json:"sourceInterval,omitempty"`
	ExchangeDataDelayedBy             int     `json:"exchangeDataDelayedBy,omitempty"`
	Tradeable                         bool    `json:"tradeable,omitempty"`
	CryptoTradeable                   bool    `json:"cryptoTradeable,omitempty"`
	Symbol                            string  `json:"symbol"`
}

// YFinanceValue represents a Yahoo Finance value that can be either a number or null
type YFinanceValue struct {
	Raw float64 `json:"raw,omitempty"`
	Fmt string  `json:"fmt,omitempty"`
}

// APIError represents API error response
type APIError struct {
	Code        string `json:"code,omitempty"`
	Description string `json:"description,omitempty"`
}

// Helper methods for YFinanceValue

// Float64 returns the raw float64 value
func (v YFinanceValue) Float64() float64 {
	return v.Raw
}

// String returns the formatted string value
func (v YFinanceValue) String() string {
	return v.Fmt
}

// IsValid returns true if the value has valid data
func (v YFinanceValue) IsValid() bool {
	return v.Fmt != "" || v.Raw != 0
}

// UnixToTime function to convert Unix timestamp to time.Time
func UnixToTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}
