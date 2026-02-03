// file path: internal/clients/ollama/listModels.go
package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// NOT : buraya ListModels(ctx).toList() eklse

// ListModels returns the list of available models from the Ollama server.
// It decodes the response shaped like { "object":"list", "data": [ ... ] }.
func (c *OllamaClient) ListModels(ctx context.Context) ([]OllamaModel, error) {
	if c == nil {
		return nil, fmt.Errorf("ollama client is nil")
	}
	if c.BaseURL == "" {
		return nil, fmt.Errorf("ollama base URL is empty")
	}
	url := strings.TrimRight(c.BaseURL, "/") + "/v1/models"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		if len(body) > 0 {
			return nil, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	// Try the {"object":"list","data":[...]} format
	var respWrapper struct {
		Object string        `json:"object"`
		Data   []OllamaModel `json:"data"`
	}
	if err := json.Unmarshal(data, &respWrapper); err == nil {
		if len(respWrapper.Data) > 0 {
			return respWrapper.Data, nil
		}
		// return empty slice if no data
		return []OllamaModel{}, nil
	}

	// Fallback: try direct array of models
	var direct []OllamaModel
	if err := json.Unmarshal(data, &direct); err == nil {
		return direct, nil
	}

	return nil, fmt.Errorf("unable to decode ollama models response")
}
