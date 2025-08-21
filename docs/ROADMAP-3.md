# Mosychlos Execution Flow Diagrams

## Analyze Pipeline: Data → Function → Committee → Report

```mermaid
flowchart TD
classDef engine fill:#cce5ff,stroke:#007ACC,stroke-width:1px,color:#000;
classDef tool fill:#d5e8d4,stroke:#82b366,stroke-width:1px,color:#000;
classDef openai fill:#ffe6cc,stroke:#E67E22,stroke-width:1px,color:#000;
classDef shared fill:#e1d5e7,stroke:#9673A6,stroke-width:1px,color:#000;

Start(["User initiates Analysis Pipeline"])
Start --> NE["Engine: NewsEngine (data gathering)"]:::engine
NE --> NAT["Tool: NewsAPI (fetch news)"]:::tool
NAT --> NE
NE --> SB1["SharedBag: News data stored"]:::shared
SB1 --> ME["Engine: MarketDataEngine (data gathering)"]:::engine
ME --> MFT["Tool: MarketData (fetch market metrics)"]:::tool
MFT --> ME
ME --> SB2["SharedBag: Market data stored"]:::shared
SB2 --> RE1["Engine: RiskEngine (risk analysis)"]:::engine
RE1 --> SB3["SharedBag: Risk analysis results"]:::shared
SB3 --> AE["Engine: AllocationEngine (allocation analysis)"]:::engine
AE --> SB4["SharedBag: Allocation analysis results"]:::shared
SB4 --> CE["Engine: ComplianceEngine (compliance check)"]:::engine
CE --> SB5["SharedBag: Compliance findings"]:::shared
SB5 --> CoE["Engine: ReallocationEngine (Committee synthesis)"]:::engine
CoE --> OA[("OpenAI API call (Chat Completions with Tools)")]:::openai
OA --> YFT["Tool: YFinance (data lookup)"]:::tool
YFT --> OA
OA --> NWST["Tool: WebSearch (market news)"]:::tool
NWST --> OA
OA --> RPT["SharedBag: Final rebalancing report"]:::shared
RPT --> End(["Output: Investment Report Generated"])
```

## Optimize Pipeline: Macro + Tax → Reallocation → Execution Decision

```mermaid
flowchart TD
classDef engine fill:#cce5ff,stroke:#007ACC,stroke-width:1px,color:#000;
classDef tool fill:#d5e8d4,stroke:#82b366,stroke-width:1px,color:#000;
classDef openai fill:#ffe6cc,stroke:#E67E22,stroke-width:1px,color:#000;
classDef shared fill:#e1d5e7,stroke:#9673A6,stroke-width:1px,color:#000;

StartOpt(["User initiates Optimization Pipeline"])
StartOpt --> MacroE["Engine: MacroAnalysis (macro environment)"]:::engine
MacroE --> FREDT["Tool: FRED (economic indicators)"]:::tool
FREDT --> MacroE
MacroE --> SBm["SharedBag: Macro outlook stored"]:::shared
SBm --> TaxE["Engine: TaxAnalysis (tax implications)"]:::engine
TaxE --> SBt["SharedBag: Tax impact analysis"]:::shared
SBt --> ReOpt["Engine: ReallocationEngine (optimized rebalancing)"]:::engine
ReOpt --> OA2[("OpenAI API call (Chat with Tools for strategy)")]:::openai
OA2 --> YFT2["Tool: YFinance (market data)"]:::tool
YFT2 --> OA2
OA2 --> TaxT["Tool: TaxEstimator (transaction cost)"]:::tool
TaxT --> OA2
OA2 --> SBo["SharedBag: Reallocation plan"]:::shared
SBo --> ExecE["Engine: ExecutionDecision (go/no-go)"]:::engine
ExecE --> OA3[("OpenAI API call (Chat for execution recommendation)")]:::openai
OA3 --> DecOut(["Output: Execute now or wait decision"])
```

