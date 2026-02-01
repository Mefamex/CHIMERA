// file path: internal/config/config.go
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config yapısı, YAML dosyasındaki hiyerarşinin birebir aynısıdır.
type Config struct {
	App      AppConfig          `yaml:"app"`
	Server   ServerConfig       `yaml:"server"`
	Runtime  RuntimePathsConfig `yaml:"runtime_paths"`
	LLM      LLMConfig          `yaml:"llm"`
	Services ServicesConfig     `yaml:"services"`
	Telegram TelegramConfig     `yaml:"telegram"`
	Paths    PathsConfig        `yaml:"paths"`
}

type AppConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
	Debug       bool   `yaml:"debug"`
}

type ServerConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	ReadTimeout  string `yaml:"read_timeout"`
	WriteTimeout string `yaml:"write_timeout"`
}

type LLMConfig struct {
	Provider      string  `yaml:"provider"`
	Host          string  `yaml:"host"`
	ModelMain     string  `yaml:"model_main"`
	ModelCoder    string  `yaml:"model_coder"`
	Temperature   float64 `yaml:"temperature"`
	ContextWindow int     `yaml:"context_window"`
	Format        string  `yaml:"format"`
}

type ServicesConfig struct {
	PythonAnalyzer ServiceDetail `yaml:"python_analyzer"`
	NodeBrowser    ServiceDetail `yaml:"node_browser"`
}

type ServiceDetail struct {
	URL     string `yaml:"url"`
	Timeout string `yaml:"timeout"`
}

type TelegramConfig struct {
	Enabled      bool    `yaml:"enabled"`
	Token        string  `yaml:"token"`
	AllowedUsers []int64 `yaml:"allowed_users"`
}

type PathsConfig struct {
	PromptsDir string `yaml:"prompts_dir"`
	Database   string `yaml:"database"`
	Logs       string `yaml:"logs"`
	Downloads  string `yaml:"downloads"`
}

type RuntimePathsConfig struct {
	Python      string `yaml:"python"`
	NodeModules string `yaml:"node_modules"`
	NodeBin     string `yaml:"node_bin"`
}

// LoadConfig belirtilen yoldan YAML dosyasını okur ve Config struct'ına çevirir.
func LoadConfig(configPath string) (*Config, error) {
	// Dosyayı aç
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("config dosyası açılamadı: %w", err)
	}
	defer file.Close()

	// Parse işlemi
	var cfg Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("YAML decode hatası: %w", err)
	}

	return &cfg, nil
}
