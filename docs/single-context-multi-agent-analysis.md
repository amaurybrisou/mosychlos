# Single-Context Multi-Agent Implementation for Mosychlos

## Key Insight: GPT-5 Single Context Advantage

**Traditional Multi-Agent Approach**: Separate conversations, context switching, message passing
**GPT-5 Single Context Approach**: All layers (L1, L2, L3) in one continuous context with role-switching

## Architecture Implications

### **Current FinRobot Layer Separation**

```
Layer 1 (Data) → Layer 2 (Functions) → Layer 3 (Multi-Agent) → Layer 4 (Reports)
   ↓              ↓                      ↓                      ↓
Separate calls   Separate calls        Separate agent calls   Report generation
```

### **Proposed Single-Context Flow**

```
Single GPT-5 Context:
├── Layer 1: Data gathering instructions
├── Layer 2: Function execution within context
├── Layer 3: Role-based perspective analysis
└── Layer 4: Report synthesis - all in memory
```

## Implementation Strategy for Mosychlos

### **1. Enhanced Engine Architecture**

```go
// internal/engine/unified_context_engine.go
type UnifiedContextEngine struct {
    name           string
    personas       []PersonaConfig
    tools          []models.Tool
    contextBuilder *ContextBuilder
    constraints    models.ToolConstraints
}

type ContextBuilder struct {
    systemPrompt    string
    dataLayer       DataLayerConfig
    functionsLayer  FunctionsLayerConfig
    agentLayer      AgentLayerConfig
    reportLayer     ReportLayerConfig
}

type AgentLayerConfig struct {
    Personas        []PersonaConfig `yaml:"personas"`
    WorkflowType    string         `yaml:"workflow_type"` // "committee", "sequential", "debate"
    MaxPerspectives int            `yaml:"max_perspectives"`
    SynthesisMode   string         `yaml:"synthesis_mode"` // "consensus", "majority", "chairperson"
}

func (e *UnifiedContextEngine) Execute(ctx context.Context, client models.AiClient, sharedBag bag.SharedBag) error {
    // Build comprehensive single prompt with all layers
    contextPrompt, err := e.contextBuilder.BuildUnifiedPrompt(sharedBag)
    if err != nil {
        return fmt.Errorf("failed to build unified prompt: %w", err)
    }

    // Single AI call with complete context
    session := client.NewSession(contextPrompt)

    // Add portfolio data and user request
    session.Add(models.RoleUser, e.buildUserRequest(sharedBag))

    // Execute with all tools available in single context
    response, err := session.Next(ctx, e.getToolDefinitions(), nil)
    if err != nil {
        return fmt.Errorf("unified analysis failed: %w", err)
    }

    // Parse and store results in shared bag
    return e.parseAndStoreResults(response, sharedBag)
}
```

### **2. Multi-Persona Single Context Prompts**

```go
const UnifiedFinancialAnalysisPrompt = `
You are an elite Investment Committee conducting comprehensive portfolio analysis. You have access to institutional-grade data sources and analysis tools.

## COMMITTEE COMPOSITION & ROLES

You will analyze this portfolio from multiple expert perspectives in a single comprehensive analysis:

### 1. SENIOR FINANCIAL ANALYST
- **Expertise**: 15+ years fundamental analysis, valuation modeling, financial statements
- **Focus**: Income statement, balance sheet, cash flow analysis, competitive positioning
- **Tools**: SEC filings, financial statements, peer comparison data

### 2. QUANTITATIVE ANALYST
- **Expertise**: Statistical modeling, risk metrics, backtesting, portfolio optimization
- **Focus**: Risk-adjusted returns, VaR analysis, correlation analysis, factor exposure
- **Tools**: Historical price data, statistical analysis, Monte Carlo simulation

### 3. MARKET ANALYST
- **Expertise**: Market sentiment, technical analysis, sector trends, timing
- **Focus**: Current market conditions, sentiment indicators, technical levels
- **Tools**: Market data, sentiment analysis, technical indicators

### 4. PORTFOLIO MANAGER
- **Expertise**: Strategic allocation, risk budgeting, client objectives
- **Focus**: Portfolio construction, position sizing, strategic recommendations
- **Tools**: Portfolio analytics, risk management, performance attribution

### 5. RISK ANALYST
- **Expertise**: Enterprise risk, regulatory compliance, scenario analysis
- **Focus**: Comprehensive risk assessment, stress testing, mitigation strategies
- **Tools**: Risk modeling, scenario analysis, regulatory analysis

## ANALYSIS WORKFLOW

Execute this analysis in phases, maintaining context throughout:

### PHASE 1: DATA LAYER (Use available tools to gather data)
1. **Financial Data**: Get latest financial statements, SEC filings, analyst estimates
2. **Market Data**: Obtain current prices, historical data, volatility metrics
3. **Sentiment Data**: Gather news sentiment, analyst ratings, institutional flows
4. **Risk Data**: Collect correlation data, sector exposures, macro indicators

### PHASE 2: FUNCTIONAL ANALYSIS (Apply analytical tools)
1. **Financial Analysis**: Calculate ratios, growth rates, profitability metrics
2. **Quantitative Analysis**: Perform statistical analysis, risk calculations
3. **Technical Analysis**: Identify chart patterns, support/resistance levels
4. **Risk Analysis**: Model scenarios, calculate VaR, assess correlations

