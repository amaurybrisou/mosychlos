package agents

// import (
// 	"context"
// 	"fmt"

// 	"github.com/amaurybrisou/mosychlos/pkg/bag"
// 	"github.com/amaurybrisou/mosychlos/pkg/models"
// 	"github.com/nlpodyssey/openai-agents-go/agents"
// )

// // TriageResult represents the structured output from triage analysis
// type TriageResult struct {
// 	AnalysisType    string   `json:"analysis_type" jsonschema:"description=Type of analysis recommended: risk, allocation, compliance, research"`
// 	Priority        string   `json:"priority" jsonschema:"description=Priority level: High, Medium, Low"`
// 	Reasoning       string   `json:"reasoning" jsonschema:"description=Explanation for the analysis type selection"`
// 	RequiredTools   []string `json:"required_tools" jsonschema:"description=List of tools needed for the analysis"`
// 	EstimatedTime   string   `json:"estimated_time" jsonschema:"description=Estimated analysis completion time"`
// 	Complexity      string   `json:"complexity" jsonschema:"description=Analysis complexity: Simple, Medium, Complex"`
// 	Recommendations []string `json:"recommendations" jsonschema:"description=Initial recommendations before detailed analysis"`
// 	NextSteps       []string `json:"next_steps" jsonschema:"description=Specific next steps for the chosen analysis path"`
// }

// const TriageAnalysisPrompt = `You are a senior portfolio management triage specialist responsible for routing analysis requests to the most appropriate specialist teams.

// Your role is to:
// 1. Analyze the portfolio composition and investment context
// 2. Identify the primary analytical needs (risk, allocation, compliance, research)
// 3. Route the request to the appropriate specialist agent
// 4. Provide initial guidance and recommendations

// Analysis routing guidelines:
// - **Risk Analysis**: For volatility assessment, downside protection, correlation analysis, regulatory compliance
// - **Allocation Analysis**: For asset allocation optimization, rebalancing recommendations, diversification strategy
// - **Compliance Analysis**: For regulatory adherence, jurisdiction-specific requirements, ESG compliance
// - **Investment Research**: For new investment opportunities, market analysis, sector research

// Consider:
// - Portfolio size and complexity
// - Investment objectives and constraints
// - Regulatory environment and jurisdiction
// - Current market conditions
// - Available analytical tools and data sources

// Provide structured output with clear reasoning for your routing decision and initial analytical insights.`

// // AgentTriageEngine implements portfolio analysis triage using agent handoffs
// type AgentTriageEngine struct {
// 	*BaseAgentEngine
// 	riskAgent       *agents.Agent
// 	allocationAgent *agents.Agent
// 	complianceAgent *agents.Agent
// 	researchAgent   *agents.Agent
// 	promptBuilder   models.PromptBuilder
// }

// // NewAgentTriageEngine creates a new agent-based triage engine with handoff capabilities
// func NewAgentTriageEngine(name string, promptBuilder models.PromptBuilder, tools []agents.Tool) *AgentTriageEngine {
// 	if name == "" {
// 		name = "agent-triage-engine"
// 	}

// 	// Create specialist agents for handoffs
// 	riskAgent := agents.New("RiskSpecialist").
// 		WithInstructions("You are a risk management specialist. Focus on portfolio risk analysis, volatility assessment, and risk mitigation strategies.").
// 		WithModel("gpt-5-nano")

// 	allocationAgent := agents.New("AllocationSpecialist").
// 		WithInstructions("You are an asset allocation specialist. Focus on portfolio optimization, rebalancing strategies, and diversification analysis.").
// 		WithModel("gpt-5-nano")

// 	complianceAgent := agents.New("ComplianceSpecialist").
// 		WithInstructions("You are a regulatory compliance specialist. Focus on regulatory adherence, jurisdiction requirements, and ESG compliance.").
// 		WithModel("gpt-5-nano")

