// file path: internal/server/websocket.go
package server

import (
	"encoding/json"
	"log"
	"time"

	"CHIMERA/internal/models"

	"github.com/gofiber/websocket/v2"
)

// WebSocketHandler: İstemci bağlandığında çalışacak fonksiyon
func WebSocketHandler(c *websocket.Conn) {
	// Bağlantı
	log.Println("🔌 Yeni bir istemci bağlandı:", c.RemoteAddr())
	defer c.Close() // Bağlantı kapandığında temizleme

	// Client oluştur ve Hub'e kaydet
	clientID := c.RemoteAddr().String()
	client := &Client{ID: clientID, Send: make(chan models.Envelope, 16)}
	HubInstance.Register <- client
	defer func() { HubInstance.Unregister <- client }()

	// Writer: yalnızca bu goroutine WriteJSON çağırır
	go func() {
		for env := range client.Send {
			if err := c.WriteJSON(env); err != nil {
				log.Println("Hub'dan client'a gönderim hatası: ", err)
				return
			}
		}
	}()

	// Hoşgeldin mesajı (hedefli)
	HubInstance.SendToClient(clientID, models.Envelope{
		TargetID:  clientID,
		Payload:   models.AgentResponse{Type: models.AgentTypeSystem, Content: "Secure Uplink Established."},
		Timestamp: time.Now(),
	})

	inbound := make(chan models.UserMessage)
	done := make(chan struct{})

	// Okuma ayrı goroutine'da (bloklamayı izole ediyoruz)
	go func() {
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("Bağlantı koptu (reader):", err)
				close(done)
				return
			}
			var userMsg models.UserMessage
			if err := json.Unmarshal(msg, &userMsg); err != nil {
				log.Println("JSON Hatası:", err)
				continue
			}
			inbound <- userMsg
		}
	}()

	// Ana döngü: gelen mesajları işleme / sunucu tarafından gönderim
	for {
		select {
		case <-done:
			return
		case userMsg := <-inbound:
			log.Printf("📩 Gelen Mesaj: %s\n", userMsg.Content)

			// düşünce -> hedefli gönder
			HubInstance.SendToClient(clientID, models.Envelope{
				TargetID:  clientID,
				Payload:   models.AgentResponse{Type: models.AgentTypeThought, Content: "Kullanıcı girdisi analiz ediliyor: " + userMsg.Content},
				Timestamp: time.Now(),
			})
			time.Sleep(1 * time.Second)

			// cevap -> hedefli gönder
			response := "Client is here! Henüz beynim yok ama seni duyabiliyorum. \n Mesajın: " + userMsg.Content[:min(len(userMsg.Content), 50)] + "..."
			HubInstance.SendToClient(clientID, models.Envelope{
				TargetID:  clientID,
				Payload:   models.AgentResponse{Type: models.AgentTypeResponse, Content: response},
				Timestamp: time.Now(),
			})
		}
	}
}


func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
