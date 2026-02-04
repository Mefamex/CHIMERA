// file path: internal/server/websocket.go
package server

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"CHIMERA/internal/agent"
	"CHIMERA/internal/clients/ollama"
	"CHIMERA/internal/models"

	"github.com/google/uuid"

	"github.com/gofiber/websocket/v2"
)

const maxConversationMessages = 25

func trimConversation(history []ollama.OllamaChatMessage) []ollama.OllamaChatMessage {
	if len(history) <= maxConversationMessages {
		return history
	}
	start := 0
	if len(history) > 0 && history[0].Role == "system" {
		start = 1
	}
	keep := maxConversationMessages
	if start == 1 {
		keep = maxConversationMessages - 1
		if keep < 1 {
			keep = 1
		}
	}
	if len(history)-start <= keep {
		return history
	}
	trimmed := make([]ollama.OllamaChatMessage, 0, maxConversationMessages)
	if start == 1 {
		trimmed = append(trimmed, history[0])
	}
	trimmed = append(trimmed, history[len(history)-keep:]...)
	return trimmed
}

func sendTextResponse(clientID string, msgID string, text string) {
	HubInstance.SendToClient(clientID, models.Envelope{
		ID:        msgID,
		TargetID:  clientID,
		Payload:   models.AgentResponse{Type: models.AgentTypeResponseStart, Content: ""},
		Timestamp: time.Now(),
	})
	if strings.TrimSpace(text) != "" {
		HubInstance.SendToClient(clientID, models.Envelope{
			ID:        msgID,
			TargetID:  clientID,
			Payload:   models.AgentResponse{Type: models.AgentTypeResponseDelta, Content: text},
			Timestamp: time.Now(),
		})
	}
	HubInstance.SendToClient(clientID, models.Envelope{
		ID:        msgID,
		TargetID:  clientID,
		Payload:   models.AgentResponse{Type: models.AgentTypeResponseEnd, Content: ""},
		Timestamp: time.Now(),
	})
}

