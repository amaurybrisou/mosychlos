# LLM Batch Processing Architecture (Revised)

**Status:** Proposed – ready to replace the previous draft
**Scope:** Introduces batch capabilities via a new `internal/llm` package while preserving current synchronous paths in `internal/ai`.
**Audience:** Backend engineers, infra/ops, prompt authors, CLI consumers.

---

## 1) Executive Summary

This document presents a clean, incremental architecture to add **batch LLM processing** while keeping all existing synchronous flows intact. It fits the current repository layout, reuses the **prompt system**, and adds **observability**, **cost controls**, and **model-class–aware parameters** (e.g., different behavior for GPT‑5 / o1 vs standard models).

**Key outcomes**

- New `internal/llm` package with batch job lifecycle (create, monitor, aggregate results).
- Reuse of `internal/prompt` (no template duplication) and engines via optional batch entrypoints.
- Cost estimation + guardrails (pricing tables, `--dry-run`, estimate vs actual).
- Metrics + logs for SRE observability (submitted/failed jobs, durations, token usage).
- CLI UX for batch: `--batch`, `status`, `results`, `logs`.

---

## 2) Repository Context (today)

- **Synchronous LLM**: `internal/ai` (providers, sessions, streaming).
- **Prompts**: `internal/prompt` (manager + regional overlays + templates).
- **Engines**: `internal/engine/*` (risk, investment_research, compliance).
- **Tools/Data**: `internal/tools/*` (FMP, FRED, SEC, yfinance, newsapi).
- **Shared infra**: `pkg/models`, `pkg/bag`, `pkg/persist`, `pkg/log`, `pkg/errors`, `pkg/cli`, etc.

**Goal:** introduce `internal/llm` for batch without breaking `internal/ai`. Engines can opt‑in to batch mode; later we may deprecate `internal/ai` if desired.

---

## 3) Target Structure (`internal/llm`)

```
internal/llm/
├── client.go                   # Facade: sync + batch (wraps providers)
├── batch/
│   ├── manager.go              # Job lifecycle + prompts integration
│   ├── job.go                  # BatchJob, Status, Filters
│   ├── queue.go                # Optional local status cache
│   ├── result_aggregator.go    # Parse outputs, correlate, partials
│   └── cost_optimizer.go       # Estimate & optimize requests
├── openai/
│   ├── provider.go             # Wraps OpenAI client/config reuse
│   ├── batch_client.go         # File upload + batch create/retrieve
│   ├── batch_formatter.go      # JSONL formatter & parser
│   └── batch_monitor.go        # Polling, eventing, partials
├── middleware/
│   ├── middleware.go           # Interfaces
│   ├── rate_limiting.go
│   └── retry.go
├── validation/
│   └── config_validator.go
└── templates/
    ├── portfolio_analysis.go
    ├── market_research.go
    └── compliance_reports.go
```

**Design principles**

- **Thin provider adapters**: keep OpenAI specifics in `internal/llm/openai`.
- **Provider-agnostic contracts**: types live in `pkg/models`.
- **Prompt reuse**: build requests via `internal/prompt.Manager` & `RegionalManager`.
- **No duplication** of `.tmpl` files; batch templates are _builders_, not content.

---

## 4) Provider-Agnostic Batch Contracts (`pkg/models`)

Create `pkg/models/ai_batch.go`:

