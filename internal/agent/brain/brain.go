// file path: internal/agent/brain/brain.go
package brain

import (
	"context"
	"fmt"
	"strings"
	"time"

	"CHIMERA/internal/agent"
	"CHIMERA/internal/clients/ollama"
)

func NewBrain() ( error) {
	if strings.TrimSpace(config.LLM.Host) == "" {
		return fmt.Errorf("llm.host is empty")
	}
	model := strings.TrimSpace(config.LLM.ModelMain)
	if model == "" {
		return fmt.Errorf("llm.model_main is empty")
	}

	ps, err := agent.LoadPrompts(config.Paths.PromptsDir)
	if err != nil {
		return err
	}

	Brain = BrainStruct{
		Ollama:       ollama.NewOllamaClient(),
		Model:        model,
		SystemPrompt: strings.TrimSpace(ps.Start),
		Temperature:  config.LLM.Temperature,
	}
	if strings.TrimSpace(config.LLM.Format) != "" {
		Brain.Format = strings.TrimSpace(config.LLM.Format)
	}

	log.Info("BrainStruct initialized ", "component", "BRAIN", "model", Brain.Model, "temperature", Brain.Temperature)
	return nil
}

func (b *BrainStruct) Chat(ctx context.Context, userText string) (string, error) {
	if b == nil || b.Ollama == nil {
		return "", fmt.Errorf("brain not initialized")
	}
	userText = strings.TrimSpace(userText)
	if userText == "" {
		return "", fmt.Errorf("empty user message")
	}

	// WS handler tarafında ayrı timeout olsa da burada da güvenli bir default ekliyoruz.
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 600*time.Second)
		defer cancel()
	}

	msgs := []ollama.OllamaChatMessage{
		{Role: "system", Content: b.SystemPrompt},
		{Role: "user", Content: userText},
	}

	options := map[string]any{
		"temperature": b.Temperature,
	}

	// Chat endpoint'i için JSON format zorlamıyoruz (format=nil).
	content, err := b.Ollama.Chat(ctx, b.Model, msgs, nil, options)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(content), nil
}

// NewConversation creates a new conversation history seeded with the system prompt.
// The returned slice can be kept per-connection and passed back to ChatStreamWithHistory.
func (b *BrainStruct) NewConversation() []ollama.OllamaChatMessage {
	if b == nil {
		return nil
	}
	return []ollama.OllamaChatMessage{{Role: "system", Content: b.SystemPrompt}}
}

func trimConversation(history []ollama.OllamaChatMessage) []ollama.OllamaChatMessage {
	maxConversationMessages := config.LLM.MaxConversationMessages
	if len(history) <= maxConversationMessages {
		return history
	}
	// Always keep the very first system message if present.
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

// ChatStreamWithHistory streams a response while including conversation history.
// On success it returns the assistant full text plus the updated history (history+user+assistant).
// On error, it returns the original history unchanged.
func (b *BrainStruct) ChatStreamWithHistory(
	ctx context.Context,
	history []ollama.OllamaChatMessage,
	userText string,
	onDelta func(delta string),
) (string, []ollama.OllamaChatMessage, error) {
	if b == nil || b.Ollama == nil {
		return "", history, fmt.Errorf("brain not initialized")
	}
	userText = strings.TrimSpace(userText)
	if userText == "" {
		return "", history, fmt.Errorf("empty user message")
	}

	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 600*time.Second)
		defer cancel()
	}

	base := history
	if len(base) == 0 {
		base = b.NewConversation()
	}
	base = trimConversation(base)

	msgs := make([]ollama.OllamaChatMessage, 0, len(base)+2)
	msgs = append(msgs, base...)
	msgs = append(msgs, ollama.OllamaChatMessage{Role: "user", Content: userText})

	options := map[string]any{
		"temperature": b.Temperature,
	}

	full, err := b.Ollama.ChatStream(ctx, b.Model, msgs, nil, options, onDelta)
	if err != nil {
		// History'yi değiştirmeyelim; caller isterse retry edebilir.
		return "", history, err
	}
	full = strings.TrimSpace(full)
	updated := append(msgs, ollama.OllamaChatMessage{Role: "assistant", Content: full})
	updated = trimConversation(updated)
	return full, updated, nil
}

// ChatStream tries to stream deltas from Ollama. If streaming fails, it falls back to Chat() and emits a single delta.
func (b *BrainStruct) ChatStream(ctx context.Context, userText string, onDelta func(delta string)) (string, error) {
	if b == nil || b.Ollama == nil {
		return "", fmt.Errorf("brain not initialized")
	}
	userText = strings.TrimSpace(userText)
	if userText == "" {
		return "", fmt.Errorf("empty user message")
	}

	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 600*time.Second)
		defer cancel()
	}

	msgs := []ollama.OllamaChatMessage{
		{Role: "system", Content: b.SystemPrompt},
		{Role: "user", Content: userText},
	}

	options := map[string]any{
		"temperature": b.Temperature,
	}

	// Web chat için JSON format zorlamıyoruz; streaming'de de format=nil.
	full, err := b.Ollama.ChatStream(ctx, b.Model, msgs, nil, options, onDelta)
	if err == nil {
		return strings.TrimSpace(full), nil
	}

	// Fallback: streaming desteklenmiyorsa veya hata olursa tek seferlik yanıt üret.
	content, err2 := b.Ollama.Chat(ctx, b.Model, msgs, nil, options)
	if err2 != nil {
		return "", err2
	}
	content = strings.TrimSpace(content)
	if content != "" && onDelta != nil {
		onDelta(content)
	}
	return content, nil
}
