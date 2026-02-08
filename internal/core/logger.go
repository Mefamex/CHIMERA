// file path : internal/core/logger.go
/*
Durum / Senaryo,Fonksiyon Adı,Parametreler (Özet),Hedef Dosya,Örnek Kullanım
"🧠 Beyin Düşünüyor(Planlama, Karar)",TraceBrain,"Phase, Thought, Args...",trace_brain_user.jsonl,"core.GetLogger().TraceBrain(""Planning"", ""Dosya aranıyor..."", ""file"", ""main.go"")"
🤖 LLM ile Konuşma(Ham Prompt/Cevap),TraceLLM,"Model, Duration, Prompt, Response",trace_llmbot.jsonl,"core.GetLogger().TraceLLM(""llama3"", 1200, promptStr, responseStr)"
💬 Sohbet (Chat)(User veya Agent),Chat,"Sender, Msg, IsFinal",chat.jsonl,"core.GetLogger().Chat(""User"", ""React kur"", false)"
"🛠️ Araç Çalıştı(Cmd, File Ops)",Tool,"Name, Args, Output, Success, Duration",tools.jsonl,"core.GetLogger().Tool(""runCmd"", ""ls -la"", ""total 4..."", true, 45)"
🚨 HATA (Panic)(Go tarafında),Error,"Msg, Err, Args...",errors.jsonlsystem.jsonl,"core.GetLogger().Error(""Dosya yok"", err, ""path"", path)"
🐍 Python/Node Çıktısı(Script logları),SubProcess,"Source, Msg, IsError",internal_python.jsonlinternal_node.jsonl,"core.GetLogger().SubProcess(""Python"", ""Analiz bitti"", false)"
⚙️ Sistem Olayı(Start/Stop),System,"Event, Msg, Args...",system.jsonl,"core.GetLogger().System(""Startup"", ""Server Başladı"", ""port"", 8080)"
📡 Alt Servis Çağrısı(Delegation),Internal,"Target, Cmd, Success",internals.jsonl,"core.GetLogger().Internal(""PythonService"", ""analyze"", true)"*/

package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// -- GLOBAL LOGGER --
var (
	Logger *ChimeraLogger
	once   sync.Once
)

// -- LOG TİPLERİ (Dosya Adları ile Eşleşir) --
type LogType string

const (
	LogSystem       LogType = "system"
	LogErrors       LogType = "errors"
	LogChain        LogType = "chain"
	LogChat         LogType = "chat"
	LogTraceBrain   LogType = "trace_brain"
	LogTraceLLM     LogType = "trace_llmbot"
	LogTools        LogType = "tools"
	LogInternals    LogType = "internals"
	LogInternalPy   LogType = "internal_python"
	LogInternalNode LogType = "internal_node"
)

type LogSink func(logType LogType, data map[string]any)

var (
	logSinksMu sync.RWMutex
	logSinks   []LogSink
)

func RegisterLogSink(sink LogSink) {
	if sink == nil {
		return
	}
	logSinksMu.Lock()
	logSinks = append(logSinks, sink)
	logSinksMu.Unlock()
}

// -- INIT (Otomatik) --
func init() {
	// empty init for not blocking GetLogger calls before InitLogger is explicitly called.
	Logger = &ChimeraLogger{
		consoleOutput: true,
		initialized:   false,
	}
}

// -- LOGGER YAPISI --
type ChimeraLogger struct {
	mu            sync.Mutex
	logDir        string
	files         map[LogType]*os.File
	consoleOutput bool
	initialized   bool
}

// GetLogger: Global logger'ı döndürür. Eğer başlatılmadıysa başlatır.
func GetLogger() *ChimeraLogger {
	if !Logger.initialized {
		InitLogger()
	}
	return Logger
}

