// file path: internal/server/server.go
package server

import (
	"fmt"
	"time"

	"CHIMERA/internal/agent"
	"CHIMERA/internal/core"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// Start starts the web server, hub and API routes. This function blocks until shutdown signal.
func Start(brain *agent.Brain) (*fiber.App, error) {
	if brain == nil {
		return nil, fmt.Errorf("brain is nil")
	}

	config := core.Config
	log := core.Log

	// Fiber uygulaması
	app := fiber.New(fiber.Config{AppName: config.App.Name})

	// Basit request logging middleware
	app.Use(func(c *fiber.Ctx) error {
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
		app.Use(func(c *fiber.Ctx) error {
			c.Set("Cache-Control", "no-store")
			c.Set("Pragma", "no-cache")
			c.Set("Expires", "0")
			return c.Next()
		})
	}

	// Static files
	app.Static("/", "./web/static")
	log.Info("Static dosya yolu ayarlandı", "component", "WEB")
	time.Sleep(200 * time.Millisecond)

	// WebSocket upgrade path
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws", websocket.New(NewWebSocketHandler(brain)))
	log.Info("WebSocket Yolu Ayarlandı", "component", "WEB")
	time.Sleep(200 * time.Millisecond)

	// API
	setupApiStackts(app)
	log.Info("API Yolları Ayarlandı", "component", "API")
	time.Sleep(200 * time.Millisecond)

	// Run hub
	go HubInstance.Run()
	
	return app,nil
}
