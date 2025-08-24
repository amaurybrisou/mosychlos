package secedgar

import (
	"context"
	"testing"
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvider_New(t *testing.T) {
	cases := []struct {
		name      string
		userAgent string
		baseURL   string
		wantError bool
	}{
		{
			name:      "valid config",
			userAgent: "test-agent",
			baseURL:   "https://data.sec.gov",
		},
		{
			name:      "missing user agent",
			userAgent: "",
			baseURL:   "https://data.sec.gov",
			wantError: true,
		},
		{
			name:      "default base url",
			userAgent: "test-agent",
			baseURL:   "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			provider, err := new(c.userAgent, c.baseURL, bag.NewSharedBag())

			if c.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, provider)
		})
	}
}

func TestProvider_NewFromConfig(t *testing.T) {
	cfg := &config.SECEdgarConfig{
		UserAgent: "test-user-agent",
		BaseURL:   "https://data.sec.gov",
	}

	provider, err := new(cfg.UserAgent, cfg.BaseURL, bag.NewSharedBag())
	require.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestProvider_ToolInterface(t *testing.T) {
	provider, err := new("test-agent", "https://data.sec.gov", bag.NewSharedBag())
	require.NoError(t, err)

	// Test tool interface methods
	assert.Equal(t, "sec_edgar_filings", provider.Name())
	assert.Equal(t, bag.SECFilings, provider.Key())
	assert.NotEmpty(t, provider.Description())
	assert.NotEmpty(t, provider.Tags())

	// Test definition structure
	def := provider.Definition()

	typedDef, ok := def.(*models.CustomToolDef)
	assert.True(t, ok)
	assert.Equal(t, models.CustomToolDefType, typedDef.Type)
	assert.Equal(t, provider.Name(), typedDef.FunctionDef.Name)
	assert.Equal(t, provider.Description(), typedDef.FunctionDef.Description)
	assert.NotNil(t, typedDef.FunctionDef.Parameters)

	// Check required parameters structure
	params, ok := typedDef.FunctionDef.Parameters["properties"].(map[string]any)
	require.True(t, ok, "parameters should have properties field")

	// Check action parameter
	action, ok := params["action"].(map[string]any)
	require.True(t, ok, "should have action parameter")
	assert.Equal(t, "string", action["type"])

	enum, ok := action["enum"].([]string)
	require.True(t, ok, "action should have enum values")
	expectedActions := []string{"tickers", "facts", "filings", "insider_transactions"}
	assert.ElementsMatch(t, expectedActions, enum)
}

func TestProvider_Run_InvalidArgs(t *testing.T) {
	provider, err := new("test-agent", "https://data.sec.gov", bag.NewSharedBag())
	require.NoError(t, err)

	ctx := context.Background()

	cases := []struct {
		name string
		args string
	}{
		{
			name: "invalid json",
			args: `{"action":}`,
		},
		{
			name: "missing action",
			args: `{"cik": "0000320193"}`,
		},
		{
			name: "invalid action",
			args: `{"action": "invalid"}`,
		},
		{
			name: "facts without cik or ticker",
			args: `{"action": "facts"}`,
		},
		{
			name: "filings without cik or ticker",
			args: `{"action": "filings"}`,
		},
		{
			name: "insider_transactions without cik or ticker",
			args: `{"action": "insider_transactions"}`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := provider.Run(ctx, c.args)
			assert.Error(t, err)
		})
	}
}

func TestProvider_Run_ValidArgs(t *testing.T) {
	provider, err := new("test-agent", "https://data.sec.gov", bag.NewSharedBag())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test tickers action (doesn't require additional params)
	args := `{"action": "tickers"}`
	_, err = provider.Run(ctx, args)
	// This might fail due to network/API but shouldn't fail on validation
	// The error should be from the SEC client, not parameter validation
	if err != nil {
		assert.Contains(t, err.Error(), "SEC Edgar tool execution failed")
	}
}
