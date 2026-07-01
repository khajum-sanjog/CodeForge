//go:build windows

package cmd

import (
	"os"
	"os/exec"
)

// setDaemonProcess is a no-op on Windows (Setsid does not exist).
func setDaemonProcess(cmd *exec.Cmd) {}

// sendTermSignal kills the process on Windows (no SIGTERM concept).
func sendTermSignal(proc *os.Process) error {
	return proc.Kill()
}

// isProcessRunning on Windows: FindProcess always succeeds, so we return true.
// A more robust check would use OpenProcess, but this is sufficient for CLI use.
func isProcessRunning(proc *os.Process) bool {
	return true
}
