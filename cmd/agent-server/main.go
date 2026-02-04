// file path: cmd/agent-server/main.go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"CHIMERA/cmd"
	"CHIMERA/internal/agent"
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
	brain, err := agent.NewBrain()
	if err != nil {
		log.Error("❌ Kritik Hata: Brain başlatılamadı!", "error", err)
		os.Exit(1)
	}
	time.Sleep(200 * time.Millisecond)

	cmd.TryOllamaConnection()

	// Web Sunucusunu Başlat (Hub ve API dahil)
	app, err := server.Start(brain)
	if err != nil {
		log.Error("Sunucu başlatılamadı!", "error", err)
		os.Exit(1)
	}
	addr := fmt.Sprintf("%s:%d", core.Config.Server.Host, core.Config.Server.Port)
	errCh := make(chan error, 1)
	go func() {
		if err := app.Listen(addr); err != nil {
			errCh <- err
		}
	}()
	log.Info("Web Arayüz Yayında", "component", "WEB", "url", "http://"+addr)

	// Wait for signals or server error
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	select {
	case sig := <-sigCh:
		log.Warn("Kapanış sinyali alındı", "signal", sig)
		server.HubInstance.Shutdown()
		if err := app.Shutdown(); err != nil {
			log.Error("Shutdown hatası", "error", err)
		}
	case err := <-errCh:
		server.HubInstance.Shutdown()
		log.Error("Sunucu hatası", "error", err)
	}
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
