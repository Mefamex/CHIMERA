// file path: internal/core/config.go
package core

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
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
var isInitialized bool = false

// mu: Eşzamanlı okuma/yazma (Concurrency) güvenliği için kilit mekanizması.
var mu sync.RWMutex

func init() {
	Config.Init()
}

func (*configContainer) Init() {
	if isInitialized {
		fmt.Printf("[CONFIG] Config initialized already\n")
		return
	}
	Config = &configContainer{}
	if err := Config.Reload(); err != nil {
		fmt.Printf("[CONFIG] ⚠️ Config init hatası: %v\n", err)
	} else {
		fmt.Printf("[CONFIG] ✅ Config initialized successfully from %s\n", configPath)
	}
	isInitialized = true
}

func (c *configContainer) Initialized() bool {
	return isInitialized
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

	Runtimes struct {
		Python      string `yaml:"python"`
		NodeModules string `yaml:"node_modules"`
		NodeBin     string `yaml:"node_bin"`
	} `yaml:"runtimes"`

	LLM struct {
		Provider                string  `yaml:"provider"`
		Host                    string  `yaml:"host"`
		ModelMain               string  `yaml:"model_main"`
		FormatterModel          string  `yaml:"formatter_model"`
		Temperature             float64 `yaml:"temperature"`
		FormatterTemperature    float64 `yaml:"formatter_temperature"`
		Format                  string  `yaml:"format"`
		FormatterFormat         string  `yaml:"formatter_format"`
		MaxConversationMessages int     `yaml:"max_conversation_messages"`
	} `yaml:"llm"`

	Paths struct {
		PromptsDir string `yaml:"prompts_dir"`
		Database   string `yaml:"database"`
		LogsDir    string `yaml:"logsdir"`
	} `yaml:"paths"`
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

	if err := yaml.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("yaml hatası: %w", err)
	}

	// compare YAML vs struct and log/report issues
	cmpErr := c.CompareYAML(configPath)
	if cmpErr != nil {
		fmt.Printf("[CONFIG] ⚠️ compare error: %v\n[CONFIG] ⚠️ Check file: %s\n\n", cmpErr, configPath)
		return fmt.Errorf("Unexpected yaml keys present or struct fields unset.\n")
	} else {
		fmt.Printf("[CONFIG] ✅ YAML-Struct comparison passed without issues.\n")
	}

	// *c = temp diyerek mevcut pointer'ın gösterdiği veriyi DEĞİŞTİRİYORUZ.
	// Böylece core.Config adresi değişmiyor ama içindeki veri değişiyor.
	*c = temp

	// Burası sıkıntılıydı ha.. çok kullanma bu yapıyı (*x = y -> Deep Copy)
	fmt.Printf("[CONFIG] ✅ Config OK: %s\n", configPath)
	// ATTEMPTION: write a checker every required field and return error if missing (after unmarshal)

	// print with indent
	out, err := yaml.Marshal(c)
	if err != nil {
		fmt.Printf("[CONFIG] ⚠️ print marshal error: %v\n", err)
	} else {
		fmt.Println(string(out))
	}

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

// CompareYAML compares the YAML file with configContainer struct
func (c *configContainer) CompareYAML(filePath string) error {
	structUnset := []string{}
	yamlUnmapped := []string{}
	if filePath == "" {
		filePath = configPath
	}

	if _, statErr := os.Stat(filePath); os.IsNotExist(statErr) {
		return fmt.Errorf("dosya bulunamadı: %s", filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("dosya okunamadı: %w", err)
	}

	// Parse raw YAML into generic map to collect keys present in YAML
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("yaml parse hatası: %w", err)
	}

	yamlKeys := map[string]struct{}{}
	var collectY func(node interface{}, prefix string)
	collectY = func(node interface{}, prefix string) {
		switch v := node.(type) {
		case map[string]interface{}:
			for k, val := range v {
				nextPrefix := k
				if prefix != "" {
					nextPrefix = prefix + "." + k
				}
				if _, ok := val.(map[string]interface{}); ok {
					collectY(val, nextPrefix)
				} else {
					yamlKeys[nextPrefix] = struct{}{}
				}
			}
		default:
			key := strings.TrimSuffix(prefix, "")
			if key != "" {
				yamlKeys[key] = struct{}{}
			}
		}
	}
	collectY(raw, "")

	// Unmarshal into a temp config to inspect which struct fields remain zero
	var temp configContainer
	if err := yaml.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("yaml parse hatası: %w", err)
	}

	structKeys := map[string]struct{}{}

	// Recursively collect struct keys (from `yaml` tags) and record zero-valued fields
	var collectStruct func(t reflect.Type, v reflect.Value, prefix string)
	collectStruct = func(t reflect.Type, v reflect.Value, prefix string) {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			tag := f.Tag.Get("yaml")
			name := strings.Split(tag, ",")[0]
			if name == "" || name == "-" {
				continue
			}

			fv := v.Field(i)
			if f.Type.Kind() == reflect.Struct {
				collectStruct(f.Type, fv, prefix+name+".")
				continue
			}

			key := strings.TrimSuffix(prefix+name, ".")
			structKeys[key] = struct{}{}
			if fv.IsZero() {
				structUnset = append(structUnset, key)
			}
		}
	}

	t := reflect.TypeOf(temp)
	v := reflect.ValueOf(temp)
	collectStruct(t, v, "")

	// Keys present in YAML but not in struct
	for k := range yamlKeys {
		if _, ok := structKeys[k]; !ok {
			yamlUnmapped = append(yamlUnmapped, k)
		}
	}

	sort.Strings(yamlUnmapped)
	sort.Strings(structUnset)
	if len(yamlUnmapped) == 0 && len(structUnset) == 0 {
		fmt.Printf("[CONFIG] ✅ YAML-Struct Comparison: Perfect Match!\n")
		return nil
	}
	if len(yamlUnmapped) > 0 {
		fmt.Printf("[CONFIG] ⚠️ YAML'de struct'ta olmayan anahtarlar: %v\n", yamlUnmapped)
	}
	if len(structUnset) > 0 {
		fmt.Printf("[CONFIG] ⚠️ Struct'ta eksik değerler: %v\n", structUnset)
	}

	return fmt.Errorf("yaml-struct comparison found issues: %d unmapped keys, %d unset fields", len(yamlUnmapped), len(structUnset))
}
