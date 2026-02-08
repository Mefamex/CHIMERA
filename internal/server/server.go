// file path: internal/server/server.go
package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"CHIMERA/internal/common"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// Start starts the web server, hub and API routes. This function blocks until shutdown signal.
func Start() {

	// Fiber uygulaması
	setupFiberApp()

	// Set routes
	setupRoutes()

	// Start the Hub (message center) in the background first
	go HubInstance.Run()

	// Start the Fiber app
	addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
	go func() {
		if err := App.Listen(addr); err != nil {
			log.Error("Sunucu başlatılamadı veya durdu!", err, "component", "WEB")
			// Burada os.Exit yapmıyoruz, sadece hatayı logluyoruz.
		}
	}()
	time.Sleep(50 * time.Millisecond)
	// Test the server is up
	resp, err := http.Get(fmt.Sprintf("http://%s/ping", addr))
	if err != nil {
		log.Error("Web server is not responding!", err, "component", "WEB")
	} else {
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Error("Web server returned non-200 status, server is not healthy", fmt.Errorf("status code: %d", resp.StatusCode), "component", "WEB")
		} else {
			log.System("STARTUP", "🌍 Web server is up", "address", "http://"+addr, "component", "WEB")
		}
	}
	StartTime = time.Now()
}

func Shutdown() error {
	log.System("SHUTDOWN", "🛑 Sunucu kapatılıyor...", "component", "HUB")
	HubInstance.Shutdown()
	return App.Shutdown()
}

func setupFiberApp() {
	App = fiber.New(fiber.Config{AppName: config.App.Name, ReadBufferSize: 8192})

	// Basit request logging middleware
	App.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		stop := time.Now()
		latency := stop.Sub(start)

		if c.Path() == "/ws" {
			return err
		}
		status := c.Response().StatusCode()
		msg := fmt.Sprintf("%s %s -> %d", c.Method(), c.Path(), status)
		if status >= 400 {
			log.Error(msg, fmt.Errorf("status code: %d", status), "component", "WEB", "latency", latency.String(), "ip", c.IP())
		} else if !strings.Contains(msg, " /data/log_2026") {
			log.System("REQUEST", msg, "component", "WEB", "latency", latency.String())
		}
		return err
	})
	if config.App.Environment == "development" {
		App.Use(func(c *fiber.Ctx) error {
			c.Set("Cache-Control", "no-store")
			c.Set("Pragma", "no-cache")
			c.Set("Expires", "0")
			return c.Next()
		})
	}
	log.System("Startup", "Fiber uygulaması başlatıldı", "component", "WEB")
}

func setupRoutes() {
	App.Static("/", "./web/static")
	App.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "uptime": fmt.Sprintf("%.2fs", time.Since(StartTime).Seconds())})
	})

	// Serve log files for the frontend at /data/* using net/http FileServer so
	// directory listing HTML is returned (same behavior as cmd/logViewer/main.go).
	logDir := resolveLogDir()
	fs := http.StripPrefix("/data/", http.FileServer(http.Dir(logDir)))
	App.Use("/data", adaptor.HTTPHandler(fs))

	App.Get("/api/stacks", func(c *fiber.Ctx) error {
		status, err := common.StackInfo.Get()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(status)
	})

	App.Get("/api/model", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"provider": config.LLM.Provider,
			"model":    config.LLM.ModelMain,
		})
	})
	log.System("Startup", "Stackts API Yolu Ayarlandı", "component", "WEB")
	time.Sleep(10 * time.Millisecond)

	App.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	App.Get("/ws", websocket.New(NewWebSocketHandler()))
	log.System("Startup", "WebSocket Yolu Ayarlandı", "component", "WEB")
	time.Sleep(50 * time.Millisecond)
}

func resolveLogDir() string {
	if config != nil && config.Paths.LogsDir != "" {
		return config.Paths.LogsDir
	}
	return "./data/logs"
}
