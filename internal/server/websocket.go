// file path: internal/server/websocket.go
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"CHIMERA/internal/agent"
	"CHIMERA/internal/models"

	"github.com/google/uuid"
	"github.com/ollama/ollama/api"

	"github.com/gofiber/websocket/v2"
)

// NewWebSocketHandler: Websocket bağlantısını yöneten ana handler
func NewWebSocketHandler() func(c *websocket.Conn) {
	eng := agent.NewEngine(&agent.Brain)

	return func(c *websocket.Conn) {
		// 1. Bağlantı Kurulumu
		clientID := c.RemoteAddr().String()
		log.System("Connection", "🔌 Yeni İstemci Bağlandı", "component", "WS", "id", clientID)
		defer func() {
			log.System("Connection", "🔌 Bağlantı Kesildi", "component", "WS", "id", clientID)
			c.Close()
		}()

		// Hub'a kaydol
		client := &Client{ID: clientID, Send: make(chan models.Envelope, 256)}
		HubInstance.Register <- client
		defer func() { HubInstance.Unregister <- client }()

		// Yazma (Write) Goroutine'i
		go func() {
			for env := range client.Send {
				if err := c.WriteJSON(env); err != nil {
					return
				}
			}
		}()

		// Hoşgeldin Mesajı
		sendSystemNotification(clientID, "Secure Uplink Established. System Ready.")

		// Her bağlantı için özel hafıza (History)
		var conversation []api.Message

		// Okuma (Read) Goroutine'i
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				break
			}

			// Gelen JSON'ı UserMessage modeline çevir
			var userMsg models.UserMessage
			if err := json.Unmarshal(msg, &userMsg); err != nil {
				log.Error("JSON Parse Hatası", err)
				continue
			}

			if userMsg.Content == "" {
				continue
			}

			// Ana İşlem Başlatılıyor
			log.System("Message", "📩 Mesaj Alındı", "content", userMsg.Content)
			msgID := uuid.NewString()

			sendStartSignal(clientID, msgID)

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)

			updatedHistory, runErr := eng.Run(ctx, conversation, userMsg.Content, func(ev agent.Event) {
				switch ev.Type {
				case agent.EventThought:
					sendThought(clientID, msgID, ev.Content)
				case agent.EventReply:
					sendReplyDelta(clientID, msgID, ev.Content)
				case agent.EventToolCall:
					sendToolCallInfo(clientID, msgID, ev.ToolName, ev.Args)
				case agent.EventToolResult:
					sendToolResult(clientID, msgID, ev.ToolName, ev.Content)
				case agent.EventError:
					sendError(clientID, msgID, ev.Error)
				default:
					// ignore
				}
			})
			conversation = updatedHistory
			_ = runErr

			cancel()
			sendEndSignal(clientID, msgID)
		}
	}
}

// --- YARDIMCI GÖNDERİM FONKSİYONLARI ---

func sendSystemNotification(clientID, text string) {
	HubInstance.SendToClient(clientID, models.Envelope{
		TargetID:  clientID,
		Timestamp: time.Now(),
		Payload:   models.AgentResponse{Type: models.AgentTypeSystem, Content: text},
	})
}

func sendStartSignal(clientID, msgID string) {
	HubInstance.SendToClient(clientID, models.Envelope{
		ID:        msgID,
		TargetID:  clientID,
		Timestamp: time.Now(),
		Payload:   models.AgentResponse{Type: models.AgentTypeResponseStart},
	})
}

func sendEndSignal(clientID, msgID string) {
	HubInstance.SendToClient(clientID, models.Envelope{
		ID:        msgID,
		TargetID:  clientID,
		Timestamp: time.Now(),
		Payload:   models.AgentResponse{Type: models.AgentTypeResponseEnd},
	})
}

func sendReplyDelta(clientID, msgID, text string) {
	// Arayüz streaming beklediği için tam metni delta olarak gönderiyoruz
	HubInstance.SendToClient(clientID, models.Envelope{
		ID:        msgID,
		TargetID:  clientID,
		Timestamp: time.Now(),
		Payload:   models.AgentResponse{Type: models.AgentTypeResponseDelta, Content: text},
	})
}

func sendThought(clientID, msgID, thought string) {
	// Arayüzde "Düşünüyor..." veya gri yazı olarak göstermek için
	HubInstance.SendToClient(clientID, models.Envelope{
		ID:        msgID,
		TargetID:  clientID,
		Timestamp: time.Now(),
		Payload:   models.AgentResponse{Type: models.AgentTypeThought, Content: thought},
	})
}

func sendToolCallInfo(clientID, msgID, toolName string, args map[string]interface{}) {
	argsJson, _ := json.Marshal(args)
	content := fmt.Sprintf("🛠️ Calling Tool: %s | Args: %s", toolName, string(argsJson))

	HubInstance.SendToClient(clientID, models.Envelope{
		ID:        msgID,
		TargetID:  clientID,
		Timestamp: time.Now(),
		Payload: models.AgentResponse{
			Type:     models.AgentTypeTool,
			Content:  content,
			ToolName: toolName,
		},
	})
}

func sendToolResult(clientID, msgID, toolName, content string) {
	payload := fmt.Sprintf("Tool Output: %s", content)
	HubInstance.SendToClient(clientID, models.Envelope{
		ID:        msgID,
		TargetID:  clientID,
		Timestamp: time.Now(),
		Payload: models.AgentResponse{
			Type:     models.AgentTypeTool,
			Content:  payload,
			ToolName: toolName,
		},
	})
}

func sendError(clientID, msgID, errText string) {
	HubInstance.SendToClient(clientID, models.Envelope{
		ID:        msgID,
		TargetID:  clientID,
		Timestamp: time.Now(),
		Payload:   models.AgentResponse{Type: models.AgentTypeError, Content: errText},
	})
}
