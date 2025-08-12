// internal/llm/openai/batch_formatter_test.go
package openai

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

func TestRequestsToJSONL(t *testing.T) {
	cases := []struct {
		name     string
		requests []models.BatchRequest
		wantErr  bool
		validate func(t *testing.T, output string)
	}{
		{
			name: "single request",
			requests: []models.BatchRequest{
				{
					CustomID: "test_1",
					Method:   "POST",
					URL:      "/v1/chat/completions",
					Body: map[string]any{
						"model": "gpt-4o-mini",
						"messages": []map[string]any{
							{"role": "user", "content": "Hello"},
						},
					},
				},
			},
			validate: func(t *testing.T, output string) {
				lines := strings.Split(strings.TrimSpace(output), "\n")
				if len(lines) != 1 {
					t.Errorf("expected 1 line, got %d", len(lines))
				}

				var req map[string]any
				if err := json.Unmarshal([]byte(lines[0]), &req); err != nil {
					t.Errorf("failed to parse JSON: %v", err)
				}

				if req["custom_id"] != "test_1" {
					t.Errorf("expected custom_id 'test_1', got %v", req["custom_id"])
				}
			},
		},
		{
			name: "multiple requests",
			requests: []models.BatchRequest{
				{
					CustomID: "req_1",
					Method:   "POST",
					URL:      "/v1/chat/completions",
					Body:     map[string]any{"model": "gpt-4o-mini"},
				},
				{
					CustomID: "req_2",
					Method:   "POST",
					URL:      "/v1/chat/completions",
					Body:     map[string]any{"model": "gpt-4o"},
				},
			},
			validate: func(t *testing.T, output string) {
				lines := strings.Split(strings.TrimSpace(output), "\n")
				if len(lines) != 2 {
					t.Errorf("expected 2 lines, got %d", len(lines))
				}

				// Validate each line is valid JSON
				for i, line := range lines {
					var req map[string]any
					if err := json.Unmarshal([]byte(line), &req); err != nil {
						t.Errorf("line %d: failed to parse JSON: %v", i, err)
					}
				}
			},
		},
		{
			name:     "empty requests",
			requests: []models.BatchRequest{},
			validate: func(t *testing.T, output string) {
				if strings.TrimSpace(output) != "" {
					t.Errorf("expected empty output, got: %s", output)
				}
			},
		},
		{
			name: "deterministic output",
			requests: []models.BatchRequest{
				{
					CustomID: "deterministic_test",
					Method:   "POST",
					URL:      "/v1/chat/completions",
					Body: map[string]any{
						"model":       "gpt-4o-mini",
						"temperature": 0.5,
						"max_tokens":  100,
					},
				},
			},
			validate: func(t *testing.T, output string) {
				// Run the same conversion multiple times to ensure deterministic output
				reader1, err1 := RequestsToJSONL([]models.BatchRequest{
					{
						CustomID: "deterministic_test",
						Method:   "POST",
						URL:      "/v1/chat/completions",
						Body: map[string]any{
							"model":       "gpt-4o-mini",
							"temperature": 0.5,
							"max_tokens":  100,
						},
					},
				})
				if err1 != nil {
					t.Fatalf("failed to convert requests: %v", err1)
				}

				data1, err := io.ReadAll(reader1)
				if err != nil {
					t.Fatalf("failed to read data: %v", err)
				}

				if string(data1) != output {
					t.Errorf("output not deterministic")
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			reader, err := RequestsToJSONL(c.requests)

			if c.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Read the output
			data, err := io.ReadAll(reader)
			if err != nil {
				t.Errorf("failed to read output: %v", err)
				return
			}

			output := string(data)
			c.validate(t, output)

			// Test that reader can be reset and read again
			_, err = reader.Seek(0, io.SeekStart)
			if err != nil {
				t.Errorf("failed to seek to start: %v", err)
			}

			data2, err := io.ReadAll(reader)
			if err != nil {
				t.Errorf("failed to read output after seek: %v", err)
				return
			}

			if string(data2) != output {
				t.Errorf("output changed after seek")
			}
		})
	}
}

func TestRequestsToJSONL_StreamingMemorySafety(t *testing.T) {
	// Create a large number of requests to test memory efficiency
	requests := make([]models.BatchRequest, 1000)
	for i := range requests {
		requests[i] = models.BatchRequest{
			CustomID: fmt.Sprintf("req_%d", i),
			Method:   "POST",
			URL:      "/v1/chat/completions",
			Body: map[string]any{
				"model": "gpt-4o-mini",
				"messages": []map[string]any{
					{"role": "user", "content": strings.Repeat("Large content ", 100)},
				},
			},
		}
	}

	reader, err := RequestsToJSONL(requests)
	if err != nil {
		t.Fatalf("failed to convert requests: %v", err)
	}

	// Read line by line to simulate streaming
	scanner := bufio.NewScanner(reader)
	lineCount := 0
	for scanner.Scan() {
		line := scanner.Bytes()

		// Validate each line is valid JSON
		var req map[string]any
		if err := json.Unmarshal(line, &req); err != nil {
			t.Errorf("line %d: invalid JSON: %v", lineCount, err)
		}

		lineCount++
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner error: %v", err)
	}

	if lineCount != 1000 {
		t.Errorf("expected 1000 lines, got %d", lineCount)
	}
}

func TestRequestsToJSONL_ValidJSONLFormat(t *testing.T) {
	requests := []models.BatchRequest{
		{
			CustomID: "format_test_1",
			Method:   "POST",
			URL:      "/v1/chat/completions",
			Body: map[string]any{
				"model": "gpt-4o-mini",
				"messages": []map[string]any{
					{"role": "user", "content": "Test message"},
				},
			},
		},
		{
			CustomID: "format_test_2",
			Method:   "POST",
			URL:      "/v1/chat/completions",
			Body: map[string]any{
				"model":       "gpt-4o",
				"max_tokens":  500,
				"temperature": 0.7,
			},
		},
	}

	reader, err := RequestsToJSONL(requests)
	if err != nil {
		t.Fatalf("failed to convert requests: %v", err)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read data: %v", err)
	}

	output := string(data)
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Check JSONL format requirements
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}

	// Each line should be valid JSON
	for i, line := range lines {
		var obj map[string]any
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			t.Errorf("line %d is not valid JSON: %v", i, err)
		}

		// Check required fields
		requiredFields := []string{"custom_id", "method", "url", "body"}
		for _, field := range requiredFields {
			if _, ok := obj[field]; !ok {
				t.Errorf("line %d missing required field: %s", i, field)
			}
		}
	}

	// Output should end with newline
	if !strings.HasSuffix(output, "\n") {
		t.Error("output should end with newline")
	}
}