// InitLogger: Klasörleri oluşturur ve dosya bağlantılarını açar.
// sync.Once kullanılarak sadece 1 kere çalışması garanti altına alınır.
func InitLogger() *ChimeraLogger {
	once.Do(func() {
		baseDir := Config.Paths.LogsDir
		if baseDir == "" {
			fmt.Printf("\nLog klasoru belirtilmedi. Config dosyasindaki 'paths.logsdir' alanini kontrol edin.\n")
			os.Exit(1)
		}
		baseDir = filepath.Clean(baseDir)
		if !filepath.IsAbs(baseDir) {
			if absDir, err := filepath.Abs(baseDir); err == nil {
				baseDir = absDir
			}
		}

		sessionName := fmt.Sprintf("log_%s", time.Now().Format("2006_01_02_15_04_05"))
		sessionDir := filepath.Join(baseDir, sessionName)

		if err := os.MkdirAll(sessionDir, 0755); err != nil {
			fmt.Printf("\n KRİTİK HATA: Log klasörü oluşturulamadı: %v\n", err)
			os.Exit(1)
		}

		filesMap := make(map[LogType]*os.File)
		logTypes := []LogType{
			LogSystem, LogErrors, LogChain, LogChat,
			LogTraceBrain, LogTraceLLM, LogTools,
			LogInternals, LogInternalPy, LogInternalNode,
		}

		for _, t := range logTypes {
			// Dosya adı: logs/log_2026.../system.jsonl
			filename := filepath.Join(sessionDir, string(t)+".jsonl")
			f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Printf("\n[LOGGER] UYARI: Log dosyası açılamadı (%s): %v\n", t, err)
				continue
			}
			filesMap[t] = f
		}

		Logger = &ChimeraLogger{logDir: sessionDir, files: filesMap, consoleOutput: true, initialized: true}

		Logger.System("Startup", "✅ Logger sistemi başarıyla başlatıldı", map[string]any{
			"path": sessionDir,
		})
	})
	return Logger
}

// -- CORE WRITE FUNCTION --
// Tüm helper fonksiyonlar burayı kullanır.
func (l *ChimeraLogger) write(logType LogType, data map[string]any) {
	const (
		ColorReset  = "\033[0m"
		ColorRed    = "\033[31m"
		ColorGreen  = "\033[32m"
		ColorYellow = "\033[33m"
		ColorBlue   = "\033[34m"
		ColorCyan   = "\033[36m"
		ColorGray   = "\033[90m"
	)
	l.mu.Lock()
	defer l.mu.Unlock()

	// Eğer dosya sistemi başlatılmadıysa (InitLogger çağrılmadıysa)
	if !l.initialized {
		if l.consoleOutput {
			fmt.Printf("[PRE-INIT] %v\n", data)
		}
		return
	}

	// 1. Zorunlu Alanlar
	if _, ok := data["timestamp"]; !ok {
		data["timestamp"] = time.Now().Format("2006-01-02T15:04:05.000")
	}
	level, _ := data["level"].(string)
	if level == "" {
		level = "INFO"
		data["level"] = level
	}

	// 2. JSONL Serialization
	bytes, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Log Marshal Error: %v\n", err)
		return
	}
	line := append(bytes, '\n')

	// 3. Dosyaya Yaz
	if f, ok := l.files[logType]; ok {
		if _, err := f.Write(line); err != nil {
			fmt.Printf("Dosyaya yazılamadı (%s): %v\n", logType, err)
		}
	} else {
		fmt.Printf("MISSING LOG FILE (%s): %s", logType, string(line))
	}

	l.emitToSinks(logType, data)

	if l.consoleOutput && (logType == LogSystem || logType == LogErrors || logType == LogChat || logType == LogTools) {

		color := ColorGreen
		if level == "ERROR" {
			color = ColorRed
		} else if level == "WARN" {
			color = ColorYellow
		} else if logType == LogChat {
			color = ColorCyan
		}

		timeStr := data["timestamp"].(string)
		if len(timeStr) > 11 {
			timeStr = timeStr[11:19]
		}

		comp, _ := data["component"].(string)
		if len(comp) > 7 {
			comp = comp[:7]
		} else {
			comp = fmt.Sprintf("%-7s", comp)
		}

		msg, _ := data["msg"].(string)

		extra := ""
		for k, v := range data {
			if k == "timestamp" || k == "level" || k == "component" || k == "msg" {
				continue
			} else if k == "event" {
				comp += fmt.Sprintf("| %v", v)
			} else {
				extra += fmt.Sprintf("%s=%v ", k, v)
			}
		}

		fmt.Printf("%s%s%s %s%s%s %s[%s]%s %s %s%s%s\n",
			ColorGray, timeStr, ColorReset,
			color, fmt.Sprintf("%-5s", level), ColorReset,
			ColorBlue, comp, ColorReset,
			msg,
			ColorGray, extra, ColorReset,
		)
	}
}

func (l *ChimeraLogger) emitToSinks(logType LogType, data map[string]any) {
	logSinksMu.RLock()
	if len(logSinks) == 0 {
		logSinksMu.RUnlock()
		return
	}
	sinks := make([]LogSink, len(logSinks))
	copy(sinks, logSinks)
	logSinksMu.RUnlock()

	for _, sink := range sinks {
		payload := cloneMap(data)
		s := sink
		go func() {
			defer func() {
				_ = recover()
			}()
			s(logType, payload)
		}()
	}
}

// ============================================================================
// 🛠️ HELPER FONKSİYONLAR (Kullanacağın Metodlar)
// ============================================================================

