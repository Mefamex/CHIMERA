// file path : internal/clients/ollama/types.go
package ollama

type AgentResponse struct {
	Thought string            `json:"thought"` // Modelin kendi kendine düşünmesi (Chain of Thought)
	Tool    string            `json:"tool"`    // Hangi aracı kullanacak? (Boşsa tool yok)
	Args    map[string]string `json:"args"`    // Aracın parametreleri
	Reply   string            `json:"reply"`   // Kullanıcıya gösterilecek mesaj
}
