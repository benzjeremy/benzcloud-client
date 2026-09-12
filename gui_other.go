//go:build (!linux && !windows) || (linux && !cgo)

package main

import (
	"log"
	"os/exec"
	"runtime"
)

// LaunchGUI opens default browser as fallback.
func LaunchGUI(title, url string, width, height int) {
	log.Printf("[BenzCloud Client] Opening interface at %s\n", url)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
