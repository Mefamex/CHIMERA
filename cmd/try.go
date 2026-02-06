package cmd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"CHIMERA/internal/clients/ollama"
	"CHIMERA/internal/core"
	"CHIMERA/internal/skills"

	"gopkg.in/yaml.v3"
)

var log = core.Log

func TryOllamaConnection() {
	time.Sleep(5 * time.Second)
	baseURL := "http://localhost:11434"
	client := ollama.OllamaClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}

	// Check if the Ollama server is online
	online, err := client.IsOnline(context.Background())
	if err != nil {
		fmt.Printf("failed to check if Ollama server is online: %v\n", err)
	} else {
		if !online {
			fmt.Println("Ollama server is not online")
		} else {
			models, err := client.ListModels(context.Background())
			if err != nil {
				fmt.Printf("failed to list models from Ollama server: %v\n", err)
			}
			if len(models) == 0 {
				fmt.Println("no models available on Ollama server")
			} else {
				fmt.Printf("✅ Ollama server is online. Available models count: %d\n", len(models))
			}
		}
	}

	// List available models
	tools := skills.GetTools()
	for name, tool := range tools {
		result := fmt.Sprintf("Tool: %s,\n\tCategory: %s,\n\tPath: %s,\n\tDesc: %s,\n\tParams: %v,\n\tReturns: %v\n",
			name, tool.Category, tool.ScriptPath, tool.Description, tool.Parameters, tool.Returns)
		log.Info(result, "component", "Try")
	}

	path := "~/Masaüstü/example.txt"
	result := skills.UseTool("readfile", path)
	out, err := yaml.Marshal(result)
	log.Info("Tool Result:\n"+string(out), "component", "Try")
}
