package wiremin

import (
	"encoding/json"
)

type TSPoint struct {
	T int64 // epoch seconds
	O float64
	H float64
	L float64
	C float64
	V int64
}

type TSMeta struct {
	Symbol      string
	Currency    string
	Exchange    string
	Granularity string
	Timezone    string // "Z" for UTC
}

func PackTimeseries(v int, meta TSMeta, pts []TSPoint) ([]byte, error) {
	p := Payload{
		V: v,
		K: KTimeseries,
		M: []any{meta.Symbol, meta.Currency, meta.Exchange, meta.Granularity, meta.Timezone},
		D: make([][]any, 0, len(pts)),
	}
	for _, x := range pts {
		p.D = append(p.D, []any{x.T, x.O, x.H, x.L, x.C, x.V})
	}
	// compact JSON
	return json.Marshal(p) // already compact (no indent)
}
