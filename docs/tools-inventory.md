# Mosychlos Tools Inventory

This document provides a comprehensive list of all tools in the Mosychlos ecosystem. Tools are **data providers and service integrations** that supply data to engines for analysis.

## **🎉 Current Status: 9 Active Tools**

As of **August 16, 2025**, Mosychlos has **9 fully implemented and active tools**:

- **5 YFinance Tools**: Complete Yahoo Finance market data suite ⭐ **FIXED & VALIDATED**
- **2 FMP Tools**: Financial Modeling Prep data and analyst estimates
- **1 FRED Tool**: Federal Reserve economic indicators
- **1 NewsAPI Tool**: Market news and sentiment data
- **1 Weather Tool**: Basic weather data integration

All tools are registered, configured, cached, and monitored with institutional-grade error handling.

## **Architecture Overview**

```
Tools (Data Providers)          Engines (Analysis Logic)
┌─────────────────────┐        ┌──────────────────────┐
│ yfinance/           │───────►│ Financial Analysis   │
│ ├── stock_data.go   │        │ Engine               │
│ ├── stock_info.go   │        └──────────────────────┘
│ └── options.go      │        ┌──────────────────────┐
└─────────────────────┘        │ Risk Analysis        │
┌─────────────────────┐───────►│ Engine               │
│ fmp/                │        └──────────────────────┘
│ ├── fundamentals.go │        ┌──────────────────────┐
│ └── estimates.go    │───────►│ Investment Committee │
└─────────────────────┘        │ Engine               │
                               └──────────────────────┘
```

**Tools** provide raw data → **Engines** perform analysis using multiple tools

## **✅ Current Tools (Implemented)**

### **fmp** - Financial Modeling Prep

- **Provider**: Financial Modeling Prep API
- **Functions**: Company profiles, financial metrics, sector data, ratios
- **Files**: `internal/tools/fmp/` directory
- **Status**: ✅ Active and registered

### **fred** - Federal Reserve Economic Data

- **Provider**: Federal Reserve Bank of St. Louis
- **Functions**: Economic indicators, interest rates, GDP, inflation
- **Files**: `internal/tools/fred/` directory
- **Status**: ✅ Active and registered

### **newsapi** - News API

- **Provider**: NewsAPI.org
- **Functions**: Market news, sentiment data, headlines
- **Files**: `internal/tools/newsapi/` directory
- **Status**: ✅ Active and registered

### **fmp_estimates** - FMP Analyst Estimates

- **Provider**: Financial Modeling Prep Estimates API
- **Functions**: Analyst forecasts, earnings estimates, price targets
- **Files**: `internal/tools/fmp_estimates/` directory
- **Status**: ✅ Active and registered

### **yfinance** - Yahoo Finance (NEW! ⭐)

- **Provider**: Yahoo Finance API
- **Functions**: Complete market data suite across multiple asset classes
- **Files**: `internal/tools/yfinance/` directory
- **Status**: ✅ Active and registered (5 specialized tools)

#### **YFinance Tool Suite**:

```
internal/tools/yfinance/
├── stock_data.go      # ✅ Real-time and historical OHLCV data
├── stock_info.go      # ✅ Company profiles, metrics, and fundamentals [RECENTLY FIXED]
├── dividends.go       # ✅ Dividend history, yields, and distribution analysis
├── financials.go      # ✅ Financial statements (income, balance, cash flow)
├── market_data.go     # ✅ Market indices, sector performance, broad market
├── yfinance_test.go   # ✅ Comprehensive test suite [ALL TESTS PASSING]
└── README.md         # ✅ Documentation and usage examples
```

**Individual YFinance Tools**:

- `yfinance_stock_data` - Retrieve real-time quotes and historical OHLCV price data
- `yfinance_stock_info` - **[RECENTLY FIXED]** Fetch comprehensive company profiles and financial metrics
- `yfinance_dividends` - Analyze dividend payments, yields, and distribution history
- `yfinance_financials` - Access complete financial statements and reports
- `yfinance_market_data` - Monitor broad market indices and sector performance

**Recent Technical Improvements**:

- **Stock Info Tool Fixed**: Resolved 401 authentication errors by switching from `/v10/finance/quoteSummary/` to `/v8/finance/chart/` endpoint
- **Enhanced HTTP Headers**: Added comprehensive browser-like headers to avoid API restrictions
- **Removed External Dependencies**: Eliminated `piquette/finance-go` dependency for better reliability
- **All Tests Passing**: Complete YFinance test suite validates all 5 tools work correctly

### **weather** - Weather Data

- **Provider**: Weather API service
- **Functions**: Weather conditions (basic implementation)
- **Files**: `internal/tools/weather/` directory
- **Status**: ✅ Active and registered (basic functionality)

## **🚧 Future Tools (Potential Implementation)**

### **yfinance_extensions** - Extended Yahoo Finance Tools

