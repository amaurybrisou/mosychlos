package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFileCache_BasicOperations(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	cache := NewFileCache(tmpDir)

	// Test Set and Get
	key := "test:key"
	value := []byte("test value")

	cache.Set(key, value, time.Hour)

	retrieved, found := cache.Get(key)
	if !found {
		t.Error("Expected to find cached value")
	}

	if string(retrieved) != string(value) {
		t.Errorf("Expected %s, got %s", string(value), string(retrieved))
	}

	// Test stats
	stats := cache.Stats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}
}

func TestFileCache_Expiration(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewFileCache(tmpDir)

	key := "expiring:key"
	value := []byte("expiring value")

	// Set with short TTL
	cache.Set(key, value, 100*time.Millisecond)

	// Should be available immediately
	_, found := cache.Get(key)
	if !found {
		t.Error("Expected to find value before expiration")
	}

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Should be expired
	_, found = cache.Get(key)
	if found {
		t.Error("Expected value to be expired")
	}

	stats := cache.Stats()
	if stats.Expired == 0 {
		t.Error("Expected at least one expired entry")
	}
}

func TestFileCache_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewFileCache(tmpDir)

	key := "deletable:key"
	value := []byte("deletable value")

	cache.Set(key, value, time.Hour)

	// Verify it exists
	_, found := cache.Get(key)
	if !found {
		t.Error("Expected to find value before deletion")
	}

	// Delete it
	cache.Delete(key)

	// Should not be found
	_, found = cache.Get(key)
	if found {
		t.Error("Expected value to be deleted")
	}
}

func TestFileCache_KeySanitization(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewFileCache(tmpDir)

	testCases := []struct {
		name        string
		key         string
		expectValid bool
	}{
		{
			name:        "Simple key",
			key:         "simple",
			expectValid: true,
		},
		{
			name:        "Key with colons",
			key:         "namespace:action:id",
			expectValid: true,
		},
		{
			name:        "Key with special characters",
			key:         "key<>:\"/\\|?*with special chars",
			expectValid: true,
		},
		{
			name:        "Key with brackets and spaces",
			key:         "news:[AAPL MSFT GOOGL]",
			expectValid: true,
		},
		{
			name:        "Empty key",
			key:         "",
			expectValid: true, // Should be handled gracefully
		},
		{
			name:        "Very long key",
			key:         strings.Repeat("very-long-key-", 50), // 650+ chars
			expectValid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value := []byte("test value for " + tc.name)

			// This should not panic or fail
			cache.Set(tc.key, value, time.Hour)

			retrieved, found := cache.Get(tc.key)
			if !found {
				t.Errorf("Expected to find value for key: %s", tc.key)
			}

			if string(retrieved) != string(value) {
				t.Errorf("Retrieved value doesn't match for key %s", tc.key)
			}
		})
	}
}

func TestFileCache_NoDateSubdirs(t *testing.T) {
	tmpDir := t.TempDir()

	opts := FileCacheOptions{
		DateSubdirs:  false,
		MaxKeyLength: 200,
	}
	cache := NewFileCacheWithOptions(tmpDir, opts)

	key := "no-date:key"
	value := []byte("no date value")

	cache.Set(key, value, time.Hour)

	// Check that file is created directly in base directory (no date subdirs)
	filePath := cache.keyToPath(key)
	if strings.Contains(filePath, time.Now().Format("2006-01-02")) {
		t.Error("Expected no date subdirectory when DateSubdirs is false")
	}

	// Should still be retrievable
	retrieved, found := cache.Get(key)
	if !found || string(retrieved) != string(value) {
		t.Error("Failed to retrieve value from cache without date subdirs")
	}
}

