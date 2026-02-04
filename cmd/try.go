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

	// Ping the Ollama server
	if err := client.Ping(context.Background()); err != nil {
		return fmt.Errorf("failed to ping Ollama server: %w", err)
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
		fmt.Printf("Ollama Model Available: %s\n", models)
	}
	return nil
}
