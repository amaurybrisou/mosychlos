package cache

import (
	"sync"
	"time"
)

//go:generate mockgen -source=cache.go -destination=mocks/cache_mock.go -package=mocks

// Entry represents a cached value with expiration.
type Entry struct {
	val []byte
	exp time.Time
}

// Cache is a simple TTL cache interface for reuse.
type Cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, val []byte, ttl time.Duration)
	Delete(key string)
	Stats() Stats
}

// Stats exposes lightweight metrics for diagnostics.
type Stats struct {
	Size     int
	Hits     int
	Misses   int
	Evicted  int
	Expired  int
	Capacity int
}

// ttlCache implements Cache with size bound + TTL eviction (lazy on access / periodic cleanup trigger).
type ttlCache struct {
	mu      sync.RWMutex
	items   map[string]Entry
	cap     int
	hits    int
	misses  int
	evicted int
	expired int
}

// NewTTL returns a bounded TTL cache. cap<=0 means unbounded.
func NewTTL(capacity int) Cache {
	return &ttlCache{items: make(map[string]Entry, 64), cap: capacity}
}

// NewContextCache creates a file-based cache for context data persistence
// Data will be stored in baseDir organized by date (e.g., data/context/2025-08-11/)
func NewContextCache(baseDir string) Cache {
	return NewFileCache(baseDir)
}

func (c *ttlCache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.items[key]
	if !ok {
		c.misses++
		return nil, false
	}
	if !e.exp.IsZero() && time.Now().After(e.exp) {
		delete(c.items, key)
		c.expired++
		c.misses++
		return nil, false
	}
	c.hits++
	return append([]byte(nil), e.val...), true
}

func (c *ttlCache) Set(key string, val []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cap > 0 && len(c.items) >= c.cap {
		// naive eviction: remove random (first) item
		for k := range c.items {
			delete(c.items, k)
			c.evicted++
			break
		}
	}
	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	c.items[key] = Entry{val: append([]byte(nil), val...), exp: exp}
}

func (c *ttlCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

func (c *ttlCache) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return Stats{Size: len(c.items), Hits: c.hits, Misses: c.misses, Evicted: c.evicted, Expired: c.expired, Capacity: c.cap}
}
