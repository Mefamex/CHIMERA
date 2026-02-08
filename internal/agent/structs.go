package agent

import (
	"CHIMERA/internal/core"

	"github.com/ollama/ollama/api"
)

var log = core.GetLogger()
var config = core.Config

// BrainStruct: Yapay zeka motorunun ana yapisi
type BrainStruct struct {
	Ollama               *api.Client
	Model                string
	ModelFormatter       string
	SystemPrompt         string
	FormatterPrompt      string
	Temperature          float64
	FormatterTemperature float64
}

// Global Brain instance
var Brain BrainStruct

// AgentResponse: Modelin JSON ciktisini karsilayan yapi
type AgentResponse struct {
	Thought string                 `json:"thought"`
	Tool    string                 `json:"tool"`
	Args    map[string]interface{} `json:"args"`
	Reply   string                 `json:"reply"`
}
