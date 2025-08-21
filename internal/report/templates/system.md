# System Health & Diagnostics Report

**Generated:** {{.GeneratedAt.Format "January 2, 2006 15:04 MST"}}

---

## 🚦 Application Status

### Overall Health: **{{.ApplicationHealth.Status | toUpper}}**

| Metric                | Value                                                           |
| --------------------- | --------------------------------------------------------------- |
| **Uptime**            | {{formatDuration .ApplicationHealth.Uptime}}                    |
| **Memory Usage**      | {{printf "%.2f" .ApplicationHealth.MemoryUsageMB}} MB           |
| **Error Rate**        | {{printf "%.2f%%" (multiply .ApplicationHealth.ErrorRate 100)}} |
| **Last Health Check** | {{.ApplicationHealth.LastHealthCheck.Format "15:04:05"}}        |

#### Component Health

{{range $component, $status := .ApplicationHealth.ComponentsHealth}}

- **{{$component}}:** {{if eq $status "healthy"}}✅{{else if eq $status "warning"}}⚠️{{else}}❌{{end}} {{$status}}
  {{end}}

---

## ⚡ Performance Metrics

{{if .ToolMetrics}}

### Tool Execution Statistics

| Metric              | Value                                                       |
| ------------------- | ----------------------------------------------------------- |
| **Total API Calls** | {{.ToolMetrics.TotalCalls}}                                 |
| **Success Rate**    | {{printf "%.2f%%" (multiply .ToolMetrics.SuccessRate 100)}} |
| **Avg Duration**    | {{formatDuration .ToolMetrics.AverageDuration}}             |
| **Total Cost**      | ${{printf "%.4f" .ToolMetrics.TotalCost}}                   |
| **Total Tokens**    | {{formatNumber .ToolMetrics.TotalTokens}}                   |

{{if .ToolMetrics.ByTool}}

#### By Tool:

{{range $toolName, $stats := .ToolMetrics.ByTool}}

- **{{$toolName}}:** Calls: {{$stats.Calls}}, Success: {{$stats.Successes}}, Errors: {{$stats.Errors}}, Avg Duration: {{formatDuration $stats.AverageDuration}}, Cost: ${{printf "%.4f" $stats.Cost}}, Tokens: {{$stats.Tokens}}
  {{end}}
  {{end}}
  {{end}}

{{if .CacheStats}}

### 💾 Cache Performance

| Metric                       | Value                                                                                                                                              |
| ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------- | ------- |
| **Hit Ratio**                | {{printf "%.2f%%" (multiply .CacheStats.HitRatio 100)}}                                                                                            |
| **Total Hits**               | {{formatNumber .CacheStats.TotalHits}}                                                                                                             |
| **Total Misses**             | {{formatNumber .CacheStats.TotalMisses}}                                                                                                           |
| **Expired Entries**          | {{formatNumber .CacheStats.TotalExpired}}                                                                                                          |
| **Storage Health**           | {{if eq .CacheStats.StorageHealth "healthy"}}✅{{else if eq .CacheStats.StorageHealth "warning"}}⚠️{{else}}❌{{end}} {{.CacheStats.StorageHealth}} |
| {{if .CacheStats.CacheSize}} | **Cache Size**                                                                                                                                     | {{formatBytes .CacheStats.CacheSize}} | {{end}} |

{{end}}

---

## 🌐 External Data Sources

{{if and .ExternalDataHealth .ExternalDataHealth.Providers}}
{{range $provider, $health := .ExternalDataHealth.Providers}}

### {{$provider | toUpper}}

| Metric           | Value                                                                                                              |
| ---------------- | ------------------------------------------------------------------------------------------------------------------ |
| **Status**       | {{if eq $health.Status "healthy"}}✅{{else if eq $health.Status "degraded"}}⚠️{{else}}❌{{end}} {{$health.Status}} |
| **Success Rate** | {{printf "%.2f%%" (multiply $health.SuccessRate 100)}}                                                             |
| **Avg Latency**  | {{formatDuration $health.AverageLatency}}                                                                          |
| **Last Success** | {{if not $health.LastSuccess.IsZero}}{{$health.LastSuccess.Format "January 2, 15:04:05"}}{{else}}Never{{end}}      |
| **Last Failure** | {{if not $health.LastFailure.IsZero}}{{$health.LastFailure.Format "January 2, 15:04:05"}}{{else}}Never{{end}}      |

{{if $health.RecentErrors}}
**Recent Errors:**
{{range $health.RecentErrors}}

- {{.}}
  {{end}}
  {{end}}
  {{end}}
  {{else}}
  _No external data providers configured or no data available_
  {{end}}

{{if and .MarketDataFreshness .MarketDataFreshness.Sources}}

### 📊 Data Freshness Status

{{range $source, $freshness := .MarketDataFreshness.Sources}}

- **{{$source}}:** {{if eq $freshness.Quality "fresh"}}🟢{{else if eq $freshness.Quality "stale"}}🟡{{else}}🔴{{end}} {{$freshness.Quality}} ({{printf "%.1f" $freshness.AgeMins}} mins old, {{$freshness.RecordCount}} records)
  {{end}}
  {{end}}

---

## 🔧 Recent Activity Log

{{if .ToolComputations}}

### Last 10 Tool Executions

{{$recentComputations := slice .ToolComputations 0 (min 10 (len .ToolComputations))}}
{{range $recentComputations}}

- **{{.ToolName}}** ({{.StartTime.Format "15:04:05"}}): {{if .Success}}✅ Success{{else}}❌ Failed{{end}}, Duration: {{formatDuration .Duration}}{{if .Cost}}, Cost: ${{printf "%.4f" .Cost}}{{end}}{{if .TokensUsed}}, Tokens: {{.TokensUsed}}{{end}}{{if .Error}}, Error: {{.Error}}{{end}}
  {{end}}
  {{end}}

---

## 📈 Health Trends

- Uptime: {{formatDuration .ApplicationHealth.Uptime}}
- Error rate: {{printf "%.2f%%" (multiply .ApplicationHealth.ErrorRate 100)}}
  {{if .ToolMetrics}}- Total API calls: {{.ToolMetrics.TotalCalls}}{{end}}

{{if or (gt .ApplicationHealth.ErrorRate 0.1) (and .CacheStats (lt .CacheStats.HitRatio 0.7)) (and .ToolMetrics (gt .ToolMetrics.AverageDuration.Milliseconds 5000))}}

### Recommendations

{{if gt .ApplicationHealth.ErrorRate 0.1}}
⚠️ **High Error Rate**: Investigate failing components
{{end}}
{{if and .CacheStats (lt .CacheStats.HitRatio 0.7)}}
⚠️ **Low Cache Hit Ratio**: Consider cache optimization
{{end}}
{{if and .ToolMetrics (gt .ToolMetrics.AverageDuration.Milliseconds 5000)}}
⚠️ **Slow Tool Performance**: Average execution time above 5 seconds
{{end}}
{{end}}

---

_System diagnostics generated by Mosychlos Portfolio Management System_
