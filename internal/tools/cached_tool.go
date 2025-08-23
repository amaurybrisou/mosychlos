package tools

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/cache"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// CachedTool wraps an models.Tool with caching functionality
// Caches tool responses by date and parameter hash to reduce API calls and improve performance
type CachedTool struct {
	tool  models.Tool
	cache cache.Cache
	ttl   time.Duration
}

var _ models.Tool = &CachedTool{}

// NewCachedTool creates a new cached tool wrapper
func NewCachedTool(tool models.Tool, cacheDir string, ttl time.Duration) *CachedTool {
	toolCacheDir := filepath.Join(cacheDir, "tools")
	fileCache := cache.NewFileCache(toolCacheDir)

	return &CachedTool{
		tool:  tool,
		cache: fileCache,
		ttl:   ttl,
	}
}

// NewCachedToolWithMonitoring creates a new cached tool wrapper with cache monitoring
func NewCachedToolWithMonitoring(tool models.Tool, cacheDir string, ttl time.Duration, sharedBag bag.SharedBag) *CachedTool {
	toolCacheDir := filepath.Join(cacheDir, "tools")
	fileCache := cache.NewFileCache(toolCacheDir)

	var monitoredCache cache.Cache = fileCache
	if sharedBag != nil {
		monitoredCache = cache.NewMonitor(fileCache, sharedBag, tool.Key())
	}

	return &CachedTool{
		tool:  tool,
		cache: monitoredCache,
		ttl:   ttl,
	}
}

// Name returns the wrapped tool's name
func (ct *CachedTool) Name() string {
	return ct.tool.Name()
}

// Key returns the wrapped tool's key
func (ct *CachedTool) Key() bag.Key {
	return ct.tool.Key()
}

// Description returns the wrapped tool's description
func (ct *CachedTool) Description() string {
	return ct.tool.Description()
}

func (ct *CachedTool) IsExternal() bool {
	return ct.tool.IsExternal()
}

// Definition returns the wrapped tool's definition
func (ct *CachedTool) Definition() models.ToolDef {
	return ct.tool.Definition()
}

// Tags returns the wrapped tool's tags
func (ct *CachedTool) Tags() []string {
	return ct.tool.Tags()
}

// Run executes the tool with caching
// First checks cache for existing result, otherwise executes the tool and caches the result
func (ct *CachedTool) Run(ctx context.Context, args string) (string, error) {
	// Generate cache key based on tool name, date, and parameter hash
	cacheKey := ct.generateCacheKey(args)

	// Try to get cached result first
	if cached, found := ct.cache.Get(cacheKey); found {
		return string(cached), nil
	}

	// Execute the actual tool if not cached
	result, err := ct.tool.Run(ctx, args)
	if err != nil {
		// Don't cache errors
		return result, err
	}

	// Cache the successful result
	ct.cache.Set(cacheKey, []byte(result), ct.ttl)

	return result, nil
}

// generateCacheKey creates a deterministic cache key based on tool name, date, and arguments
// Format: "tool:{toolname}:{date}:{args_hash}"
func (ct *CachedTool) generateCacheKey(args string) string {
	// Include date for daily invalidation
	date := time.Now().Format("2006-01-02")

	// Create hash of arguments for deterministic, compact keys
	hasher := sha256.New()
	hasher.Write([]byte(args))
	argsHash := hex.EncodeToString(hasher.Sum(nil))[:16] // Use first 16 chars for compact key

	return fmt.Sprintf("tool:%s:%s:%s", ct.tool.Name(), date, argsHash)
}

// ToolCacheConfig holds configuration for tool caching
type ToolCacheConfig struct {
	Enabled bool          `yaml:"enabled"`
	TTL     time.Duration `yaml:"ttl"`
	BaseDir string        `yaml:"base_dir"`
}

func wrapWithCache(tool models.Tool, cacheConfig *ToolCacheConfig) models.Tool {
	if cacheConfig == nil {
		cacheConfig = &ToolCacheConfig{}
	}

	if !cacheConfig.Enabled {
		return tool
	}

	return NewCachedTool(tool, cacheConfig.BaseDir, cacheConfig.TTL)
}
