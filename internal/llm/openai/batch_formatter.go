// internal/llm/openai/batch_formatter.go
package openai

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// jsonlReq represents a single JSONL request line
type jsonlReq struct {
	CustomID string         `json:"custom_id"`
	Method   string         `json:"method"`
	URL      string         `json:"url"`
	Body     map[string]any `json:"body"`
}

// RequestsToJSONL converts batch requests to JSONL format
// Must produce exactly len(reqs) lines; each line is a valid JSON object.
func RequestsToJSONL(reqs []models.BatchRequest) (io.ReadSeeker, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	for _, r := range reqs {
		line := jsonlReq{
			CustomID: r.CustomID,
			Method:   r.Method,
			URL:      r.URL,
			Body:     r.Body,
		}

		if err := enc.Encode(&line); err != nil {
			return nil, err
		}
	}

	return bytes.NewReader(buf.Bytes()), nil
}
