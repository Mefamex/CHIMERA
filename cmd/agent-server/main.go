package main

import (
	"fmt"
	"log"
	
	"CHIMERA/internal/config"
)

func main() {
	fmt.Println("🚀 Chimera Agent Başlatılıyor...")

	// 1. Config Yükle
	cfg, err := config.LoadConfig("./configs/config.yaml")
	if err != nil {
		log.Fatalf("❌ Kritik Hata: Config yüklenemedi! \n%v", err)
	}

	// Test Çıktısı
	fmt.Printf("✅ Konfigürasyon Yüklendi:\n")
	fmt.Printf("   - Uygulama: %s (v%s)\n", cfg.App.Name, cfg.App.Version)
	fmt.Printf("   - LLM Modeli: %s\n", cfg.LLM.ModelMain)
	fmt.Printf("   - Python Servisi: %s\n", cfg.Services.PythonAnalyzer.URL)
	fmt.Printf("   - Port: %d\n", cfg.Server.Port)
	
	// Buradan sonra sunucu başlatma kodları gelecek...
}