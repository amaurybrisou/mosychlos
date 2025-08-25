package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"log/slog"
)

// fileEntry represents a cached value with metadata for file storage
type fileEntry struct {
	Value     []byte    `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// FileCache is a persistent cache that stores data in the filesystem
// organized by date and context type for easy access and transparency
type FileCache struct {
	baseDir         string
	dateSubdirs     bool // Whether to organize files by date subdirectories
	maxKeyLength    int  // Maximum key length before hashing
	mu              sync.RWMutex
	stats           Stats
	filenamePattern *regexp.Regexp // Regex for sanitizing filenames
}

// FileCacheOptions configures FileCache behavior
type FileCacheOptions struct {
	DateSubdirs  bool   // Organize files in date subdirectories (default: true)
	MaxKeyLength int    // Max key length before hashing (default: 200)
	DateFormat   string // Date format for subdirectories (default: "2006-01-02")
}

// DefaultFileCacheOptions returns sensible defaults
func DefaultFileCacheOptions() FileCacheOptions {
	return FileCacheOptions{
		DateSubdirs:  true,
		MaxKeyLength: 200,
		DateFormat:   "2006-01-02",
	}
}

// NewFileCache creates a new file-based cache that persists to the given directory
func NewFileCache(baseDir string) *FileCache {
	return NewFileCacheWithOptions(baseDir, DefaultFileCacheOptions())
}

// NewFileCacheWithOptions creates a new file-based cache with custom options
func NewFileCacheWithOptions(baseDir string, opts FileCacheOptions) *FileCache {
	// Ensure the base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create cache directory: dir=%s err=%v", baseDir, err))
	}

	// Create regex pattern for filename sanitization
	filenamePattern := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)

	return &FileCache{
		baseDir:         baseDir,
		dateSubdirs:     opts.DateSubdirs,
		maxKeyLength:    opts.MaxKeyLength,
		stats:           Stats{},
		filenamePattern: filenamePattern,
	}
}

// keyToPath converts a cache key to a file path using generalized rules
func (fc *FileCache) keyToPath(key string) string {
	var pathParts []string

	// Add date subdirectory if enabled
	if fc.dateSubdirs {
		today := time.Now().Format("2006-01-02")
		pathParts = append(pathParts, fc.baseDir, today)
	} else {
		pathParts = append(pathParts, fc.baseDir)
	}

	// Generate filename from key
	filename := fc.sanitizeKey(key)

	// Hash long keys to prevent filesystem issues
	if len(filename) > fc.maxKeyLength {
		hasher := sha256.New()
		hasher.Write([]byte(key))
		hash := hex.EncodeToString(hasher.Sum(nil))[:16]
		// Keep prefix for readability
		prefix := filename[:min(50, len(filename))]
		filename = fmt.Sprintf("%s_%s", prefix, hash)
	}

	// Ensure .json extension
	if !strings.HasSuffix(filename, ".json") {
		filename += ".json"
	}

	pathParts = append(pathParts, filename)
	return filepath.Join(pathParts...)
}

// sanitizeKey converts a cache key to a filesystem-safe filename
func (fc *FileCache) sanitizeKey(key string) string {
	// Replace filesystem-unsafe characters
	safe := fc.filenamePattern.ReplaceAllString(key, "_")

	// Replace common delimiters with underscores for readability
	safe = strings.ReplaceAll(safe, ":", "_")
	safe = strings.ReplaceAll(safe, "[", "")
	safe = strings.ReplaceAll(safe, "]", "")
	safe = strings.ReplaceAll(safe, " ", "_")
	safe = strings.ReplaceAll(safe, ",", "_")

	// Remove multiple consecutive underscores
	for strings.Contains(safe, "__") {
		safe = strings.ReplaceAll(safe, "__", "_")
	}

	// Trim underscores from start/end
	safe = strings.Trim(safe, "_")

	// Ensure it's not empty
	if safe == "" {
		safe = "cache_key"
	}

	return safe
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Get retrieves a value from the file cache
func (fc *FileCache) Get(key string) ([]byte, bool) {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	filePath := fc.keyToPath(key)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fc.stats.Misses++
		return nil, false
	}

	// Read and parse the file
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		slog.Debug("cache: failed to read file", "path", filePath, "err", err)
		fc.stats.Misses++
		return nil, false
	}

	var entry fileEntry
	if err := json.Unmarshal(fileData, &entry); err != nil {
		slog.Debug("cache: failed to unmarshal cache entry", "path", filePath, "err", err)
		fc.stats.Misses++
		return nil, false
	}

	// Check if expired
	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		// Remove expired file
		os.Remove(filePath)
		fc.stats.Expired++
		fc.stats.Misses++
		return nil, false
	}

	fc.stats.Hits++
	return entry.Value, true
}

// Set stores a value in the file cache
func (fc *FileCache) Set(key string, val []byte, ttl time.Duration) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	filePath := fc.keyToPath(key)

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		slog.Warn("cache: failed to create directory", "dir", dir, "err", err)
		return
	}

	// Create entry with expiration
	entry := fileEntry{
		Value:     val,
		CreatedAt: time.Now(),
	}

	if ttl > 0 {
		entry.ExpiresAt = time.Now().Add(ttl)
	}

	// Marshal and write to file
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		slog.Warn("cache: failed to marshal entry", "key", key, "err", err)
		return
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		slog.Warn("cache: failed to write file", "path", filePath, "err", err)
		return
	}

	slog.Debug("cache: stored to file", "key", key, "path", filePath, "size", len(val))
}

// Delete removes a value from the file cache
func (fc *FileCache) Delete(key string) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	filePath := fc.keyToPath(key)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		slog.Debug("cache: failed to delete file", "path", filePath, "err", err)
	}
}

// Stats returns cache statistics
func (fc *FileCache) Stats() Stats {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.stats
}

// CleanExpired removes expired cache files (useful for maintenance)
func (fc *FileCache) CleanExpired() error {
	return filepath.Walk(fc.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.Mode().IsRegular() || filepath.Ext(path) != ".json" {
			return err
		}

		// Read and check expiration
		fileData, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip problematic files
		}

		var entry fileEntry
		if err := json.Unmarshal(fileData, &entry); err != nil {
			return nil // Skip corrupted files
		}

		// Remove if expired
		if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
			slog.Debug("cache: removing expired file", "path", path)
			os.Remove(path)
			fc.mu.Lock()
			fc.stats.Expired++
			fc.mu.Unlock()
		}

		return nil
	})
}
