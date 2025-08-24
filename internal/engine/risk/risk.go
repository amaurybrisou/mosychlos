// Package risk
package risk

import (
	"context"
	"log/slog"

	"github.com/amaurybrisou/mosychlos/internal/budget"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	pkgopenai "github.com/amaurybrisou/mosychlos/pkg/openai"
)

// RiskEngine is a simple example engine that prompts the model for a risk analysis
// and stores the result into the SharedBag. It forces a ToolConsumer budget
// so tool usage is controlled by constraints.
type RiskEngine struct {
	name          string
	constraints   models.BaseToolConstraints
	promptManager models.PromptBuilder // must provide BuildPrompt(ctx, kind) (string, error)
}

func New(name string, pm models.PromptBuilder, constraints models.BaseToolConstraints) *RiskEngine {
	if name == "" {
		name = "risk-engine"
	}
	return &RiskEngine{
		name:          name,
		promptManager: pm,
		constraints:   constraints,
	}
}

func (r *RiskEngine) Name() string { return r.name }

func (r *RiskEngine) ResultKey() bag.Key { return bag.KRiskAnalysisResult }

// Execute runs one synchronous Responses-API request and stores the content in KRiskAnalysisResult.
func (r *RiskEngine) Execute(ctx context.Context, client models.AiClient, sharedBag bag.SharedBag) error {
	// 1) Force budget constraints for tools used by this engine.
	client.SetToolConsumer(budget.NewToolConsumer(&r.constraints))

	// 2) Build the prompt content via your prompt manager.
	prompt, err := r.promptManager.BuildPrompt(ctx, models.AnalysisRisk)
	if err != nil {
		return err
	}

	// 3) Send to LLM via the new sync path (Responses API).
	req := models.PromptRequest{
		// leave Model empty to use the configured default
		Messages: []map[string]any{
			// optional system:
			{"role": "system", "content": `
			You are a Portfolio Risk Analysis Expert.
Your role is to evaluate investment portfolios with a focus on risk exposure, resilience, and alignment with the client’s objectives.
Always reason step by step, highlight assumptions, and provide transparent justifications.

Your output must:
- Identify key risk factors (market, credit, liquidity, concentration, geopolitical, regulatory).
- Quantify exposure where possible (ratios, percentages, stress test scenarios).
- Flag diversification issues and correlations.
- Provide both short-term risk alerts and long-term structural risks.
- Suggest concrete mitigation strategies (hedging, rebalancing, asset class adjustments).
- Keep explanations clear and actionable for decision-makers, avoiding jargon when possible.

Constraints:
- Do not invent data; only analyze based on the provided portfolio, market context, and tool outputs.
- If information is missing, clearly state assumptions or request clarification.
- Format responses in structured sections: **Summary**, **Detailed Risk Breakdown**, **Recommendations**.
			`},
			{"role": "user", "content": prompt},
		},
		// You can also set MaxTokens / Temperature from engine config if you want:
		MaxTokens: 2000,
		// Temperature: ptr.To(0.2),
		Tools: r.constraints.Tools, // use the tools defined in constraints,
		ResponseFormat: &models.ResponseFormat{
			Format: models.Format{
				Type:      bag.ResponseFormatJSON,
				Name:      bag.KRiskAnalysisResult.String(),
				Schema:    pkgopenai.BuildSchema[models.InvestmentResearchResult](),
				Verbosity: "	",
			},
		},
	}

	resp, err := client.DoSync(ctx, req)
	if err != nil {
		return err
	}

	// DEBUG: Log the response content to see what we're getting
	contentLen := len(resp.Content)
	preview := resp.Content
	if contentLen > 200 {
		preview = resp.Content[:200]
	}
	slog.Info("Risk analysis response received", "content_length", contentLen, "content_preview", preview)

	// 4) Store the result into the bag for downstream engines / reporters.
	sharedBag.Set(r.ResultKey(), resp.Content)
	// Optionally:
	sharedBag.Set(bag.KRiskMetrics, resp.Usage)

	return nil
}