// 1. SYSTEM: Altyapı olayları
func (l *ChimeraLogger) System(event, msg string, args ...any) {
	data := map[string]any{
		"component": "System",
		"event":     event,
		"msg":       msg,
	}
	mergeMeta(data, argsToMap(args...))
	l.write(LogSystem, data)
}

// 2. ERROR: Hata yönetimi (Stack trace içerir)
func (l *ChimeraLogger) Error(msg string, err error, args ...any) {
	_, file, line, _ := runtime.Caller(1)

	errStr := ""
	if err != nil {
		errStr = err.Error()
	}

	data := map[string]any{
		"level":     "ERROR",
		"component": "Brain",
		"msg":       msg,
		"error_msg": errStr,
		"culprit":   fmt.Sprintf("%s:%d", filepath.Base(file), line),
	}
	mergeMeta(data, argsToMap(args...))

	l.write(LogErrors, data)
	
	l.write(LogSystem, map[string]any{
		"level":     "ERROR",
		"event":     "Exception",
		"msg":       msg,
		"error_msg": errStr,
	})
}

// 3. CHAIN: Akış diyagramı için
func (l *ChimeraLogger) Chain(from, to, action, status string) {
	l.write(LogChain, map[string]any{
		"component": "Flow",
		"from":      from,
		"to":        to,
		"action":    action,
		"status":    status,
	})
}

// 4. CHAT: Kullanıcı ile konuşma
func (l *ChimeraLogger) Chat(sender, msg string, isFinal bool) {
	l.write(LogChat, map[string]any{
		"component": "ChatInterface",
		"sender":    sender, // "User" veya "Agent"
		"msg":       msg,
		"is_final":  isFinal,
	})
}

// 5. TRACE BRAIN: Düşünce süreci
func (l *ChimeraLogger) TraceBrain(phase, thought string, args ...any) {
	l.write(LogTraceBrain, map[string]any{
		"component": "Brain",
		"phase":     phase, // Planning, Reflection, Execution
		"msg":       thought,
		"context":   argsToMap(args...), // Array veya Map olabilir
	})
}

// 6. TRACE LLM: Ham model verisi
func (l *ChimeraLogger) TraceLLM(model string, durationMs int64, prompt, response string) {
	l.write(LogTraceLLM, map[string]any{
		"component":    "OllamaClient",
		"model":        model,
		"duration_ms":  durationMs,
		"raw_prompt":   prompt,
		"raw_response": response,
	})
}

// 7. TOOLS: Araç kullanımları
func (l *ChimeraLogger) Tool(name string, args any, success bool, durationMs int64 , output string) {
	l.write(LogTools, map[string]any{
		"component":   "ToolExecutor",
		"tool_name":   name,
		"args":        args,
		"success":     success,
		"duration_ms": durationMs,
		"output":      output,
	})
}

// 8. INTERNALS: Orkestrasyon emirleri
func (l *ChimeraLogger) Internal(target, cmd string, success bool) {
	l.write(LogInternals, map[string]any{
		"component": "Orchestrator",
		"target":    target, // "Python" veya "Node"
		"cmd":       cmd,
		"success":   success,
	})
}

// 9. SUB PROCESS: Python ve Node çıktıları
func (l *ChimeraLogger) SubProcess(source string, msg string, isError bool) {
	level := "INFO"
	if isError {
		level = "ERROR"
	}

	data := map[string]any{
		"level":     level,
		"component": source, // "PythonService" veya "NodeService"
		"msg":       msg,
	}

	// Kaynağa göre doğru dosyaya yönlendir
	if source == "Node" || source == "NodeService" {
		l.write(LogInternalNode, data)
	} else {
		// Varsayılan Python
		l.write(LogInternalPy, data)
	}
}

// -- Close: Uygulama kapanırken dosyaları kapatır --
func (l *ChimeraLogger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.initialized {
		return
	}

	for _, f := range l.files {
		f.Close()
	}
	l.initialized = false
}

// -- Helper: Map birleştirme --
func mergeMeta(dest, src map[string]any) {
	if src == nil {
		return
	}
	for k, v := range src {
		dest[k] = v
	}
}

func argsToMap(args ...any) map[string]any {
	if len(args) == 0 {
		return nil
	}

	m := make(map[string]any)
	for i := 0; i < len(args); i += 2 {
		// Key mutlaka string olmalı
		key, ok := args[i].(string)
		if !ok {
			// Eğer key string değilse atla veya logla
			continue
		}

		// Value var mı kontrol et
		var val any = "MISSING_VALUE"
		if i+1 < len(args) {
			val = args[i+1]
		}

		m[key] = val
	}
	return m
}

func cloneMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
