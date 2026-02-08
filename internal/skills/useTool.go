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
	"strings"
	"sync"
	"time"
)

var toolsCache = make(map[string]Tool)
var toolsMu sync.RWMutex

type Tool struct {
	Name        string            `json:"-"`
	Category    string            `json:"Category"`
	ScriptPath  string            `json:"Path"`
	Description string            `json:"Desc"`
	Parameters  []string          `json:"Params"`
	TimeoutSec  int               `json:"TimeoutSec,omitempty"`
	Returns     map[string]string `json:"Returns"`
}
type ToolResult struct {
	ToolName string                 `json:"tool_name"`
	Data     map[string]interface{} `json:"data"`      // Python'dan gelen asıl JSON verisi
	Stdout   string                 `json:"stdout"`    // Ham standart çıktı
	Stderr   string                 `json:"stderr"`    // Hata logları
	Duration time.Duration          `json:"duration"`  // Çalışma süresi
	ExitCode int                    `json:"exit_code"` // 0: Başarılı, Diğerleri: Hata
	Error    error                  `json:"-"`         // Go tarafındaki çalıştırma hatası
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

// ReloadTools: tools.json dosyasını okur ve belleğe yükler.
func ReloadTools() error {
	// tools.json yolunu mutlak yola çevirmeye çalışalım.
	// Varsayım: Program proje kök dizininden çalışıyor.
	basePath := "internal/skills/tools.json"

	absPath, err := filepath.Abs(basePath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path for tools.json: %w", err)
	}

	file, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("failed to open tools.json at %s: %w", absPath, err)
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

	return nil
}

// UseTool: Aracı çalıştırır ve detaylı bir rapor döner.
func UseTool(toolName string, args ...string) *ToolResult {
	start := time.Now()
	log := core.GetLogger()

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
		result.Error = fmt.Errorf("parameter mismatch: expected %d (%v), got %d", len(tool.Parameters), tool.Parameters, len(args))
		return result
	}

	// 3. Yolları Çözümle
	if core.Config.Runtimes.Python == "" {
		result.Error = fmt.Errorf("python runtime path is not configured")
		return result
	}
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

	// 4. Zaman Aşımı (Timeout)
	// Varsayılan 30 saniye; tool JSON içinde TimeoutSec verilirse onu kullan.
	timeout := 30 * time.Second
	if tool.TimeoutSec > 0 {
		timeout = time.Duration(tool.TimeoutSec) * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 5. Komut Hazırlığı
	cmdArgs := append([]string{scriptPath}, args...)
	cmd := exec.CommandContext(ctx, pythonExec, cmdArgs...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	// 6. ÇALIŞTIR
	log.Tool(toolName, strings.Join(args, " "), false, 0, "calling")

	err = cmd.Run()

	result.Duration = time.Since(start)
	result.Stdout = stdoutBuf.String()
	result.Stderr = stderrBuf.String()
	result.ExitCode = cmd.ProcessState.ExitCode()

	// 7. Hata Analizi
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			result.Error = fmt.Errorf("execution timed out (30s)")
		} else {
			result.Error = fmt.Errorf("execution failed with code %d: %s", result.ExitCode, result.Stderr)
		}
		// Hata olsa bile JSON dönmüş olabilir (Python tarafında handle edildiyse)
		// O yüzden aşağıda unmarshal deniyoruz.
	}

	// 8. JSON Parse İşlemi
	// Eğer stdout boşsa ve hata varsa, zaten yukarıda error set ettik.
	if stdoutBuf.Len() > 0 {
		if jsonErr := json.Unmarshal(stdoutBuf.Bytes(), &result.Data); jsonErr != nil {
			// Sadece eğer execution hatası yoksa bunu hata sayalım.
			// Execution hatası varsa zaten asıl sorun odur.
			if result.Error == nil {
				result.Error = fmt.Errorf("invalid json output from tool: %w | Raw: %s", jsonErr, stdoutBuf.String())
			}
		}
	} else if result.Error == nil {
		// Başarılı (exit 0) ama boş çıktı?
		result.Error = fmt.Errorf("tool returned empty output")
	}

	// 9. Mantıksal Hata Kontrolü ("error" key'i)
	if errStr, ok := result.Data["error"].(string); ok && errStr != "" {
		if log != nil {
			log.Error("Tool returned logic error", fmt.Errorf("%s", errStr), "tool", toolName, "component", "skills")
		}
		// Go hatası olarak set etmiyoruz, çünkü tool başarıyla çalıştı ve hatayı raporladı.
		// Ancak çağırana kolaylık olsun diye bir flag eklenebilir.
	}
	log.Tool(toolName, args, result.Error == nil, result.Duration.Milliseconds(), result.Stdout)
	return result
}

func GetToolsPrompt() string {
	toolsMu.RLock()
	defer toolsMu.RUnlock()

	if len(toolsCache) == 0 {
		return "No tools available. (Check tools.json)"
	}

	var builder bytes.Buffer
	builder.WriteString("Mevcut Araçlar (Tools):\n")

	for _, tool := range toolsCache {
		paramsJSON, _ := json.Marshal(tool.Parameters)
		line := fmt.Sprintf("- %s (Params: %s): %s\n", tool.Name, paramsJSON, tool.Description)
		builder.WriteString(line)
	}

	return builder.String()
}
