# LLM Batch Processing – Final Implementation Guide (Claude/Copilot Ready)
**Status:** Final – drop-in plan to extend your project with batch LLM processing  
**Repo module:** `github.com/amaurybrisou/mosychlos`  
**Focus:** Keep it simple, no security section. Reuse existing components. Minimize ambiguity so Copilot/Claude do the right thing.

> This guide **assumes your current tree** (from `tree.json`) with `internal/ai`, `internal/prompt`, `internal/engine/*`, `pkg/models`, `pkg/cli`, `pkg/bag`, `pkg/persist`, `internal/tools/*`, etc. It adds a new `internal/llm` for batch without breaking sync flows.

---

## 0) Why this doc?
- Single source of truth for **how** to implement batch LLM processing in *this* repo.
- Designed for **Claude 3.5/4** and **GitHub Copilot**: exact files, signatures, invariants, test data, and acceptance criteria.
- Keep it incremental and compatible with your engines, prompts, and CLI.

---

## 1) What stays the same
- **Sync path**: `internal/ai` (providers, sessions, streaming) remains intact.
- **Prompts**: `internal/prompt` (manager + regional overlays + templates) is the **only** place for templates.
- **Engines**: `internal/engine/*` keep their current sync logic; we add optional batch entrypoints.
- **Tools**: `internal/tools/*` are unchanged and can be invoked in sync flows as before.

---

## 2) What we add (new package)
Create `internal/llm` to host **batch** functionality:
```
internal/llm/
├── client.go                   # Facade for sync (proxy) + batch (new)
├── batch/
│   ├── manager.go              # Job lifecycle + prompt integration
│   ├── job.go                  # BatchJob, Status, Filters
│   ├── result_aggregator.go    # Parse outputs, correlate, partials
│   └── cost_optimizer.go       # Estimate & optimize requests
└── openai/
    ├── provider.go             # Wrap OpenAI client/config reuse
    ├── batch_client.go         # File upload + batch create/retrieve
    ├── batch_formatter.go      # JSONL formatter
    └── batch_monitor.go        # Polling for status/results
```

**Principles**
- **No template duplication:** keep `.tmpl` in `internal/prompt/templates/*`.
- **Provider specifics** live under `internal/llm/openai`.
- **Provider-agnostic contracts** go to `pkg/models`.

---

## 3) Provider‑agnostic contracts (add this file)
### pkg/models/ai_batch.go
```go
// pkg/models/ai_batch.go
// File: pkg/models/ai_batch.go
package models

import (
	"context"
	"io"
)

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
	Method     string                 `json:"method"`               // "POST"
	URL        string                 `json:"url"`                  // "/v1/chat/completions" or "/v1/responses"
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
	GetBatchResults(ctx context.Context, jobID string) (io.ReadCloser, error) // stream success JSONL
	GetBatchErrors(ctx context.Context, jobID string) (io.ReadCloser, error)  // stream error JSONL
	CancelBatch(ctx context.Context, jobID string) error
	ListBatches(ctx context.Context, filters map[string]string) ([]BatchJob, error)
}
```

---

## 4) Model‑class helper (don’t let models be guessed)
### internal/llm/util.go
```go
// internal/llm/util.go
// File: internal/llm/util.go
package llm

import (
	"strings"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// Invariant: reasoning models must NOT set unsupported params like temperature/tool_choice.
func DetectModelClass(model string) models.ModelClass {
	if strings.HasPrefix(model, "gpt-5") || strings.HasPrefix(model, "o1-") {
		return models.ModelClassReasoning
	}
	return models.ModelClassStandard
}
```

> Replace `{github.com/amaurybrisou/mosychlos}` above with your module if your editor doesn’t auto‑fill.

---

## 5) Minimal stubs (Claude/Copilot will fill bodies)

### internal/llm/openai/batch_formatter.go
```go
// internal/llm/openai/batch_formatter.go
// File: internal/llm/openai/batch_formatter.go
package openai

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// Must produce exactly len(reqs) lines; each line is a valid JSON object.
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

### internal/llm/openai/batch_client.go
```go
// internal/llm/openai/batch_client.go
// File: internal/llm/openai/batch_client.go
package openai