```go
package models

type ModelClass string
const (
    ModelClassStandard  ModelClass = "standard"
    ModelClassReasoning ModelClass = "reasoning"
)

type BatchStatus string
const (
    BatchStatusValidating BatchStatus = "validating"
    BatchStatusFailed     BatchStatus = "failed"
    BatchStatusInProgress BatchStatus = "in_progress"
    BatchStatusFinalizing BatchStatus = "finalizing"
    BatchStatusCompleted  BatchStatus = "completed"
    BatchStatusExpired    BatchStatus = "expired"
    BatchStatusCancelled  BatchStatus = "cancelled"
)

type BatchRequest struct {
    CustomID   string                 `json:"custom_id"`
    Method     string                 `json:"method"`
    URL        string                 `json:"url"`
    Body       map[string]any         `json:"body"`
    ModelClass ModelClass             `json:"model_class,omitempty"`
}

type BatchOptions struct {
    CompletionWindow string            `json:"completion_window"` // e.g. "24h"
    Metadata         map[string]string `json:"metadata,omitempty"`
    Priority         string            `json:"priority,omitempty"` // "low|normal|high"
    CostOptimize     bool              `json:"cost_optimize"`
    ModelClass       ModelClass        `json:"model_class,omitempty"`
}

type BatchJob struct {
    ID            string            `json:"id"`
    Status        BatchStatus       `json:"status"`
    InputFileID   string            `json:"input_file_id"`
    OutputFileID  *string           `json:"output_file_id"`
    ErrorFileID   *string           `json:"error_file_id"`
    CreatedAt     int64             `json:"created_at_unix"`
    CompletedAt   *int64            `json:"completed_at_unix"`
    RequestCounts struct {
        Total, Completed, Failed int `json:"total","completed","failed"`
    } `json:"request_counts"`
    Metadata     map[string]string `json:"metadata"`
    CostEstimate *CostEstimate     `json:"cost_estimate,omitempty"`
}

type CostEstimate struct {
    EstimatedCost      float64 `json:"estimated_cost"`
    SavingsVsSync      float64 `json:"savings_vs_sync"`
    EstimatedTokensIn  int     `json:"estimated_tokens_in"`
    EstimatedTokensOut int     `json:"estimated_tokens_out"`
}

type AiBatchClient interface {
    SubmitBatch(ctx context.Context, reqs []BatchRequest, opts BatchOptions) (*BatchJob, error)
    GetBatchStatus(ctx context.Context, jobID string) (*BatchJob, error)
    GetBatchResults(ctx context.Context, jobID string) (io.ReadCloser, error) // stream output file
    GetBatchErrors(ctx context.Context, jobID string) (io.ReadCloser, error)  // stream error file
    CancelBatch(ctx context.Context, jobID string) error
    ListBatches(ctx context.Context, filters map[string]string) ([]BatchJob, error)
}
```

**Why here?** Engines and CLI depend on these interfaces without importing any provider SDKs.

---

## 5) Batch vs. Sync: Clear Differences

| Aspect        | Sync (`internal/ai`) | Batch (`internal/llm`)                    |
| ------------- | -------------------- | ----------------------------------------- |
| Transport     | direct API call      | JSONL file → batch job                    |
| Latency       | immediate            | minutes to hours (≤24h window)            |
| Scale         | per request          | up to 50k requests / 200MB                |
| Pricing       | standard             | **discounted** (provider‑dependent)       |
| Parameters    | model‑dependent      | also model‑dependent; _different_ rules   |
| Errors        | immediate            | per‑line failures + error file            |
| Observability | request logs         | jobs, status, output+error files, metrics |

**Model‑class handling**

- **Standard** models: allow `temperature`, `tools`, `tool_choice`.
- **Reasoning** (e.g., gpt‑5, o1): restrict parameters (often no `temperature`/`tool_choice`), may require embedding “tool context” in the prompt.

---

## 6) Request/Result Formats (OpenAI Batch)

**JSONL request line**

```json
{"custom_id":"risk_portfolio_123","method":"POST","url":"/v1/chat/completions","body":{ "...openai payload..." }}
```

**Success output file (JSONL)**

- Each line contains the model response and `custom_id` to correlate.

**Error file (JSONL)**

- Each line contains `custom_id`, HTTP status, and error payload for failed requests.

**Correlating results**
Use `custom_id` to stitch outputs back to portfolios, store a consolidated object in `SharedBag` and/or `pkg/persist`.

---

## 7) Core Components & Responsibilities

### 7.1 `internal/llm/openai/batch_formatter.go`

- Convert `[]models.BatchRequest` → JSONL stream (`io.ReadSeeker`).
- Avoid loading entire files in memory for huge batches (streaming writer).

### 7.2 `internal/llm/openai/batch_client.go`

- Upload JSONL (`purpose=batch`).
- Create batch job (`endpoint=/v1/chat/completions` or `/v1/responses`).
- Retrieve job, fetch output & error files as streams.

### 7.3 `internal/llm/batch/manager.go`

