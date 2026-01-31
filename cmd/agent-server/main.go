package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"

	"CHIMERA/internal/config"
	"CHIMERA/internal/server"
	"CHIMERA/scripts"
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
	app.Get("/ws", websocket.New(server.WebSocketHandler))
	fmt.Println("✅ WebSocket Yolu Ayarlandı.")

	// 5. API Yolları
	// STACKS_INFO
	app.Get("/api/stacks", func(c *fiber.Ctx) error {
		status, err := scripts.Stacks_info()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(status)
	})
	fmt.Println("✅ API Yolları Ayarlandı.")

	// 6. Sunucuyu Başlat
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("🌍 Arayüz Yayında: http://%s\n", addr)

	errCh := make(chan error, 1)
	go func() {
		if err := app.Listen(addr); err != nil {
			errCh <- err
		}
	}()

	// Ctrl+C veya SIGTERM bekle
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		fmt.Println("\n\nAlındı:", sig, "- sunucu kapatılıyor...")
		if err := app.Shutdown(); err != nil {
			log.Printf("Shutdown hatası: %v", err)
		}
	case err := <-errCh:
		log.Fatalf("\n\nSunucu hatası: %v", err)
	}
}

/*
func sendStacksInfo() {
	time.Sleep(2 * time.Second)
	status, err := scripts.Stacks_info()
	if err != nil {
		log.Printf("stacks_info error: %v", err)
		return
	}
	server.payload, _ := json.Marshal(status)
	// server.Broadcast fonksiyonunun mevcut olduğunu varsayıyoruz
	// yoksa internal/server içinde Hub/Broadcast implementasyonu ekleyin

}
*/
