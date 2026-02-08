// file path: internal/clients/ollama/client.go
package ollama

import (
	"CHIMERA/internal/core"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ollama/ollama/api"
)

var log = core.GetLogger()
var config = core.Config

// NewClient: Resmi Ollama istemcisini başlatır ve docker kontrolü yapar.
func NewClient() (*api.Client, error) {
	hostUrl := strings.TrimRight(strings.TrimSpace(config.LLM.Host), "/")
	if hostUrl == "" {
		hostUrl = "http://127.0.0.1:11434"
	}

	log.System("Startup", "Initializing Ollama Client", "component", "OLLAMA", "host", hostUrl)

	u, err := url.Parse(hostUrl)
	if err != nil {
		log.Error("Invalid Ollama URL", fmt.Errorf("%v", err), "component", "OLLAMA")
		return nil, err
	}

	client := api.NewClient(u, http.DefaultClient)

	// Bağlantı kontrolü
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Heartbeat(ctx); err != nil {
		log.Error("Ollama is not reachable. Attempting to start Docker...", fmt.Errorf("%v", err), "component", "OLLAMA")
		startOllamaDocker()
	} else {
		log.System("Startup", "Connected to Ollama server successfully", "component", "OLLAMA")
	}

	return client, nil
}

func startOllamaDocker() {
	log.System("Action", "Trying to activate with Docker", "component", "OLLAMA")

	cmdStart := exec.Command("systemctl", "start", "--now", "docker")
	if err := cmdStart.Run(); err != nil {
		log.Error("❌ Docker service start failed", fmt.Errorf("%v", err), "component", "OLLAMA")
		return
	}

	log.System("Action", "🐳 Starting Ollama container...", "component", "OLLAMA")
	cmdRun := exec.Command("docker", "start", "ollama")
	cmdRun.Stdout = os.Stdout
	cmdRun.Stderr = os.Stderr

	if err := cmdRun.Run(); err != nil {
		log.Error("❌ Docker container start failed", fmt.Errorf("%v", err), "component", "OLLAMA")
	}
}