// NewWebSocketHandler returns a websocket handler wired to a Brain.
func NewWebSocketHandler(brain *agent.Brain) func(c *websocket.Conn) {
	return func(c *websocket.Conn) {
		// Bağlantı
		log.Info("🔌 Yeni bir istemci bağlandı:", "component", "WEBSOCKET", "remote_addr", c.RemoteAddr())
		defer c.Close() // Bağlantı kapandığında temizleme

		// Client oluştur ve Hub'e kaydet
		clientID := c.RemoteAddr().String()
		client := &Client{ID: clientID, Send: make(chan models.Envelope, 256)}
		HubInstance.Register <- client
		defer func() { HubInstance.Unregister <- client }()

		// Writer: yalnızca bu goroutine WriteJSON çağırır
		go func() {
			for env := range client.Send {
				if err := c.WriteJSON(env); err != nil {
					log.Info("Hub'dan client'a gönderim hatası", "component", "WEBSOCKET", "error", err)
					return
				}
			}
		}()

		// Hoşgeldin mesajı (hedefli) (istemci hakkında kısaca yazdır)
		HubInstance.SendToClient(clientID, models.Envelope{
			TargetID:  clientID,
			Payload:   models.AgentResponse{Type: models.AgentTypeResponse, Content: "Secure Uplink Established to " + clientID +"\n System ready for commands."},
			Timestamp: time.Now(),
		})

		// Per-connection conversation history: this is why the model can remember earlier messages.
		conversation := []ollama.OllamaChatMessage(nil)
		if brain != nil {
			conversation = brain.NewConversation()
		}
		conversation = trimConversation(conversation)

		// Per-connection state for simple deterministic behaviors.
		countingMode := false
		lastMeaningfulUserRequest := ""

		inbound := make(chan models.UserMessage)
		done := make(chan struct{})

		// Okuma ayrı goroutine'da (bloklamayı izole ediyoruz)
		go func() {
			for {
				_, msg, err := c.ReadMessage()
				if err != nil {
					log.Info("Bağlantı koptu (reader)", "component", "WEBSOCKET", "error", err)
					close(done)
					return
				}
				var userMsg models.UserMessage
				if err := json.Unmarshal(msg, &userMsg); err != nil {
					log.Info("JSON Hatası", "component", "WEBSOCKET", "error", err)
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
				if strings.TrimSpace(userMsg.Type) != "user_message" {
					continue
				}
				userText := strings.TrimSpace(userMsg.Content)
				if userText == "" {
					continue
				}
				log.Info("📩 Gelen Mesaj", "component", "WEBSOCKET", "message", userText)

				lower := strings.ToLower(userText)

				msgID := uuid.NewString()

				// 4b) "en son istediğimi yap" gibi komutlar: önce son isteği hatırlat.
				if strings.Contains(lower, "en son") && strings.Contains(lower, "yap") {
					resp := lastMeaningfulUserRequest
					if strings.TrimSpace(resp) == "" {
						resp = "En son net bir isteğin yoktu. Ne yapmamı istersin?"
					} else {
						resp = "En son isteğin şuydu: \"" + resp + "\". Bunu mu yapayım?"
					}
					sendTextResponse(clientID, msgID, resp)
					conversation = append(conversation,
						ollama.OllamaChatMessage{Role: "user", Content: userText},
						ollama.OllamaChatMessage{Role: "assistant", Content: resp},
					)
					conversation = trimConversation(conversation)
					continue
				}

				// 5) Counting mode: user wants alternating counting.
				if strings.Contains(lower, "say") && (strings.Contains(lower, "beraber") || strings.Contains(lower, "karşı") || strings.Contains(lower, "sayal") || strings.Contains(lower, "sayıc") || strings.Contains(lower, "sayac")) {
					countingMode = true
					resp := "Tamam. Sen 1 yaz, ben 2 diyeceğim."
					sendTextResponse(clientID, msgID, resp)
					conversation = append(conversation,
						ollama.OllamaChatMessage{Role: "user", Content: userText},
						ollama.OllamaChatMessage{Role: "assistant", Content: resp},
					)
					conversation = trimConversation(conversation)
					continue
				}
				if countingMode {
					if lower == "dur" || lower == "stop" {
						countingMode = false
						resp := "Tamam, saymayı durdurdum."
						sendTextResponse(clientID, msgID, resp)
						conversation = append(conversation,
							ollama.OllamaChatMessage{Role: "user", Content: userText},
							ollama.OllamaChatMessage{Role: "assistant", Content: resp},
						)
						conversation = trimConversation(conversation)
						continue
					}
					if n, err := strconv.Atoi(lower); err == nil {
						resp := strconv.Itoa(n + 1)
						sendTextResponse(clientID, msgID, resp)
						conversation = append(conversation,
							ollama.OllamaChatMessage{Role: "user", Content: userText},
							ollama.OllamaChatMessage{Role: "assistant", Content: resp},
						)
						conversation = trimConversation(conversation)
						continue
					}
				}

				// Fallback to LLM.
				if brain == nil {
					sendTextResponse(clientID, msgID, "Beyin (LLM) hazır değil.")
					continue
				}

				ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
				full, updatedHistory, err := brain.ChatStreamWithHistory(ctx, conversation, userText, func(delta string) {
					if strings.TrimSpace(delta) == "" {
						return
					}
					HubInstance.SendToClient(clientID, models.Envelope{
						ID:        msgID,
						TargetID:  clientID,
						Payload:   models.AgentResponse{Type: models.AgentTypeResponseDelta, Content: delta},
						Timestamp: time.Now(),
					})
				})
				cancel()
				if err != nil {
					HubInstance.SendToClient(clientID, models.Envelope{
						ID:        msgID,
						TargetID:  clientID,
						Payload:   models.AgentResponse{Type: models.AgentTypeError, Content: "Ollama hatası: " + err.Error()},
						Timestamp: time.Now(),
					})
					HubInstance.SendToClient(clientID, models.Envelope{
						ID:        msgID,
						TargetID:  clientID,
						Payload:   models.AgentResponse{Type: models.AgentTypeResponseEnd, Content: ""},
						Timestamp: time.Now(),
					})
					continue
				}

				_ = full
				conversation = updatedHistory
				conversation = trimConversation(conversation)
				HubInstance.SendToClient(clientID, models.Envelope{
					ID:        msgID,
					TargetID:  clientID,
					Payload:   models.AgentResponse{Type: models.AgentTypeResponseEnd, Content: ""},
					Timestamp: time.Now(),
				})
			}
		}
	}
}
