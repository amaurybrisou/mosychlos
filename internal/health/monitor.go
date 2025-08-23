package health

import (
	"runtime"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// ApplicationMonitor tracks overall application health and performance
type ApplicationMonitor struct {
	sharedBag bag.SharedBag
	startTime time.Time
}

// NewApplicationMonitor creates a new health monitor
func NewApplicationMonitor(sharedBag bag.SharedBag) *ApplicationMonitor {
	return &ApplicationMonitor{
		sharedBag: sharedBag,
		startTime: time.Now(),
	}
}

// UpdateApplicationHealth updates the application health status in the shared bag
func (ahm *ApplicationMonitor) UpdateApplicationHealth() {
	if ahm.sharedBag == nil {
		return
	}

	// Get memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Calculate uptime
	uptime := time.Since(ahm.startTime)

	// Get tool metrics for error rate calculation
	var errorRate float64
	if toolMetricsData, exists := ahm.sharedBag.Get(bag.KToolMetrics); exists {
		if toolMetrics, ok := toolMetricsData.(models.ToolMetrics); ok {
			if toolMetrics.TotalCalls > 0 {
				errorRate = float64(toolMetrics.ErrorCount) / float64(toolMetrics.TotalCalls)
			}
		}
	}

	// Determine overall status based on error rate and other factors
	status := "healthy"
	if errorRate > 0.2 { // More than 20% errors
		status = "error"
	} else if errorRate > 0.1 { // More than 10% errors
		status = "warning"
	}

	// Calculate memory usage in MB
	memoryUsageMB := float64(memStats.Alloc) / 1024 / 1024

	// Get component health from external data health
	componentsHealth := make(map[string]string)
	if externalDataHealthData, exists := ahm.sharedBag.Get(bag.KExternalDataHealth); exists {
		if externalDataHealth, ok := externalDataHealthData.(models.ExternalDataHealth); ok {
			for name, provider := range externalDataHealth.Providers {
				componentsHealth[name] = provider.Status
			}
		}
	}

	// Add cache health to components
	if cacheHealthData, exists := ahm.sharedBag.Get(bag.KCacheStats); exists {
		if cacheStatsMap, ok := cacheHealthData.(models.CacheStatsMap); ok {
			componentsHealth["cache"] = cacheStatsMap.Aggregated.StorageHealth
		} else if cacheHealth, ok := cacheHealthData.(models.CacheHealthStatus); ok {
			// Fallback for old structure
			componentsHealth["cache"] = cacheHealth.StorageHealth
		}
	}

	// Calculate performance metrics
	performanceMetrics := models.PerformanceMetrics{
		AverageResponseTime: 0, // Would need to track this separately
		P95ResponseTime:     0, // Would need to track this separately
		RequestsPerSecond:   0, // Would need to track this separately
		ActiveConnections:   0, // Would need to track this separately
		LastUpdated:         time.Now(),
	}

	// Create application health record
	appHealth := models.ApplicationHealth{
		Status:             status,
		Uptime:             uptime,
		ErrorRate:          errorRate,
		MemoryUsageMB:      memoryUsageMB,
		ComponentsHealth:   componentsHealth,
		LastHealthCheck:    time.Now(),
		PerformanceMetrics: performanceMetrics,
	}

	// Update application health in shared bag
	ahm.sharedBag.Set(bag.KApplicationHealth, appHealth)
}

// StartPeriodicHealthCheck starts a goroutine that periodically updates health status
func (ahm *ApplicationMonitor) StartPeriodicHealthCheck(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			ahm.UpdateApplicationHealth()
		}
	}()
}
