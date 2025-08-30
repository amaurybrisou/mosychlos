package normalize

import (
	"context"
	"encoding/json"
	"time"
)

// NewsAPINormalizer normalizes tool "news_api" into articles.
type NewsAPINormalizer struct{}

func (NewsAPINormalizer) Can(tool string) bool { return tool == "news_api" }

type newsResp struct {
	Status       string `json:"status"`
	TotalResults int    `json:"totalResults"`
	Articles     []struct {
		Source struct {
			ID   *string `json:"id"`
			Name string  `json:"name"`
		} `json:"source"`
		Author      *string `json:"author"`
		Title       string  `json:"title"`
		Description *string `json:"description"`
		URL         string  `json:"url"`
		PublishedAt string  `json:"publishedAt"` // often RFC3339
	} `json:"articles"`
}

func (NewsAPINormalizer) Normalize(
	_ context.Context,
	toolName string,
	args json.RawMessage,
	raw json.RawMessage,
) (*Envelope, Status, error) {
	var r newsResp
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, "", err
	}

	env := &Envelope{
		SchemaVersion: "1.0",
		Provider:      "news_api",
		Tool:          "news",
		Kind:          KindNews,
		ReceivedAt:    time.Now().UTC(),
		Meta:          map[string]any{},
		Data:          &NewsData{Articles: make([]NewsArticle, 0, len(r.Articles))},
		Args:          args,
		RawResult:     raw,
	}

	for _, a := range r.Articles {
		// parse publishedAt (fallback to ReceivedAt if parse fails)
		pub, err := time.Parse(time.RFC3339, a.PublishedAt)
		if err != nil {
			pub = env.ReceivedAt
		}
		src := a.Source.Name
		env.Data.(*NewsData).Articles = append(env.Data.(*NewsData).Articles, NewsArticle{
			Source:      src,
			Title:       a.Title,
			URL:         a.URL,
			PublishedAt: toUTC(pub),
		})
	}

	if len(env.Data.(*NewsData).Articles) == 0 {
		env.Meta["empty_reason"] = "no_articles"
		return env, StatusEmpty, nil
	}
	return env, StatusOK, nil
}
