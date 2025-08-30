package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/amaurybrisou/mosychlos/pkg/normalize"
)

type NormalizeWrapper struct {
	tool     models.Tool
	registry normalize.Registry
	shared   bag.SharedBag
}

func NewNormalizeWrapper(tool models.Tool, reg normalize.Registry, shared bag.SharedBag) *NormalizeWrapper {
	return &NormalizeWrapper{tool: tool, registry: reg, shared: shared}
}

func (w *NormalizeWrapper) Name() string               { return w.tool.Name() }
func (w *NormalizeWrapper) Key() bag.Key               { return w.tool.Key() }
func (w *NormalizeWrapper) Description() string        { return w.tool.Description() }
func (w *NormalizeWrapper) IsExternal() bool           { return w.tool.IsExternal() }
func (w *NormalizeWrapper) Definition() models.ToolDef { return w.tool.Definition() }
func (w *NormalizeWrapper) Tags() []string             { return w.tool.Tags() }

// Run pass-through result; side-effect: store normalized envelope in SharedBag.
func (w *NormalizeWrapper) Run(ctx context.Context, args any) (any, error) {
	out, err := w.tool.Run(ctx, args)
	if err != nil {
		return out, fmt.Errorf("tool %s invocation error: %w", w.tool.Name(), err)
	}

	raw, ok := toRawJSON(out)
	if !ok {
		// Not JSON; record a small note, keep flowing.
		w.appendNormalized(map[string]any{
			"tool":    w.tool.Name(),
			"status":  "error",
			"error":   "non_json_tool_result",
			"at":      time.Now().UTC().Format(time.RFC3339),
			"preview": preview(out),
		})
		return out, nil
	}

	var argsRaw json.RawMessage
	if b, ok := toRawJSON(args); ok {
		argsRaw = b
	}

	n, ok := w.registry.Find(w.tool.Name())
	if !ok {
		w.appendNormalized(map[string]any{
			"tool":   w.tool.Name(),
			"status": "error",
			"error":  "no_normalizer",
			"raw":    json.RawMessage(raw),
			"at":     time.Now().UTC().Format(time.RFC3339),
		})
		return out, nil
	}

	env, status, normErr := n.Normalize(ctx, w.tool.Name(), argsRaw, raw)
	rec := map[string]any{
		"tool":   w.tool.Name(),
		"status": string(status),
		"at":     time.Now().UTC().Format(time.RFC3339),
		"raw":    json.RawMessage(raw),
	}
	if normErr != nil {
		rec["error"] = normErr.Error()
	} else {
		rec["normalized"] = env
	}
	w.appendNormalized(rec)

	return out, nil
}

func (w *NormalizeWrapper) appendNormalized(rec map[string]any) {
	key := bag.Key("normalized_tool_results")
	w.shared.Update(key, func(cur any) any {
		switch s := cur.(type) {
		case nil:
			return []map[string]any{rec}
		case []map[string]any:
			return append(s, rec)
		case []any:
			return append(s, rec)
		default:
			return []any{rec}
		}
	})
}

func toRawJSON(v any) (json.RawMessage, bool) {
	switch x := v.(type) {
	case json.RawMessage:
		return x, json.Valid(x)
	case []byte:
		if json.Valid(x) {
			return json.RawMessage(x), true
		}
		return nil, false
	case string:
		b := []byte(x)
		if json.Valid(b) {
			return json.RawMessage(b), true
		}
		return nil, false
	default:
		b, err := json.Marshal(x)
		if err != nil || !json.Valid(b) {
			return nil, false
		}
		return json.RawMessage(b), true
	}
}

func preview(v any) string {
	b, _ := json.Marshal(v)
	if len(b) > 256 {
		return string(b[:256]) + "…"
	}
	return string(b)
}
