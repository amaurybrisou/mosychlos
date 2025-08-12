# Batch LLM Function Implementation Guide

This document provides detailed implementation guidance for the new batch-related functions created under `internal/llm` and related packages. It complements the architecture document and fills in placeholders with concrete implementation details.

---

## 1. internal/llm/client.go

### `NewLLMClient()`
- **Purpose**: Factory to initialize an LLM client with support for both synchronous and batch APIs.
- **Implementation**:
  - Accept configuration from `internal/config`.
  - Inject provider (`internal/llm/openai.Provider`).
  - Return a struct with methods `DoSync(req)` and `DoBatch(reqs)`.

### `DoSync(req models.PromptRequest)`
- Direct call to the provider’s synchronous API (chat/completions).
- Returns a `models.LLMResponse`.

### `DoBatch(reqs []models.PromptRequest)`
- Delegates to `batch.Manager.SubmitBatch`.

---

## 2. internal/llm/batch/manager.go

### `SubmitBatch(ctx context.Context, prompts []models.PromptRequest) (*BatchJob, error)`
- **Steps**:
  1. Convert `prompts` → JSONL using `openai.BatchFormatter`.
  2. Call `openai.BatchClient.UploadAndStart`.
  3. Persist metadata in `pkg/persist.Manager`.
  4. Return `BatchJob{ID, Status=Pending}`.

### `PollStatus(ctx, jobID string) (*BatchJob, error)`
- Calls `openai.BatchMonitor.CheckStatus` until `Completed` or `Failed`.

---

## 3. internal/llm/batch/job.go

### `BatchJob`
```go
type BatchJob struct {
  ID       string
  Status   string
  Created  time.Time
  Finished *time.Time
  CostEst  float64
}
```

---

## 4. internal/llm/batch/result_aggregator.go

### `AggregateResults(ctx, jobID string) (*Aggregated, error)`
- **Steps**:
  1. Fetch output + error files via `openai.BatchClient.GetResults`.
  2. Stream-scan JSONL.
  3. Correlate `custom_id` with original requests.
  4. Build `Aggregated{Successes, Failures, Items, Errors}`.

---

## 5. internal/llm/batch/cost_optimizer.go

### `EstimateCost(reqs []models.PromptRequest) float64`
- Use `pkg/models.TokenEstimates` to calculate input/output tokens × pricing table.
- Return approximate USD/EUR cost.

---

## 6. internal/llm/openai/batch_formatter.go

### `RequestsToJSONL(reqs []models.BatchRequest) (io.ReadSeeker, error)`
- Encode each request as JSON per line:
```json
{"custom_id":"risk_p1_0","method":"POST","url":"/v1/chat/completions","body":{...}}
```

---

## 7. internal/llm/openai/batch_client.go

### `UploadAndStart(file io.Reader) (string, error)`
- POST `/v1/files` with JSONL, then `/v1/batches` to start job.
- Return batch job ID.

### `GetResults(jobID string) (io.ReadCloser, io.ReadCloser, error)`
- Download `output.jsonl` and `errors.jsonl`.

---

## 8. internal/llm/openai/batch_monitor.go

### `CheckStatus(jobID string) (string, error)`
- Poll `/v1/batches/{id}` until `completed` or `failed`.

---

## 9. CLI: pkg/cli/portfolio.go

### `func runPortfolioAnalyzeRisk(cmd *cobra.Command, args []string)`
- Support `--batch` flag.
- If `--batch`:
  1. Build prompts → `SubmitBatch`.
  2. Poll with spinner.
  3. Fetch + render aggregated report (`internal/report`).

---

## 10. Test Strategy

- **Golden tests**:
  - `testdata/batch/requests.jsonl`
  - `testdata/batch/output.jsonl`
  - `testdata/batch/errors.jsonl`
- **Unit tests**:
  - `BatchFormatter`: ensures valid JSONL.
  - `ResultAggregator`: counts successes + failures correctly.
  - `CostEstimator`: matches pricing table.

---

## 11. Acceptance Criteria

- Streaming-safe: no large buffers in memory.
- Deterministic JSONL formatting.
- Partial results tolerated, errors logged.
- CLI shows cost estimation and progress.
