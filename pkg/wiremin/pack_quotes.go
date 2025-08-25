package wiremin

import "encoding/json"

type QuoteRow struct {
	Symbol   string
	Price    float64
	Exchange string
	Currency string
	AsOf     int64 // epoch
}

func PackQuotes(v int, rows []QuoteRow) ([]byte, error) {
	p := Payload{V: v, K: KQuotes, D: make([][]any, 0, len(rows))}
	for _, r := range rows {
		p.D = append(p.D, []any{r.Symbol, r.Price, r.Exchange, r.Currency, r.AsOf})
	}
	return json.Marshal(p)
}
