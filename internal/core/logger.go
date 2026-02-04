// file path: internal/core/logger.go
package core

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// -- GLOBAL LOGGER --
var Log *slog.Logger

// -- INIT --
func init() {
	// İlk açılışta basit başlat
	Log = slog.New(NewDevHandler(os.Stdout))
}

// -- YAPILANDIRMA --
func InitLogger() *slog.Logger {
	// Dosya Handler'ı (JSON Formatında, "logs" klasörüne)
	logDir := Config.Paths.LogsDir
	if logDir == "" {
		logDir = "./data/logs"
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(fmt.Sprintf("Log klasörü oluşturulamadı: %v", err))
	}
	now := time.Now()
	timestamp := now.Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("chimera-%s.json", timestamp)
	fullPath := filepath.Join(logDir, filename)

	file, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Sprintf("Log dosyası açılamadı: %v", err))
	}

	// Dosya Handler (JSON)
	jsonHandler := slog.NewJSONHandler(file, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	// Terminal Handler (Renkli / Development)
	var consoleHandler slog.Handler
	if Config.App.Environment == "development" {
		consoleHandler = NewDevHandler(os.Stdout)
	} else {
		consoleHandler = slog.NewJSONHandler(os.Stdout, nil)
	}

	// Çoklayıcı (Tee) İkisini Birleştir (Tee / Fanout)
	// Gelen logu hem JSON dosyasına hem de Terminale gönderir.
	multiHandler := &TeeHandler{
		handlers: []slog.Handler{consoleHandler, jsonHandler},
	}

	// Global Logger'ı güncelle
	Log = slog.New(multiHandler)
	Log.Info("Logger sistemi hazır", "env", Config.App.Environment, "path", fullPath)
	return Log
}

//  CUSTOM DEV HANDLER

type DevHandler struct {
	w  io.Writer
	mu sync.Mutex
}

func NewDevHandler(w io.Writer) *DevHandler {
	return &DevHandler{w: w}
}

func (h *DevHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= slog.LevelDebug // Debug ve üstünü göster
}

func (h *DevHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 1. ZAMAN FORMATI (2026/02/04 01:56:13)
	timeStr := r.Time.Format("2006/01/02 15:04:05")

	// 2. RENKLER VE LEVEL
	levelStr := r.Level.String()
	colorCode := "\033[37m" // Beyaz (Default)

	switch r.Level {
	case slog.LevelDebug:
		levelStr = "DEBG"
		colorCode = "\033[36m" // Cyan
	case slog.LevelInfo:
		levelStr = "INFO"
		colorCode = "\033[32m" // Yeşil
	case slog.LevelWarn:
		levelStr = "WARN"
		colorCode = "\033[33m" // Sarı
	case slog.LevelError:
		levelStr = "ERRO"
		colorCode = "\033[31m" // Kırmızı
	}

	// 3. COMPONENT AYIKLAMA ([WEB   ] Kısmı)
	// Log atarken: Log.Info("Mesaj", "component", "WEB") dersek buraya düşer.
	component := "SYS " // Varsayılan
	var attrs []string

	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "component" {
			component = a.Value.String()
			return true
		}
		// Diğer attribute'ları (key=value) sakla
		attrs = append(attrs, fmt.Sprintf("%s=%v", a.Key, a.Value.Any()))
		return true
	})

	// Component'i 6 karaktere tamamla (Hizalama için)
	if len(component) < 6 {
		component = component + strings.Repeat(" ", 6-len(component))
	} else if len(component) > 6 {
		component = component[:6]
	}

	// 4. ÇIKTIYI BİRLEŞTİR
	// Format: ZAMAN [COMP  ] MESAJ ATTR

	// Attribute'lar varsa gri ekle
	attrStr := ""
	if len(attrs) > 0 {
		attrStr = fmt.Sprintf(" \033[90m%s\033[0m", strings.Join(attrs, " "))
	}

	// Ekrana bas
	fmt.Fprintf(h.w, "%s %s %s %s %s\n",
		"\033[90m"+timeStr+"\033[0m",                  // Tarih Gri
		colorCode+levelStr+"\033[0m",                  // Level Renkli
		fmt.Sprintf("\033[34m[%s]\033[0m", component), // WEB   Mavi
		r.Message, // Mesaj
		attrStr,   // key=value (Gri)
	)

	return nil
}

func (h *DevHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h } // Basitleştirildi
func (h *DevHandler) WithGroup(name string) slog.Handler       { return h } // Basitleştirildi

// TEE HANDLER (Çoklayıcı)

type TeeHandler struct {
	handlers []slog.Handler
}

func (t *TeeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range t.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (t *TeeHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range t.handlers {
		// Her handler kendi işini yapsın (Biri ekrana, biri dosyaya)
		_ = h.Handle(ctx, r)
	}
	return nil
}

func (t *TeeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Bu kısım biraz complex, şimdilik geçtimm
	return t
}

func (t *TeeHandler) WithGroup(name string) slog.Handler {
	return t
}