// 	researchAgent := agents.New("ResearchSpecialist").
// 		WithInstructions("You are an investment research specialist. Focus on market analysis, investment opportunities, and sector research.").
// 		WithModel("gpt-5-nano")

// 	// Add tools to specialist agents
// 	if len(tools) > 0 {
// 		riskAgent = riskAgent.WithTools(tools...)
// 		allocationAgent = allocationAgent.WithTools(tools...)
// 		complianceAgent = complianceAgent.WithTools(tools...)
// 		researchAgent = researchAgent.WithTools(tools...)
// 	}

// 	config := BaseAgentEngineConfig{
// 		Name:      name,
// 		ResultKey: bag.Key("triage_analysis_result"),
// 		Instructions: func(ctx context.Context, _ *agents.Agent) (string, error) {
// 			return GetTriagePrompt(ctx, promptBuilder)
// 		},
// 		Model:      "gpt-4o",
// 		Tools:      tools,
// 		OutputType: nil, // Disable structured output for now to avoid schema issues
// 		MaxTurns:   20,
// 	}

// 	baseEngine := NewBaseAgentEngine(config)

// 	// Add handoff capabilities to the base agent
// 	baseEngine.agent = baseEngine.agent.WithAgentHandoffs(
// 		riskAgent,
// 		allocationAgent,
// 		complianceAgent,
// 		researchAgent,
// 	)

// 	return &AgentTriageEngine{
// 		riskAgent:       riskAgent,
// 		allocationAgent: allocationAgent,
// 		complianceAgent: complianceAgent,
// 		researchAgent:   researchAgent,
// 		promptBuilder:   promptBuilder,
// 	}
// }

// // GetTriagePrompt creates a comprehensive triage analysis input
// func GetTriagePrompt(ctx context.Context, promptBuilder models.PromptBuilder) (string, error) {
// 	//TODO: use the promptBuilder
// 	return "", nil
// }

// // storeResult stores the triage analysis result and any handoff results
// func (e *AgentTriageEngine) storeResult(result *agents.RunResult, sharedBag bag.SharedBag) error {
// 	// Since we're not using structured output, store the string result
// 	resultStr, ok := result.FinalOutput.(string)
// 	if !ok {
// 		resultStr = fmt.Sprintf("%+v", result.FinalOutput)
// 	}

// 	// Store in the expected format
// 	sharedBag.Set(e.resultKey, map[string]any{
// 		"analysis_recommendation": resultStr,
// 		"raw_output":              result.FinalOutput,
// 		"analysis_type":           "triage_routing",
// 	})

// 	// Check if there were any handoffs and store their results too
// 	if len(result.NewItems) > 0 {
// 		// Store information about handoffs that occurred
// 		handoffInfo := map[string]any{
// 			"handoff_occurred": true,
// 			"final_agent":      result.LastAgent.Name,
// 			"total_items":      len(result.NewItems),
// 		}

// 		handoffKey := bag.Key(string(e.resultKey) + "_handoffs")
// 		sharedBag.Set(handoffKey, handoffInfo)

// 		// If the final agent was a specialist, store their specialized result
// 		if result.LastAgent.Name != e.name {
// 			specialistResultKey := bag.Key(fmt.Sprintf("%s_specialist_result", result.LastAgent.Name))
// 			sharedBag.Set(specialistResultKey, result.FinalOutput)
// 		}
// 	}

// 	// Store metadata
// 	metadata := map[string]any{
// 		"agent_name":      e.name,
// 		"analysis_type":   "triage_routing",
// 		"output_type":     "text_with_handoffs",
// 		"new_items":       len(result.NewItems),
// 		"responses":       len(result.RawResponses),
// 		"final_agent":     result.LastAgent.Name,
// 		"handoff_enabled": true,
// 	}

// 	metadataKey := bag.Key(string(e.resultKey) + "_metadata")
// 	sharedBag.Set(metadataKey, metadata)

// 	return nil
// }
