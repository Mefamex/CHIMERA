// file path: internal/clients/ollama/client.go
package ollama

import (
	"strings"
	"net/http"
	"time"
)

func NewOllamaClient(baseURL string) *OllamaClient {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	return &OllamaClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 600 * time.Second,
		},
	}
}
