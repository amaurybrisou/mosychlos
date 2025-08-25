package fmpestimates

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestAnalystEstimatesProvider_New(t *testing.T) {
	cases := []struct {
		name      string
		apiKey    string
		cacheDir  string
		wantError bool
	}{
		{
			name:     "valid config",
			apiKey:   "test-key",
			cacheDir: "/tmp",
		},
		{
			name:      "missing api key",
			apiKey:    "",
			cacheDir:  "/tmp",
			wantError: true,
		},
	}

	sharedBag := bag.NewSharedBag()

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			provider, err := new(c.apiKey, c.cacheDir, sharedBag)

			if c.wantError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if provider == nil {
				t.Error("expected provider but got nil")
			}
		})
	}
}

func TestAnalystEstimatesProvider_ToolInterface(t *testing.T) {
	sharedBag := bag.NewSharedBag()
	provider, err := new("test-key", "/tmp", sharedBag)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	// Test tool interface methods
	if provider.Name() == "" {
		t.Error("Name() returned empty string")
	}

	if provider.Description() == "" {
		t.Error("Description() returned empty string")
	}

	if len(provider.Tags()) == 0 {
		t.Error("Tags() returned empty slice")
	}

	def := provider.Definition()

	typedDef, ok := def.(*models.CustomToolDef)
	if !ok {
		t.Errorf("Definition() = %T, want *CustomToolDef", def)
		return
	}

	if typedDef.Name != provider.Name() {
		t.Errorf("Definition() Name = %v, want %v", typedDef.Name, provider.Name())
	}

	require.Equal(t, models.CustomToolDefType, typedDef.Type)
}

// Integration test - requires FMP_API_KEY environment variable
func TestAnalystEstimatesProvider_Run_Integration(t *testing.T) {
	// Skip if no API key available
	apiKey := os.Getenv("FMP_API_KEY")
	if apiKey == "" {
		t.Skip("FMP_API_KEY not set, skipping integration test")
	}

	sharedBag := bag.NewSharedBag()

	provider, err := new(apiKey, "/tmp", sharedBag)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	args := `{"tickers": ["AAPL"]}`

	result, err := provider.Run(ctx, args)
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	if result == "" {
		t.Error("Run() returned empty result")
	}

	// t.Logf("Result length: %d characters", len(result))
}

func TestAnalystEstimatesProvider_Run_InvalidArgs(t *testing.T) {
	sharedBag := bag.NewSharedBag()
	provider, err := new("test-key", "/tmp", sharedBag)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	ctx := context.Background()

	// Test with invalid JSON
	_, err = provider.Run(ctx, "invalid json")
	if err == nil {
		t.Error("expected error for invalid JSON but got nil")
	}

	// Test with missing tickers
	_, err = provider.Run(ctx, `{"tickers": []}`)
	if err == nil {
		t.Error("expected error for empty tickers but got nil")
	}
}

func TestAnalystEstimatesProvider_Run_Defaults(t *testing.T) {
	sharedBag := bag.NewSharedBag()
	provider, err := new("test-key", "/tmp", sharedBag)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	ctx := context.Background()

	// Test with empty args (should fail due to missing tickers)
	_, err = provider.Run(ctx, "")
	if err == nil {
		t.Error("expected error for missing tickers but got nil")
	}
}
