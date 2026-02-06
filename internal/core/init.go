// file path: internal/core/init.go
package core

import (
	"fmt"
	"time"
)

// Setup: Core modülünü (Config, Logger vb.) doğru sırayla ayağa kaldırır.

func SetupCore() error {
	if err := Config.Reload(); err != nil {
		return fmt.Errorf("config yükleme hatası: %w", err)
	}
	time.Sleep(200 * time.Millisecond)

	InitLogger()
	time.Sleep(200 * time.Millisecond)

	if err := InitRuntimes(); err != nil {
		return fmt.Errorf("runtime kontrol hatası: %w", err)
	}
	time.Sleep(200 * time.Millisecond)

	Log.Info("CHIMERA Core Sistemleri Hazır", "component", "SYS")
	time.Sleep(200 * time.Millisecond)

	return nil
	// maybe errors.go
}
