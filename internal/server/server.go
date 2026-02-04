// file path: internal/server/server.go
package server

import (
	"fmt"
	"net/http"
	"time"

	"CHIMERA/scripts"

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
			log.Error("Sunucu başlatılamadı veya durdu!", "error", err, "component", "WEB")
			// Burada os.Exit yapmıyoruz, sadece hatayı logluyoruz.
		}
	}()
	time.Sleep(200 * time.Millisecond)
	// Test the server is up
	resp, err := http.Get(fmt.Sprintf("http://%s/ping", addr))
	if err != nil {
		log.Error("!!! Web server is not responding!", "error", err, "component", "WEB")
	}
	if resp == nil || resp.StatusCode != 200 {
		log.Error("!!! Web server returned non-200 status, server is not healthy", "status_code", resp.StatusCode, "component", "WEB")
	} else {
		log.Info("🌍 Web server is up", "address", "http://"+addr, "component", "WEB")
	}
	StartTime = time.Now()
}

func Shutdown() error {
	fmt.Println("🛑 Sunucu kapatılıyor...")
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
			log.Error(msg, "component", "WEB", "latency", latency.String(), "ip", c.IP())
		} else {
			log.Info(msg, "component", "WEB", "latency", latency.String())
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
	log.Info("Fiber uygulaması başlatıldı", "component", "WEB")
}

func setupRoutes() {
	App.Static("/", "./web/static")
	App.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "uptime": fmt.Sprintf("%.2fs", time.Since(StartTime).Seconds())})
	})

	App.Get("/api/stacks", func(c *fiber.Ctx) error {
		status, err := scripts.Stacks_info()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(status)
	})
	log.Info("Stackts API Yolu Ayarlandı", "component", "WEB")
	time.Sleep(100 * time.Millisecond)

	App.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	App.Get("/ws", websocket.New(NewWebSocketHandler()))
	log.Info("WebSocket Yolu Ayarlandı", "component", "WEB")
	time.Sleep(100 * time.Millisecond)
}
