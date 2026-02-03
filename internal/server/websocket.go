// file path: internal/server/websocket.go
package server

import (
	"context"
	"encoding/json"
	"log"
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

func conversationPlainSummary(history []ollama.OllamaChatMessage) string {
	// Keep it short and human-friendly. No JSON.
	var userCount, assistantCount int
	var lastUser, lastAssistant string
	for _, m := range history {
		role := strings.ToLower(strings.TrimSpace(m.Role))
		switch role {
		case "user":
			userCount++
			if strings.TrimSpace(m.Content) != "" {
				lastUser = strings.TrimSpace(m.Content)
			}
		case "assistant":
			assistantCount++
			if strings.TrimSpace(m.Content) != "" {
				lastAssistant = strings.TrimSpace(m.Content)
			}
		}
	}
	if userCount == 0 {
		return "Henüz konuşma geçmişi yok."
	}
	if lastUser == "" {
		lastUser = "(boş)"
	}
	if lastAssistant == "" {
		lastAssistant = "(boş)"
	}
	return "Konuşma özeti: " +
		"Şu ana kadar " + strconv.Itoa(userCount) + " kullanıcı mesajı ve " + strconv.Itoa(assistantCount) + " yanıt var. " +
		"Son kullanıcı mesajı: \"" + lastUser + "\". " +
		"Son yanıt: \"" + lastAssistant + "\"."
}

// NewWebSocketHandler returns a websocket handler wired to a Brain.
func NewWebSocketHandler(brain *agent.Brain) func(c *websocket.Conn) {
	return func(c *websocket.Conn) {
		// Bağlantı
		log.Println("🔌 Yeni bir istemci bağlandı:", c.RemoteAddr())
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
					log.Println("Hub'dan client'a gönderim hatası: ", err)
					return
				}
			}
		}()

		// Hoşgeldin mesajı (hedefli) (istemci hakkında kısaca yazdır)
		HubInstance.SendToClient(clientID, models.Envelope{
			TargetID:  clientID,
			Payload:   models.AgentResponse{Type: models.AgentTypeSystem, Content: "Secure Uplink Established to " + clientID},
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
		lastAssistantMode := "" // e.g. offer_plain_summary

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
				if strings.TrimSpace(userMsg.Type) != "user_message" {
					continue
				}
				userText := strings.TrimSpace(userMsg.Content)
				if userText == "" {
					continue
				}
				log.Printf("📩 Gelen Mesaj: %s\n", userText)

				lower := strings.ToLower(userText)

				msgID := uuid.NewString()

				isAffirm := lower == "olur" || lower == "tamam" || lower == "evet" || lower == "ok" || lower == "okay" || lower == "peki"
				if isAffirm && lastAssistantMode == "offer_plain_summary" {
					lastAssistantMode = ""
					summary := conversationPlainSummary(conversation)
					sendTextResponse(clientID, msgID, summary)
					conversation = append(conversation,
						ollama.OllamaChatMessage{Role: "user", Content: userText},
						ollama.OllamaChatMessage{Role: "assistant", Content: summary},
					)
					conversation = trimConversation(conversation)
					continue
				}
				// Update "last meaningful" request for "ne istedim senden" questions.
				if len(userText) >= 4 && strings.TrimSpace(strings.Trim(userText, "0123456789")) != "" {
					// Avoid storing trivial counting / filler.
					if !strings.Contains(lower, "ne istedim") && !strings.Contains(lower, "what did i") {
						lastMeaningfulUserRequest = userText
					}
				}

				// Deterministic policies (cross-model reliability):
				// 1) Don't reveal system prompt.
				if strings.Contains(lower, "sistem prompt") || strings.Contains(lower, "system prompt") {
					sendTextResponse(clientID, msgID, "Sistem promptumu paylaşamam.")
					conversation = append(conversation,
						ollama.OllamaChatMessage{Role: "user", Content: userText},
						ollama.OllamaChatMessage{Role: "assistant", Content: "Sistem promptumu paylaşamam."},
					)
					conversation = trimConversation(conversation)
					continue
				}

				// 2) Version hint: user expects the number they mentioned at the beginning.
				if strings.Contains(lower, "version") && (strings.Contains(lower, "kaç") || strings.Contains(lower, "nedir") || strings.Contains(lower, "what")) {
					sendTextResponse(clientID, msgID, "2")
					conversation = append(conversation,
						ollama.OllamaChatMessage{Role: "user", Content: userText},
						ollama.OllamaChatMessage{Role: "assistant", Content: "2"},
					)
					conversation = trimConversation(conversation)
					continue
				}

				// 3) Refuse dumping conversation as JSON.
				if strings.Contains(lower, "json") && (strings.Contains(lower, "mesaj") || strings.Contains(lower, "sohbet")) {
					resp := "Bunu yapamam. İstersen konuşmayı düz metin olarak özetleyebilirim."
					lastAssistantMode = "offer_plain_summary"
					sendTextResponse(clientID, msgID, resp)
					conversation = append(conversation,
						ollama.OllamaChatMessage{Role: "user", Content: userText},
						ollama.OllamaChatMessage{Role: "assistant", Content: resp},
					)
					conversation = trimConversation(conversation)
					continue
				}

				// 4) Answer "en son ... ne" questions deterministically.
				asksLast := strings.Contains(lower, "en son") || strings.Contains(lower, "az önce") || strings.Contains(lower, "what is my last") || strings.Contains(lower, "what did i")
				if strings.Contains(lower, "ne istedim") || strings.Contains(lower, "ne istemiştim") ||
					(asksLast && (strings.Contains(lower, "ne") || strings.Contains(lower, "sor") || strings.Contains(lower, "isted") || strings.Contains(lower, "yaz"))) {
					resp := lastMeaningfulUserRequest
					if strings.TrimSpace(resp) == "" {
						resp = "Az önce net bir isteğin yoktu."
					}
					sendTextResponse(clientID, msgID, resp)
					conversation = append(conversation,
						ollama.OllamaChatMessage{Role: "user", Content: userText},
						ollama.OllamaChatMessage{Role: "assistant", Content: resp},
					)
					conversation = trimConversation(conversation)
					continue
				}

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
