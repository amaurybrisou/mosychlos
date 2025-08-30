package wiremin

import "encoding/json"

type NewsRow struct {
	Source string
	Title  string
	URL    string
	Pub    int64 // epoch
}

func PackNews(v int, rows []NewsRow) ([]byte, error) {
	p := Payload{V: v, K: KNews, D: make([][]any, 0, len(rows))}
	for _, r := range rows {
		p.D = append(p.D, []any{r.Source, r.Title, r.URL, r.Pub})
	}
	return json.Marshal(p)
}
