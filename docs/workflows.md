Here’s the full roadmap of **Mosychlos Engine & Tool Execution Flows** in **Markdown** with embedded Mermaid diagrams:

# Engine and Tool Execution Flows in Mosychlos

Mosychlos uses a **chained engine orchestration** approach for multi-layer analysis workflow.
Each engine corresponds to a phase (Data, Function, Multi-Agent, Report) and runs sequentially in a single AI session, sharing one context state (the `SharedBag`).

---

## Figure 1: Sequential Engine Chain

```mermaid
flowchart LR
    Engine1[Data Gathering Engine]:::data -- "Financial data & context" --> Engine2[Financial Analysis Engine]:::function
    Engine2 -- "Analytical insights" --> Engine3[Investment Committee Engine]:::multiagent
    Engine3 -- "Final decision" --> Engine4[Report Generation Engine]:::report
    %% All engines share a common SharedBag context throughout

classDef data fill:#4DA6FF,stroke:#000,color:#fff
classDef function fill:#4DFF91,stroke:#000,color:#000
classDef multiagent fill:#FFB84D,stroke:#000,color:#000
classDef report fill:#B266FF,stroke:#000,color:#fff
```

---

## Figure 2: Batch Engine Iterative Flow

```mermaid
flowchart TD
    A[Execute Called] --> B[GetInitialPrompt Hook]
    B --> C[Initialize First BatchJob]
    C --> D{Iteration Loop}
    D --> E[PreIteration Hook]
    E --> F[Submit Batch to AI]
    F --> G[Wait for Completion]
    G --> H[Process Results]
    H --> I{Tool Calls?}
    I -- Yes --> J[ProcessToolResult Hook]:::tool
    I -- No --> K[ProcessFinalResult Hook]
    J --> L[Generate Next Jobs]
    K --> L
    L --> M[PostIteration Hook]
    M --> N{Continue?}
    N -- Yes --> D
    N -- No --> O[Store Final Results]
    O --> P[Complete]

classDef tool fill:#CCCCCC,stroke:#000,color:#000
```

---

## Figure 3: Parallel Multi-Persona Flow

```mermaid
flowchart LR
    CE[Investment Committee Engine]:::multiagent -->|spawn| FA(Financial Analyst Persona)
    CE -->|spawn| QA(Quantitative Analyst Persona)
    CE -->|spawn| MA(Market Analyst Persona)
    CE -->|spawn| PM(Portfolio Manager Persona)
    FA -->|insight| SYNT[Synthesis & Recommendation]
    QA -->|insight| SYNT
    MA -->|insight| SYNT
    PM -->|insight| SYNT
    SYNT -->|final output| OUT[Committee Decision]:::report

classDef multiagent fill:#FFB84D,stroke:#000,color:#000
classDef report fill:#B266FF,stroke:#000,color:#fff
```

---

## Figure 4: Tool Calling within an Engine

```mermaid
flowchart TD
    EngineStart([Engine Execution Start]) --> LLM[AI Client Session]:::function
    LLM -->|Call| Tool1[Tool A]:::tool
    LLM -->|Call| Tool2[Tool B]:::tool
    Tool1 -->|Result| LLM
    Tool2 -->|Result| LLM
    LLM --> Answer[AI Final Answer]
    Answer --> EngineEnd([Engine Stores Result])

classDef function fill:#4DFF91,stroke:#000,color:#000
classDef tool fill:#CCCCCC,stroke:#000,color:#000
```

---

## Figure 5: Complete Engine & Tool Execution Flow (Mosychlos Pipeline)

```mermaid
flowchart TD
    subgraph Mosychlos Full Pipeline
      direction LR
      Start([User Query]) --> DE[Data Engine]:::data
      DE -- consolidates data --> AE[Analysis Engine]:::function
      AE -- insights --> CE[Committee Engine]:::multiagent
      CE -- recommendation --> RE[Report Engine]:::report
      RE --> Output([Professional Report]):::report

      %% Tools used by Data Engine (examples)
      DE -.-> T1[SEC Filings API]:::tool
      DE -.-> T2[Market Data API]:::tool
      DE -.-> T3[News Sentiment API]:::tool

      %% Tools used by Analysis Engine (examples)
      AE -.-> T4[Financial Ratios Tool]:::tool
      AE -.-> T5[Valuation Models]:::tool

      %% Parallel personas in Committee Engine
      CE -.-> P1(Financial Analyst)
      CE -.-> P2(Quant Analyst)
      CE -.-> P3(Market Analyst)
      CE -.-> P4(Portfolio Manager)
    end

    subgraph Legend[Legend]
      direction TB
      LData[Data Layer]:::data
      LFunc[Function Layer]:::function
      LMulti[Multi-Agent Layer]:::multiagent
      LRep[Report Layer]:::report
      LTool[Tool/External API]:::tool
    end

classDef data fill:#4DA6FF,stroke:#000,color:#fff
classDef function fill:#4DFF91,stroke:#000,color:#000
classDef multiagent fill:#FFB84D,stroke:#000,color:#000
classDef report fill:#B266FF,stroke:#000,color:#fff
classDef tool fill:#CCCCCC,stroke:#000,color:#000
```

---

```

```
