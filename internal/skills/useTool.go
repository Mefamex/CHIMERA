// file path : internal/skills/useTool.go
package skills

import (
	"CHIMERA/internal/core"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

var toolsCache = make(map[string]Tool)
var toolsMu sync.RWMutex
var log = core.Log

type Tool struct {
	Name        string            `json:"-"`
	Category    string            `json:"Category"`
	ScriptPath  string            `json:"Path"`
	Description string            `json:"Desc"`
	Parameters  []string          `json:"Params"`
	Returns     map[string]string `json:"Returns"`
}
type ToolResult struct {
	ToolName string                 `json:"tool_name"`
	Data     map[string]interface{} `json:"data"`      // Python'dan gelen asıl JSON verisi
	Stdout   string                 `json:"stdout"`    // Ham standart çıktı (Debug için)
	Stderr   string                 `json:"stderr"`    // Hata logları (Python traceback vb.)
	Duration time.Duration          `json:"duration"`  // Çalışma süresi
	ExitCode int                    `json:"exit_code"` // 0: Başarılı, Diğerleri: Hata
	Error    error                  `json:"-"`         // Go tarafındaki çalıştırma hatası (Timeout vb.)
}

func init() {
	if err := ReloadTools(); err != nil {
		fmt.Printf("WARNING: Initial skills load failed: %v\n", err)
		if log != nil {
			log.Error("Initial skills load failed:", "error", err, "component", "skills")
		}
	}
}

func GetTools() map[string]Tool {
	toolsMu.RLock()
	defer toolsMu.RUnlock()
	return toolsCache
}

func GetTool(name string) (Tool, bool) {
	toolsMu.RLock()
	defer toolsMu.RUnlock()
	tool, exists := toolsCache[name]
	return tool, exists
}

func ReloadTools() error {
	file, err := os.Open(filepath.Join("internal", "skills", "tools.json"))
	if err != nil {
		return fmt.Errorf("failed to open tools.json: %w", err)
	}
	defer file.Close()
	var rawMap map[string]Tool
	if err := json.NewDecoder(file).Decode(&rawMap); err != nil {
		return fmt.Errorf("failed to decode tools.json: %w", err)
	}
	newTools := make(map[string]Tool, len(rawMap))
	for key, tool := range rawMap {
		tool.Name = key
		newTools[key] = tool
	}
	toolsMu.Lock()
	toolsCache = newTools
	toolsMu.Unlock()
	if log != nil {
		log.Info("Skills reloaded successfully:", "total_loaded", len(newTools), "component", "skills")
	}
	return nil
}

// UseTool: Aracı çalıştırır ve detaylı bir rapor döner.
// Hata olsa bile Result nesnesi dönmeli ki loglara bakabilelim.
func UseTool(toolName string, args ...string) *ToolResult {
	start := time.Now()

	result := &ToolResult{
		ToolName: toolName,
		Data:     make(map[string]interface{}),
	}

	// 1. Tool var mı?
	tool, exists := GetTool(toolName)
	if !exists {
		result.Error = fmt.Errorf("tool not found: %s", toolName)
		return result
	}

	// 2. Parametre Kontrolü
	if len(args) != len(tool.Parameters) {
		result.Error = fmt.Errorf("parameter mismatch: expected %d, got %d", len(tool.Parameters), len(args))
		return result
	}

	// 3. Yolları Çözümle
	pythonExec, err := filepath.Abs(core.Config.Runtimes.Python)
	if err != nil {
		result.Error = fmt.Errorf("invalid python path: %w", err)
		return result
	}

	scriptPath, err := filepath.Abs(tool.ScriptPath)
	if err != nil {
		result.Error = fmt.Errorf("invalid script path: %w", err)
		return result
	}

	// 4. Zaman Aşımı (Timeout) - 10 Saniye
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 5. Komut Hazırlığı
	cmdArgs := append([]string{scriptPath}, args...)
	cmd := exec.CommandContext(ctx, pythonExec, cmdArgs...)

	// Çıktıları yakalamak için Buffer
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	// 6. ÇALIŞTIR
	log.Info("Using Tool", "toolName", toolName, "args", args, "component", "skills")
	err = cmd.Run()

	// Süreyi kaydet
	result.Duration = time.Since(start)
	result.Stdout = stdoutBuf.String()
	result.Stderr = stderrBuf.String()
	result.ExitCode = cmd.ProcessState.ExitCode()

	// 7. Hata Analizi
	if err != nil {
		// Timeout mu?
		if ctx.Err() == context.DeadlineExceeded {
			result.Error = fmt.Errorf("execution timed out (10s)")
			return result
		}

		// Script hata koduyla mı çıktı? (Exit Code != 0)
		result.Error = fmt.Errorf("execution failed with code %d", result.ExitCode)
		return result // Çıktıyı parse etmeye çalışmıyoruz çünkü patladı.
	}

	// 8. JSON Parse İşlemi (Sadece exit code 0 ise)
	if err := json.Unmarshal(stdoutBuf.Bytes(), &result.Data); err != nil {
		// Python başarılı bitmiş ama JSON basmamış (Örn: Sadece print("bitti") demiş)
		result.Error = fmt.Errorf("invalid json output: %w", err)
		return result
	}

	// 9. Python içinden gelen mantıksal hataları kontrol et ("error" key'i var mı?)
	if errStr, ok := result.Data["error"].(string); ok && errStr != "" {
		// Bu bir Go hatası değil, Tool'un kendi mantık hatasıdır (Örn: Dosya yok)
		// İsteğe bağlı olarak result.Error set edilebilir veya Data içinde bırakılabilir.
		// Biz loglara düşmesi için buraya not düşelim.
		log.Error("Tool logic error", "error", errStr, "component", "skills")
	}

	return result
}