## Tool Chaining with NEXT_PATTERN (SEC → YFinance → FRED)

```mermaid
flowchart TD
classDef engine fill:#cce5ff,stroke:#007ACC,stroke-width:1px,color:#000;
classDef tool fill:#d5e8d4,stroke:#82b366,stroke-width:1px,color:#000;
classDef openai fill:#ffe6cc,stroke:#E67E22,stroke-width:1px,color:#000;
classDef shared fill:#e1d5e7,stroke:#9673A6,stroke-width:1px,color:#000;
QStart(["User query (requires multi-step data)"])
QStart --> Agent1["AI Agent: decides to retrieve SEC filings"]:::openai
Agent1 --> SECT["Tool: SEC Filings (fetch company report)"]:::tool
SECT --> Agent2["AI Agent: analyzes SEC data, then requests stock info"]:::openai
Agent2 --> YFT_chain["Tool: YFinance (fetch stock data)"]:::tool
YFT_chain --> Agent3["AI Agent: analyzes stock data, then requests macro data"]:::openai
Agent3 --> FREDT_chain["Tool: FRED (fetch macro indicator)"]:::tool
FREDT_chain --> Agent4["AI Agent: integrates all data into final answer"]:::openai
Agent4 --> AnswerOut(["Final Answer (analysis using SEC, market, macro data)"])
```

## Batch Tool Iteration (Multi-item Processing)

```mermaid
flowchart TD
classDef engine fill:#cce5ff,stroke:#007ACC,stroke-width:1px;
classDef tool fill:#d5e8d4,stroke:#82b366,stroke-width:1px;
classDef openai fill:#ffe6cc,stroke:#E67E22,stroke-width:1px;
classDef shared fill:#e1d5e7,stroke:#9673A6,stroke-width:1px;
BEStart(["Batch Engine started"])
BEStart --> PromptGen["Hook: GetInitialPrompt (batch job template)"]:::engine
PromptGen --> Jobs["Batch jobs created (initial)"]:::engine
Jobs --> SubmitBatch[("OpenAI Batch API call - parallel requests")]:::openai
SubmitBatch --> WaitResults["Wait for batch completion"]:::engine
WaitResults --> ProcRes["Process batch results"]:::engine
ProcRes --> ToolCheck{"Any tool calls in results?"}
ToolCheck --> |"Yes"| ToolProc["Hook: ProcessToolResult (execute tools, update job data)"]:::engine
ToolProc --> PostIter["Hook: PostIteration (custom analysis)"]:::engine
PostIter --> ContinueCheck{"More jobs for next iteration?"}
ContinueCheck --> |"Yes"| NextJobs["Prepare next iteration jobs"]:::engine
NextJobs --> SubmitBatch
ToolCheck --> |"No"| NoTool["No tool usage"]:::engine
NoTool --> PostIter
ContinueCheck --> |"No"| Finalize["Hook: ProcessFinalResult (store outputs in SharedBag)"]:::engine
Finalize --> BatchEnd(["Batch processing complete"])
```

## Tool Failure with Fallback Handling

```mermaid
flowchart TD
classDef engine fill:#cce5ff,stroke:#007ACC,stroke-width:1px,color:#000;
classDef tool fill:#d5e8d4,stroke:#82b366,stroke-width:1px,color:#000;
classDef openai fill:#ffe6cc,stroke:#E67E22,stroke-width:1px,color:#000;
classDef shared fill:#e1d5e7,stroke:#9673A6,stroke-width:1px,color:#000;

StartFail(["AI triggers a tool call"])
StartFail --> ToolX["Tool X (primary data source)"]:::tool
ToolX --> SuccessX{"Tool X succeeded?"}
SuccessX --> |"Yes"| ResultOK["Tool result returned to AI"]:::openai
SuccessX --> |"No"| RetryCheck{"Retry attempt remaining?"}
RetryCheck --> |"Yes"| RetryCall["Retry Tool X call"]:::tool
RetryCall --> ToolX
RetryCheck --> |"No"| FallbackCheck{"Fallback tool available?"}
FallbackCheck --> |"Yes"| ToolY["Tool Y (fallback data source)"]:::tool
ToolY --> SuccessY{"Tool Y succeeded?"}
SuccessY --> |"Yes"| ResultOK
SuccessY --> |"No"| FailOut["Failure returned (no data)"]:::openai
FallbackCheck --> |"No"| FailOut
```

