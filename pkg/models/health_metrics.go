package models

import (
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/keys"
)

// APICallStatus represents the status of an API call
type APICallStatus struct {
	Timestamp time.Time     `json:"timestamp"`
	Success   bool          `json:"success"`
	Latency   time.Duration `json:"latency"`
	Error     string        `json:"error,omitempty"`
	Endpoint  string        `json:"endpoint,omitempty"`
}

// CacheHealthStatus represents cache performance and health
type CacheHealthStatus struct {
	TotalHits     int64     `json:"total_hits"`
	TotalMisses   int64     `json:"total_misses"`
	TotalExpired  int64     `json:"total_expired"`
	HitRatio      float64   `json:"hit_ratio"`
	LastUpdated   time.Time `json:"last_updated"`
	CacheSize     int64     `json:"cache_size_bytes,omitempty"`
	StorageHealth string    `json:"storage_health"` // "healthy", "warning", "error"
}

// CacheStatsMap represents cache statistics organized by tool key
type CacheStatsMap struct {
	ToolCaches  map[keys.Key]CacheHealthStatus `json:"tool_caches"`
	Aggregated  CacheHealthStatus              `json:"aggregated"`
	LastUpdated time.Time                      `json:"last_updated"`
}

// ApplicationHealth represents overall system health
type ApplicationHealth struct {
	Status             string             `json:"status"` // "healthy", "warning", "error"
	Uptime             time.Duration      `json:"uptime"`
	ErrorRate          float64            `json:"error_rate"`
	MemoryUsageMB      float64            `json:"memory_usage_mb,omitempty"`
	ComponentsHealth   map[string]string  `json:"components_health"`
	LastHealthCheck    time.Time          `json:"last_health_check"`
	PerformanceMetrics PerformanceMetrics `json:"performance_metrics"`
}

// PerformanceMetrics represents application performance data
type PerformanceMetrics struct {
	AverageResponseTime time.Duration `json:"average_response_time"`
	P95ResponseTime     time.Duration `json:"p95_response_time"`
	RequestsPerSecond   float64       `json:"requests_per_second"`
	ActiveConnections   int           `json:"active_connections"`
	LastUpdated         time.Time     `json:"last_updated"`
}

// ExternalDataHealth tracks health of external data providers
type ExternalDataHealth struct {
	Providers map[string]DataProviderHealth `json:"providers"`
	LastCheck time.Time                     `json:"last_check"`
}

// DataProviderHealth represents health status of an external data provider
type DataProviderHealth struct {
	Name           string               `json:"name"`
	Status         string               `json:"status"` // "healthy", "degraded", "down"
	LastSuccess    time.Time            `json:"last_success"`
	LastFailure    time.Time            `json:"last_failure"`
	SuccessRate    float64              `json:"success_rate"`
	AverageLatency time.Duration        `json:"average_latency"`
	RecentErrors   []string             `json:"recent_errors,omitempty"`
	DataFreshness  map[string]time.Time `json:"data_freshness"` // Type -> last updated
}

// MarketDataFreshness tracks age and quality of market data
type MarketDataFreshness struct {
	Sources   map[string]DataSourceFreshness `json:"sources"`
	LastCheck time.Time                      `json:"last_check"`
}

// DataSourceFreshness represents freshness of data from a specific source
type DataSourceFreshness struct {
	Source        string    `json:"source"`
	LastUpdated   time.Time `json:"last_updated"`
	AgeMins       float64   `json:"age_minutes"`
	Quality       string    `json:"quality"` // "fresh", "stale", "expired"
	RecordCount   int       `json:"record_count"`
	LastRefresh   time.Time `json:"last_refresh"`
	RefreshStatus string    `json:"refresh_status"` // "success", "partial", "failed"
}
