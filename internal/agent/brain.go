package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"CHIMERA/internal/clients"
	"CHIMERA/internal/config"
)

type Brain struct {
	ollama       *clients.OllamaClient
	model        string
	systemPrompt string
	temperature  float64
	format       any
}

func NewBrain(cfg *config.Config) (*Brain, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
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
	// Web sohbetinde zincirleme düşünce/plan JSON dökmeyelim; sadece kullanıcıya cevap verelim.
	systemPrompt := strings.TrimSpace(ps.Start) + "\n\n" +
		"ÖNEMLİ: Kullanıcıya asla düşünce zinciri, plan/act/observe/critic adımları veya JSON dökme. " +
		"Sadece kısa ve net nihai cevabı düz metin olarak ver."

	b := &Brain{
		ollama:       clients.NewOllamaClient(cfg.LLM.Host),
		model:        model,
		systemPrompt: systemPrompt,
		temperature:  cfg.LLM.Temperature,
	}
	// Ollama 'format' alanı string ("json") olabiliyor; boşsa göndermeyelim.
	if strings.TrimSpace(cfg.LLM.Format) != "" {
		b.format = strings.TrimSpace(cfg.LLM.Format)
	}
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
		ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
	}

	msgs := []clients.OllamaChatMessage{
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
