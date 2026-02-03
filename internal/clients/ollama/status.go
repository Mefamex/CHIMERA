// file path: internal/clients/ollama/status.go
package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Ping
// IsOnline
// Version

func (c *OllamaClient) Ping(ctx context.Context) error {
	// Sadece bağlantı kurulabiliyor mu diye bakarız
	_, err := c.IsOnline(ctx)
	return err
}

func (c *OllamaClient) IsOnline(ctx context.Context) (bool, error) {
	const timeoutResponse = 5 * time.Second

	ctx, cancel := context.WithTimeout(ctx, timeoutResponse)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL, nil)
	if err != nil {
		return false, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Read and print response body for debugging
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("failed to read response body:", err)
	} else {
		var obj any
		content := ""
		if json.Unmarshal(data, &obj) == nil {
			if b, err := json.MarshalIndent(obj, "", "  "); err == nil {
				content = string(b)
			} else {
				content = string(data)
			}
		} else {
			content = string(data)
		}
		if !strings.Contains(content, "Ollama is running"){
			return false, nil
		}
	}

	return resp.StatusCode == http.StatusOK, nil
}

func (c *OllamaClient) Version(ctx context.Context) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/api/version", nil)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Version, nil
}