import (
	"context"
	"io"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

type BatchClient interface {
	SubmitBatch(ctx context.Context, reqs []models.BatchRequest, opts models.BatchOptions) (*models.BatchJob, error)
	RetrieveStatus(ctx context.Context, jobID string) (*models.BatchJob, error)
	OpenResults(ctx context.Context, jobID string) (io.ReadCloser, error)
	OpenErrors(ctx context.Context, jobID string) (io.ReadCloser, error)
	Cancel(ctx context.Context, jobID string) error
}
```

### internal/llm/batch/manager.go
```go
// internal/llm/batch/manager.go
// File: internal/llm/batch/manager.go
package batch

import (
	"context"

	"github.com/amaurybrisou/mosychlos/internal/prompt"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

type Manager struct {
	client        models.AiBatchClient
	promptManager prompt.Manager
}

func NewManager(c models.AiBatchClient, pm prompt.Manager) *Manager {
	return &Manager{client: c, promptManager: pm}
}

// TODO: Add functions like SubmitPortfolioAnalysisBatch(ctx, portfolios, analysisType) using promptManager.
```

### internal/llm/batch/result_aggregator.go
```go
// internal/llm/batch/result_aggregator.go
// File: internal/llm/batch/result_aggregator.go
package batch

import (
	"bufio"
	"context"
	"encoding/json"
	"io"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

type Aggregated struct {
	JobID     string
	Successes int
	Failures  int
	Items     map[string]any    // by CustomID
	Errors    map[string]string // by CustomID: raw error JSON
}

type ResultsReader interface {
	GetBatchResults(ctx context.Context, jobID string) (io.ReadCloser, error)
	GetBatchErrors(ctx context.Context, jobID string) (io.ReadCloser, error)
}

func scanJSONLLines(r io.Reader, fn func(map[string]any)) error {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		var m map[string]any
		if err := json.Unmarshal(sc.Bytes(), &m); err == nil {
			fn(m)
		}
	}
	return sc.Err()
}
```

---

## 6) JSONL samples (commit under `testdata/batch/`)
These anchor Copilot’s output and enable golden tests.

### testdata/batch/requests.jsonl
```json
{"custom_id":"risk_p1_0","method":"POST","url":"/v1/chat/completions","body":{"model":"gpt-4o-mini","messages":[{"role":"system","content":"...risk sys..."},{"role":"user","content":"...portfolio p1..."}],"max_tokens":1200,"temperature":0.1}}
{"custom_id":"risk_p2_0","method":"POST","url":"/v1/chat/completions","body":{"model":"gpt-5-mini","messages":[{"role":"system","content":"...risk sys + tool ctx..."},{"role":"user","content":"...portfolio p2..."}],"max_tokens":1200}}
```

### testdata/batch/output.jsonl
```json
{"custom_id":"risk_p1_0","response":{"choices":[{"message":{"role":"assistant","content":"...analysis..."}}]}}
{"custom_id":"risk_p2_0","response":{"choices":[{"message":{"role":"assistant","content":"...analysis reasoning..."}}]}}
```

### testdata/batch/errors.jsonl
```json
{"custom_id":"risk_p2_0","status":429,"error":{"message":"rate limited"}}
```

---

## 7) CLI contract (update in `pkg/cli`)
- `--batch` – submit as batch  
- `--wait` – wait until completed and show summary  
- `--regional` – use RegionalManager overlays  
- `--template` – choose prompt variant  
- `--batch-priority` – "low|normal|high" (metadata only)  
- `--dry-run` – **do not submit**; print cost estimate and first 3 JSONL lines

**Expected output** (use for snapshot tests)
```text
$ mosychlos portfolio analyze risk --batch --dry-run
Estimate: €3.12 (≈ –50% vs sync)
Preview JSONL (3 lines):
{"custom_id":"risk_p1_0",...}
{"custom_id":"risk_p2_0",...}
{"custom_id":"risk_p3_0",...}
```

```text
$ mosychlos portfolio batch status batch_abc123
Status: completed | Total: 250 | Completed: 247 | Failed: 3
```

```text
$ mosychlos portfolio batch results batch_abc123 --output md
# Batch Results (batch_abc123)
- Completed: 247
- Failed: 3
...
```

---

## 8) Acceptance criteria (so Copilot/Claude don’t hallucinate)

### Batch Formatter
- Produces exactly **N lines** for **N requests**.
- Each line is a valid JSON object; newline‑terminated.
- Streaming‑friendly; avoid large allocations.

### Batch Client
- `SubmitBatch` uploads JSONL and creates job; returns `BatchJob` with ID.
- `RetrieveStatus` returns status + request counts.
- `OpenResults` returns a reader for success JSONL; `OpenErrors` for errors JSONL.
- Never panic; always wrap errors with context.

### Manager
- Builds requests from the **existing prompt manager** (`internal/prompt`), not new templates.
- Injects **model‑class–aware params**: do **not** set `temperature` / `tool_choice` for reasoning models.
- Adds metadata (analysis type, counts) and supports `--regional` overlays.

### Result Aggregator
- Stream‑parses output JSONL; correlates by `custom_id`.
- If error file exists, map failures by `custom_id`; allow partial success.
- Returns counts: total, successes, failures + per‑ID items.

### Config (minimal)
- `LLMBatchEnabled bool`
- `BatchPollingInterval string` (e.g. `5s`)
- `BatchEndpoint string` (default `/v1/chat/completions`)
- `DefaultModel string`, `DefaultReasoningModel string`
- `PricingTablePath string` (JSON with model pricing + batch_discount)

---

## 9) Unit test skeletons (names ready for Copilot)

### pkg/models/ai_batch_test.go
```go
// pkg/models/ai_batch_test.go
// File: pkg/models/ai_batch_test.go
package models_test

import "testing"

func TestModelClassDetection(t *testing.T) {
	// Implement using llm.DetectModelClass once available.
}
```

### internal/llm/openai/batch_formatter_test.go
```go
// internal/llm/openai/batch_formatter_test.go
// File: internal/llm/openai/batch_formatter_test.go
package openai

import (
	"testing"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

func TestRequestsToJSONL(t *testing.T) {
	reqs := []models.BatchRequest{
		{CustomID: "a", Method: "POST", URL: "/v1/chat/completions", Body: map[string]any{{"x":1}}},
		{CustomID: "b", Method: "POST", URL: "/v1/chat/completions", Body: map[string]any{{"y":2}}},
	}
	r, err := RequestsToJSONL(reqs)
	if err != nil { t.Fatal(err) }
	_ = r // TODO: Copilot to add assertions on content/line count
}
```

### internal/llm/batch/result_aggregator_test.go
```go
// internal/llm/batch/result_aggregator_test.go
// File: internal/llm/batch/result_aggregator_test.go
package batch

import "testing"

func TestScanJSONLLines(t *testing.T) {
	// Provide a small JSONL buffer; ensure lines are decoded and visited.
}
```

---

## 10) Migration (2–3 PRs)

**PR1 – Contracts + Stubs**
- Add `pkg/models/ai_batch.go`.
- Add `internal/llm` skeleton files above.
- Add `testdata/batch/*` and CLI flags (`--batch`, `--dry-run`) wiring only.

**PR2 – Happy Path**
- Implement OpenAI batch client (upload/create/retrieve).
- Implement manager build+submit and status monitor.
- Implement aggregator success path; CLI `status` and `results`.

**PR3 – Reliability**
- Parse error file; handle partials; retry policy.
- Cost estimation (pricing table) + `--dry-run` preview.

---

## 11) Developer notes
- Keep sync flows (`internal/ai`) as-is; engines opt into batch gradually.
- Reuse `internal/prompt` exclusively for composing batch request messages.
- Keep param rules simple: **reasoning models** = no `temperature` / no `tool_choice`.

---

**End of Guide**
