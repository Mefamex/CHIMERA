// file: internal/common/stacksInfo.go
/*
check go, runtimes and versions; output json to file and stdout
*/
package common

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type stacksInfo struct {
	SystemStatus systemStatus
	RuntimeInfo  runtimeInfo
}

var StackInfo = stacksInfo{}

type systemStatus struct {
	Timestamp string        `json:"timestamp"`
	HostOS    string        `json:"host_os"`
	GoVersion string        `json:"go_version"`
	Runtimes  []runtimeInfo `json:"runtimes"`
}

type runtimeInfo struct {
	Type    string `json:"type"`    // python, node
	Path    string `json:"path"`    // runtimes/python-3.14.2
	Version string `json:"version"` // 3.14.2 veya v25.4.0
	Status  string `json:"status"`  // OK, ERROR
	Message string `json:"message"` // Hata mesajı veya detay
}

// --- Renk Kodları (Terminal Şıklığı İçin) ---
const (
	colorReset = "\033[0m"
	colorGreen = "\033[32m"
	colorRed   = "\033[31m"
	//colorYellow = "\033[33m"
	//colorBlue   = "\033[34m"
	colorCyan = "\033[36m"
)

// --- Ana Fonksiyon ---
func (s *stacksInfo) Get() (systemStatus, error) {
	// 1. Taramayı Başlat
	fmt.Println(colorCyan + "🔍 CHIMERA System Diagnostic Tool Başlatılıyor..." + colorReset)
	status := s.performSystemCheck()

	// 2. JSON Çıktısını Oluştur
	jsonData, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		fmt.Println(colorRed + "JSON Hatası:" + err.Error() + colorReset)
		return systemStatus{}, err
	}
	s.SystemStatus = status
	// 3. JSON Dosyasına Kaydet
	savePath := "data/system_stacks_info.json"
	// Klasör yoksa oluştur (Garanti olsun)
	os.MkdirAll("data", 0755)

	err = os.WriteFile(savePath, jsonData, 0644)
	if err != nil {
		fmt.Printf("%s❌ Dosya yazılamadı: %s%s\n", colorRed, err.Error(), colorReset)
	} else {
		fmt.Printf("%s💾 Rapor kaydedildi: %s%s\n", colorGreen, savePath, colorReset)
	}

	// Bu scripti başka bir Go programından çağırırsan JSON'ı stdout'a basabilirsin
	// fmt.Println(string(jsonData))
	return status, nil
}

// --- Mantık Katmanı ---

func (s *stacksInfo) performSystemCheck() systemStatus {
	status := systemStatus{
		Timestamp: time.Now().Format(time.RFC3339),
		HostOS:    runtime.GOOS + "/" + runtime.GOARCH,
		GoVersion: runtime.Version(),
		Runtimes:  []runtimeInfo{},
	}

	// Runtimes klasörünü tara
	entries, err := os.ReadDir("runtimes")
	if err != nil {
		fmt.Printf(colorRed+"HATA: runtimes klasörü okunamadı: %v"+colorReset, err)
		return status
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirName := entry.Name()
		fullPath := filepath.Join("runtimes", dirName)

		var info runtimeInfo

		// Klasör ismine göre algılama
		if strings.HasPrefix(dirName, "python") {
			info = s.CheckPythonRuntime(fullPath)
		} else if strings.HasPrefix(dirName, "node") {
			info = s.CheckNodeRuntime(fullPath)
		} else {
			info = runtimeInfo{
				Type:    "unknown",
				Path:    fullPath,
				Status:  "SKIPPED",
				Message: "Bilinmeyen runtime tipi",
			}
		}
		status.Runtimes = append(status.Runtimes, info)
	}

	return status
}

func (s *stacksInfo) CheckPythonRuntime(pythonBin string) runtimeInfo {
	// Sanal ortamdaki binary yolu
	basePath := filepath.Join(pythonBin, "..", "..", "..")
	info := runtimeInfo{
		Type: "python",
		Path: basePath,
	}
	// 1. Binary var mı?
	if _, err := os.Stat(pythonBin); os.IsNotExist(err) {
		info.Status = "ERROR"
		info.Message = "venv/bin/python bulunamadı. 'venv' kurulu mu? "
		// check python*/ or Python*/ folders for version
		for _, name := range []string{"python", "Python"} {
			matches, _ := filepath.Glob(filepath.Join(basePath, name+"*"))
			for _, match := range matches {
				if fi, err := os.Stat(match); err == nil && fi.IsDir() {
					info.Message += fmt.Sprintf(" Bulunan klasör: %s. ", match)
				}
			}
			if len(matches) > 0 {
				info.Message += " (Dikkat: Config dosyasını python runtimes ile uyumlu hale getirin.)"
			}
		}
		return info
	}

	// 2. Çalışıyor mu? (--version)
	cmd := exec.Command(pythonBin, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		info.Status = "ERROR"
		info.Message = fmt.Sprintf("Çalıştırma hatası: %v", err)
		info.Version = " "
		return info
	}
	rawOutput := string(output)
	cleanOutput := strings.TrimSpace(rawOutput)
	info.Version = strings.TrimPrefix(cleanOutput, "Python ")

	// 3. python içinde çalışıyor mu? (sys.executable)
	cmd2 := exec.Command(pythonBin, "-c", "import sys; print(sys.executable)")
	if _, err := cmd2.CombinedOutput(); err != nil {
		info.Status = "ERROR"
		info.Message = fmt.Sprintf("Python içi çalışma hatası: %v", err)
		return info
	}

	info.Status = "OK"
	info.Message = "Ready. Executable: " + pythonBin

	// check python version and folder name is same
	exceptedName := "python-" + info.Version
	if strings.Contains(filepath.Base(info.Path), exceptedName) == false {
		info.Status = "WARNING"
		info.Message += " (Dikkat: Klasör ismi ve Python versiyonu uyuşmuyor.)"
	}
	return info
}

func (s *stacksInfo) CheckNodeRuntime(basePath string) runtimeInfo {
	// Node.js modülleri kontrolü
	packageJson := filepath.Join(basePath, "package.json")
	nodeModules := filepath.Join(basePath, "node_modules")

	info := runtimeInfo{
		Type: "node",
		Path: basePath,
	}

	// 1. Yapı kontrolü
	if _, err := os.Stat(packageJson); os.IsNotExist(err) {
		info.Status = "ERROR"
		info.Message = "package.json eksik."
		return info
	}
	if _, err := os.Stat(nodeModules); os.IsNotExist(err) {
		info.Status = "WARNING" // Modüller yok ama package.json var
		info.Message = "node_modules eksik (pnpm install gerekli olabilir)."
	} else {
		info.Status = "OK"
		info.Message = "Ready"
	}

	// 2. Sistem Node Versiyonunu Al (Referans için)
	cmd := exec.Command("node", "--version")
	output, err := cmd.CombinedOutput()
	if err == nil {
		info.Version = strings.TrimSpace(string(output))
	} else {
		info.Version = "Node Not Found in PATH"
	}

	return info
}
