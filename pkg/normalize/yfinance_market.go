package normalize

import (
	"context"
	"encoding/json"
	"time"
)

// YFinanceMarketData normalizes tool "yfinance_market_data" (quote snapshots).
type YFinanceMarketData struct{}

func (YFinanceMarketData) Can(tool string) bool { return tool == "yfinance_market_data" }

type yfSnapResp struct {
	QuoteResponse struct {
		Result []struct {
			Symbol string   `json:"symbol"`
			Price  *float64 `json:"regularMarketPrice"`
			Exch   string   `json:"exchange"`
			Curr   string   `json:"currency"`
		} `json:"result"`
		Error any `json:"error"`
	} `json:"quoteResponse"`
}

func (YFinanceMarketData) Normalize(
	_ context.Context,
	toolName string,
	args json.RawMessage,
	raw json.RawMessage,
) (*Envelope, Status, error) {
	var r yfSnapResp
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, "", err
	}

	env := &Envelope{
		SchemaVersion: "1.0",
		Provider:      "yfinance",
		Tool:          "market_data",
		Kind:          KindSnapshot,
		ReceivedAt:    time.Now().UTC(),
		Meta:          map[string]any{},
		Data:          &SnapshotData{Quotes: make([]SnapshotQuote, 0)},
		Args:          args,
		RawResult:     raw,
	}

	now := time.Now().UTC()
	for _, it := range r.QuoteResponse.Result {
		if it.Price == nil {
			// skip quotes with no price
			continue
		}
		env.Data.(*SnapshotData).Quotes = append(env.Data.(*SnapshotData).Quotes, SnapshotQuote{
			Symbol:   it.Symbol,
			Price:    *it.Price,
			Exchange: it.Exch,
			Currency: it.Curr,
			AsOf:     now,
		})
	}

	if len(env.Data.(*SnapshotData).Quotes) == 0 {
		env.Meta["empty_reason"] = "no_quotes"
		return env, StatusEmpty, nil
	}
	return env, StatusOK, nil
}