- Build requests from **existing** `internal/prompt.Manager` (+ `RegionalManager` if enabled).
- Inject model‑class–aware params (skip unsupported fields for reasoning models).
- Submit, persist metadata (`pkg/persist`), publish monitoring events.
- Implement domain helpers, e.g. `SubmitPortfolioAnalysisBatch(...)`.

### 7.4 `internal/llm/batch/result_aggregator.go`

- Stream‑parse output JSONL, map by `custom_id`.
- If error file exists, collect failures and attach diagnostics.
- Produce `AggregatedResults` summary: counts, per‑portfolio items, partials.
- Store in `SharedBag` and optionally persist (for reports).

### 7.5 `internal/llm/batch/cost_optimizer.go`

- Load **pricing table** (JSON, versioned under `internal/config/pricing/`).
- Estimate tokens (heuristic) & cost; choose optimal model when allowed.
- Guardrails: block submission if estimate exceeds ceiling (unless `--force`).
- Feedback loop: log estimate vs actual (when available).

### 7.6 Middleware & Validation

- Retries on transient errors; rate limiting tuned for batch endpoints.
- Validate that **reasoning** model requests do not set unsupported params.

---

## 8) Prompt System Reuse (no duplication)

- Keep `.tmpl` files in `internal/prompt/templates/*` as the **single source of truth**.
- Batch templates (`internal/llm/templates/*.go`) are _builders_ that assemble system/user prompts and model params, given a portfolio/context.
- For reasoning models, convert tool definitions to a **“capability context”** paragraph appended to the system prompt.

---

## 9) CLI UX (in `pkg/cli`)

Add to `analyze` commands:

- `--batch` – submit as batch.
- `--wait` – block until completed and display summary.
- `--regional` – use `RegionalManager` overlays.
- `--template` – choose prompt variant.
- `--batch-priority` – hint for scheduling.
- `--dry-run` – **show cost estimate** and preview first N JSONL lines; do not submit.

New commands:

```
mosychlos portfolio batch status <job-id>
mosychlos portfolio batch results <job-id> --output json|md
mosychlos portfolio batch logs <job-id>    # summarize error file
```

Examples:

```
mosychlos portfolio analyze risk --batch --dry-run
mosychlos portfolio analyze performance --batch --wait --template=conservative
mosychlos portfolio batch status batch_abc123
mosychlos portfolio batch results batch_abc123 --output md
```

---

## 10) Observability (Metrics & Logs)

Export Prometheus metrics (from `internal/health/monitor.go` or a new package):

- `llm_batch_jobs_submitted_total`
- `llm_batch_jobs_completed_total`
- `llm_batch_jobs_failed_total`
- `llm_batch_duration_seconds`
- `llm_batch_tokens_in_total`
- `llm_batch_tokens_out_total`
- `llm_batch_estimated_cost_total`
- `llm_batch_actual_cost_total` (if available)

**Alerts**: cost overrun, job stuck in-progress over SLO, failure ratio > threshold.

**Structured logs**: include `job_id`, `status`, counts, estimate, and model.

---

## 11) Security & Compliance

- Encrypt input files during upload; secure storage for job metadata.
- PII redaction/masking where applicable.
- Data residency boundaries for jurisdictions (EU-only paths if required).
- RBAC: who can submit, cancel, read results.
- Full audit trail (job lifecycle + cost attribution).

---

## 12) Configuration (extend `internal/config/types.go`)

- `LLMBatchEnabled bool`
- `BatchPollingInterval time.Duration`
- `BatchMaxCost float64`
- `BatchEndpoint string` (default `/v1/chat/completions`, opt `/v1/responses`)
- `DefaultModel string`, `DefaultReasoningModel string`
- `PricingTablePath string` (e.g., `internal/config/pricing/openai-2025-08.json`)

---

## 13) Migration Plan (low risk, 2–3 PRs)

**PR1 – Skeleton & contracts**

- Add `pkg/models/ai_batch.go`.
- Create `internal/llm` skeleton + TODOs.
- CLI: `--batch`, `--dry-run` (no submit yet).

**PR2 – OpenAI batch happy path**

