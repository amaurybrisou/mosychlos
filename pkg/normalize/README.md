<!-- pkg/normalize/README.md -->

# normalize package

A stdlib-only normalization layer turning heterogeneous tool outputs into a small, typed, provider-agnostic `Envelope`.

## Features

- Timeseries (`yfinance_stock_data`) → `TimeseriesData` with UTC RFC3339 points
- Snapshot quotes (`yfinance_market_data`) → `SnapshotData`
- News (`news_api`) → `NewsData`
- Handles nulls in yfinance quote arrays gracefully
- Returns a `Status` (`ok` / `empty`) to simplify downstream branching
- Keeps `Args` and `RawResult` for provenance/audit

## Usage

```go
reg := normalize.DefaultRegistry()

for i := range ctx.ToolComputations {
    tc := &ctx.ToolComputations[i]
    n, ok := reg.Find(tc.ToolName)
    if !ok {
        continue // or record an error
    }
    env, status, err := n.Normalize(context.Background(), tc.ToolName, tc.Arguments, tc.RawResult)
    if err != nil {
        // handle error
        continue
    }
    tc.Normalized = env
    tc.Status = string(status)
}
```
