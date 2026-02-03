// file path: internal/clients/ollama/chat.go
package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (c *OllamaClient) Chat(
	ctx context.Context,
	model string,
	messages []OllamaChatMessage,
	format any,
	options map[string]any,
) (string, error) {
	if c == nil {
		return "", fmt.Errorf("ollama client is nil")
	}
	if c.BaseURL == "" {
		return "", fmt.Errorf("ollama base URL is empty")
	}
	if model == "" {
		return "", fmt.Errorf("ollama model is empty")
	}

	reqBody := ollamaChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
		Options:  options,
	}
	if format != nil {
		reqBody.Format = format
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal ollama request: %w", err)
	}

	url := c.BaseURL + "/api/chat"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		if len(body) > 0 {
			return "", fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}
		return "", fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	var out ollamaChatResponse
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&out); err != nil {
		return "", fmt.Errorf("decode ollama response: %w", err)
	}

	return out.Message.Content, nil
}
