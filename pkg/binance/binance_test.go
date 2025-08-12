package binance

import (
	"testing"

	"github.com/amaurybrisou/mosychlos/internal/config"
)

func TestNew(t *testing.T) {
	t.Parallel()

	cfg := &config.BinanceConfig{
		APIKey:    "test_key",
		APISecret: "test_secret",
		BaseURL:   "https://testnet.binance.vision",
	}

	client := New(cfg)
	if client == nil {
		t.Error("expected non-nil client")
	}
}

func TestNewPortfolioProvider(t *testing.T) {
	t.Parallel()

	cfg := &config.BinanceConfig{
		APIKey:    "test_key",
		APISecret: "test_secret",
	}

	provider := NewPortfolioProvider(cfg)
	if provider == nil {
		t.Error("expected non-nil portfolio provider")
	}
}
