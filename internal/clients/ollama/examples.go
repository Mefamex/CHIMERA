// file path: internal/clients/ollama/examples.go
package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ollama/ollama/api"
)

// 1. Modelleri Listeleme
func exampleListModels(client *api.Client) {
	fmt.Println("--- MODELLER LİSTELENİYOR ---")
	resp, err := client.List(context.Background())
	if err != nil {
		fmt.Println("Hata:", err)
		return
	}

	for _, m := range resp.Models {
		fmt.Printf("Model: %s | Boyut: %.2f GB | Aile: %s\n",
			m.Name, float64(m.Size)/1024/1024/1024, m.Details.Family)
	}
}

// 2. Basit Chat (Tek Seferlik, Stream Yok)
func exampleChatBasic(client *api.Client) {
	fmt.Println("\n--- BASİT CHAT ---")

	stream := false
	req := &api.ChatRequest{
		Model: "llama3",
		Messages: []api.Message{
			{Role: "user", Content: "Go dili neden hızlıdır? Tek cümleyle söyle."},
		},
		Stream: &stream,
	}

	err := client.Chat(context.Background(), req, func(resp api.ChatResponse) error {
		fmt.Println("Cevap:", resp.Message.Content)
		return nil
	})
	if err != nil {
		fmt.Println("Chat Hatası:", err)
	}
}

// 3. Streaming Chat (Daktilo Efekti)
func exampleChatStream(client *api.Client) {
	fmt.Println("\n--- STREAMING CHAT ---")

	stream := true
	req := &api.ChatRequest{
		Model: "llama3",
		Messages: []api.Message{
			{Role: "user", Content: "Bana 3 satırlık bir şiir yaz."},
		},
		Stream: &stream,
	}

	fmt.Print("Bot: ")
	err := client.Chat(context.Background(), req, func(resp api.ChatResponse) error {
		// Her parça geldiğinde burası çalışır
		fmt.Print(resp.Message.Content)
		return nil
	})
	fmt.Println() // Satır sonu
	if err != nil {
		fmt.Println("Stream Hatası:", err)
	}
}

// 4. Konuşma Geçmişi (Context/Memory)
func exampleChatHistory(client *api.Client) {
	fmt.Println("\n--- TARİHÇELİ SOHBET ---")

	// Geçmişi bir slice içinde tutuyoruz
	history := []api.Message{
		{Role: "system", Content: "Sen huysuz bir ihtiyarsın."},
		{Role: "user", Content: "Merhaba amca, nasılsın?"},
	}

	stream := false
	req := &api.ChatRequest{
		Model:    "llama3",
		Messages: history,
		Stream:   &stream,
	}

	// İlk cevap
	client.Chat(context.Background(), req, func(resp api.ChatResponse) error {
		fmt.Println("Bot:", resp.Message.Content)
		// Cevabı geçmişe ekle
		history = append(history, resp.Message)
		return nil
	})

	// Devam ediyoruz...
	fmt.Println("User: Torunların nasıl?")
	history = append(history, api.Message{Role: "user", Content: "Torunların nasıl?"})

	req.Messages = history // Güncel geçmişi ver
	client.Chat(context.Background(), req, func(resp api.ChatResponse) error {
		fmt.Println("Bot:", resp.Message.Content)
		return nil
	})
}

// 5. Vision (Resim Gönderme)
func exampleVision(client *api.Client) {
	fmt.Println("\n--- GÖRSEL ANALİZ ---")

	// Resmi byte array olarak oku
	imgData, err := os.ReadFile("image.png")
	if err != nil {
		fmt.Println("Resim bulunamadı (test için image.png koymalısın)")
		return
	}

	stream := false
	req := &api.ChatRequest{
		Model: "llava", // Vision destekli model şart (örn: llava, llama3.2-vision)
		Messages: []api.Message{
			{
				Role:    "user",
				Content: "Bu resimde ne görüyorsun?",
				Images:  []api.ImageData{imgData}, // []byte
			},
		},
		Stream: &stream,
	}

	client.Chat(context.Background(), req, func(resp api.ChatResponse) error {
		fmt.Println("Analiz:", resp.Message.Content)
		return nil
	})
}

// 6. Tool Calling (Fonksiyon Çağırma - EN ÖNEMLİSİ)
func ExampleTools(client *api.Client) {
	fmt.Println("\n--- TOOL CALLING ---")

	// 1. JSON Şemasını Basitçe Hazırla
	// Go'nun katı struct kurallarıyla boğuşmak yerine, şemayı map[string]any
	// veya anonim struct olarak hazırlayıp, JSON üzerinden kütüphanenin tipine çevireceğiz.
	jsonSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"location": map[string]any{
				"type":        "string",
				"description": "Şehir adı, örn: Ankara",
			},
		},
		"required": []string{"location"},
	}

	// 2. Kütüphanenin İstediği Tipe Dönüştür (Marshalling Trick)
	// Bu yöntem "Unknown field" hatalarını bypass eder.
	data, _ := json.Marshal(jsonSchema)

	var params api.ToolFunctionParameters
	if err := json.Unmarshal(data, &params); err != nil {
		fmt.Println("Şema dönüştürme hatası:", err)
		return
	}

	// 3. Tool Tanımı
	myTool := api.Tool{
		Type: "function",
		Function: api.ToolFunction{
			Name:        "get_weather",
			Description: "Verilen şehir için hava durumunu getirir",
			Parameters:  params, // Artık tam olarak kütüphanenin istediği tipte!
		},
	}

	stream := false
	req := &api.ChatRequest{
		Model: "llama3.1", // Tool destekleyen model şart (llama3.1, qwen2.5 vb.)
		Messages: []api.Message{
			{Role: "user", Content: "Ankara'da hava nasıl?"},
		},
		Tools:  []api.Tool{myTool},
		Stream: &stream,
	}

	client.Chat(context.Background(), req, func(resp api.ChatResponse) error {
		// Tool çağrısı var mı kontrol et
		if len(resp.Message.ToolCalls) > 0 {
			for _, tool := range resp.Message.ToolCalls {
				fmt.Printf("Model Tool Çağırdı -> İsim: %s | Parametre: %v\n",
					tool.Function.Name, tool.Function.Arguments)
			}
		} else {
			fmt.Println("Model normal cevap verdi:", resp.Message.Content)
		}
		return nil
	})
}

func exampleEmbeddings(client *api.Client) {
	req := &api.EmbeddingRequest{
		Model:  "nomic-embed-text",
		Prompt: "Bu cümleyi sayısal vektöre çevir.",
	}
	resp, _ := client.Embeddings(context.Background(), req)
	fmt.Printf("Vektör Boyutu: %d\n", len(resp.Embedding))
}
