# YFinance Tool

Yahoo Finance data provider tool for the Mosychlos portfolio management system.

## Overview

The YFinance tool provides access to Yahoo Finance data through multiple specialized sub-tools:

- **Stock Data**: Real-time and historical stock prices (OHLCV data)
- **Stock Info**: Company information and key metrics
- **Dividends**: Dividend history and yield information
- **Financials**: Financial statements (income, balance sheet, cash flow)
- **Market Data**: Market indices and sector performance

## Configuration

```yaml
tools:
  yfinance:
    base_url: 'https://query1.finance.yahoo.com' # Optional, uses default if not specified
    cache_enable: true
    max_daily: 1000
    timeout: 30 # seconds
```

## Environment Variables

No API key required - Yahoo Finance provides free access to basic financial data.

## Available Tools

### 1. Stock Data Tool (`yfinance_stock_data`)

Get real-time and historical stock price data.

**Function Call:**

```json
{
  "name": "yfinance_stock_data",
  "arguments": {
    "symbol": "AAPL",
    "period": "1y",
    "interval": "1d"
  }
}
```

**Parameters:**

- `symbol` (required): Stock symbol (e.g., 'AAPL', 'MSFT', 'SPY')
- `period` (optional): Time period - 1d, 5d, 1mo, 3mo, 6mo, 1y, 2y, 5y, 10y, ytd, max (default: 1y)
- `interval` (optional): Data interval - 1m, 2m, 5m, 15m, 30m, 60m, 90m, 1h, 1d, 5d, 1wk, 1mo, 3mo (default: 1d)

### 2. Stock Info Tool (`yfinance_stock_info`)

Get comprehensive company information and key statistics.

**Function Call:**

```json
{
  "name": "yfinance_stock_info",
  "arguments": {
    "symbol": "GOOGL"
  }
}
```

**Parameters:**

- `symbol` (required): Stock symbol (e.g., 'AAPL', 'MSFT', 'GOOGL')

### 3. Dividends Tool (`yfinance_dividends`)

Get dividend history and yield information.

**Function Call:**

```json
{
  "name": "yfinance_dividends",
  "arguments": {
    "symbol": "JNJ",
    "period": "5y"
  }
}
```

**Parameters:**

- `symbol` (required): Stock symbol (e.g., 'AAPL', 'MSFT', 'JNJ')
- `period` (optional): Time period - 1y, 2y, 5y, 10y, max (default: 5y)

### 4. Financials Tool (`yfinance_financials`)

Get financial statements including income statement, balance sheet, and cash flow.

**Function Call:**

```json
{
  "name": "yfinance_financials",
  "arguments": {
    "symbol": "TSLA",
    "statement_type": "income",
    "frequency": "annual"
  }
}
```

**Parameters:**

- `symbol` (required): Stock symbol (e.g., 'AAPL', 'MSFT', 'GOOGL')
- `statement_type` (optional): Type of statement - income, balance, cashflow (default: income)
- `frequency` (optional): Frequency - annual, quarterly (default: annual)

### 5. Market Data Tool (`yfinance_market_data`)

Get market indices and sector performance data.

**Function Call:**

```json
{
  "name": "yfinance_market_data",
  "arguments": {
    "symbols": ["^GSPC", "^DJI", "^IXIC", "SPY", "QQQ"],
    "period": "1d"
  }
}
```

**Parameters:**

- `symbols` (required): Array of market symbols (max 20 items)
- `period` (optional): Time period - 1d, 5d, 1mo, 3mo, 6mo, 1y, 2y, 5y (default: 1d)

## Response Format

All tools return JSON responses with this structure:

```json
{
  "status": "success",
  "symbol": "AAPL",
  "data": {
    // Tool-specific data structure
  },
  "metadata": {
    "timestamp": "2025-08-16T10:00:00Z",
    "source": "yahoo_finance",
    "tool": "yfinance_stock_data"
  }
}
```

## Usage Examples

### Get Apple Stock Price History

```json
{
  "name": "yfinance_stock_data",
  "arguments": {
    "symbol": "AAPL",
    "period": "1mo",
    "interval": "1d"
  }
}
```

### Get Microsoft Company Information

```json
{
  "name": "yfinance_stock_info",
  "arguments": {
    "symbol": "MSFT"
  }
}
```

### Get Johnson & Johnson Dividend History

```json
{
  "name": "yfinance_dividends",
  "arguments": {
    "symbol": "JNJ",
    "period": "10y"
  }
}
```

### Get Tesla Financial Statements

```json
{
  "name": "yfinance_financials",
  "arguments": {
    "symbol": "TSLA",
    "statement_type": "income",
    "frequency": "quarterly"
  }
}
```

### Get Market Indices Performance

```json
{
  "name": "yfinance_market_data",
  "arguments": {
    "symbols": ["^GSPC", "^DJI", "^IXIC"],
    "period": "1mo"
  }
}
```

## Error Handling

The tools handle various error scenarios:

- **Missing required parameters**: Returns descriptive error messages
- **Invalid symbols**: Returns error for non-existent or invalid stock symbols
- **Network timeouts**: 30-second timeout with proper error handling
- **API failures**: Graceful handling of Yahoo Finance API errors
- **Invalid response formats**: Validates response structure before processing

## Data Sources

- **Provider**: Yahoo Finance (query1.finance.yahoo.com)
- **Rate Limits**: No official rate limits, but respectful usage recommended
- **Caching**: 24-hour TTL when caching is enabled
- **Timeout**: 30 seconds default (configurable)

## Integration

The YFinance tools are automatically registered when the yfinance configuration is present:

```go
// Tool creation and registration happens automatically
// Individual tools can be accessed through the tools registry
```

## Testing

```bash
# Unit tests
go test ./internal/tools/yfinance/ -v

# Integration tests (requires internet)
go test ./internal/tools/yfinance/ -v -run Integration

# Benchmarks
go test ./internal/tools/yfinance/ -bench=. -run=^$

# Run tests with race detection
go test ./internal/tools/yfinance/ -race -v
```

## Architecture

```
internal/tools/yfinance/
├── stock_data.go         # YFinanceStockDataTool
├── stock_info.go         # YFinanceStockInfoTool
├── dividends.go          # YFinanceDividendsTool
├── financials.go         # YFinanceFinancialsTool
├── market_data.go        # YFinanceMarketDataTool
├── yfinance_test.go      # Comprehensive tests
└── README.md            # This documentation
```

Each tool is a separate implementation of the `models.Tool` interface, providing focused functionality while sharing the same Yahoo Finance API infrastructure.

## Notes

- No API key required - Yahoo Finance provides free access
- Rate limiting is handled through respectful usage patterns
- All tools support context cancellation for timeout handling
- Comprehensive error handling for network and API issues
- Structured logging for debugging and monitoring
- Full test coverage including integration tests

---

_Part of the Mosychlos portfolio management system_
