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
	// 1. Bağlantı Başarılı Logu
	log.Println("🔌 Yeni bir istemci bağlandı:", c.RemoteAddr())
	defer c.Close()

	// Hoşgeldin sinyali gönder
	sendJSON(c, "system", "Secure Uplink Established.", "")

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

			// --- BURASI GEÇİCİ BEYİN SİMÜLASYONU ---
			sendJSON(c, "thought", "Kullanıcı girdisi analiz ediliyor: "+userMsg.Content, "")
			time.Sleep(1 * time.Second)

			response := "Anlaşıldı. Henüz tam beynim (Ollama) bağlı değil ama seni duyabiliyorum Mefamex. \n Mesajın: " + userMsg.Content[:min(len(userMsg.Content), 50)] + "..."
			sendJSON(c, "response", response, "")
			// ----------------------------------------
		}
	}
}

// Yardımcı Fonksiyon: JSON göndermeyi kolaylaştırır
func sendJSON(c *websocket.Conn, msgType, content, toolName string) {
	resp := models.AgentResponse{
		Type:     msgType,
		Content:  content,
		ToolName: toolName,
	}
	if err := c.WriteJSON(resp); err != nil {
		log.Println("Gönderme hatası:", err)
	}
}
