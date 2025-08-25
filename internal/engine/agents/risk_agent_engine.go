package agents

import (
	"context"
	"fmt"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/nlpodyssey/openai-agents-go/agents"
)

// RiskAnalysisResult represents the structured output from risk analysis
type RiskAnalysisResult struct {
	RiskScore          float64                `json:"risk_score" jsonschema_description:"Overall risk score from 0-100"`
	RiskLevel          string                 `json:"risk_level" jsonschema_description:"Risk level: Low, Medium, High, or Critical"`
	KeyRisks           []string               `json:"key_risks" jsonschema_description:"List of primary risks identified"`
	RiskMetrics        map[string]interface{} `json:"risk_metrics" jsonschema_description:"Detailed risk metrics and calculations"`
	Recommendations    []string               `json:"recommendations" jsonschema_description:"Risk mitigation recommendations"`
	ComplianceIssues   []string               `json:"compliance_issues" jsonschema_description:"Identified compliance concerns"`
	MarketExposure     map[string]float64     `json:"market_exposure" jsonschema_description:"Exposure breakdown by market/sector"`
	Methodology        string                 `json:"methodology" jsonschema_description:"Description of analysis methodology used"`
	ConfidenceLevel    float64                `json:"confidence_level" jsonschema_description:"Confidence in analysis from 0-100"`
	LastUpdated        string                 `json:"last_updated" jsonschema_description:"Timestamp of analysis"`
}

const RiskAnalysisPrompt = `You are a senior risk analyst specializing in portfolio risk assessment for institutional investors.
Your role is to provide comprehensive risk analysis that meets regulatory and fiduciary standards.

Key responsibilities:
1. Quantify portfolio risks using established methodologies (VaR, CVaR, beta analysis, etc.)
2. Identify concentration risks, liquidity risks, and market exposures
3. Assess compliance with risk management policies and regulatory requirements
4. Provide actionable risk mitigation strategies
5. Evaluate correlation risks and tail risks in the portfolio

Analysis framework:
- Use quantitative metrics where possible (volatility, Sharpe ratio, maximum drawdown, etc.)
- Consider macro-economic factors and market conditions
- Assess geographic and sector concentration risks
- Evaluate liquidity and operational risks
- Consider ESG and regulatory compliance risks

Output requirements:
- Provide specific, measurable risk assessments
- Include confidence intervals and methodology explanations
- Suggest concrete risk mitigation actions
- Prioritize risks by impact and likelihood
- Ensure compliance with institutional risk management standards

Use the available tools to gather market data, news sentiment, and economic indicators to inform your analysis.
Base your risk assessment on current market conditions and the specific portfolio composition provided.`

// AgentRiskEngine implements risk analysis using the agents SDK
type AgentRiskEngine struct {
	*BaseAgentEngine
	promptBuilder models.PromptBuilder
}

// NewAgentRiskEngine creates a new agent-based risk engine
func NewAgentRiskEngine(name string, promptBuilder models.PromptBuilder, tools []agents.Tool) *AgentRiskEngine {
	if name == "" {
		name = "agent-risk-engine"
	}

	config := BaseAgentEngineConfig{
		Name:         name,
		ResultKey:    bag.KRiskAnalysisResult,
		Instructions: RiskAnalysisPrompt,
		Model:        "gpt-4o",
		Tools:        tools,
		OutputType:   agents.OutputType[RiskAnalysisResult](),
		MaxTurns:     15,
	}

	baseEngine := NewBaseAgentEngine(config)

	return &AgentRiskEngine{
		BaseAgentEngine: baseEngine,
		promptBuilder:   promptBuilder,
	}
}

// buildInput creates a comprehensive risk analysis input from SharedBag data
func (e *AgentRiskEngine) buildInput(ctx context.Context, sharedBag bag.SharedBag) (string, error) {
	input := "# Portfolio Risk Analysis Request\n\n"

	// Get portfolio data
	if portfolioData, exists := sharedBag.Get(bag.KPortfolio); exists {
		input += fmt.Sprintf("## Portfolio Data\n%+v\n\n", portfolioData)
	}

	// Get investment profile if available
	if profile, exists := sharedBag.Get(bag.KProfile); exists {
		input += fmt.Sprintf("## Investment Profile\n%+v\n\n", profile)
	}

	// Get jurisdiction data if available
	if jurisdiction, exists := sharedBag.Get(bag.KJurisdiction); exists {
		input += fmt.Sprintf("## Regulatory Jurisdiction\n%+v\n\n", jurisdiction)
	}

	// Get configuration if available
	if config, exists := sharedBag.Get(bag.KAnalysisConfig); exists {
		input += fmt.Sprintf("## Analysis Configuration\n%+v\n\n", config)
	}

	// Add analysis instructions
	input += `## Analysis Requirements

Please provide a comprehensive risk analysis of this portfolio including:

1. **Quantitative Risk Assessment**
   - Calculate overall risk score (0-100 scale)
   - Determine risk level classification
   - Provide key risk metrics

2. **Risk Identification**
   - Identify primary risk factors
   - Assess concentration risks
   - Evaluate market exposures
   - Check for correlation risks

3. **Compliance Review**
   - Review against regulatory requirements
   - Identify compliance issues
   - Assess policy adherence

4. **Risk Mitigation**
   - Provide specific recommendations
   - Suggest hedging strategies
   - Recommend portfolio adjustments

5. **Confidence Assessment**
   - Rate confidence in analysis
   - Explain methodology used
   - Note any limitations

Use the available tools to gather current market data, economic indicators, and news sentiment to inform your analysis. 
Ensure your analysis meets institutional risk management standards and provides actionable insights.`

	return input, nil
}

// storeResult stores the structured risk analysis result
func (e *AgentRiskEngine) storeResult(result *agents.RunResult, sharedBag bag.SharedBag) error {
	// Store the structured result
	riskResult := result.FinalOutput.(RiskAnalysisResult)
	
	// Store in the standard format expected by the orchestrator
	sharedBag.Set(e.resultKey, map[string]any{
		"result":            fmt.Sprintf("Risk Level: %s (Score: %.1f/100)", riskResult.RiskLevel, riskResult.RiskScore),
		"risk_score":        riskResult.RiskScore,
		"risk_level":        riskResult.RiskLevel,
		"key_risks":         riskResult.KeyRisks,
		"risk_metrics":      riskResult.RiskMetrics,
		"recommendations":   riskResult.Recommendations,
		"compliance_issues": riskResult.ComplianceIssues,
		"market_exposure":   riskResult.MarketExposure,
		"methodology":       riskResult.Methodology,
		"confidence_level":  riskResult.ConfidenceLevel,
		"structured_output": riskResult,
	})

	// Store metadata
	metadata := map[string]any{
		"agent_name":       e.name,
		"analysis_type":    "risk_assessment",
		"output_type":      "structured",
		"new_items":        len(result.NewItems),
		"responses":        len(result.RawResponses),
		"confidence_level": riskResult.ConfidenceLevel,
		"risk_score":       riskResult.RiskScore,
	}
	
	metadataKey := bag.Key(string(e.resultKey) + "_metadata")
	sharedBag.Set(metadataKey, metadata)

	return nil
}