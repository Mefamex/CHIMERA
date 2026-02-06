// file path : internal/agent/brain/structs.go
package brain

import (
	"CHIMERA/internal/clients/ollama"
	"CHIMERA/internal/core"
)

// globals of brain
var Brain BrainStruct

// locals of package
var log = core.Log
var config = core.Config

type BrainStruct struct {
	Ollama       *ollama.OllamaClient
	Model        string
	SystemPrompt string
	Temperature  float64
	Format       any
}