### PHASE 3: MULTI-PERSPECTIVE ANALYSIS (Role-based insights)
For each major finding, provide perspectives from ALL committee members:

**Financial Analyst Perspective**: [Fundamental view with specific metrics]
**Quantitative Analyst Perspective**: [Statistical significance and risk metrics]
**Market Analyst Perspective**: [Market timing and sentiment implications]
**Portfolio Manager Perspective**: [Strategic allocation and implementation]
**Risk Analyst Perspective**: [Risk assessment and mitigation recommendations]

### PHASE 4: COMMITTEE SYNTHESIS
1. **Consensus Areas**: Where all perspectives align
2. **Debate Points**: Where perspectives differ and why
3. **Weighted Assessment**: Final recommendation considering all viewpoints
4. **Implementation Plan**: Specific actions and timeline
5. **Monitoring Framework**: Key metrics and triggers for review

## OUTPUT FORMAT

Structure your response as a comprehensive investment committee report:

1. **EXECUTIVE SUMMARY** (Committee Consensus)
2. **DATA ANALYSIS SUMMARY** (Layer 1 findings)
3. **FUNCTIONAL ANALYSIS RESULTS** (Layer 2 findings)
4. **MULTI-PERSPECTIVE ANALYSIS** (Layer 3 insights)
5. **COMMITTEE RECOMMENDATION** (Synthesized decision)
6. **RISK ASSESSMENT & MITIGATION**
7. **IMPLEMENTATION PLAN**
8. **MONITORING & REVIEW FRAMEWORK**

Maintain the context of being a single committee with multiple expert perspectives rather than separate agents having a conversation.
`
```

### **3. Context State Management**

```go
// internal/engine/context_state.go
type ContextState struct {
    DataGathered    map[string]interface{} `json:"data_gathered"`
    FunctionsUsed   []string              `json:"functions_used"`
    Perspectives    map[string]string     `json:"perspectives"`
    Synthesis       string                `json:"synthesis"`
    Confidence      float64               `json:"confidence"`
}

func (e *UnifiedContextEngine) maintainContextState(session models.Session) {
    // Track what data has been gathered
    // Track which functions have been used
    // Track perspective analysis completeness
    // Ensure all required viewpoints are covered
}
```

### **4. Benefits of Single Context Approach**

#### **Technical Advantages:**

1. **Context Coherence**: No context loss between "agents"
2. **Efficiency**: Single API call instead of multiple agent conversations
3. **Cost Optimization**: Reduced token usage vs. multiple separate calls
4. **Simplicity**: No complex message passing or state synchronization

#### **Analytical Advantages:**

1. **Holistic Analysis**: All perspectives informed by same complete data set
2. **Real-time Synthesis**: Can cross-reference insights immediately
3. **Reduced Bias**: No sequential bias from agent ordering
4. **Natural Debate**: Can present conflicting views and resolve in context

#### **Implementation Advantages:**

1. **Leverages Existing Architecture**: Uses current engine system
2. **Tool Integration**: All tools available in single context
3. **SharedBag Compatibility**: Results stored once, no merging needed
4. **Debugging Simplicity**: Single conversation to trace vs. multiple agent logs

## Modified Implementation Timeline

### **Phase 1: Single Context Foundation (Weeks 1-2)**

1. 🔥 **UnifiedContextEngine** - Single context with multi-perspective prompts
2. 🔥 **Enhanced Prompt System** - Multi-role instructions in single prompt
3. 🔥 **Context State Management** - Track analysis completeness
4. 🟡 **Tool Integration** - All financial tools available in single context

### **Phase 2: Advanced Context Patterns (Weeks 3-4)**

1. 🔥 **Committee Synthesis Logic** - Weighted perspective integration
2. 🔥 **Dynamic Role Emphasis** - Adjust role focus based on analysis type
3. 🟡 **Context Memory Management** - Optimize for long analytical sessions
4. 🟡 **Quality Assurance** - Ensure all required perspectives covered

### **Phase 3: Professional Integration (Weeks 5-6)**

1. 🟡 **Report Generation** - Professional outputs from single context
2. 🟡 **Chart Integration** - Visual outputs within analysis flow
3. 🔶 **Context Templating** - Reusable patterns for different analysis types
4. 🔶 **Performance Optimization** - Token usage and response time optimization

## Key Implementation Decisions

### **1. Role Switching vs. Agent Simulation**

- **Approach**: Role-based perspective analysis within single context
- **Benefit**: More natural analysis flow, no artificial agent boundaries
- **Implementation**: Structured prompts with clear role sections

### **2. Tool Usage Pattern**

- **Approach**: All tools available throughout entire analysis
- **Benefit**: Can gather additional data based on emerging insights
- **Implementation**: Progressive tool usage with context building

### **3. Synthesis Method**

- **Approach**: Integrated analysis with explicit consensus/disagreement tracking
- **Benefit**: More nuanced final recommendations
- **Implementation**: Weighted perspective integration with confidence scoring

This approach leverages GPT-5's context capabilities while maintaining the institutional-grade analysis quality that FinRobot provides, but in a more efficient and coherent single-context implementation.
