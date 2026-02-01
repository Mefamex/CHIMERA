// file path: cmd/agent-server/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"

	"CHIMERA/internal/config"
	"CHIMERA/internal/models"
	"CHIMERA/internal/server"
	"CHIMERA/scripts"
)

func main() {
	// Config Yükle
	cfg, err := config.LoadConfig("./configs/config.yaml")
	if err != nil {
		log.Fatalf("❌ Kritik Hata: Config yüklenemedi! \n%v", err)
	}

	// Bilgi Yazdır
	fmt.Println("🚀 CHIMERA AGENT BAŞLATILIYOR...")
	fmt.Printf("✅ Konfigürasyon: %s (v%s) | Port: %d\n", cfg.App.Name, cfg.App.Version, cfg.Server.Port)

	// Fiber Uygulamasını Oluştur
	app := fiber.New(fiber.Config{
		AppName: cfg.App.Name,
	})

	// Loglama Middleware'i (Terminalde HTTP isteklerini görmek için)
	app.Use(logger.New())

	// Statik Dosyalar (HTML/CSS/JS)
	app.Static("/", "./web/static")

	// WebSocket Yolu
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws", websocket.New(server.WebSocketHandler))
	fmt.Println("✅ WebSocket Yolu Ayarlandı.")

	// API Yolları
	app.Get("/api/stacks", func(c *fiber.Ctx) error {
		status, err := scripts.Stacks_info()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(status)
	})
	fmt.Println("✅ API Yolları Ayarlandı.")

	// Context ile iptal yönetimi (randomNotifications durdurmak için)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Hub'i çalıştır
	go server.HubInstance.Run()

	// Rasgele bildirim göndericiyi başlat (kontrollü, ctx ile durdurulabilir)
	go randomNotifications(ctx)

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
		// rastgele bildirim goroutine'ini durdur
		cancel()
		// hub'i kapat
		server.HubInstance.Shutdown()
		if err := app.Shutdown(); err != nil {
			log.Printf("Shutdown hatası: %v", err)
		}
	case err := <-errCh:
		// sunucu hatası durumunda önce temizleme
		cancel()
		server.HubInstance.Shutdown()
		log.Fatalf("\n\nSunucu hatası: %v", err)
	}
}

// rasgele zamanlarda tüm bağlılara bildirim gönderici (ctx ile durdurulabilir)
func randomNotifications(ctx context.Context) {
	rand.Seed(time.Now().UnixNano())
	for {
		// 5-25 saniye arası rastgele bekle veya iptal olmasını kontrol et
		sleep := time.Duration(5+rand.Intn(20)) * time.Second
		select {
		case <-time.After(sleep):
			env := models.Envelope{
				Broadcast: true,
				Payload: models.AgentResponse{
					Type:    models.AgentTypeSystem,
					Content: "Sunucudan rastgele bildirim.",
				},
				Timestamp: time.Now(),
			}
			// non-blocking: hub.Run okuyorsa gönder, yoksa düşsün
			select {
			case server.HubInstance.Broadcast <- env:
			default:
			}
		case <-ctx.Done():
			return
		}
	}
}
