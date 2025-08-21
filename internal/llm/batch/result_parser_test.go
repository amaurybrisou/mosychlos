// internal/llm/batch/result_aggregator_test.go
package batch

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// mockResultsReader implements ResultsReader for testing
type mockResultsReader struct {
	resultsData string
	errorsData  string
	resultsErr  error
	errorsErr   error
}

func (m *mockResultsReader) GetBatchResults(ctx context.Context, jobID string) (io.ReadCloser, error) {
	if m.resultsErr != nil {
		return nil, m.resultsErr
	}
	return io.NopCloser(strings.NewReader(m.resultsData)), nil
}

func (m *mockResultsReader) GetBatchErrors(ctx context.Context, jobID string) (io.ReadCloser, error) {
	if m.errorsErr != nil {
		return nil, m.errorsErr
	}
	return io.NopCloser(strings.NewReader(m.errorsData)), nil
}

func TestResultAggregator_AggregateResults(t *testing.T) {
	cases := []struct {
		name           string
		resultsData    string
		errorsData     string
		resultsErr     error
		errorsErr      error
		wantSuccesses  int
		wantFailures   int
		validateResult func(t *testing.T, result *models.BatchResult)
	}{
		{
			name: "successful results only",
			resultsData: `{"id": "batch_req_1", "custom_id": "portfolio_analysis_1", "response": {"status_code": 200, "request_id": "req_1", "body": {"id": "chatcmpl-123", "object": "chat.completion", "created": 1703073600, "model": "gpt-4o-mini-2024-07-18", "choices": [{"index": 0, "message": {"role": "assistant", "content": "Portfolio analysis complete."}, "finish_reason": "stop"}], "usage": {"prompt_tokens": 150, "completion_tokens": 75, "total_tokens": 225}}}}
{"id": "batch_req_2", "custom_id": "risk_assessment_1", "response": {"status_code": 200, "request_id": "req_2", "body": {"id": "chatcmpl-456", "object": "chat.completion", "created": 1703073650, "model": "gpt-4o-2024-05-13", "choices": [{"index": 0, "message": {"role": "assistant", "content": "Risk assessment complete."}, "finish_reason": "stop"}], "usage": {"prompt_tokens": 200, "completion_tokens": 100, "total_tokens": 300}}}}`,
			errorsData:    "",
			wantSuccesses: 2,
			wantFailures:  0,
			validateResult: func(t *testing.T, result *models.BatchResult) {
				if len(result.Items) != 2 {
					t.Errorf("expected 2 items, got %d", len(result.Items))
				}
				if len(result.Errors) != 0 {
					t.Errorf("expected 0 errors, got %d", len(result.Errors))
				}

				// Check specific items exist
				if _, ok := result.Items["portfolio_analysis_1"]; !ok {
					t.Error("expected portfolio_analysis_1 in results")
				}
				if _, ok := result.Items["risk_assessment_1"]; !ok {
					t.Error("expected risk_assessment_1 in results")
				}
			},
		},
		{
			name:        "mixed success and errors",
			resultsData: `{"id": "batch_req_1", "custom_id": "success_req", "response": {"status_code": 200, "request_id": "req_1", "body": {"id": "chatcmpl-123", "object": "chat.completion", "created": 1703073600, "model": "gpt-4o-mini-2024-07-18", "choices": [{"index": 0, "message": {"role": "assistant", "content": "Success"}, "finish_reason": "stop"}], "usage": {"prompt_tokens": 100, "completion_tokens": 50, "total_tokens": 150}}}}`,
			errorsData: `{"id": "batch_req_2", "custom_id": "error_req", "error": {"code": "invalid_request_error", "message": "Invalid model specified", "param": "model", "type": "invalid_request_error"}}
{"id": "batch_req_3", "custom_id": "rate_limit_error", "error": {"code": "rate_limit_exceeded", "message": "Rate limit exceeded", "param": null, "type": "rate_limit_error"}}`,
			wantSuccesses: 1,
			wantFailures:  2,
			validateResult: func(t *testing.T, result *models.BatchResult) {
				if len(result.Items) != 1 {
					t.Errorf("expected 1 item, got %d", len(result.Items))
				}
				if len(result.Errors) != 2 {
					t.Errorf("expected 2 errors, got %d", len(result.Errors))
				}

				// Check specific items
				if _, ok := result.Items["success_req"]; !ok {
					t.Error("expected success_req in results")
				}
				if _, ok := result.Errors["error_req"]; !ok {
					t.Error("expected error_req in errors")
				}
				if _, ok := result.Errors["rate_limit_error"]; !ok {
					t.Error("expected rate_limit_error in errors")
				}
			},
		},
		{
			name:        "errors only",
			resultsData: "",
			errorsData: `{"id": "batch_req_1", "custom_id": "failed_req_1", "error": {"code": "content_filter", "message": "Content filtered", "param": null, "type": "content_filter"}}
{"id": "batch_req_2", "custom_id": "failed_req_2", "error": {"code": "insufficient_quota", "message": "Insufficient quota", "param": null, "type": "insufficient_quota"}}`,
			wantSuccesses: 0,
			wantFailures:  2,
			validateResult: func(t *testing.T, result *models.BatchResult) {
				if len(result.Items) != 0 {
					t.Errorf("expected 0 items, got %d", len(result.Items))
				}
				if len(result.Errors) != 2 {
					t.Errorf("expected 2 errors, got %d", len(result.Errors))
				}
			},
		},
		{
			name:          "empty data",
			resultsData:   "",
			errorsData:    "",
			wantSuccesses: 0,
			wantFailures:  0,
			validateResult: func(t *testing.T, result *models.BatchResult) {
				if len(result.Items) != 0 {
					t.Errorf("expected 0 items, got %d", len(result.Items))
				}
				if len(result.Errors) != 0 {
					t.Errorf("expected 0 errors, got %d", len(result.Errors))
				}
			},
		},
		{
			name: "malformed JSON skipped",
			resultsData: `{"id": "batch_req_1", "custom_id": "good_req", "response": {"status_code": 200}}
invalid_json_line_should_be_skipped
{"id": "batch_req_2", "custom_id": "another_good_req", "response": {"status_code": 200}}`,
			errorsData:    "",
			wantSuccesses: 2,
			wantFailures:  0,
			validateResult: func(t *testing.T, result *models.BatchResult) {
				if len(result.Items) != 2 {
					t.Errorf("expected 2 items (malformed line skipped), got %d", len(result.Items))
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mock := &mockResultsReader{
				resultsData: c.resultsData,
				errorsData:  c.errorsData,
				resultsErr:  c.resultsErr,
				errorsErr:   c.errorsErr,
			}

			aggregator := NewBatchResultParser(mock)
			result, err := aggregator.AggregateBatchResult(context.Background(), "test-job-id")

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Validate basic counts
			if result.Successes != c.wantSuccesses {
				t.Errorf("expected %d successes, got %d", c.wantSuccesses, result.Successes)
			}
			if result.Failures != c.wantFailures {
				t.Errorf("expected %d failures, got %d", c.wantFailures, result.Failures)
			}

			// Validate JobID is set
			if result.JobID != "test-job-id" {
				t.Errorf("expected JobID 'test-job-id', got '%s'", result.JobID)
			}

			// Run custom validations
			if c.validateResult != nil {
				c.validateResult(t, result)
			}
		})
	}
}

func TestResultAggregator_ErrorHandling(t *testing.T) {
	cases := []struct {
		name       string
		resultsErr error
		errorsErr  error
		wantErr    bool
	}{
		{
			name:       "results reader error",
			resultsErr: io.EOF,
			wantErr:    true,
		},
		{
			name:      "errors reader error ignored",
			errorsErr: io.EOF,
			wantErr:   false, // errors are ignored as per implementation
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			mock := &mockResultsReader{
				resultsData: "",
				errorsData:  "",
				resultsErr:  c.resultsErr,
				errorsErr:   c.errorsErr,
			}

			aggregator := NewBatchResultParser(mock)
			_, err := aggregator.AggregateBatchResult(context.Background(), "test-job-id")

			if c.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !c.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestScanJSONLLines(t *testing.T) {
	cases := []struct {
		name      string
		data      string
		wantCalls int
	}{
		{
			name: "valid JSONL",
			data: `{"id": 1}
{"id": 2}
{"id": 3}`,
			wantCalls: 3,
		},
		{
			name: "mixed valid and invalid lines",
			data: `{"id": 1}
invalid_json
{"id": 2}`,
			wantCalls: 2, // invalid lines are skipped
		},
		{
			name:      "empty input",
			data:      "",
			wantCalls: 0,
		},
		{
			name: "only invalid lines",
			data: `invalid
also_invalid`,
			wantCalls: 0,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			callCount := 0
			reader := strings.NewReader(c.data)

			err := scanJSONLLines(reader, func(item map[string]any) {
				callCount++
			})

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if callCount != c.wantCalls {
				t.Errorf("expected %d function calls, got %d", c.wantCalls, callCount)
			}
		})
	}
}
