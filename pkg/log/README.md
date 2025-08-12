# `internal/log`

The **log** package provides a centralized, structured logging system for **mosychlos** based on Go's `log/slog` standard library.

---

# Logging

Useful, quiet by default.

## What you get

- Clear INFO messages for major actions (loaded portfolio, generated plan).
- Errors that point to fixes.
- Optional DEBUG for deeper troubleshooting when needed.

## Why it matters

- Understand what happened without reading code or stack traces.

## Tips

- Keep logs with your saved plans in `data/history/` when auditing changes.

---

## 1 · Overview

| Feature                 | Purpose                                                     |
| ----------------------- | ----------------------------------------------------------- |
| **Structured logging**  | Consistent key-value pairs for easy filtering and querying. |
| **Multiple formats**    | Human-readable text or machine-processable JSON.            |
| **Configurable levels** | Debug, Info, Warn, Error to control verbosity.              |
| **Environment control** | Set log level and format via environment variables.         |
| **Source locations**    | Optional code position annotations for debugging.           |
| **Context scoping**     | Create loggers with preset fields or groups.                |

---

## 2 · Simple Usage

```go
// Initialize logging (typically in main.go or app startup)
import "github.com/amaurybrisou/mosychlos/internal/log"

func main() {
    log.Init()

    // Use standard slog API for logging
    slog.Info("Portfolio loaded", "assets", 42, "total_value", 125000.75)
    slog.Error("API connection failed", "service", "pricing", "err", err)
}
```

## 3 · Log Levels

```go
// Available levels in increasing order of severity
slog.Debug("Detailed information for debugging", "iteration", i)
slog.Info("Normal operational information", "profile", "balanced")
slog.Warn("Potential issue detected", "drift", 0.12)
slog.Error("Operation failed", "operation", "rebalance", "err", err)
```

## 4 · Environment Configuration

| Variable      | Values                           | Default | Purpose                       |
| ------------- | -------------------------------- | ------- | ----------------------------- |
| `SLOG_LEVEL`  | `debug`, `info`, `warn`, `error` | `info`  | Sets minimum log level        |
| `SLOG_FORMAT` | `text`, `json`                   | `text`  | Sets output format            |
| `SLOG_SOURCE` | `true`, `false`                  | `false` | Include source code positions |

## 5 · Advanced Configuration

```go
// For more control, initialize with a custom configuration
log.InitWithConfig(log.Config{
    Level:     log.Debug,
    Format:    log.JSON,
    Output:    os.Stdout,  // or any io.Writer
    AddSource: true,
})
```

## 6 · Component Loggers

```go
// Create a logger with predefined attributes
rebalancerLogger := log.With(
    "component", "rebalancer",
    "profile", profile.Name(),
)

rebalancerLogger.Info("Starting rebalance operation")
rebalancerLogger.Info("Evaluating drift", "current_drift", drift)

// Create a logger with grouped attributes
portfolioLogger := log.WithGroup("portfolio")
portfolioLogger.Info("Updated positions",
    "stocks", stockValue,
    "bonds", bondValue,
    "cash", cashValue,
)
```

## 7 · JSON Output Example

```json
{
  "time": "2025-08-08T15:30:45.123Z",
  "level": "INFO",
  "msg": "Portfolio loaded",
  "assets": 42,
  "total_value": 125000.75
}
```

---

## 8 · Best Practices

- Include contextual information as structured key-value pairs
- Use consistent keys across the application
- Log at appropriate levels - debug for developers, info for operations
- Group related attributes for better readability
- Include error details when logging errors

---

## 9 · Integration with Other Packages

The `log` package is designed to be used by all other mosychlos components:

```go
// In the analyze package
import "log/slog"

func AnalyzePortfolio(p *Portfolio) {
    slog.Info("Starting portfolio analysis",
        "portfolio", p.Name,
        "assets_count", len(p.Assets),
    )

    // Business logic...

    if someError != nil {
        slog.Error("Analysis failed", "err", someError)
        return
    }
}
```
