package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"CHIMERA/internal/clients/ollama"
	"CHIMERA/internal/core"
)

var log = core.Log

type Brain struct {
	ollama       *ollama.OllamaClient
	model        string
	systemPrompt string
	temperature  float64
	format       any
}

const maxConversationMessages = 25

func NewBrain() (*Brain, error) {
	cfg := core.Config
	if strings.TrimSpace(cfg.LLM.Host) == "" {
		return nil, fmt.Errorf("llm.host is empty")
	}
	model := strings.TrimSpace(cfg.LLM.ModelMain)
	if model == "" {
		return nil, fmt.Errorf("llm.model_main is empty")
	}

	ps, err := LoadPrompts(cfg.Paths.PromptsDir)
	if err != nil {
		return nil, err
	}

	b := &Brain{
		ollama:       ollama.NewOllamaClient(cfg.LLM.Host),
		model:        model,
		systemPrompt: strings.TrimSpace(ps.Start),
		temperature:  cfg.LLM.Temperature,
	}
	if strings.TrimSpace(cfg.LLM.Format) != "" {
		b.format = strings.TrimSpace(cfg.LLM.Format)
	}

	log.Info("Brain initialized ", "component", "BRAIN", "model", b.model, "temperature", b.temperature)
	return b, nil
}

func (b *Brain) Chat(ctx context.Context, userText string) (string, error) {
	if b == nil || b.ollama == nil {
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
		{Role: "system", Content: b.systemPrompt},
		{Role: "user", Content: userText},
	}

	options := map[string]any{
		"temperature": b.temperature,
	}

	// Chat endpoint'i için JSON format zorlamıyoruz (format=nil).
	content, err := b.ollama.Chat(ctx, b.model, msgs, nil, options)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(content), nil
}

// NewConversation creates a new conversation history seeded with the system prompt.
// The returned slice can be kept per-connection and passed back to ChatStreamWithHistory.
func (b *Brain) NewConversation() []ollama.OllamaChatMessage {
	if b == nil {
		return nil
	}
	return []ollama.OllamaChatMessage{{Role: "system", Content: b.systemPrompt}}
}

func trimConversation(history []ollama.OllamaChatMessage) []ollama.OllamaChatMessage {
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
func (b *Brain) ChatStreamWithHistory(
	ctx context.Context,
	history []ollama.OllamaChatMessage,
	userText string,
	onDelta func(delta string),
) (string, []ollama.OllamaChatMessage, error) {
	if b == nil || b.ollama == nil {
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
		"temperature": b.temperature,
	}

	full, err := b.ollama.ChatStream(ctx, b.model, msgs, nil, options, onDelta)
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
func (b *Brain) ChatStream(ctx context.Context, userText string, onDelta func(delta string)) (string, error) {
	if b == nil || b.ollama == nil {
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
		{Role: "system", Content: b.systemPrompt},
		{Role: "user", Content: userText},
	}

	options := map[string]any{
		"temperature": b.temperature,
	}

	// Web chat için JSON format zorlamıyoruz; streaming'de de format=nil.
	full, err := b.ollama.ChatStream(ctx, b.model, msgs, nil, options, onDelta)
	if err == nil {
		return strings.TrimSpace(full), nil
	}

	// Fallback: streaming desteklenmiyorsa veya hata olursa tek seferlik yanıt üret.
	content, err2 := b.ollama.Chat(ctx, b.model, msgs, nil, options)
	if err2 != nil {
		return "", err2
	}
	content = strings.TrimSpace(content)
	if content != "" && onDelta != nil {
		onDelta(content)
	}
	return content, nil
}
