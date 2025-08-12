// pkg/openai/json.go
// File: pkg/openai/json.go
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func (c *Client) DoJSON(ctx context.Context, method, url string, headers http.Header, in any, out any) (*http.Response, error) {
	var buf bytes.Buffer
	if in != nil {
		if err := json.NewEncoder(&buf).Encode(in); err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		return nil, err
	}

	for k, vv := range headers {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}

	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.Do(ctx, req)
	if err != nil {
		return resp, err
	}
	if out != nil {
		defer resp.Body.Close()
		return resp, json.NewDecoder(resp.Body).Decode(out)
	}
	return resp, nil
}
