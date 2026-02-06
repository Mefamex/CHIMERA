// file path : internal/core/initRuntimes.go
package core

import (
	"fmt"
	"os/exec"
)

func InitRuntimes() error {
	ver, err := checkPython()
	if err != nil {
		Log.Error("Python runtime check failed\n", "error", err, "component", "Runtime")
	} else {
		Log.Info("✅ Python runtime ", "version", ver, "component", "Runtime")
	}
	return err
}

func checkPython() (string, error) {
	executable := Config.Runtimes.Python
	if executable == "" {
		return "", fmt.Errorf("python executable not configured (Config.Runtime.Python is empty)")
	}
	cmd := exec.Command(executable, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Python is not installed or not in PATH: %w", err)
	}
	return string(output), nil
}
