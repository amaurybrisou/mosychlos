package cache

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkFileCache_Set(b *testing.B) {
	tmpDir := b.TempDir()
	cache := NewFileCache(tmpDir)
	value := []byte("benchmark test value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("benchmark:set:%d", i)
		cache.Set(key, value, time.Hour)
	}
}

func BenchmarkFileCache_Get(b *testing.B) {
	tmpDir := b.TempDir()
	cache := NewFileCache(tmpDir)
	value := []byte("benchmark test value")

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("benchmark:get:%d", i)
		cache.Set(key, value, time.Hour)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("benchmark:get:%d", i%1000)
		cache.Get(key)
	}
}

func BenchmarkFileCache_SetGet(b *testing.B) {
	tmpDir := b.TempDir()
	cache := NewFileCache(tmpDir)
	value := []byte("benchmark test value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("benchmark:setget:%d", i)
		cache.Set(key, value, time.Hour)
		cache.Get(key)
	}
}

func BenchmarkFileCache_KeySanitization(b *testing.B) {
	tmpDir := b.TempDir()
	cache := NewFileCache(tmpDir)

	// Test with keys that require heavy sanitization
	dirtyKey := "namespace<>:\"/\\|?*with[MANY SPECIAL] chars,and,commas"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.sanitizeKey(dirtyKey)
	}
}

func BenchmarkFileCache_LongKeyHashing(b *testing.B) {
	tmpDir := b.TempDir()
	cache := NewFileCacheWithOptions(tmpDir, FileCacheOptions{
		DateSubdirs:  true,
		MaxKeyLength: 50, // Force hashing
	})

	longKey := fmt.Sprintf("very-long-key-that-will-trigger-hashing-%s", string(make([]byte, 200)))
	value := []byte("test value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(longKey, value, time.Hour)
		cache.Get(longKey)
	}
}
