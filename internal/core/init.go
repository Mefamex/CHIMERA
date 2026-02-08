// file path: internal/core/init.go
package core

import (
	"CHIMERA/internal/common"
	"encoding/json"
	"fmt"
	"time"
)

// Setup: Core modülünü (Config, Logger vb.) doğru sırayla ayağa kaldırır.

func SetupCore() error {
	Config.Init()
	if !Config.Initialized() {
		return fmt.Errorf("config yükleme hatası")
	}
	time.Sleep(100 * time.Millisecond)

	InitLogger()
	time.Sleep(100 * time.Millisecond)

	if err := checkPython(); err != nil {
		return fmt.Errorf("runtime kontrol hatası: %w", err)
	}
	time.Sleep(100 * time.Millisecond)

	Logger.System("Startup", "✅ CHIMERA Core Sistemleri ready", "component", "Core")
	time.Sleep(100 * time.Millisecond)

	return nil
	// maybe errors.go
}

func checkPython() error {

	out := common.StackInfo.CheckPythonRuntime(Config.Runtimes.Python)
	if out.Status != "OK" {
		jsonData, _ := json.Marshal(out)
		Logger.Error("Python runtime check failed\n", fmt.Errorf("Python runtime check failed"), "error", string(jsonData), "component", "Runtime")
	} else {
		Logger.System("Startup", "✅ Python runtime ready", "path", out.Path, "version", out.Version, "component", "Runtime")
	}
	return nil
}
