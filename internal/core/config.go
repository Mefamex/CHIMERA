// file path: internal/core/config.go
package core

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// -- Global Değişkenler --

// Config: Dışarıya AÇIK (Public). Okumak için buna ihtiyacımız var.
// Kullanım: core.Config.App.Name
var Config *configContainer

// Path: Config dosyasının varsayılan yolu.
// configPath: Dışarıya KAPALI (Private). Sadece bu dosya içinden erişilebilir.
var configPath = "./configs/config.yaml"

// mu: Eşzamanlı okuma/yazma (Concurrency) güvenliği için kilit mekanizması.
var mu sync.RWMutex

// -- Init --

// init fonksiyonu paket import edildiği an otomatik çalışır.
func init() {
	// İlk güvenli açılış. Önce boş bir struct oluştur (Nil pointer hatasını önlemek için)
	Config = &configContainer{}
	Config.setDefaults()

	// initilize configs in init.go with Reload
}

// Private Helper
func (c *configContainer) setDefaults() {
	c.Server.Port = 8080
	c.Server.Host = "localhost"
	c.App.Environment = "development"
	c.Paths.LogsDir = "data/logs"
}

func (c *configContainer) Initialized() (*configContainer, error) {
	mu.RLock()
	defer mu.RUnlock()

	// Eğer Config boşsa, yükleme başarısız demektir.
	if Config == nil {
		return nil, fmt.Errorf("configs not initialized or is nil")
	}

	return c, nil
}

// -- CONFIG YAPISI --

type configContainer struct {
	App struct {
		Name        string `yaml:"name"`
		Version     string `yaml:"version"`
		Environment string `yaml:"environment"`
		Debug       bool   `yaml:"debug"`
	} `yaml:"app"`

	Server struct {
		Host         string `yaml:"host"`
		Port         int    `yaml:"port"`
		ReadTimeout  string `yaml:"read_timeout"`
		WriteTimeout string `yaml:"write_timeout"`
	} `yaml:"server"`

	Runtime struct {
		Python      string `yaml:"python"`
		NodeModules string `yaml:"node_modules"`
		NodeBin     string `yaml:"node_bin"`
	} `yaml:"runtime"`

	LLM struct {
		Provider    string  `yaml:"provider"`
		Host        string  `yaml:"host"`
		ModelMain   string  `yaml:"model_main"`
		ModelCoder  string  `yaml:"model_coder"`
		Temperature float64 `yaml:"temperature"`
		Format      string  `yaml:"format"`
	} `yaml:"llm"`

	Paths struct {
		PromptsDir string `yaml:"prompts_dir"`
		Database   string `yaml:"database"`
		LogsDir    string `yaml:"logs"`
	} `yaml:"paths"`
}

//  METODLAR (Yetenekler)

// SetConfigPath: Config dosyasının yolunu değiştirir.
// Kullanım: core.Config.SetConfigPath("/yeni/yol.yaml")
func (c *configContainer) SetConfigPath(path string) {
	mu.Lock()
	defer mu.Unlock()
	configPath = path
}

// GetConfigPath: Mevcut config dosya yolunu döner.
// Kullanım: yol := core.Config.GetConfigPath()
func (c *configContainer) GetConfigPath() string {
	mu.RLock()
	defer mu.RUnlock()
	return configPath
}

// Reload: Dosyayı tekrar okur ve verileri günceller.
// Kullanım: core.Config.Reload()
func (c *configContainer) Reload() error {
	mu.Lock()
	defer mu.Unlock()

	// Dosya var mı?
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("dosya bulunamadı: %s", configPath)
	}

	// Oku
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	// Geçici yapıya yükle
	var temp configContainer
	temp.setDefaults() // Varsayılanları koy

	if err := yaml.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("yaml hatası: %w", err)
	}

	// *c = temp diyerek mevcut pointer'ın gösterdiği veriyi DEĞİŞTİRİYORUZ.
	// Böylece core.Config adresi değişmiyor ama içindeki veri değişiyor.
	*c = temp
	// Burası sıkıntılıydı ha.. çok kullanma bu yapıyı (*x = y -> Deep Copy)
	fmt.Printf("[CONFIG] ✅ Config OK: %s\n", configPath)
	return nil
}

// Export: Mevcut ayarları başka bir dosyaya kaydeder.
// Kullanım: core.Config.Export("./backup_config.yaml")
func (c *configContainer) Export(outputPath string) error {
	mu.RLock()
	defer mu.RUnlock()

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, data, 0644)
}
