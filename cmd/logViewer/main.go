// file : path : cmd/logViewer/main.go
package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"CHIMERA/internal/core"
)

const (
	StaticDir = "./web/static"
	StartPage = "/viewLogs/index.html"
)

func main() {
	logDir := resolveLogDir()

	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		log.Printf("UYARI: Log dizini bulunamadi: %s (Henuz log olusmamis olabilir)\n", logDir)
	}

	if _, err := os.Stat(StaticDir); os.IsNotExist(err) {
		log.Fatalf("KRİTİK HATA: Statik dosya dizini bulunamadı: %s\nLütfen projeyi kök dizinden çalıştırdığınızdan emin olun.", StaticDir)
	}
	mux := http.NewServeMux()
	logFS := http.FileServer(http.Dir(logDir))
	mux.Handle("/data/", http.StripPrefix("/data/", logFS))
	staticFS := http.FileServer(http.Dir(StaticDir))
	mux.Handle("/", staticFS)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal("Port bağlanamadı:", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	fullURL := baseURL + StartPage
	fmt.Println("==================================================")
	fmt.Printf("🚀 CHIMERA Log Viewer Başlatılıyor...\n")
	fmt.Printf("📂 Log Kaynagi : %s\n", logDir)
	fmt.Printf("🎨 Arayüz      : %s\n", StaticDir)
	fmt.Printf("🌍 Adres       : %s\n", fullURL)
	fmt.Println("==================================================")
	fmt.Println("Durdurmak için CTRL+C tuşlarına basın.")
	go func() {
		time.Sleep(1000 * time.Millisecond)
		openBrowser(fullURL)
	}()
	if err := http.Serve(listener, mux); err != nil {
		log.Fatal("Sunucu hatası:", err)
	}
}

func resolveLogDir() string {
	if core.Config != nil && core.Config.Paths.LogsDir != "" {
		return core.Config.Paths.LogsDir
	}
	return "./data/logs"
}

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("desteklenmeyen işletim sistemi")
	}
	if err != nil {
		fmt.Printf("⚠️  Tarayıcı otomatik açılamadı. Lütfen şu linke tıklayın: %s\n", url)
	}
}
