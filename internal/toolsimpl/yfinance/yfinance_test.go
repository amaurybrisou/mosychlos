package yfinance

import (
	"context"
	"testing"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/stretchr/testify/assert"
)

var testConfig = &config.YFinanceConfig{
	BaseURL: "https://query1.finance.yahoo.com",
	Timeout: 30,
}

func TestYFinanceStockDataTool(t *testing.T) {
	sharedBag := bag.NewSharedBag()
	tool, err := newStockDataFromConfig(testConfig, sharedBag)
	assert.NoError(t, err)
	assert.NotNil(t, tool)

	t.Run("Tool Properties", func(t *testing.T) {
		assert.Equal(t, "yfinance_stock_data", tool.Name())
		assert.Equal(t, false, tool.IsExternal())
		assert.Contains(t, tool.Tags(), "finance")
	})

	t.Run("Tool Definition", func(t *testing.T) {
		def := tool.Definition()

		typedDef, ok := def.(*models.CustomToolDef)
		assert.True(t, ok)
		assert.Equal(t, models.CustomToolDefType, typedDef.Type)
		assert.Equal(t, "yfinance_stock_data", typedDef.FunctionDef.Name)
		assert.NotEmpty(t, typedDef.FunctionDef.Description)
	})

	t.Run("Run Should Execute", func(t *testing.T) {
		result, err := tool.Run(context.Background(), `{"symbol": "AAPL"}`)
		assert.NoError(t, err) // Should succeed
		assert.NotEmpty(t, result)
	})
}

func TestYFinanceMarketDataTool(t *testing.T) {
	sharedBag := bag.NewSharedBag()
	tool, err := newMarketDataFromConfig(testConfig, sharedBag)
	assert.NoError(t, err)
	assert.NotNil(t, tool)

	t.Run("Tool Properties", func(t *testing.T) {
		assert.Equal(t, "yfinance_market_data", tool.Name())
		assert.Equal(t, false, tool.IsExternal())
		assert.Contains(t, tool.Tags(), "financial")
	})

	t.Run("Tool Definition", func(t *testing.T) {
		def := tool.Definition()

		typedDef, ok := def.(*models.CustomToolDef)
		assert.True(t, ok)
		assert.Equal(t, models.CustomToolDefType, typedDef.Type)
		assert.Equal(t, "yfinance_market_data", typedDef.FunctionDef.Name)
		assert.NotEmpty(t, typedDef.FunctionDef.Description)
	})

	t.Run("Run Should Execute", func(t *testing.T) {
		result, err := tool.Run(context.Background(), `{"indices": ["^GSPC"]}`)
		assert.Error(t, err) // Should error because not implemented
		assert.Empty(t, result)
	})
}
