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

### Component Health Status

{{range $component, $status := .ApplicationHealth.ComponentsHealth}}

- **{{$component}}:** {{if eq $status "healthy"}}✅{{else if eq $status "warning"}}⚠️{{else}}❌{{end}} {{$status}}
  {{end}}

---

## ⚡ Performance Metrics

{{if .ToolMetrics}}

### Tool Execution Statistics

| Metric               | Value                                                       |
| -------------------- | ----------------------------------------------------------- |
| **Total API Calls**  | {{.ToolMetrics.TotalCalls}}                                 |
| **Success Rate**     | {{printf "%.2f%%" (multiply .ToolMetrics.SuccessRate 100)}} |
| **Average Duration** | {{formatDuration .ToolMetrics.AverageDuration}}             |
| **Total Cost**       | ${{printf "%.4f" .ToolMetrics.TotalCost}}                   |
| **Total Tokens**     | {{formatNumber .ToolMetrics.TotalTokens}}                   |

{{if .ToolMetrics.ByTool}}

#### Performance by Tool:

{{range $toolName, $stats := .ToolMetrics.ByTool}}
**{{$toolName}}:**

- Calls: {{$stats.Calls}}, Success: {{$stats.Successes}}, Errors: {{$stats.Errors}}
- Avg Duration: {{formatDuration $stats.AverageDuration}}
- Cost: ${{printf "%.4f" $stats.Cost}}, Tokens: {{$stats.Tokens}}

{{end}}
{{end}}
{{end}}

{{if .CacheStats}}

### 💾 Cache Performance

#### System-wide Cache Statistics

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

| Metric              | Value                                                                                                              |
| ------------------- | ------------------------------------------------------------------------------------------------------------------ |
| **Status**          | {{if eq $health.Status "healthy"}}✅{{else if eq $health.Status "degraded"}}⚠️{{else}}❌{{end}} {{$health.Status}} |
| **Success Rate**    | {{printf "%.2f%%" (multiply $health.SuccessRate 100)}}                                                             |
| **Average Latency** | {{formatDuration $health.AverageLatency}}                                                                          |
| **Last Success**    | {{if not $health.LastSuccess.IsZero}}{{$health.LastSuccess.Format "January 2, 15:04:05"}}{{else}}Never{{end}}      |
| **Last Failure**    | {{if not $health.LastFailure.IsZero}}{{$health.LastFailure.Format "January 2, 15:04:05"}}{{else}}Never{{end}}      |

{{if $health.RecentErrors}}
**Recent Errors:**
{{range $health.RecentErrors}}

- {{.}}
  {{end}}
  {{end}}

{{if and $.ToolComputations (gt (len $.ToolComputations) 0)}}

#### Recent API Calls

{{$callCount := 0}}
{{range $.ToolComputations}}
{{if and (eq .ToolName $provider) (lt $callCount 5)}}

{{if isWebSearchTool .ToolName}}
**{{.StartTime.Format "15:04:05"}}**
{{formatWebSearchData .ToolName .Arguments .Result}}
{{else}}
**{{.StartTime.Format "15:04:05"}}** - {{if .Success}}✅ Success{{else}}❌ Failed{{end}} ({{formatDuration .Duration}})
{{if not .Success}}

- **Error:** {{.Error}}
  {{end}}
  {{if .Arguments}}
- **Request:** {{truncate .Arguments 100}}
  {{end}}
  {{if and .Result .Success}}
- **Response:** {{truncate .Result 150}}
  {{end}}
  {{if .Cost}}
- **Cost:** ${{printf "%.4f" .Cost}}
  {{end}}
  {{if .TokensUsed}}
- **Tokens:** {{.TokensUsed}}
  {{end}}
  {{end}}
  {{$callCount = add $callCount 1}}
  {{end}}
  {{end}}
  {{if eq $callCount 0}}
  _No recent calls recorded for this provider_
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

#### {{.ToolName}} - {{.StartTime.Format "15:04:05"}}

{{if isWebSearchTool .ToolName}}
{{formatWebSearchData .ToolName .Arguments .Result}}
{{else}}

- **Duration:** {{formatDuration .Duration}}
- **Status:** {{if .Success}}✅ Success{{else}}❌ Failed{{end}}
  {{if not .Success}}- **Error:** {{.Error}}{{end}}
  {{if .Cost}}- **Cost:** ${{printf "%.4f" .Cost}}{{end}}
  {{if .TokensUsed}}- **Tokens:** {{.TokensUsed}}{{end}}
  {{if .DataConsumed}}- **Data Read:** {{join .DataConsumed ", "}}{{end}}
  {{if .DataProduced}}- **Data Written:** {{join .DataProduced ", "}}{{end}}
  {{end}}

{{end}}
{{end}}

---

## 📈 Health Trends

### System Performance Over Time

- Application has been running for {{formatDuration .ApplicationHealth.Uptime}}
- Current error rate: {{printf "%.2f%%" (multiply .ApplicationHealth.ErrorRate 100)}}
  {{if .ToolMetrics}}- Total API calls processed: {{.ToolMetrics.TotalCalls}}{{end}}

{{$hasRecommendations := false}}
{{if gt .ApplicationHealth.ErrorRate 0.1}}{{$hasRecommendations = true}}{{end}}
{{if and .CacheStats (lt .CacheStats.HitRatio 0.7)}}{{$hasRecommendations = true}}{{end}}
{{if and .ToolMetrics (gt .ToolMetrics.AverageDuration.Milliseconds 5000)}}{{$hasRecommendations = true}}{{end}}

{{if $hasRecommendations}}

### Recommendations

{{if gt .ApplicationHealth.ErrorRate 0.1}}
⚠️ **High Error Rate Detected** - Error rate is above 10%, investigate failing components
{{end}}
{{if and .CacheStats (lt .CacheStats.HitRatio 0.7)}}
⚠️ **Low Cache Hit Ratio** - Cache efficiency is below 70%, consider cache optimization
{{end}}
{{if and .ToolMetrics (gt .ToolMetrics.AverageDuration.Milliseconds 5000)}}
⚠️ **Slow Tool Performance** - Average tool execution time is above 5 seconds
{{end}}
{{end}}

---

_System diagnostics generated by Mosychlos Portfolio Management System_
