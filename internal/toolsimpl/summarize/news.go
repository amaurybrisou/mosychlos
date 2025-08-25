// Package summarize implements a news summarization tool that follows the same pattern as FMP.
package summarize

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/nlpodyssey/openai-agents-go/agents"
)

// ---- OUTPUT TYPE (typed struct for your wrappers/IO/metrics) ----

type SummaryOut struct {
	Bullets  []string       `json:"bullets"`
	Topics   []string       `json:"topics"`
	Sources  []SourceBucket `json:"sources"`
	TopLinks []string       `json:"top_links,omitempty"`
}

type SourceBucket struct {
	Source string `json:"source"`
	Count  int    `json:"count"`
}

// ---- TOOL ----

type Tool struct {
	shared bag.SharedBag
	model  string
}

// Ensure it implements models.Tool
var _ models.Tool = (*Tool)(nil)

func new(shared bag.SharedBag, model string) (*Tool, error) {
	if model == "" {
		model = "gpt-5-nano"
	}
	return &Tool{shared: shared, model: model}, nil
}

func (t *Tool) Name() string { return bag.SummarizeNews.String() }
func (t *Tool) Key() bag.Key { return bag.SummarizeNews }

func (t *Tool) Tags() []string {
	return []string{"news", "summarize", "nlp", "reports"}
}

func (t *Tool) Description() string {
	return "Summarizes a compact wire-min news payload into concise bullets, topics, and source counts"
}

func (t *Tool) IsExternal() bool { return false }

// Definition JSON Schema for function tool params
func (t *Tool) Definition() models.ToolDef {
	return &models.CustomToolDef{
		Type: models.CustomToolDefType,
		FunctionDef: models.FunctionDef{
			Name:        t.Name(),
			Description: t.Description(),
			Parameters: map[string]any{
				"title":                t.Name(),
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"wire": map[string]any{
						"type":        "string",
						"description": "wire-min news JSON: {\"v\":1,\"k\":\"n\",\"d\":[[source,title,url,epoch],...]}",
					},
					"max_bullets": map[string]any{
						"type":        "integer",
						"minimum":     1,
						"maximum":     7,
						"description": "max bullets (default 7)",
					},
					"max_topics": map[string]any{
						"type":        "integer",
						"minimum":     1,
						"maximum":     8,
						"description": "max topics (default 8)",
					},
				},
				// "wire" is optional: if empty we'll fallback to latest from SharedBag
				"required": []string{"wire", "max_bullets", "max_topics"},
			},
		},
	}
}

// Run executes the tool. It runs a tiny agent to produce a JSON summary,
// then returns a typed SummaryOut for your pipeline. The LLM boundary
// (your agents adapter) will stringify it for the model.
func (t *Tool) Run(ctx context.Context, args any) (any, error) {
	// 1) Parse args
	var in struct {
		Wire       string `json:"wire"`
		MaxBullets int    `json:"max_bullets"`
		MaxTopics  int    `json:"max_topics"`
	}
	if args != nil {
		_ = json.Unmarshal([]byte(fmt.Sprintf("%v", args)), &in) // tolerate non-strict inputs
	}
	if in.MaxBullets <= 0 || in.MaxBullets > 7 {
		in.MaxBullets = 7
	}
	if in.MaxTopics <= 0 || in.MaxTopics > 8 {
		in.MaxTopics = 8
	}

	wire := strings.TrimSpace(in.Wire)
	if wire == "" {
		wire = latestNewsWire(t.shared) // fallback to bag
	}
	if wire == "" {
		// return a typed, minimal error payload (adapter will stringify it)
		return map[string]any{"error": "no news payload available"}, nil
	}

	// 2) Build the tiny summarizer agent
	sys := `You are a news summarizer.
Input: a compact "wire-min" news JSON under the variable news_wire with k="n" and rows [source,title,url,epoch].
Return ONLY a single JSON object:
{"bullets":[<=7], "topics":[<=8], "sources":[{"source":string,"count":number}], "top_links":[<=5]?}
Rules:
- One sentence per bullet, specific & deduplicated; prefer recency and diverse sources.
- Topics: short lowercase keywords.
- No commentary, no extra fields.`

	sum := agents.New("news-summarizer").
		WithModel(t.model).
		WithTools(agents.WebSearchTool{}).
		WithInstructions(sys)

	// 3) Build a single message input
	input := fmt.Sprintf(
		"max_bullets=%d max_topics=%d\nnews_wire=%s\nSummarize per the system schema and return JSON only.",
		in.MaxBullets, in.MaxTopics, wire,
	)

	// 4) Run the agent
	res, err := agents.Run(ctx, sum, input)
	if err != nil {
		return map[string]any{"error": "summarizer_failed"}, nil
	}

	// 5) Extract text and parse to our typed struct
	outText := agents.ItemHelpers().TextMessageOutputs(res.NewItems)
	if outText == "" && res.FinalOutput != nil {
		if s, ok := res.FinalOutput.(string); ok {
			outText = s
		}
	}
	if outText == "" {
		return map[string]any{"error": "empty_summary"}, nil
	}

	var parsed SummaryOut
	if err := json.Unmarshal([]byte(outText), &parsed); err != nil {
		// If model returned non-JSON, return a typed error (adapter will stringify)
		return map[string]any{"error": "invalid_summary_json", "raw": outText}, nil
	}

	return parsed, nil
}

// ---- helpers ----

func latestNewsWire(shared bag.SharedBag) string {
	v, ok := shared.Get(bag.Key("wiremin_tool_payloads"))
	if !ok {
		return ""
	}
	var best string
	var bestAt time.Time

	// The bag holds []map[string]any or []any; we handle both.
	switch rows := v.(type) {
	case []map[string]any:
		for _, rec := range rows {
			k, _ := rec["kind"].(string)
			if k != "news" {
				continue
			}
			atStr, _ := rec["at"].(string)
			data, _ := rec["data"].(json.RawMessage)
			at, _ := time.Parse(time.RFC3339, atStr)
			if len(data) > 0 && json.Valid(data) && at.After(bestAt) {
				best, bestAt = string(data), at
			}
		}
	case []any:
		for _, it := range rows {
			if rec, ok := it.(map[string]any); ok {
				k, _ := rec["kind"].(string)
				if k != "news" {
					continue
				}
				atStr, _ := rec["at"].(string)
				data, _ := rec["data"].(json.RawMessage)
				at, _ := time.Parse(time.RFC3339, atStr)
				if len(data) > 0 && json.Valid(data) && at.After(bestAt) {
					best, bestAt = string(data), at
				}
			}
		}
	}
	return best
}
