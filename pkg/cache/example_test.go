package cache_test

import (
	"fmt"
	"log"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/cache"
)

func ExampleFileCache() {
	// Create a new file cache in a temporary directory
	fileCache := cache.NewFileCache("/tmp/example-cache")

	// Store some data with a 1-hour expiration
	key := "user:123:profile"
	data := []byte(`{"name": "John Doe", "email": "john@example.com"}`)
	fileCache.Set(key, data, time.Hour)

	// Retrieve the data
	retrieved, found := fileCache.Get(key)
	if found {
		fmt.Printf("Found data: %s\n", string(retrieved))
	} else {
		fmt.Println("Data not found")
	}

	// Output: Found data: {"name": "John Doe", "email": "john@example.com"}
}

func ExampleFileCache_withOptions() {
	// Create a cache with custom options
	opts := cache.FileCacheOptions{
		DateSubdirs:  false, // Don't organize by date
		MaxKeyLength: 100,   // Hash keys longer than 100 characters
	}
	fileCache := cache.NewFileCacheWithOptions("/tmp/custom-cache", opts)

	// Store data with no expiration (permanent until manually deleted)
	key := "config:app:settings"
	data := []byte(`{"theme": "dark", "language": "en"}`)
	fileCache.Set(key, data, 0) // 0 duration = no expiration

	// Check cache statistics
	stats := fileCache.Stats()
	fmt.Printf("Cache stats - Hits: %d, Misses: %d\n", stats.Hits, stats.Misses)
}

func ExampleFileCache_specialCharacters() {
	fileCache := cache.NewFileCache("/tmp/special-cache")

	// The cache handles special characters in keys automatically
	specialKey := "news:[AAPL,MSFT,GOOGL] <market/analysis>"
	data := []byte("Market analysis data")
	fileCache.Set(specialKey, data, time.Hour)

	// Retrieve using the same key
	retrieved, found := fileCache.Get(specialKey)
	if found {
		fmt.Printf("Successfully stored and retrieved data with special characters\n")
		fmt.Printf("Data length: %d bytes\n", len(retrieved))
	}

	// Output:
	// Successfully stored and retrieved data with special characters
	// Data length: 20 bytes
}

func ExampleFileCache_cleanup() {
	fileCache := cache.NewFileCache("/tmp/cleanup-cache")

	// Add some entries with short expiration
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("temp:item:%d", i)
		fileCache.Set(key, []byte("temporary data"), 100*time.Millisecond)
	}

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Clean up expired entries
	err := fileCache.CleanExpired()
	if err != nil {
		log.Printf("Cleanup failed: %v", err)
	} else {
		fmt.Println("Expired entries cleaned up successfully")
	}

	// Output: Expired entries cleaned up successfully
}
