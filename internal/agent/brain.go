package agent

import (
	"CHIMERA/internal/clients/ollama"
	"CHIMERA/internal/skills"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ollama/ollama/api"
)

// NewBrain: Brain'i baslatir
func NewBrain() error {
	if strings.TrimSpace(config.LLM.Host) == "" {
		return fmt.Errorf("llm.host is empty")
	}
	model := strings.TrimSpace(config.LLM.ModelMain)

	// Promptlari yukle
	ps, err := LoadPrompts(config.Paths.PromptsDir)
	if err != nil {
		return err
	}

	// Ollama istemcisini al (internal/clients/ollama/client.go)
	cli, err := ollama.NewClient()
	if err != nil {
		return fmt.Errorf("failed to initialize Ollama client: %w", err)
	}

	Brain = BrainStruct{
		Ollama:               cli,
		Model:                model,
		ModelFormatter:       strings.TrimSpace(config.LLM.FormatterModel),
		SystemPrompt:         strings.TrimSpace(ps.Start),
		FormatterPrompt:      strings.TrimSpace(ps.Formatter),
		Temperature:          config.LLM.Temperature,
		FormatterTemperature: config.LLM.FormatterTemperature,
	}

	log.System("Init", "BrainStruct initialized", "component", "BRAIN", "model", Brain.Model)
	return nil
}

// ChatAgent: Stateless (Durumsuz) konusma fonksiyonu.
// History disaridan gelir, guncellenmis history ve karar (AgentResponse) doner.
func (b *BrainStruct) ChatAgent(
	ctx context.Context,
	history []api.Message,
	userText string,
) (*AgentResponse, []api.Message, error) {

	if b == nil || b.Ollama == nil {
		return nil, history, fmt.Errorf("brain not initialized")
	}
	userText = strings.TrimSpace(userText)
	if userText == "" {
		return nil, history, fmt.Errorf("empty user message")
	}

	// 1. Sistem Promptunu Hazirla (Tool listesiyle birlikte)
	currentSystemPrompt := b.prepareSystemPrompt()

	// 2. Tarihce Yonetimi
	var currentHistory []api.Message
	if len(history) > 0 && history[0].Role == "system" {
		// Sistem promptunu tazele
		currentHistory = make([]api.Message, len(history))
		copy(currentHistory, history)
		currentHistory[0].Content = currentSystemPrompt
	} else {
		// Tarihce yoksa olustur
		currentHistory = []api.Message{
			{Role: "system", Content: currentSystemPrompt},
		}
		if len(history) > 0 {
			currentHistory = append(currentHistory, history...)
		}
	}

	// 3. Mesaj Sinirla ve Kullanici Mesajini Ekle
	currentHistory = trimConversation(currentHistory, config.LLM.MaxConversationMessages)
	currentHistory = append(currentHistory, api.Message{Role: "user", Content: userText})

	// Chat gecmisini kaydet - Brain'e giren mesaj (Debug)
	log.TraceBrain("Receive", "Brain: User message received", "component", "BRAIN", "content", userText)

	// 4. Istegi Hazirla
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 300*time.Second)
		defer cancel()
	}

	stream := false
	agentResponseSchema := json.RawMessage(`{
	"type": "object",
	"properties": {
		"thought": {"type": "string"},
		"tool": {"type": "string", "description": "Tool name to call, or empty if no tool."},
		"args": {"type": "object", "description": "Tool arguments object.", "additionalProperties": true},
		"reply": {"type": "string", "description": "Assistant reply to the user."}
	},
	"required": ["thought", "tool", "args", "reply"],
	"additionalProperties": false
}`)
	req := &api.ChatRequest{
		Model:    b.Model,
		Messages: currentHistory,
		Format:   agentResponseSchema,
		Options: map[string]any{
			"temperature": b.Temperature,
			"num_predict": 1024,
		},
		Stream: &stream,
	}

	// 5. Ollama'ya Gonder
	var responseThinking string
	var responseContent string
	var responseRaw string
	err := b.Ollama.Chat(ctx, req, func(resp api.ChatResponse) error {
		// Bazi modeller/adapter'lar bos chunk'lar yollayabiliyor; dolu veriyi ezmeyelim.
		if strings.TrimSpace(resp.Message.Thinking) != "" {
			responseThinking = resp.Message.Thinking
		}
		if strings.TrimSpace(resp.Message.Content) != "" {
			responseContent = resp.Message.Content
		}
		// Parse edilecek icerik sadece modelin JSON content'i olmali.
		responseRaw = responseContent
		return nil
	})

	if err != nil {
		return nil, currentHistory, fmt.Errorf("ollama API error: %w", err)
	}

	rawForFormatter := strings.TrimSpace(responseRaw)
	if rawForFormatter == "" {
		rawForFormatter = strings.TrimSpace(responseThinking)
	}
	if rawForFormatter != "" && b.ModelFormatter != "" && b.FormatterPrompt != "" {
		formatted, formatErr := b.formatWithHelper(ctx, rawForFormatter, agentResponseSchema)
		if formatErr == nil && strings.TrimSpace(formatted) != "" {
			responseRaw = formatted
		} else if formatErr != nil {
			log.TraceBrain("Error", "Formatter failed", "component", "BRAIN", "error", formatErr.Error())
		}
	}

	if strings.TrimSpace(responseRaw) == "" && strings.TrimSpace(responseThinking) != "" {
		// Bazi modeller JSON'u only-thinking kanalina basabiliyor; ilk JSON objesini kurtar.
		if extracted, ok := extractFirstJSONObject(responseThinking); ok {
			responseRaw = extracted
			log.TraceBrain("LLM", "Brain: JSON extracted from thinking", "component", "BRAIN", "raw_response", responseRaw)
		}
	}

	if strings.TrimSpace(responseRaw) == "" {
		log.TraceBrain("Error", "Empty model response", "component", "BRAIN", "thinking", responseThinking)
		return nil, currentHistory, InvalidAgentResponseError{
			Raw:    "",
			Clean:  "",
			Reason: "Empty response content from model (no JSON produced).",
		}
	}

	// Chat gecmisini kaydet - Ollama raw response (Debug)
	if strings.TrimSpace(responseThinking) != "" {
		log.TraceBrain("LLM", "Brain: Ollama raw thinking", "component", "BRAIN", "thinking", responseThinking)
	}
	log.TraceBrain("LLM", "Brain: Ollama raw response", "component", "BRAIN", "raw_response", responseRaw)

	// 6. JSON Parse Et
	cleanRaw := cleanJSON(responseRaw)
	var agentResp AgentResponse
	if err := json.Unmarshal([]byte(cleanRaw), &agentResp); err != nil {
		log.TraceBrain("Parse", "JSON Parse Error", "raw", responseRaw, "clean", cleanRaw, "error", err.Error())
		currentHistory = append(currentHistory, api.Message{
			Role:    "assistant",
			Content: cleanRaw,
		})
		return nil, currentHistory, InvalidJSONError{Raw: responseRaw, Clean: cleanRaw}
	}

	// JSON gecerli olsa bile bazi modeller `{}` gibi bos objeler donup akisi kilitleyebilir.
	// En azindan bir tool cagrisi veya bir reply zorunlu.
	if strings.TrimSpace(agentResp.Tool) == "" && strings.TrimSpace(agentResp.Reply) == "" {
		return nil, currentHistory, InvalidAgentResponseError{
			Raw:    responseRaw,
			Clean:  cleanRaw,
			Reason: "Missing both 'tool' and 'reply'.",
		}
	}

	// 7. Cevabi Tarihceye Ekle
	currentHistory = append(currentHistory, api.Message{
		Role:    "assistant",
		Content: cleanRaw,
	})

	// Chat gecmisini kaydet - Parsed AgentResponse (Debug)
	log.TraceBrain("Parse", "Brain: Parsed agent response",
		"component", "BRAIN",
		"thought", agentResp.Thought,
		"tool", agentResp.Tool,
		"reply", agentResp.Reply,
	)

	return &agentResp, currentHistory, nil
}

