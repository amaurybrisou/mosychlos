# Binance Data Provider

Clean interface for fetching raw data from Binance API.

## What it does

- **Spot Portfolio Data** - Retrieve account balances and current market values
- **Price Data** - Get current or historical prices for trading pairs
- **Market Data** - Access 24hr ticker statistics and candlestick data

## Why it matters

- **Raw Data Access** - Returns unprocessed API responses for flexible usage
- **Clean Interface** - Simple methods focused on data retrieval only
- **No Dependencies** - Doesn't depend on filesystem or persistence layers

## Key Features

✅ **Portfolio Fetching** - Get spot account balances with current valuations
✅ **Price Discovery** - Retrieve current and historical price data
✅ **Market Statistics** - Access 24hr ticker and volume data
✅ **Flexible Configuration** - Support for testnet and custom endpoints

## Usage Examples

### Get Spot Portfolio

```go
provider := binance.NewPortfolioProvider(cfg)
portfolio, err := provider.GetSpotPortfolio(ctx)
// Returns processed balances, prices, and total values
```

### Fetch Current Prices

```go
client := binance.New(cfg)
price, err := client.GetPrice(ctx, "BTCUSDT")
prices, err := client.GetPrices(ctx, []string{"BTCUSDT", "ETHUSDT"})
```

### Get Market Data

```go
ticker, err := client.GetTicker24hr(ctx, "BTCUSDT")
klines, err := client.GetKlines(ctx, "BTCUSDT", "1h", 100, nil, nil)
```

## Configuration

Add to your config:

```yaml
binance:
  api_key: 'your_api_key'
  api_secret: 'your_api_secret'
  base_url: 'https://api.binance.com' # optional, defaults to mainnet
  cache_enable: true # optional
```

## Data Flow

Raw Binance API → Clean Interface → Processed Models → Your Application

This package handles the first two steps, providing clean access to Binance data without imposing any specific application patterns.
