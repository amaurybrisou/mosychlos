package risk

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	pkgagents "github.com/amaurybrisou/mosychlos/pkg/agents"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/nlpodyssey/openai-agents-go/agents"
)

type RiskAgentEngine struct {
	agent         *agents.Agent
	promptBuilder models.PromptBuilder
	resultKey     bag.Key
}

func NewRiskAgentEngine(sharedBag bag.SharedBag, pb models.PromptBuilder, toolProvider models.ToolProvider) *RiskAgentEngine {
	tools := toolProvider.List()

	// TODO: develop the tool error function
	agentTools := pkgagents.FromToolsToAgent(tools)

	agent := agents.New("risk-engine").
		WithModel("gpt-5-nano").
		WithInstructions(`Be a investment portfolio risk analyst.`).
		WithTools(agentTools...)

	return &RiskAgentEngine{
		agent:         agent,
		promptBuilder: pb,
		resultKey:     bag.KRiskAnalysisResult,
	}
}

func (r *RiskAgentEngine) Execute(ctx context.Context, _ models.AiClient, sb bag.SharedBag) error {
	prompt, err := r.promptBuilder.BuildPrompt(ctx, models.AnalysisRisk)
	if err != nil {
		return fmt.Errorf("failed to build prompt: %w", err)
	}

	runner := agents.Runner{
		Config: agents.RunConfig{
			MaxTurns: 30,
		},
	}

	result, err := runner.Run(ctx, r.agent, prompt)
	if err != nil {
		return fmt.Errorf("failed to run agent: %w", err)
	}

	sb.Set(r.resultKey, result.FinalOutput)

	return nil
}

func (r *RiskAgentEngine) Name() string {
	return "risk-agent-engine"
}

func (r *RiskAgentEngine) ResultKey() bag.Key {
	return r.resultKey
}

// NewsHandoffMessageFilter We’ll look into the bag for the latest wire-min news payload we stashed earlier.
// In the wrapper example we used key "wiremin_tool_payloads" with records like:
// { "tool": "...", "kind": "news", "v": 1, "data": <json.RawMessage>, "at": "..." }
func NewsHandoffMessageFilter(callerAgent *agents.Agent, sharedBag bag.SharedBag) func(context.Context, agents.HandoffInputData) (agents.HandoffInputData, error) {
	return func(_ context.Context, in agents.HandoffInputData) (agents.HandoffInputData, error) {
		sharedBag.Incr("stats.handoff.news_normalization.count")

		// Clean up: remove tool spam from history (like in the example)
		// in = handoff_filters.RemoveAllTools(in)

		wire := latestNewsWire(sharedBag) // returns compact JSON string or ""
		if wire == "" {
			// No payload available—just pass through.
			return in, nil
		}

		// Inject a single user message that contains the payload.
		// payload := fmt.Sprintf("news_wire=%s", wire)
		// in.NewItems = append(in.NewItems, agents.HandoffCallItem{
		// 	Agent:   callerAgent,
		// 	RawItem: pkgagents.NewUserMessage(callerAgent, payload),
		// })

		// You can prepend to PreHandoffItems so the callee sees it immediately.
		return agents.HandoffInputData{
			InputHistory:    in.InputHistory,
			PreHandoffItems: slices.Clone(in.PreHandoffItems),
			NewItems:        slices.Clone(in.NewItems),
		}, nil
	}
}

func latestNewsWire(shared bag.SharedBag) string {
	v, ok := shared.Get(bag.Key("wiremin_tool_payloads"))
	if !ok {
		return ""
	}

	appendIf := func(kind string, at time.Time, data json.RawMessage) (string, time.Time) {
		if kind == "news" && len(data) > 0 && json.Valid(data) {
			return string(data), at
		}
		return "", time.Time{}
	}

	var best string
	var bestAt time.Time

	switch rows := v.(type) {
	case []map[string]any:
		for _, rec := range rows {
			k, _ := rec["kind"].(string)
			atStr, _ := rec["at"].(string)
			data, _ := rec["data"].(json.RawMessage)
			at, _ := time.Parse(time.RFC3339, atStr)
			if s, t := appendIf(k, at, data); t.After(bestAt) {
				best, bestAt = s, t
			}
		}
	case []any:
		for _, it := range rows {
			if rec, ok := it.(map[string]any); ok {
				k, _ := rec["kind"].(string)
				atStr, _ := rec["at"].(string)
				data, _ := rec["data"].(json.RawMessage)
				at, _ := time.Parse(time.RFC3339, atStr)
				if s, t := appendIf(k, at, data); t.After(bestAt) {
					best, bestAt = s, t
				}
			}
		}
	}
	return best
}
