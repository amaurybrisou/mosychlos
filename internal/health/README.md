# Application Health Monitoring

This package provides comprehensive application health monitoring capabilities that track system performance, external API health, cache performance, and overall application status.

## Features

- **Real-time Health Tracking**: Monitor application uptime, memory usage, and error rates
- **Component Health Monitoring**: Track health of external data providers and cache systems
- **Automatic Alerting**: Status classification (healthy/warning/error) based on metrics
- **Periodic Updates**: Configurable health check intervals
- **Shared State Integration**: All metrics stored in shared bag for cross-component access

## Usage

### Basic Setup

```go
import (
    "github.com/amaurybrisou/mosychlos/internal/health"
    "github.com/amaurybrisou/mosychlos/pkg/bag"
)

// Create shared bag and health monitor
sharedBag := bag.NewSharedBag()
healthMonitor := health.NewApplicationMonitor(sharedBag)

// Start periodic health checks every 30 seconds
healthMonitor.StartPeriodicHealthCheck(30 * time.Second)

// Manual health update
healthMonitor.UpdateApplicationHealth()
```

### Retrieving Health Data

```go
// Get application health from shared bag
if healthData, exists := sharedBag.Get(bag.KApplicationHealth); exists {
    if appHealth, ok := healthData.(models.ApplicationHealth); ok {
        fmt.Printf("System Status: %s\n", appHealth.Status)
        fmt.Printf("Uptime: %v\n", appHealth.Uptime)
        fmt.Printf("Error Rate: %.2f%%\n", appHealth.ErrorRate * 100)
        fmt.Printf("Memory Usage: %.2f MB\n", appHealth.MemoryUsageMB)

        // Check component health
        for component, status := range appHealth.ComponentsHealth {
            fmt.Printf("Component %s: %s\n", component, status)
        }
    }
}
```

## Health Status Levels

- **healthy**: Error rate < 10%, all systems operational
- **warning**: Error rate 10-20%, some degradation detected
- **error**: Error rate > 20%, significant issues detected

## Integration Points

The health monitor automatically integrates with:

- **Tool Metrics**: Tracks API call success/failure rates from tool usage
- **Cache Statistics**: Monitors cache hit rates and storage health
- **External Data Providers**: Tracks health of FMP, FRED, NewsAPI, and other data sources
- **System Resources**: Monitors memory usage and runtime statistics

## Shared Bag Keys

Health data is stored in the shared bag under these keys:

- `KApplicationHealth`: Overall application health status
- `KCacheStats`: Aggregated cache performance metrics
- `KExternalDataHealth`: Health status of external data providers
- `KToolMetrics`: Tool execution metrics and statistics

## Background Monitoring

The health monitor runs in the background using a goroutine with configurable intervals. It automatically updates health status without blocking the main application flow.

```go
// Start monitoring with custom interval
healthMonitor.StartPeriodicHealthCheck(60 * time.Second) // Every minute
```

This provides continuous health visibility for monitoring, alerting, and debugging purposes.
