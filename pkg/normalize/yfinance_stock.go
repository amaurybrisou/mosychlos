package normalize

import (
	"context"
	"encoding/json"
	"time"
)

// YFinanceStockData normalizes tool "yfinance_stock_data" into a timeseries Envelope.
// This handles columnar arrays with possible nulls in the quote arrays.
type YFinanceStockData struct{}

func (YFinanceStockData) Can(tool string) bool { return tool == "yfinance_stock_data" }

type yfStockResp struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Currency     string `json:"currency"`
				Symbol       string `json:"symbol"`
				ExchangeName string `json:"exchangeName"`
				Timezone     string `json:"timezone"`
				DataGran     string `json:"dataGranularity"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []*float64 `json:"open"`
					High   []*float64 `json:"high"`
					Low    []*float64 `json:"low"`
					Close  []*float64 `json:"close"`
					Volume []*int64   `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Error any `json:"error"`
	} `json:"chart"`
}

func (YFinanceStockData) Normalize(
	ctx context.Context,
	toolName string,
	args json.RawMessage,
	raw json.RawMessage,
) (*Envelope, Status, error) {
	var r yfStockResp
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, "", err
	}

	env := &Envelope{
		SchemaVersion: "1.0",
		Provider:      "yfinance",
		Tool:          "stock_data",
		Kind:          KindTimeseries,
		ReceivedAt:    time.Now().UTC(),
		Meta:          map[string]any{},
		Data:          &TimeseriesData{Points: make([]TimeseriesPoint, 0)},
		Args:          args,
		RawResult:     raw,
	}

	if len(r.Chart.Result) == 0 {
		env.Meta["empty_reason"] = "no_result"
		return env, StatusEmpty, nil
	}
	res := r.Chart.Result[0]
	env.Meta["symbol"] = res.Meta.Symbol
	env.Meta["currency"] = res.Meta.Currency
	env.Meta["exchange"] = res.Meta.ExchangeName
	env.Meta["granularity"] = res.Meta.DataGran

	tz := res.Meta.Timezone
	if tz == "" {
		tz = "UTC"
	}
	env.Meta["timezone"] = tz

	// if no timestamp or no quote arrays -> empty
	if len(res.Timestamp) == 0 || len(res.Indicators.Quote) == 0 {
		env.Meta["empty_reason"] = "provider_null_series"
		return env, StatusEmpty, nil
	}

	q := res.Indicators.Quote[0]
	n := len(res.Timestamp)

	// defensively clamp all arrays to the shortest
	clamp := func(a int, lens ...int) int {
		min := a
		for _, l := range lens {
			if l < min {
				min = l
			}
		}
		return min
	}
	n = clamp(n, len(q.Open), len(q.High), len(q.Low), len(q.Close), len(q.Volume))

	points := make([]TimeseriesPoint, 0, n)
	for i := 0; i < n; i++ {
		// skip rows where OHLCV are entirely nil
		if (q.Open[i] == nil) && (q.High[i] == nil) && (q.Low[i] == nil) && (q.Close[i] == nil) && (q.Volume[i] == nil) {
			continue
		}

		// use zero values for missing fields but keep row (you can change behavior if you prefer strict completeness)
		var o, h, l, c float64
		var v int64
		if q.Open[i] != nil {
			o = *q.Open[i]
		}
		if q.High[i] != nil {
			h = *q.High[i]
		}
		if q.Low[i] != nil {
			l = *q.Low[i]
		}
		if q.Close[i] != nil {
			c = *q.Close[i]
		}
		if q.Volume[i] != nil {
			v = *q.Volume[i]
		}

		points = append(points, TimeseriesPoint{
			T: epochToUTC(res.Timestamp[i]),
			O: o, H: h, L: l, C: c, V: v,
		})
	}

	if len(points) == 0 {
		env.Meta["empty_reason"] = "no_points_after_zip"
		return env, StatusEmpty, nil
	}

	// sort ascending by time
	sortPointsAsc(points)
	env.Data = &TimeseriesData{Points: points}
	return env, StatusOK, nil
}
