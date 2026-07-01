package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"codeforge/internal/daemon"
	"codeforge/internal/env"
	"codeforge/internal/logger"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	daemonPort int
	workers    int
	configDir  string
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Manage the background CodeForge daemon",
}

var daemonRunCmd = &cobra.Command{
	Use:    "run",
	Short:  "Run the daemon in the foreground",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		home, _ := os.UserHomeDir()
		logDir := filepath.Join(home, ".codeforge", "logs")
		l := logger.NewLogger(logDir)

		d := daemon.NewDaemon(configDir, daemonPort, workers, l)
		err := d.Start()
		if err != nil {
			l.Log("daemon", "ERROR", "Failed to start daemon: %v", err)
			os.Exit(1)
		}

		// Wait for SIGINT or SIGTERM
		c := make(chan os.Signal, 1)
		// bind signal notify
		// We'll write standard loop waiting for interrupt
		// In a CLI command, blocking until interrupt is standard
		<-c
		d.Stop()
	},
}

var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the CodeForge daemon in the background",
	Run: func(cmd *cobra.Command, args []string) {
		pid, running := getDaemonStatus()
		if running {
			color.Yellow("CodeForge daemon is already running (PID: %d)", pid)
			return
		}

		// Spawn detached process
		// We run: codeforge daemon run --port P --workers W --config C
		selfPath, err := os.Executable()
		if err != nil {
			selfPath = "codeforge"
		}

		argsList := []string{"daemon", "run",
			"--port", strconv.Itoa(daemonPort),
			"--workers", strconv.Itoa(workers),
			"--config", configDir,
		}

		backgroundCmd := exec.Command(selfPath, argsList...)
		backgroundCmd.Stdout = nil
		backgroundCmd.Stderr = nil
		// Detach process on Unix/macOS (no-op on Windows)
		setDaemonProcess(backgroundCmd)

		err = backgroundCmd.Start()
		if err != nil {
			color.Red("Error: failed to start background daemon: %v", err)
			os.Exit(1)
		}

		// Wait briefly to check if it started
		time.Sleep(500 * time.Millisecond)
		pid, running = getDaemonStatus()
		if running {
			color.Green("CodeForge daemon started successfully in background (PID: %d)", pid)
		} else {
			color.Red("Error: daemon started but exited immediately. Check logs.")
		}
	},
}

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the background CodeForge daemon",
	Run: func(cmd *cobra.Command, args []string) {
		pid, running := getDaemonStatus()
		if !running {
			color.Yellow("CodeForge daemon is not running.")
			return
		}

		proc, err := os.FindProcess(pid)
		if err != nil {
			color.Red("Error: process %d not found", pid)
			return
		}

		color.Cyan("Stopping CodeForge daemon (PID: %d)...", pid)
		err = sendTermSignal(proc)
		if err != nil {
			// fallback
			_ = proc.Kill()
		}

		// Wait for PID file cleanup
		home, _ := os.UserHomeDir()
		pidPath := filepath.Join(home, ".codeforge", "daemon.pid")
		for i := 0; i < 20; i++ {
			if _, err := os.Stat(pidPath); os.IsNotExist(err) {
				break
			}
			time.Sleep(250 * time.Millisecond)
		}

		color.Green("CodeForge daemon stopped.")
	},
}

var daemonRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the background CodeForge daemon",
	Run: func(cmd *cobra.Command, args []string) {
		daemonStopCmd.Run(cmd, args)
		time.Sleep(1 * time.Second)
		daemonStartCmd.Run(cmd, args)
	},
}

var daemonStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of pipelines in the running daemon",
	Run: func(cmd *cobra.Command, args []string) {
		pid, running := getDaemonStatus()
		if !running {
			color.Red("Daemon status: ● Stopped")
			return
		}

		color.Green("Daemon status: ● Running (PID: %d)\n", pid)

		// Query HTTP status
		url := fmt.Sprintf("%s/status", env.GetAPIURL(daemonPort))
		resp, err := http.Get(url)
		if err != nil {
			color.Yellow("Could not retrieve pipelines status table (API disconnected)")
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		type PipelineStatus struct {
			Project      string `json:"project"`
			Status       string `json:"status"`
			LastRun      string `json:"last_run"`
			LastDuration string `json:"last_duration"`
		}

		var list []PipelineStatus
		if err := json.Unmarshal(body, &list); err != nil {
			fmt.Println("Error parsing status JSON:", err)
			return
		}

		if len(list) == 0 {
			fmt.Println("No pipelines registered.")
			return
		}

		fmt.Println("Registered Pipelines:")
		fmt.Printf("%-20s %-12s %-25s %-15s\n", "PROJECT", "STATUS", "LAST RUN", "DURATION")
		fmt.Println(strings.Repeat("-", 75))
		for _, item := range list {
			var coloredStatus string
			switch strings.ToUpper(item.Status) {
			case "SUCCESS":
				coloredStatus = color.GreenString("SUCCESS")
			case "FAILED":
				coloredStatus = color.RedString("FAILED")
			case "ROLLBACK":
				coloredStatus = color.YellowString("ROLLBACK")
			case "RUNNING":
				coloredStatus = color.CyanString("RUNNING")
			default:
				coloredStatus = color.HiBlackString(item.Status)
			}

			lastRunFormatted := item.LastRun
			if t, err := time.Parse(time.RFC3339, item.LastRun); err == nil {
				lastRunFormatted = t.Format("2006-01-02 15:04:05")
			} else {
				lastRunFormatted = "never"
			}

			fmt.Printf("%-20s %-12s %-25s %-15s\n", item.Project, coloredStatus, lastRunFormatted, item.LastDuration)
		}
	},
}

func getDaemonStatus() (int, bool) {
	home, _ := os.UserHomeDir()
	pidPath := filepath.Join(home, ".codeforge", "daemon.pid")
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return 0, false
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, false
	}

	// Check if process is running
	process, err := os.FindProcess(pid)
	if err != nil {
		return 0, false
	}

	// Check if process is alive (platform-specific)
	if isProcessRunning(process) {
		return pid, true
	}

	return 0, false
}

func init() {
	home, _ := os.UserHomeDir()
	defaultCfg := filepath.Join(home, ".codeforge")

	defaultPort := env.GetPort(7080)
	daemonCmd.PersistentFlags().IntVarP(&daemonPort, "port", "p", defaultPort, "HTTP API port")
	daemonCmd.PersistentFlags().IntVarP(&workers, "workers", "w", 3, "Number of worker threads")
	daemonCmd.PersistentFlags().StringVarP(&configDir, "config", "c", defaultCfg, "Configuration directory path")

	daemonCmd.AddCommand(daemonRunCmd)
	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonStopCmd)
	daemonCmd.AddCommand(daemonRestartCmd)
	daemonCmd.AddCommand(daemonStatusCmd)

	rootCmd.AddCommand(daemonCmd)
}
