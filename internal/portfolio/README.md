# Portfolio Management

Centralized portfolio operations with shared state management.

## What it does

- **Portfolio Storage** - Load and save portfolios using the file system
- **External Fetching** - Retrieve portfolio data from external sources like Binance
- **Validation** - Ensure portfolio data integrity and business rules
- **State Management** - Track portfolio state using the shared bag

## Why it matters

- **Centralized Operations** - Single point for all portfolio-related operations
- **Shared State** - Portfolio state is available across the application via the bag
- **Flexible Storage** - Uses `pkg/fs` for swappable file system operations
- **Validation Ready** - Built-in validation ensures data quality

## Key Features

✅ **File-based Storage** - YAML portfolio files in configured data directory
✅ **External Fetching** - Interface for fetching from APIs like Binance
✅ **Data Validation** - Comprehensive validation with business rules
✅ **State Tracking** - Metadata like validation time and fetch source
✅ **Bag Integration** - Current portfolio available application-wide

## Usage Examples

### Basic Service Setup

```go
cfg := &config.Config{DataDir: "./data/portfolios"}
filesystem := fs.OS{}
validator := portfolio.NewBasicValidator()
sharedBag := bag.New()

service := portfolio.NewService(cfg, filesystem, validator, sharedBag)
```

### Load and Save Portfolios

```go
// Load from file
portfolio, err := service.Load(ctx, "my_portfolio")

// Save to file
err = service.Save(ctx, portfolio, "my_portfolio")

// List all portfolios
names, err := service.List(ctx)
```

### Fetch from External Source

```go
// Implement fetcher interface
type BinanceFetcher struct {
    client binance.PortfolioProvider
}

func (f *BinanceFetcher) Fetch(ctx context.Context) (*models.Portfolio, error) {
    data, err := f.client.GetSpotPortfolio(ctx)
    // ... convert to Portfolio model
    return portfolio, nil
}

// Fetch and validate
fetcher := &BinanceFetcher{client: binanceClient}
portfolio, err := service.Fetch(ctx, fetcher, true)
```

### Access Current State

```go
// Get current portfolio from bag
current := service.GetCurrentPortfolio()

// Get metadata
validationTime := service.GetLastValidationTime()
fetchTime := service.GetLastFetchTime()
source := service.GetFetchSource()
```

## Configuration

Add to your config:

```yaml
data_dir: './data/portfolios' # Directory for portfolio files
```

## Data Flow

External Source → Fetcher → Service → Validation → File System + Bag State

The service coordinates between external data sources, validation, persistence, and shared application state.
