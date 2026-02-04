// file path: internal/clients/ollama/client.go
package ollama

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func NewOllamaClient() *OllamaClient {
	url := strings.TrimRight(strings.TrimSpace(config.LLM.Host), "/")
	log.Info("Created Ollama Client", "component", "OLLAMA", "host", url)
	cli := OllamaClient{
		BaseURL: url,
		HTTPClient: &http.Client{
			Timeout: 600 * time.Second,
		},
	}
	res, err := cli.IsOnline(context.Background())
	if err != nil || !res {
		log.Error("Failed to connect to Ollama server", "component", "OLLAMA", "error", err)
		ollamaStart()
	} else {
		log.Info("Connected to Ollama server successfully", "component", "OLLAMA")
	}
	return &cli
}

func ollamaStart() {
	log.Info("Trying to activate with Docker", "component", "OLLAMA")
	// 1. Docker servisini başlat (systemctl start --now docker)
	cmdStart := exec.Command("systemctl", "start", "--now", "docker")

	cmdStart.Stdout = os.Stdout
	cmdStart.Stderr = os.Stderr

	if err := cmdStart.Run(); err != nil {
		log.Info("❌ Docker başlatılamadı", "component", "OLLAMA", "error", err)
		os.Exit(1)
	}

	// 2. Ollama konteynerını çalıştır (docker run ollama)
	log.Info("🐳 Ollama konteyneri çalıştırılıyor...", "component", "OLLAMA")
	cmdRun := exec.Command("docker", "start", "ollama")

	// Konteynır çıktılarını anlık izlemek için terminale log.Info, log.Error, os.Stdin şeklinde bağlayalım
	cmdRun.Stdout = os.Stdout
	cmdRun.Stderr = os.Stderr
	cmdRun.Stdin = os.Stdin

	if err := cmdRun.Run(); err != nil {
		log.Info("❌ Docker run hatası", "component", "OLLAMA", "error", err)
		os.Exit(1)
	}
}
