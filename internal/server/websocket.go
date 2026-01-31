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

	// Hoşgeldin sinyali gönder
	sendJSON(c, "system", "Secure Uplink Established.", "")

	var (
		msg []byte
		err error
	)

	// 2. Sonsuz Döngü (Dinleme Modu)
	for {
		// Mesaj bekle (Bloklayan satır)
		if _, msg, err = c.ReadMessage(); err != nil {
			log.Println("Bağlantı koptu:", err)
			break
		}

		// Gelen JSON'ı Parse et
		var userMsg models.UserMessage
		if err := json.Unmarshal(msg, &userMsg); err != nil {
			log.Println("JSON Hatası:", err)
			continue
		}

		log.Printf("📩 Gelen Mesaj: %s\n", userMsg.Content)

		// --- BURASI GEÇİCİ BEYİN SİMÜLASYONU ---
		// Gerçek Agent entegre edilene kadar burası "Rol Yapacak"

		// A. Düşünce Gönder (Frontend sağ panele düşer)
		sendJSON(c, "thought", "Kullanıcı girdisi analiz ediliyor: "+userMsg.Content, "")
		time.Sleep(1 * time.Second) // Yapay gecikme (Gerçekçilik için)

		// B. Tool Kullanımı Simülasyonu
		if len(userMsg.Content) > 5 {
			sendJSON(c, "tool", "Kelime sayısı kontrol ediliyor...", "word_counter")
			time.Sleep(500 * time.Millisecond)
		}

		// C. Cevap Gönder (Frontend sol panele düşer)
		response := "Anlaşıldı. Henüz tam beynim (Ollama) bağlı değil ama seni duyabiliyorum Mefamex. \n Mesajın: " + userMsg.Content
		sendJSON(c, "response", response, "")
		// ----------------------------------------
	}
}

// Yardımcı Fonksiyon: JSON göndermeyi kolaylaştırır
func sendJSON(c *websocket.Conn, msgType, content, toolName string) {
	resp := models.AgentResponse{
		Type:     msgType,
		Content:  content,
		ToolName: toolName,
	}
	// JSON'a çevir ve gönder
	if err := c.WriteJSON(resp); err != nil {
		log.Println("Gönderme hatası:", err)
	}
}
