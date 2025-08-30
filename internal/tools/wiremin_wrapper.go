package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/amaurybrisou/mosychlos/pkg/normalize"
	"github.com/amaurybrisou/mosychlos/pkg/wiremin"
)

type WireMinWrapper struct {
	tool     models.Tool
	registry normalize.Registry
	shared   bag.SharedBag
}

func NewWireMinWrapper(tool models.Tool, reg normalize.Registry, shared bag.SharedBag) *WireMinWrapper {
	return &WireMinWrapper{tool: tool, registry: reg, shared: shared}
}

func (w *WireMinWrapper) Name() string               { return w.tool.Name() }
func (w *WireMinWrapper) Key() bag.Key               { return w.tool.Key() }
func (w *WireMinWrapper) Description() string        { return w.tool.Description() }
func (w *WireMinWrapper) IsExternal() bool           { return w.tool.IsExternal() }
func (w *WireMinWrapper) Definition() models.ToolDef { return w.tool.Definition() }
func (w *WireMinWrapper) Tags() []string             { return w.tool.Tags() }

// Run return token-minimized JSON to the agent/LLM.
// Also record the packed bytes in SharedBag (side-channel) if you want.
func (w *WireMinWrapper) Run(ctx context.Context, args any) (any, error) {
	out, err := w.tool.Run(ctx, args)
	if err != nil {
		return out, fmt.Errorf("tool %s invocation error: %w", w.tool.Name(), err)
	}

	raw, ok := toRawJSON(out)
	if !ok {
		return out, nil // nothing we can do
	}

	var argsRaw json.RawMessage
	if b, ok := toRawJSON(args); ok {
		argsRaw = b
	}

	// Normalize first (idempotent if NormalizeWrapper already ran; cheap re-do).
	norm, status, nerr := w.normalize(ctx, w.tool.Name(), argsRaw, raw)
	if nerr != nil || status == normalize.StatusEmpty || norm == nil {
		return out, nil
	}

	// Pack to wire-min depending on kind:
	var packed []byte
	switch norm.Kind {
	case normalize.KindTimeseries:
		td, ok := norm.Data.(*normalize.TimeseriesData)
		if !ok || td == nil {
			return out, nil
		}
		meta := wiremin.TSMeta{
			Symbol:      asString(norm.Meta["symbol"]),
			Currency:    asString(norm.Meta["currency"]),
			Exchange:    asString(norm.Meta["exchange"]),
			Granularity: asString(norm.Meta["granularity"]),
			Timezone:    "Z",
		}
		points := make([]wiremin.TSPoint, 0, len(td.Points))
		for _, p := range td.Points {
			points = append(points, wiremin.TSPoint{T: p.T.Unix(), O: p.O, H: p.H, L: p.L, C: p.C, V: p.V})
		}
		packed, _ = wiremin.PackTimeseries(1, meta, points)

	case normalize.KindSnapshot:
		sd, ok := norm.Data.(*normalize.SnapshotData)
		if !ok || sd == nil {
			return out, nil
		}
		rows := make([]wiremin.QuoteRow, 0, len(sd.Quotes))
		for _, q := range sd.Quotes {
			rows = append(rows, wiremin.QuoteRow{
				Symbol: q.Symbol, Price: q.Price, Exchange: q.Exchange, Currency: q.Currency, AsOf: q.AsOf.Unix(),
			})
		}
		packed, _ = wiremin.PackQuotes(1, rows)

	case normalize.KindNews:
		nd, ok := norm.Data.(*normalize.NewsData)
		if !ok || nd == nil {
			return out, nil
		}
		rows := make([]wiremin.NewsRow, 0, len(nd.Articles))
		for _, a := range nd.Articles {
			rows = append(rows, wiremin.NewsRow{Source: a.Source, Title: a.Title, URL: a.URL, Pub: a.PublishedAt.Unix()})
		}
		packed, _ = wiremin.PackNews(1, rows)

	default:
		return out, nil
	}

	if len(packed) == 0 {
		return out, nil
	}

	// Side-channel (optional): stash the packed payload
	key := bag.Key("wiremin_tool_payloads")
	rec := map[string]any{
		"tool": w.tool.Name(),
		"at":   time.Now().UTC().Format(time.RFC3339),
		"kind": string(norm.Kind),
		"v":    1,
		"data": json.RawMessage(packed),
	}
	w.shared.Update(key, func(cur any) any {
		switch s := cur.(type) {
		case nil:
			return nil
		case []map[string]any:
			return append(s, rec)
		case []any:
			return append(s, rec)
		default:
			return []any{rec}
		}
	})

	// Return wire-minified JSON to the agent. This is what the LLM will "see".
	return string(packed), nil
}

func (w *WireMinWrapper) normalize(ctx context.Context, tool string, args, raw json.RawMessage) (*normalize.Envelope, normalize.Status, error) {
	n, ok := w.registry.Find(tool)
	if !ok {
		return nil, "", fmt.Errorf("no normalizer for tool %q", tool)
	}
	return n.Normalize(ctx, tool, args, raw)
}

func asString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
