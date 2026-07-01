//go:build !windows

package cmd

import (
	"os"
	"os/exec"
	"syscall"
)

// setDaemonProcess detaches the child process into its own session (Unix/macOS only).
func setDaemonProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
}

// sendTermSignal sends SIGTERM to a process (Unix/macOS only).
func sendTermSignal(proc *os.Process) error {
	return proc.Signal(syscall.SIGTERM)
}

// isProcessRunning checks if a PID is alive by sending signal 0 (Unix/macOS only).
func isProcessRunning(proc *os.Process) bool {
	return proc.Signal(syscall.Signal(0)) == nil
}
