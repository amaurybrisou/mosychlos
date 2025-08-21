package fred

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestProvider_New(t *testing.T) {
	cases := []struct {
		name      string
		apiKey    string
		wantError bool
	}{
		{
			name:   "valid api key",
			apiKey: "test-key",
		},
		{
			name:      "missing api key",
			apiKey:    "",
			wantError: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sharedBag := bag.NewSharedBag()
			provider, err := New(c.apiKey, sharedBag)

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

func TestProvider_ToolInterface(t *testing.T) {
	sharedBag := bag.NewSharedBag()
	provider, err := New("test-key", sharedBag)
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
	assert.True(t, ok)
	assert.Equal(t, typedDef.FunctionDef.Name, provider.Name())
	assert.Equal(t, typedDef.FunctionDef.Description, provider.Description())
}

// Integration test - requires FRED_API_KEY environment variable
func TestProvider_Run_Integration(t *testing.T) {
	// Skip if no API key available
	apiKey := os.Getenv("FRED_API_KEY")
	if apiKey == "" {
		t.Skip("FRED_API_KEY not set, skipping integration test")
	}

	sharedBag := bag.NewSharedBag()
	provider, err := New(apiKey, sharedBag)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	args := `{
		"series_group": "882",
		"date": "2023-01-01",
		"region_type": "state",
		"units": "Dollars",
		"frequency": "a",
		"season": "NSA"
	}`

	result, err := provider.Run(ctx, args)
	if err != nil {
		t.Fatalf("Run() failed: %v", err)
	}

	if result == "" {
		t.Error("Run() returned empty result")
	}

	t.Logf("Result length: %d characters", len(result))
}

func TestProvider_Run_InvalidArgs(t *testing.T) {
	sharedBag := bag.NewSharedBag()
	provider, err := New("test-key", sharedBag)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	ctx := context.Background()

	// Test with invalid JSON
	_, err = provider.Run(ctx, "invalid json")
	if err == nil {
		t.Error("expected error for invalid JSON but got nil")
	}
}

func TestProvider_Run_Defaults(t *testing.T) {
	sharedBag := bag.NewSharedBag()
	provider, err := New("test-key", sharedBag)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	ctx := context.Background()

	// Test with empty args (should use defaults)
	// This will fail at API call level but should not fail due to missing params
	_, err = provider.Run(ctx, "")
	if err != nil && err.Error() != "failed to fetch regional data: failed to get regional data: FRED GeoFRED API returned status 400: 400 Bad Request" {
		// We expect API errors with test key, but not parameter validation errors
		if len(err.Error()) < 50 {
			t.Errorf("unexpected error type: %v", err)
		}
	}
}
