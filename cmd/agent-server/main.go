package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"

	"CHIMERA/internal/config"
	"CHIMERA/internal/server"
)

func main() {
	// 1. Config Yükle
	cfg, err := config.LoadConfig("./configs/config.yaml")
	if err != nil {
		log.Fatalf("❌ Kritik Hata: Config yüklenemedi! \n%v", err)
	}

	// Bilgi Yazdır
	fmt.Println("🚀 CHIMERA AGENT BAŞLATILIYOR...")
	fmt.Printf("✅ Konfigürasyon: %s (v%s) | Port: %d\n", cfg.App.Name, cfg.App.Version, cfg.Server.Port)

	// 2. Fiber Uygulamasını Oluştur
	app := fiber.New(fiber.Config{
		AppName: cfg.App.Name,
	})

	// Loglama Middleware'i (Terminalde HTTP isteklerini görmek için)
	app.Use(logger.New())

	// 3. Statik Dosyalar (HTML/CSS/JS)
	// Kullanıcı "/" adresine girdiğinde "web/static" klasörünü görsün
	app.Static("/", "./web/static")

	// 4. WebSocket Yolu
	// Önce protokol kontrolü (Upgrade)
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// Sonra Handler'ı bağla
	app.Get("/ws", websocket.New(server.WebSocketHandler))

	// 5. Sunucuyu Başlat
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("🌍 Arayüz Yayında: http://%s\n", addr)
	
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Sunucu başlatılamadı: %v", err)
	}
}