// --- YARDIMCI FONKSIYONLAR ---

func (b *BrainStruct) prepareSystemPrompt() string {
	toolsList := skills.GetToolsPrompt() // skills paketinden listeyi al
	return b.SystemPrompt + "\n\n" + toolsList
}

func cleanJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)
	if extracted, ok := extractFirstJSONObject(raw); ok {
		return extracted
	}
	return raw
}

// extractFirstJSONObject metin icinden ilk JSON objesini guvenle ayiklar.
// Modelin JSON disi aciklama/preamble basmasi durumunda parse'i kurtarmak icin kullanilir.
func extractFirstJSONObject(s string) (string, bool) {
	start := strings.IndexByte(s, '{')
	if start == -1 {
		return "", false
	}

	inString := false
	escaped := false
	depth := 0
	for i := start; i < len(s); i++ {
		c := s[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if c == '\\' {
				escaped = true
				continue
			}
			if c == '"' {
				inString = false
			}
			continue
		}

		switch c {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return strings.TrimSpace(s[start : i+1]), true
			}
		}
	}

	return "", false
}

func (b *BrainStruct) formatWithHelper(ctx context.Context, raw string, schema json.RawMessage) (string, error) {
	if b == nil || b.Ollama == nil {
		return "", fmt.Errorf("formatter client not initialized")
	}
	if strings.TrimSpace(b.ModelFormatter) == "" || strings.TrimSpace(b.FormatterPrompt) == "" {
		return "", fmt.Errorf("formatter model or prompt is empty")
	}
	if strings.TrimSpace(raw) == "" {
		return "", fmt.Errorf("formatter input is empty")
	}

	formatterCtx := ctx
	if _, hasDeadline := formatterCtx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		formatterCtx, cancel = context.WithTimeout(formatterCtx, 60*time.Second)
		defer cancel()
	}

	stream := false
	req := &api.ChatRequest{
		Model: b.ModelFormatter,
		Messages: []api.Message{
			{Role: "system", Content: b.FormatterPrompt},
			{Role: "user", Content: raw},
		},
		Format: schema,
		Options: map[string]any{
			"temperature": b.FormatterTemperature,
			"num_predict": 512,
		},
		Stream: &stream,
	}

	var responseThinking string
	var responseContent string
	err := b.Ollama.Chat(formatterCtx, req, func(resp api.ChatResponse) error {
		if strings.TrimSpace(resp.Message.Thinking) != "" {
			responseThinking = resp.Message.Thinking
		}
		if strings.TrimSpace(resp.Message.Content) != "" {
			responseContent = resp.Message.Content
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("formatter API error: %w", err)
	}

	if strings.TrimSpace(responseContent) == "" && strings.TrimSpace(responseThinking) != "" {
		if extracted, ok := extractFirstJSONObject(responseThinking); ok {
			responseContent = extracted
		}
	}
	if strings.TrimSpace(responseContent) == "" {
		return "", fmt.Errorf("formatter returned empty content")
	}

	return responseContent, nil
}
