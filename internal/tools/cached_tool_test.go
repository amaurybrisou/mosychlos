package tools

//go:generate mockgen -source=../../pkg/models/ai.go -destination=../../pkg/models/mocks/mock_ai.go -package=mocks

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/internal/tools/newsapi"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models/mocks"
	"github.com/golang/mock/gomock"
)

func TestWrapWithCache_NewsAPI_CacheEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock tool that tracks how many times it's called
	mockTool := mocks.NewMockTool(ctrl)
	mockTool.EXPECT().Key().Return(keys.NewsApi).AnyTimes()
	mockTool.EXPECT().Name().Return("newsapi").AnyTimes()

	// Set up expectations for the tool calls
	result := `{"articles":[{"title":"Test Article","source":"Test Source"}]}`

	// Create config with cache enabled
	cfg := &config.Config{
		CacheDir: t.TempDir(), // Use temp directory for testing
		Tools: config.ToolsConfig{
			NewsAPI: &config.NewsAPIConfig{
				CacheEnable: true,
			},
		},
	}

	// Wrap with cache
	wrappedTool := wrapWithCache(mockTool, &ToolCacheConfig{
		Enabled: cfg.Tools.NewsAPI.CacheEnable,
		TTL:     24 * time.Hour,
		BaseDir: filepath.Join(cfg.CacheDir, "tools"),
	})

	// Verify it's wrapped with CachedTool
	cachedTool, ok := wrappedTool.(*CachedTool)
	if !ok {
		t.Fatal("Expected tool to be wrapped with CachedTool when cache is enabled")
	}

	if cachedTool.tool != mockTool {
		t.Error("Expected wrapped tool to contain the original tool")
	}

	// Test cache functionality: call the same arguments multiple times
	ctx := context.Background()
	args := `{"topics":["bitcoin","crypto"]}`

	// First call should execute the underlying tool and cache the result
	mockTool.EXPECT().Run(gomock.Any(), args).Return(result, nil).Times(1)

	result1, err := cachedTool.Run(ctx, args)
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}

	// Second call with same args should hit cache, not call underlying tool
	// No additional EXPECT() call needed - the mock should not be called again
	result2, err := cachedTool.Run(ctx, args)
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}

	// Results should be identical
	if result1 != result2 {
		t.Error("Expected identical results from cached call")
	}

	// Verify cache stats show hits
	stats := cachedTool.cache.Stats()
	if stats.Hits == 0 {
		t.Error("Expected at least one cache hit")
	}

	// Third call with different args should call underlying tool again
	differentArgs := `{"topics":["stocks","market"]}`
	mockTool.EXPECT().Run(gomock.Any(), differentArgs).Return(result, nil).Times(1)

	result3, err := cachedTool.Run(ctx, differentArgs)
	if err != nil {
		t.Fatalf("Third call with different args failed: %v", err)
	}

	// Results should be the same (our mock returns same result)
	if result3 != result1 {
		t.Error("Expected same result from mock tool")
	}
}

func TestWrapWithCache_NewsAPI_CacheDisabled(t *testing.T) {
	// Create a NewsAPI tool
	sharedBag := bag.NewSharedBag()
	newsTool, err := newsapi.New("test-key", "https://newsapi.org/v2", "en", sharedBag)
	if err != nil {
		t.Fatalf("Failed to create NewsAPI tool: %v", err)
	}

	// Create config with cache disabled
	cfg := &config.Config{
		CacheDir: os.TempDir(),
		Tools: config.ToolsConfig{
			NewsAPI: &config.NewsAPIConfig{
				CacheEnable: false,
			},
		},
	}

	// Wrap with cache
	wrappedTool := wrapWithCache(newsTool, &ToolCacheConfig{
		Enabled: cfg.Tools.NewsAPI.CacheEnable,
		TTL:     24 * time.Hour,
		BaseDir: filepath.Join(cfg.CacheDir, "tools"),
	})

	// Verify it's NOT wrapped (should be the original tool)
	if wrappedTool != newsTool {
		t.Error("Expected tool to NOT be wrapped when cache is disabled")
	}
}

func TestWrapWithCache_CacheDisabled_NoWrap(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock tool with an unknown key
	mockTool := mocks.NewMockTool(ctrl)
	mockTool.EXPECT().Key().Return(keys.Key("unknown")).AnyTimes()

	// Wrap with cache disabled - should not be cached
	wrappedTool := wrapWithCache(mockTool, &ToolCacheConfig{
		Enabled: false, // Cache disabled
		TTL:     24 * time.Hour,
		BaseDir: os.TempDir(),
	})

	// Verify it's NOT wrapped when cache is disabled
	if wrappedTool != mockTool {
		t.Error("Expected tool to NOT be wrapped when cache is disabled")
	}
}
