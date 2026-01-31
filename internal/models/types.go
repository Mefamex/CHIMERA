package models

// Frontend'den gelen mesaj (User -> Go)
type UserMessage struct {
	Type    string `json:"type"`    // "user_message"
	Content string `json:"content"` // "Merhaba Chimera"
}

// Frontend'e giden mesaj (Go -> User)
type AgentResponse struct {
	Type     string `json:"type"`      // "response", "thought", "tool", "system", "error"
	Content  string `json:"content"`   // Mesaj içeriği
	ToolName string `json:"tool_name"` // Eğer tool kullandıysa (örn: "browser")
}