Based on finance-go library analysis, we could add these specialized YFinance tools:

```
internal/tools/yfinance/
├── options_chains.go  # Options chains, strikes, expiration dates
├── options_straddles.go # Put/call straddles for volatility strategies
├── etf_data.go        # ETF-specific data with expense ratios, holdings
├── mutualfund_data.go # Mutual fund NAV, manager info, performance
├── crypto_data.go     # Cryptocurrency data with blockchain metrics
├── forex_data.go      # Currency pairs for international portfolios
├── futures_data.go    # Futures contracts for hedging and commodities
└── index_analysis.go  # Enhanced index analysis and sector weightings
```

### **binance** - Binance Cryptocurrency Exchange

```
internal/tools/binance/
├── spot_data.go       # Spot cryptocurrency prices
├── futures_data.go    # Futures and derivatives data
├── market_depth.go    # Order book and market depth
├── trading_data.go    # Volume and trading statistics
├── wallet_data.go     # Portfolio and balance data
└── tool.go           # Tool interface implementation
```

### **alpha_vantage** - Alpha Vantage Financial Data

```
internal/tools/alpha_vantage/
├── stock_data.go      # Stock prices and technical indicators
├── forex_data.go      # Foreign exchange rates
├── crypto_data.go     # Cryptocurrency data
├── economic_data.go   # Economic indicators
└── tool.go           # Tool interface implementation
```

### **polygon** - Polygon.io Market Data

```
internal/tools/polygon/
├── real_time.go       # Real-time market data
├── historical.go      # Historical price data
├── options.go         # Options data
├── forex.go          # Forex market data
├── crypto.go         # Cryptocurrency data
└── tool.go           # Tool interface implementation
```

### **iex_cloud** - IEX Cloud Financial Data

```
internal/tools/iex_cloud/
├── stock_data.go      # Stock prices and fundamentals
├── market_data.go     # Market statistics
├── news_data.go       # Financial news
├── economic_data.go   # Economic indicators
└── tool.go           # Tool interface implementation
```

### **sec_edgar** - SEC EDGAR Filings

```
internal/tools/sec_edgar/
├── filings.go         # SEC filing retrieval and parsing
├── company_facts.go   # Company fact extraction
├── insider_trading.go # Insider trading data
├── ownership.go       # Institutional ownership
└── tool.go           # Tool interface implementation
```

### **quandl** - Quandl Financial Data

```
internal/tools/quandl/
├── economic_data.go   # Economic datasets
├── commodity_data.go  # Commodity prices
├── financial_data.go  # Financial market data
├── alternative_data.go # Alternative datasets
└── tool.go           # Tool interface implementation
```

### **morningstar** - Morningstar Investment Data

```
internal/tools/morningstar/
├── mutual_funds.go    # Mutual fund data
├── etf_data.go       # ETF information
├── stock_analysis.go  # Stock analysis and ratings
├── portfolio_tools.go # Portfolio analytics
└── tool.go           # Tool interface implementation
```

### **bloomberg** - Bloomberg Terminal API

```
internal/tools/bloomberg/
├── market_data.go     # Real-time market data
├── news.go           # Bloomberg news and research
├── analytics.go      # Bloomberg analytics
├── economic_data.go  # Economic calendar and data
└── tool.go           # Tool interface implementation
```

### **refinitiv** - Refinitiv/Reuters Data

```
internal/tools/refinitiv/
├── market_data.go     # Market prices and data
├── news.go           # Reuters news feed
├── research.go       # Analyst research
├── economic_data.go  # Economic indicators
└── tool.go           # Tool interface implementation
```

## **Tool Implementation Pattern**

Each tool follows this structure:

### **Directory Structure**

```
internal/tools/{tool_name}/
├── {function1}.go     # Specific data function
├── {function2}.go     # Another data function
├── {function3}.go     # Additional functions
├── {tool_name}_test.go # Tests
├── README.md         # Documentation
└── tool.go           # models.Tool interface implementation
```

### **Tool Interface Implementation**

Every tool must implement the `models.Tool` interface:

```go
type Tool interface {
    Name() string                              // Tool identifier
    Key() keys.Key                            // Unique key
    Description() string                      // What the tool does
    Definition() ToolDef                      // OpenAI function definition
    Tags() []string                          // Categorization tags
    Run(ctx context.Context, args string) (string, error) // Execute function
}
```

### **Configuration Pattern**

```yaml
tools:
  tool_name:
    api_key: '${TOOL_NAME_API_KEY}'
    base_url: 'https://api.provider.com'
    cache_enable: true
    max_daily: 1000
    timeout: 30
```

## **Tool Categories**

### **Market Data Providers**

- `yfinance` - Free stock market data
- `alpha_vantage` - Market data and technical indicators
- `polygon` - Professional market data feeds
- `iex_cloud` - Real-time and historical data

