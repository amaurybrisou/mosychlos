package cache

import (
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// Monitor wraps a cache and reports statistics to the shared bag
type Monitor struct {
	Cache
	sharedBag bag.SharedBag
	toolKey   keys.Key
}

// NewMonitor creates a cache monitor that reports stats to shared bag
func NewMonitor(c Cache, sharedBag bag.SharedBag, toolKey keys.Key) *Monitor {
	return &Monitor{
		Cache:     c,
		sharedBag: sharedBag,
		toolKey:   toolKey,
	}
}

// Get wraps the cache Get method and updates stats
func (cm *Monitor) Get(key string) ([]byte, bool) {
	value, found := cm.Cache.Get(key)
	cm.updateCacheStats()
	return value, found
}

// Set wraps the cache Set method and updates stats
func (cm *Monitor) Set(key string, val []byte, ttl time.Duration) {
	cm.Cache.Set(key, val, ttl)
	cm.updateCacheStats()
}

// Delete wraps the cache Delete method and updates stats
func (cm *Monitor) Delete(key string) {
	cm.Cache.Delete(key)
	cm.updateCacheStats()
}

// updateCacheStats reports current cache statistics to the shared bag
func (cm *Monitor) updateCacheStats() {
	if cm.sharedBag == nil {
		return
	}

	stats := cm.Cache.Stats()

	// Calculate hit ratio
	hitRatio := 0.0
	totalAccess := stats.Hits + stats.Misses
	if totalAccess > 0 {
		hitRatio = float64(stats.Hits) / float64(totalAccess)
	}

	// Determine storage health based on hit ratio and expired count
	storageHealth := "healthy"
	if hitRatio < 0.5 {
		storageHealth = "warning"
	}
	if stats.Expired > stats.Hits {
		storageHealth = "warning"
	}

	cacheHealth := models.CacheHealthStatus{
		TotalHits:     int64(stats.Hits),
		TotalMisses:   int64(stats.Misses),
		TotalExpired:  int64(stats.Expired),
		HitRatio:      hitRatio,
		LastUpdated:   time.Now(),
		StorageHealth: storageHealth,
	}

	// Update cache stats in the shared bag using map structure
	cm.sharedBag.Update(keys.KCacheStats, func(current any) any {
		var cacheStatsMap models.CacheStatsMap

		// Initialize or get existing map
		if current != nil {
			if existing, ok := current.(models.CacheStatsMap); ok {
				cacheStatsMap = existing
			}
		}

		// Initialize maps if needed
		if cacheStatsMap.ToolCaches == nil {
			cacheStatsMap.ToolCaches = make(map[keys.Key]models.CacheHealthStatus)
		}

		// Update this tool's cache stats
		cacheStatsMap.ToolCaches[cm.toolKey] = cacheHealth

		// Recalculate aggregated stats from all tool caches
		var aggregated models.CacheHealthStatus
		totalHits, totalMisses, totalExpired := int64(0), int64(0), int64(0)

		for _, toolCache := range cacheStatsMap.ToolCaches {
			totalHits += toolCache.TotalHits
			totalMisses += toolCache.TotalMisses
			totalExpired += toolCache.TotalExpired
		}

		// Calculate overall metrics
		totalAccess := totalHits + totalMisses
		if totalAccess > 0 {
			aggregated.HitRatio = float64(totalHits) / float64(totalAccess)
		}

		aggregated.TotalHits = totalHits
		aggregated.TotalMisses = totalMisses
		aggregated.TotalExpired = totalExpired
		aggregated.LastUpdated = time.Now()

		// Determine overall storage health
		if aggregated.HitRatio >= 0.7 {
			aggregated.StorageHealth = "healthy"
		} else if aggregated.HitRatio >= 0.4 {
			aggregated.StorageHealth = "warning"
		} else {
			aggregated.StorageHealth = "error"
		}

		cacheStatsMap.Aggregated = aggregated
		cacheStatsMap.LastUpdated = time.Now()

		return cacheStatsMap
	})
}
