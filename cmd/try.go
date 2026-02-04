package cmd

import (
	"context"
	"fmt"
	"net/http"

	"CHIMERA/internal/clients/ollama"
)

func TryOllamaConnection() error {

	baseURL := "http://localhost:11434"
	client := ollama.OllamaClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}

	// Check if the Ollama server is online
	online, err := client.IsOnline(context.Background())
	if err != nil {
		return fmt.Errorf("failed to check if Ollama server is online: %w", err)
	}
	if !online {
		return fmt.Errorf("Ollama server is not online")
	}

	// List available models
	models, err := client.ListModels(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list models from Ollama server: %w", err)
	}
	if len(models) == 0 {
		return fmt.Errorf("no models available on Ollama server")
	} else {
		// print beautifully all in the json marshal indent
		fmt.Println("✅ Ollama server is online. Available models:")
		for _, model := range models {
			fmt.Printf(" - %s\n", model.ID)
		}
	}
	return nil
}
