# FRED Macro Context Provider

Bring macro insight into your portfolio analysis using Federal Reserve Economic Data (FRED).

## What it does

- Adds a macro snapshot to analyses (latest values):
  - GDP
  - Inflation (YoY)
  - Policy/overnight rate
  - Unemployment
- Uses configurable FRED series and units.
- Avoids API overuse via a built‑in daily cache.

## Fetch rules

- Activation
  - Enabled when `context.macro.provider: "fred"` and a FRED API key is set (`FRED_API_KEY` or `providers.fred.api_key`).
- Scope
  - Macro is market‑wide; tickers are ignored. Always returns a single snapshot.
- Default series
  - GDP: `GDP`
  - Inflation: `CPIAUCSL` with units `pc1` (percent change from year ago)
  - Policy rate: `FEDFUNDS`
  - Unemployment: `UNRATE`
- Configuration
  - Override series via `providers.fred.series.{gdp,inflation,interest_rate,unemployment}`
  - Change CPI units via `providers.fred.series.inflation_units`
  - `providers.fred.country` is informational (series IDs are global in FRED)

## Cache behavior

- Storage
  - JSON files under `data/context/macro/<YYYY-MM-DD>/<hash>.json`.
- Keying
  - One cache per day and request key. For macro, tickers are ignored, so typically a single cache file per day.
- Flow
  - On fetch → read cache; on miss → call FRED, write JSON, return result.
- TTL & invalidation
  - Daily TTL (new file each day). Delete the file or the date folder to force refetch.
- Note
  - Cache key does not include provider options (country/series). If you change them mid‑day, clear the cache to avoid stale output.

## Configure

- YAML (`config/config.yaml`)
  - `context.macro.provider: "fred"`
  - `providers.fred.api_key: ""`
  - `providers.fred.series.{gdp,inflation,interest_rate,unemployment,inflation_units}`
- Environment
  - `FRED_API_KEY`, `FRED_COUNTRY`
  - `FRED_SERIES_GDP`, `FRED_SERIES_INFLATION`, `FRED_SERIES_INTEREST_RATE`, `FRED_SERIES_UNEMPLOYMENT`, `FRED_SERIES_INFLATION_UNITS`

## Try it

1. Export your API key

2. Run analyze with context enabled (uses the daily cache)

3. Inspect today’s cache at `data/context/macro/$(date +%F)`

For the broader context system and cache adapter, see `internal/context/README.md`.
