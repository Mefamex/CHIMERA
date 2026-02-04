// file path: internal/clients/ollama/structs.go
package ollama

import (
	"net/http"
	"CHIMERA/internal/core"
)

// OllamaGlobals
var log = core.Log
var config = core.Config

type OllamaClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

type OllamaModel struct {
	ID string `json:"id,omitempty"`
}

type OllamaChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatRequest struct {
	Model    string              `json:"model"`
	Messages []OllamaChatMessage `json:"messages"`
	Stream   bool                `json:"stream"`
	Format   any                 `json:"format,omitempty"`
	Options  map[string]any      `json:"options,omitempty"`
}

type ollamaChatResponse struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done  bool   `json:"done"`
	Error string `json:"error"`
}
