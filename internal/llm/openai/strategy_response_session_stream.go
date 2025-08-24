// // internal/llm/openai/sync_provider.go
package openai

// import (
// 	"bufio"
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"strings"

// 	llmutils "github.com/amaurybrisou/mosychlos/internal/llm/llm_utils"
// 	"github.com/amaurybrisou/mosychlos/pkg/models"
// )

// func (s *session) NextStream(ctx context.Context, tools []models.ToolDef, rf *models.ResponseFormat) (<-chan models.StreamChunk, error) {
// 	body := responseAPIReq{
// 		Model: s.p.cfg.Model.String(),
// 		Input: s.messages,
// 		Tools: toAnyTools(tools, nil, llmutils.IsReasoningModel(s.p.cfg.Model.String())),
// 	}

// 	// map cfg → request (mirror Next)
// 	if max := s.p.cfg.OpenAI.MaxCompletionTokens; max > 0 {
// 		body.MaxOutputTokens = max
// 	}
// 	if s.p.cfg.OpenAI.Temperature != nil && !llmutils.IsReasoningModel(s.p.cfg.Model.String()) {
// 		body.Temperature = s.p.cfg.OpenAI.Temperature
// 	}
// 	if rf != nil {
// 		body.ResponseFormat = rf
// 	}
// 	if s.p.cfg.OpenAI.ServiceTier != nil && *s.p.cfg.OpenAI.ServiceTier != "auto" {
// 		body.ServiceTier = *s.p.cfg.OpenAI.ServiceTier
// 	}
// 	if s.p.cfg.OpenAI.ParallelToolCalls {
// 		t := true
// 		body.ParallelToolCalls = &t
// 	}

// 	var buf bytes.Buffer
// 	if err := json.NewEncoder(&buf).Encode(body); err != nil {
// 		return nil, err
// 	}

// 	baseURL := s.p.cfg.BaseURL
// 	if baseURL == "" {
// 		baseURL = "https://api.openai.com"
// 	}
// 	req, err := http.NewRequest(http.MethodPost, baseURL+"/v1/responses", &buf)
// 	if err != nil {
// 		return nil, err
// 	}
// 	req = req.WithContext(ctx)
// 	req.Header.Set("Content-Type", "application/json")
// 	// Streaming via SSE
// 	req.Header.Set("Accept", "text/event-stream")
// 	req.Header.Set("Cache-Control", "no-cache")
// 	req.Header.Set("Connection", "keep-alive")
// 	if s.p.cfg.APIKey != "" {
// 		req.Header.Set("Authorization", "Bearer "+s.p.cfg.APIKey)
// 	}

// 	ch := make(chan models.StreamChunk, 64)

// 	// Helper structs for SSE payloads
// 	type textDelta struct {
// 		// Responses API emits: event: response.output_text.delta
// 		Delta string `json:"delta"`
// 	}
// 	type textDone struct {
// 		// Some streams also include the full text at *.done
// 		Text string `json:"text"`
// 	}
// 	type errPayload struct {
// 		Error struct {
// 			Message string `json:"message"`
// 			Type    string `json:"type"`
// 			Param   string `json:"param,omitempty"`
// 			Code    string `json:"code,omitempty"`
// 		} `json:"error"`
// 	}

// 	go func() {
// 		defer close(ch)

// 		resp, err := s.p.cli.Do(ctx, req)
// 		if err != nil {
// 			ch <- models.StreamChunk{Error: err}
// 			return
// 		}
// 		defer resp.Body.Close()

// 		// Non-200 → decode JSON error if possible
// 		if resp.StatusCode != http.StatusOK {
// 			var er responseAPIErr
// 			if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
// 				ch <- models.StreamChunk{Error: fmt.Errorf("stream HTTP %d", resp.StatusCode)}
// 				return
// 			}
// 			ch <- models.StreamChunk{Error: fmt.Errorf("API error: %s (type: %s, param: %s, code: %s)",
// 				er.Error.Message, er.Error.Type, er.Error.Param, er.Error.Code)}
// 			return
// 		}

