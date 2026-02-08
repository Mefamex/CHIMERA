// file path: internal/models/types.go
package models

import "time"

// Frontend'den gelen mesaj (User -> Go)
type UserMessage struct {
	Type    string `json:"type"`    // "user_message"
	Content string `json:"content"` // "Merhaba Chimera"
}

// Agent tipi için özel tip ve sabitler (typo önlemek için)
type AgentType string

const (
	AgentTypeResponse      AgentType = "response"
	AgentTypeResponseStart AgentType = "response_start"
	AgentTypeResponseDelta AgentType = "response_delta"
	AgentTypeResponseEnd   AgentType = "response_end"
	AgentTypeThought       AgentType = "thought"
	AgentTypeTool          AgentType = "tool"
	AgentTypeToolResult    AgentType = "tool_result"
	AgentTypeSystem        AgentType = "system"
	AgentTypeError         AgentType = "error"
	AgentTypeLog           AgentType = "log"
)

// Frontend'e giden mesaj (Go -> User)
type AgentResponse struct {
	Type     AgentType       `json:"type"`                 // "response", "thought", "tool", "system", "error"
	Content  string          `json:"content"`              // Mesaj içeriği
	ToolName string          `json:"tool_name,omitempty"`  // Eğer tool kullandıysa (örn: "browser")
	LogType  string          `json:"log_type,omitempty"`   // Log stream icin kaynak
	Data     map[string]any  `json:"data,omitempty"`       // Log payload
}

// Envelope : hedefli veya broadcast mesajlar için sarmalayıcı
type Envelope struct {
	ID        string        `json:"id,omitempty"`        // opsiyonel mesaj kimliği
	TargetID  string        `json:"target_id,omitempty"` // boşsa tüm client'lara (broadcast)
	Broadcast bool          `json:"broadcast,omitempty"` // explicit broadcast tercihi
	Payload   AgentResponse `json:"payload"`             // gerçek gönderilen içerik
	Timestamp time.Time     `json:"timestamp,omitempty"` // opsiyonel zaman damgası
}

// client hakkında meta veri (ileride kullanılabilir)
type ClientInfo struct {
	ID   string            `json:"id"`
	Meta map[string]string `json:"meta,omitempty"`
}
