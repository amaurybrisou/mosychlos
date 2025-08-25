// Package normalize provides types and utilities for working with normalized data.
package normalize

import (
	"context"
	"encoding/json"
	"time"
)

// Kind is the normalized data category.
type Kind string

const (
	KindTimeseries Kind = "timeseries"
	KindSnapshot   Kind = "snapshot"
	KindNews       Kind = "news"
)

// Status indicates whether normalized data contains useful content.
type Status string

const (
	StatusOK    Status = "ok"
	StatusEmpty Status = "empty"
)

// Envelope is the normalized, provider-agnostic wrapper you can store or feed to downstream engines.
type Envelope struct {
	SchemaVersion string          `json:"schema_version"`       // e.g. "1.0"
	Provider      string          `json:"provider"`             // e.g. "yfinance", "news_api"
	Tool          string          `json:"tool"`                 // e.g. "stock_data", "market_data", "news"
	Kind          Kind            `json:"kind"`                 // one of Kind*
	ReceivedAt    time.Time       `json:"received_at"`          // UTC RFC3339
	Meta          map[string]any  `json:"meta,omitempty"`       // flat, compact
	Data          any             `json:"data,omitempty"`       // typed below (TimeseriesData, SnapshotData, NewsData)
	Args          json.RawMessage `json:"args,omitempty"`       // optional: the arguments used for the tool call
	RawResult     json.RawMessage `json:"raw_result,omitempty"` // optional: vendor blob for audit/debug
}

// TimeseriesData is the strongly typed content for KindTimeseries.
type TimeseriesData struct {
	// Points sorted ascending by time.
	Points []TimeseriesPoint `json:"points"`
}

// TimeseriesPoint is a single OHLCV observation.
type TimeseriesPoint struct {
	T time.Time `json:"t"`           // UTC RFC3339
	O float64   `json:"o,omitempty"` // open
	H float64   `json:"h,omitempty"` // high
	L float64   `json:"l,omitempty"` // low
	C float64   `json:"c,omitempty"` // close
	V int64     `json:"v,omitempty"` // volume
}

// SnapshotData is the strongly typed content for KindSnapshot.
type SnapshotData struct {
	Quotes []SnapshotQuote `json:"quotes"`
}

// SnapshotQuote is a single quote snapshot.
type SnapshotQuote struct {
	Symbol   string    `json:"symbol"`
	Price    float64   `json:"price"`
	Exchange string    `json:"exchange,omitempty"`
	Currency string    `json:"currency,omitempty"`
	AsOf     time.Time `json:"as_of"` // UTC RFC3339
}

// NewsData is the strongly typed content for KindNews.
type NewsData struct {
	Articles []NewsArticle `json:"articles"`
}

// NewsArticle is a minimal normalized news record.
type NewsArticle struct {
	Source      string    `json:"source,omitempty"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	PublishedAt time.Time `json:"published_at"` // UTC RFC3339
}

// Normalizer converts (toolName, args, rawVendor) into an Envelope.
// It returns a Status indicating whether useful content exists.
type Normalizer interface {
	Can(toolName string) bool
	Normalize(ctx context.Context, toolName string, args json.RawMessage, raw json.RawMessage) (*Envelope, Status, error)
}
