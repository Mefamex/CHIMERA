// file path: internal/clients/ollama/stream.go
package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (c *OllamaClient) ChatStream(
	ctx context.Context,
	model string,
	messages []OllamaChatMessage,
	format any,
	options map[string]any,
	onDelta func(delta string),
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
		Stream:   true,
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

	var full strings.Builder
	r := bufio.NewReader(resp.Body)
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return full.String(), fmt.Errorf("read ollama stream: %w", err)
		}
		l := strings.TrimSpace(string(line))
		if l == "" {
			continue
		}

		var chunk ollamaChatResponse
		if err := json.Unmarshal([]byte(l), &chunk); err != nil {
			// bazı ortamlarda boş/partial satır gelebiliyor; tolere et
			continue
		}
		if chunk.Error != "" {
			return full.String(), fmt.Errorf("ollama stream error: %s", chunk.Error)
		}
		if chunk.Message.Content != "" {
			full.WriteString(chunk.Message.Content)
			if onDelta != nil {
				onDelta(chunk.Message.Content)
			}
		}
		if chunk.Done {
			break
		}
	}

	return full.String(), nil
}