## Persona-Based Multi-Agent Reasoning

```mermaid
flowchart TD
classDef engine fill:#cce5ff,stroke:#007ACC,stroke-width:1px,color:#000;
classDef tool fill:#d5e8d4,stroke:#82b366,stroke-width:1px,color:#000;
classDef openai fill:#ffe6cc,stroke:#E67E22,stroke-width:1px,color:#000;
classDef shared fill:#e1d5e7,stroke:#9673A6,stroke-width:1px,color:#000;
PStart(["User question received"])
PStart --> Pers1["Engine: Persona A (Financial Analyst)"]:::engine
Pers1 --> OA_P1[("OpenAI API call (Chat, Persona A analysis)")]:::openai
OA_P1 --> SBp1["SharedBag: Persona A findings"]:::shared
SBp1 --> Pers2["Engine: Persona B (Quant Analyst)"]:::engine
Pers2 --> OA_P2[("OpenAI API call (Chat, Persona B analysis)")]:::openai
OA_P2 --> SBp2["SharedBag: Persona B findings"]:::shared
SBp2 --> Pers3["Engine: Persona C (Risk Analyst)"]:::engine
Pers3 --> OA_P3[("OpenAI API call (Chat, Persona C analysis)")]:::openai
OA_P3 --> SBp3["SharedBag: Persona C findings"]:::shared
SBp3 --> Chair["Engine: CommitteeChair (synthesize all perspectives)"]:::engine
Chair --> OA_P4[("OpenAI API call (Responses API for final consensus)")]:::openai
OA_P4 --> FinalOut(["Output: Committee consensus report"])
```

## SharedBag Context Propagation Through Engines and Tools

```mermaid
flowchart TD
classDef engine fill:#cce5ff,stroke:#007ACC,stroke-width:1px,color:#000;
classDef tool fill:#d5e8d4,stroke:#82b366,stroke-width:1px,color:#000;
classDef openai fill:#ffe6cc,stroke:#E67E22,stroke-width:1px,color:#000;
classDef shared fill:#e1d5e7,stroke:#9673A6,stroke-width:1px,color:#000;
BagStart(["User provides input data"])
BagStart --> PortSvc["PortfolioService: load portfolio data"]:::engine
PortSvc --> SB_port["SharedBag: Portfolio data stored"]:::shared
SB_port --> ProfMgr["ProfileManager: load user profile"]:::engine
ProfMgr --> SB_prof["SharedBag: User profile stored"]:::shared
SB_prof --> NewsEng["Engine: NewsEngine (uses portfolio holdings)"]:::engine
NewsEng --> NewsTool["Tool: NewsAPI (fetch relevant news)"]:::tool
NewsTool --> NewsEng
NewsEng --> SB_news["SharedBag: News insights stored"]:::shared
SB_news --> RiskEng["Engine: RiskEngine (uses portfolio & news data)"]:::engine
RiskEng --> SB_risk["SharedBag: Risk metrics stored"]:::shared
SB_risk --> RebalEng["Engine: ReallocationEngine (uses profile, risk, news)"]:::engine
RebalEng --> OA_shared[("OpenAI API call (portfolio rebalancing)")]:::openai
OA_shared --> SB_rebal["SharedBag: Rebalancing plan stored"]:::shared
SB_rebal --> ReportGen["ReportGenerator: compile final report"]:::engine
ReportGen --> FinalPDF(["Output: Final portfolio analysis PDF"])
```
