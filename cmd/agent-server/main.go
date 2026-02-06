// file path: cmd/agent-server/main.go
package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"CHIMERA/cmd"
	"CHIMERA/internal/agent/brain"
	"CHIMERA/internal/core"
	"CHIMERA/internal/server"
)

var log = core.Log

func main() {

	// CORE SİSTEMİ BAŞLAT (Config, Logger, Klasörler)
	// core.Setup() içinde zaten Config yükleniyor ve Logger ayarlanıyor.
	if err := core.SetupCore(); err != nil {
		core.Log.Error("Sistem başlatılamadı!", "error", err)
		os.Exit(1)
	}

	// Brain'i Başlat
	err := brain.NewBrain()
	if err != nil {
		log.Error("❌ Kritik Hata: Brain başlatılamadı!", "error", err)
		os.Exit(1)
	}

	go cmd.TryOllamaConnection()

	// Web Sunucusunu Başlat (Hub ve API dahil)
	server.Start()

	// Block until shutdown signal
	waitForShutdown()
}

func waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	log.Warn("🔻 Kapanış sinyali alındı", "signal", <-quit)

	if err := server.Shutdown(); err != nil {
		log.Error("Shutdown sırasında hata oluştu", "error", err)
	} else {
		log.Info("✅ Uygulama güvenli şekilde kapatıldı.")
	}
	time.Sleep(100 * time.Millisecond)
}

// rasgele zamanlarda tüm bağlılara bildirim gönderici (ctx ile durdurulabilir)
/*
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
*/