- Implement formatter, batch client (upload/create/retrieve), monitor.
- Manager builds requests from prompt manager.
- Aggregator parses outputs (success path).

**PR3 – Reliability & integration**

- Error file parsing, partials, retries.
- Cost estimator + feedback logs.
- Engines add optional batch entrypoints.
- CLI: `status`, `results`, `logs`.
- Metrics and basic alerts.

Later: dashboards, advanced optimizer, multi-provider support.

---

## 14) Testing Strategy

- **Unit**: formatter (golden files), model‑class param switches, aggregator partial handling.
- **Integration (mock provider)**: upload fail, create fail, retrieve polling, large outputs streamed.
- **Load**: generate 50k requests; ensure constant memory profile with stream I/O.
- **CLI**: snapshot tests for `--dry-run` preview and result rendering.

---

## 15) Minimal Code Stubs

**Model‑class detection helper** (place in `internal/llm/util.go` or `client.go`):

```go
func DetectModelClass(model string) models.ModelClass {
    if strings.HasPrefix(model, "gpt-5") || strings.HasPrefix(model, "o1-") {
        return models.ModelClassReasoning
    }
    return models.ModelClassStandard
}
```

**JSONL formatter** (`internal/llm/openai/batch_formatter.go`):

```go
type jsonlReq struct {
    CustomID string         `json:"custom_id"`
    Method   string         `json:"method"`
    URL      string         `json:"url"`
    Body     map[string]any `json:"body"`
}

func RequestsToJSONL(reqs []models.BatchRequest) (io.ReadSeeker, error) {
    var buf bytes.Buffer
    enc := json.NewEncoder(&buf)
    for _, r := range reqs {
        line := jsonlReq{CustomID: r.CustomID, Method: r.Method, URL: r.URL, Body: r.Body}
        if err := enc.Encode(&line); err != nil {
            return nil, err
        }
    }
    return bytes.NewReader(buf.Bytes()), nil
}
```

**Aggregator (partial handling sketch)**:

```go
type Aggregated struct {
    JobID     string
    Successes int
    Failures  int
    Items     map[string]any // by CustomID
    Errors    map[string]string
}
```

---

## 16) Folder-by-Folder TODOs

- **internal/llm/**: add all files shown above; wire config + logging.
- **pkg/models/**: add `ai_batch.go`; ensure `go:build` tags not needed.
- **pkg/cli/**: extend analyze commands; add `batch` subcommands.
- **internal/engine/\*/**: add optional batch entrypoints; keep sync paths.
- **internal/health/**: export batch metrics; register in README.
- **internal/config/**: new fields + pricing table path; sample JSON.
- **internal/report/**: add a renderer for batch summaries (optional).

---

## 17) Glossary

- **Batch job**: Asynchronous group of LLM requests submitted via JSONL.
- **Output file**: Provider‑generated JSONL results (success lines only).
- **Error file**: Provider‑generated JSONL failures (per‑line errors).
- **Model class**: `standard` vs `reasoning` feature sets, parameter support.
- **Prompt overlay**: Regional or variant layer applied by `RegionalManager`.

---

## 18) Appendix – Pricing Table (example schema)

```json
{
  "provider": "openai",
  "version": "2025-08",
  "models": {
    "gpt-4o-mini": {
      "input_per_million": 0.3,
      "output_per_million": 1.2,
      "batch_discount": 0.5
    },
    "gpt-4o": {
      "input_per_million": 5.0,
      "output_per_million": 15.0,
      "batch_discount": 0.5
    },
    "gpt-5-mini": {
      "input_per_million": 6.0,
      "output_per_million": 18.0,
      "batch_discount": 0.5
    },
    "gpt-5": {
      "input_per_million": 15.0,
      "output_per_million": 45.0,
      "batch_discount": 0.5
    }
  }
}
```

> Store under `config/pricing/openai-2025-08.json` and load at startup.

---

## 19) Final Notes

- This plan is **incremental** and **non‑breaking**.
- It maximizes reuse of your prompt system and engine logic.
- Start with the skeleton + CLI `--dry-run`; then land the happy path; then add reliability, metrics, and cost feedback loops.
- Keep `internal/ai` for sync until batch paths are battle‑tested.
