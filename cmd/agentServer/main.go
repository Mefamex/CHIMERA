// file path: cmd/agentServer/main.go
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"CHIMERA/internal/agent"
	"CHIMERA/internal/core"
	"CHIMERA/internal/server"
	"CHIMERA/internal/skills"
)

const wellcomeMessage = `
______________________________________________________________________
 _______            _________  _______   _______   _______   _______
(  ____ \ |\     /| \__   __/ (       ) (  ____ \ (  ____ ) (  ___  )
| (    \/ | )   ( |    ) (    | () () | | (    \/ | (    )| | (   ) |
| |       | (___) |    | |    | || || | | (__     | (____)| | (___) |
| |       |  ___  |    | |    | |(_)| | |  __)    |     __) |  ___  |
| |       | (   ) |    | |    | |   | | | (       | (\ (    | (   ) |
| (____/\ | )   ( | ___) (___ | )   ( | | (____/\ | ) \ \__ | )   ( |
(_______/ |/     \| \_______/ |/     \| (_______/ |/   \__/ |/     \|
______________________________________________________________________
`

func main() {
	time.Sleep(200 * time.Millisecond)
	err := core.SetupCore()
	if err != nil {
		panic("\nCRITICAL: Failed to setup core system: " + err.Error())
	}
	time.Sleep(1 * time.Second)
	log := core.GetLogger()
	log.System("Startup", "🚀 CHIMERA System Boot Sequence Initiated...")

	if err := skills.ReloadTools(); err != nil {
		// Tool'lar yüklenemezse ajan konuşabilir ama iş yapamaz.
		// Bu kritik bir hata değildir, sadece uyarı veriyoruz.
		log.Error("⚠️ Skills loaded with errors (Agent will function without tools)", err)
	} else {
		tools := skills.GetTools()
		log.System("Startup", "✅ Skills loaded successfully", "count", len(tools))
	}

	log.System("Startup", "✅ Initializing Neural Interface (Brain)...")
	if err := agent.NewBrain(); err != nil {
		// Beyin yoksa ajan bitkisel hayattadır. Uygulamayı durduruyoruz.
		log.Error("❌ FATAL: Brain initialization failed", err)
		os.Exit(1)
	}
	fmt.Print(wellcomeMessage + "\n")
	time.Sleep(200 * time.Millisecond)

	log.System("Startup", "Starting Web Server and Communication Hub...")
	server.Start()

	waitForShutdown()
}

func waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	sig := <-quit
	log := core.GetLogger()
	log.System("SHUTDOWN", "🔻 Shutdown signal received", "signal", sig.String())
	if err := server.Shutdown(); err != nil {
		log.Error("SHUTDOWN: Error during shutdown", err)
	} else {
		log.System("SHUTDOWN", "✅ All systems offline. Goodbye.")
	}
	time.Sleep(500 * time.Millisecond)
}