### **Financial Data Services**

- `fmp` - Financial statements and metrics
- `morningstar` - Investment research and analytics
- `refinitiv` - Professional financial data
- `bloomberg` - Premium market data and analytics

### **Economic Data Sources**

- `fred` - Federal Reserve economic data
- `quandl` - Economic and financial datasets

### **Alternative Data**

- `newsapi` - Market news and sentiment
- `sec_edgar` - SEC filings and regulatory data
- `weather` - Weather impact analysis

### **Cryptocurrency Data**

- `binance` - Cryptocurrency exchange data

## **Current System Status**

### **Active Tools: 9 Total**

- **3 Financial Data**: fmp, fmp_estimates, fred
- **1 News Data**: newsapi
- **5 Market Data**: yfinance suite (stock_data, stock_info, dividends, financials, market_data) **[ALL VALIDATED]**
- **1 Weather Data**: weather

### **Tool Registration**

All tools are properly registered in `internal/tools/tools.go`:

```go
// Registered and active
✅ fmp                  -> fmp tool
✅ fred                 -> fred tool
✅ newsapi              -> news_api tool
✅ fmp_estimates        -> fmp_analyst_estimates tool
✅ yfinance (5 tools)   -> yfinance_* tools [ALL TESTS PASSING]
✅ weather             -> weather tool
```

### **Configuration Support**

All active tools have configuration in `config/config.default.yaml`:

- FMP: API key, caching, rate limits
- FRED: API key, caching
- NewsAPI: API key, caching
- YFinance: Base URL, caching, request limits (no API key required)

## **Implementation Priority**

### **Phase 1: Completed ✅**

1. ~~`yfinance` - Free, reliable stock data~~ **DONE**
2. ~~Core market data functionality~~ **DONE**
3. ~~Tool registration and configuration~~ **DONE**

### **Phase 2: YFinance Extensions (Next)**

### **Phase 2: YFinance Extensions (Next)**

1. `yfinance_options_chains` - Options data for risk management
2. `yfinance_etf_data` - ETF analysis and expense ratios
3. `yfinance_crypto_data` - Cryptocurrency market data

### **Phase 3: Advanced Financial Data**

1. `polygon` - Professional market data
2. `alpha_vantage` - Technical indicators
3. `iex_cloud` - Real-time market data
4. `sec_edgar` - Regulatory filings
5. `morningstar` - Investment analytics

### **Phase 4: Alternative Data**

1. `quandl` - Economic datasets
2. `binance` - Cryptocurrency exchange data
3. Enhanced weather integration

### **Phase 5: Premium Services**

1. `bloomberg` - Premium market data
2. `refinitiv` - Professional analytics

## **Next Steps**

### **For YFinance Extensions**:

1. **Choose Extension**: Start with `yfinance_options_chains` or `yfinance_etf_data`
2. **Follow Guide**: Use [tool-generation.md](../.github/instructions/tool-generation.md)
3. **Add to YFinance Package**: Extend `internal/tools/yfinance/` directory
4. **Update Registration**: Add new tool keys and registration in `tools.go`
5. **Test Integration**: Create comprehensive tests following existing patterns

### **For New Tool Providers**:

1. **Choose Provider**: Select from planned tools (polygon, alpha_vantage, etc.)
2. **Create Directory**: `internal/tools/{provider_name}/`
3. **Implement Functions**: Create specific .go files for each data endpoint
4. **Add Configuration**: Update `config.go` with tool config
5. **Register Tool**: Add to `tools.go` and `keys.go`
6. **Test Integration**: Create comprehensive tests

### **Current Status Summary**:

- **9/9 tools** successfully implemented and registered **[RECENTLY VALIDATED]**
- **YFinance suite** provides comprehensive stock market data coverage with all API issues resolved
- **Stock Info Tool Fixed**: Resolved authentication errors and switched to reliable `/v8/finance/chart/` endpoint
- **All Tests Passing**: Complete validation of tool functionality and reliability
- **Ready for extensions** with options, ETF, and cryptocurrency data
- **Tool architecture** proven and scalable for additional providers

**Latest Achievements (August 16, 2025)**:

- ✅ Fixed YFinance stock info 401 authentication errors
- ✅ Removed problematic `piquette/finance-go` dependency
- ✅ Enhanced HTTP headers for better Yahoo Finance API compatibility
- ✅ All 9 tools now working reliably with comprehensive test coverage

## **Tool vs Engine Distinction**

**🔧 Tools** (This Document):

- Data providers (yfinance, fmp, fred)
- Service integrations (OpenAI, Binance)
- Raw data access and retrieval

**⚙️ Engines** (Separate System):

- Financial analysis logic
- Risk calculations
- Investment decision making
- Portfolio optimization

Tools **supply data** → Engines **perform analysis** → Professional insights

---

_This inventory focuses on data provider tools. Analysis engines are documented separately._