// 		reader := bufio.NewReader(resp.Body)

// 		// Simple SSE frame parser: accumulate until blank line, then handle {event,data}
// 		for {
// 			select {
// 			case <-ctx.Done():
// 				ch <- models.StreamChunk{Error: ctx.Err(), IsComplete: true}
// 				return
// 			default:
// 			}

// 			var evType string
// 			var dataBuf bytes.Buffer

// 			// Read one SSE event (ends on a blank line)
// 			for {
// 				line, rerr := reader.ReadString('\n')
// 				if rerr != nil {
// 					if rerr == io.EOF {
// 						// Server closed stream
// 						ch <- models.StreamChunk{IsComplete: true}
// 					} else {
// 						ch <- models.StreamChunk{Error: rerr}
// 					}
// 					return
// 				}
// 				line = strings.TrimRight(line, "\r\n")

// 				// Comments or heartbeat
// 				if line == "" {
// 					// End of event frame
// 					break
// 				}
// 				if strings.HasPrefix(line, ":") {
// 					continue
// 				}
// 				if strings.HasPrefix(line, "event:") {
// 					evType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
// 					continue
// 				}
// 				if strings.HasPrefix(line, "data:") {
// 					// Note: multiple data: lines can appear; concat without extra newline
// 					dataLine := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
// 					dataBuf.WriteString(dataLine)
// 					continue
// 				}
// 				// Ignore other SSE fields (id:, retry:, etc.) for now
// 			}

// 			data := dataBuf.Bytes()
// 			if evType == "" && len(data) == 0 {
// 				// Empty frame / keep-alive
// 				continue
// 			}

// 			switch evType {

// 			// Text delta chunks
// 			case "response.output_text.delta":
// 				var td textDelta
// 				if err := json.Unmarshal(data, &td); err != nil {
// 					ch <- models.StreamChunk{Error: fmt.Errorf("decode delta: %w", err)}
// 					return
// 				}
// 				if td.Delta != "" {
// 					ch <- models.StreamChunk{Content: td.Delta}
// 				}

// 			// Optional final text
// 			case "response.output_text.done":
// 				var done textDone
// 				// If this fails, it's non-fatal—some providers omit "text"
// 				_ = json.Unmarshal(data, &done)
// 				// We don't need to emit anything here; completion will arrive.

// 			// Tool/function call deltas (optional handling)
// 			// You can extend these as needed to surface tool call progress.
// 			case "response.function_call.arguments.delta":
// 				// Example payload: {"call_id":"...","name":"...","arguments":"{...partial...}"}
// 				// If you want to expose to the UI/logs, forward raw data:
// 				// ch <- models.StreamChunk{ToolCallDelta: string(data)} // if your struct supports it
// 				// For now, ignore to keep behavior consistent with your non-streaming path.
// 				continue

// 			// Errors pushed by server
// 			case "response.error":
// 				var ep errPayload
// 				if err := json.Unmarshal(data, &ep); err != nil {
// 					ch <- models.StreamChunk{Error: fmt.Errorf("decode error payload: %w", err)}
// 					return
// 				}
// 				if msg := strings.TrimSpace(ep.Error.Message); msg != "" {
// 					ch <- models.StreamChunk{Error: fmt.Errorf("API error: %s (type: %s, param: %s, code: %s)",
// 						ep.Error.Message, ep.Error.Type, ep.Error.Param, ep.Error.Code)}
// 				} else {
// 					ch <- models.StreamChunk{Error: fmt.Errorf("unknown streaming error")}
// 				}
// 				return

// 			// Normal termination signal
// 			case "response.completed":
// 				ch <- models.StreamChunk{IsComplete: true}
// 				return

// 			// Early lifecycle events you might see; safe to ignore
// 			case "response.created", "response.model.started", "response.started":
// 				continue

// 			default:
// 				// Unknown/unsupported event; ignore quietly to be forward-compatible
// 				continue
// 			}
// 		}
// 	}()

// 	return ch, nil
// }