func TestFileCache_CustomMaxKeyLength(t *testing.T) {
	tmpDir := t.TempDir()

	opts := FileCacheOptions{
		DateSubdirs:  true,
		MaxKeyLength: 50, // Very short for testing
	}
	cache := NewFileCacheWithOptions(tmpDir, opts)

	longKey := strings.Repeat("a", 100) // Longer than maxKeyLength
	value := []byte("long key value")

	cache.Set(longKey, value, time.Hour)

	// Should still work (key gets hashed)
	retrieved, found := cache.Get(longKey)
	if !found || string(retrieved) != string(value) {
		t.Error("Failed to handle long key with hashing")
	}

	// Verify file path is reasonable length
	filePath := cache.keyToPath(longKey)
	filename := filepath.Base(filePath)
	// Should be much shorter due to hashing, but still readable
	if len(filename) > 100 { // Reasonable upper bound
		t.Errorf("Hashed filename too long: %s (length: %d)", filename, len(filename))
	}
}

func TestFileCache_CleanExpired(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewFileCache(tmpDir)

	// Add some entries with different expiration times
	cache.Set("short:lived", []byte("short"), 50*time.Millisecond)
	cache.Set("long:lived", []byte("long"), time.Hour)
	cache.Set("no:expiry", []byte("forever"), 0) // No expiry

	// Wait for short-lived to expire
	time.Sleep(100 * time.Millisecond)

	// Run cleanup
	err := cache.CleanExpired()
	if err != nil {
		t.Errorf("CleanExpired failed: %v", err)
	}

	// Short-lived should be gone
	_, found := cache.Get("short:lived")
	if found {
		t.Error("Expected expired entry to be cleaned up")
	}

	// Others should still exist
	_, found = cache.Get("long:lived")
	if !found {
		t.Error("Expected long-lived entry to still exist")
	}

	_, found = cache.Get("no:expiry")
	if !found {
		t.Error("Expected non-expiring entry to still exist")
	}
}

func TestFileCache_FileFormat(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewFileCache(tmpDir)

	key := "format:test"
	value := []byte("format test value")

	cache.Set(key, value, time.Hour)

	// Read the file directly to verify format
	filePath := cache.keyToPath(key)
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read cache file: %v", err)
	}

	// Parse JSON to verify structure
	var entry fileEntry
	if err := json.Unmarshal(fileData, &entry); err != nil {
		t.Fatalf("Failed to parse cache file JSON: %v", err)
	}

	// Verify entry structure
	if string(entry.Value) != string(value) {
		t.Errorf("Expected value %s, got %s", string(value), string(entry.Value))
	}

	if entry.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if entry.ExpiresAt.IsZero() {
		t.Error("Expected ExpiresAt to be set when TTL is provided")
	}
}

func TestFileCache_Stats(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewFileCache(tmpDir)

	initialStats := cache.Stats()
	if initialStats.Hits != 0 || initialStats.Misses != 0 {
		t.Error("Expected initial stats to be zero")
	}

	// Generate some hits and misses
	cache.Set("exists", []byte("value"), time.Hour)

	// Hit
	cache.Get("exists")

	// Miss
	cache.Get("nonexistent")

	// Expired (for expired count)
	cache.Set("expired", []byte("expired"), time.Nanosecond)
	time.Sleep(time.Millisecond)
	cache.Get("expired")

	stats := cache.Stats()
	if stats.Hits == 0 {
		t.Error("Expected at least one hit")
	}
	if stats.Misses == 0 {
		t.Error("Expected at least one miss")
	}
	if stats.Expired == 0 {
		t.Error("Expected at least one expired entry")
	}
}

func TestFileCache_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewFileCache(tmpDir)

	// Test concurrent access doesn't cause race conditions
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("concurrent:key:%d", i)
			value := []byte(fmt.Sprintf("value-%d", i))
			cache.Set(key, value, time.Hour)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("concurrent:key:%d", i%10) // Read some existing keys
			cache.Get(key)
		}
		done <- true
	}()

	// Wait for both to complete
	<-done
	<-done

	// Should not have panicked or crashed
	stats := cache.Stats()
	if stats.Hits == 0 && stats.Misses == 0 {
		t.Error("Expected some cache activity from concurrent access")
	}
}
