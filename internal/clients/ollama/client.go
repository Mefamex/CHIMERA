// file path: internal/clients/ollama/client.go
package ollama

import (
	"net/http"
	"strings"
	"time"
)

func NewOllamaClient() *OllamaClient {
	url := strings.TrimRight(strings.TrimSpace(config.LLM.Host), "/")
	log.Info("Created Ollama Client", "component", "OLLAMA", "host", url)
	return &OllamaClient{
		BaseURL: url,
		HTTPClient: &http.Client{
			Timeout: 600 * time.Second,
		},
	}
}
