package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"codeforge/internal/env"
	"codeforge/internal/logger"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	tailFlag  bool
	limitFlag int
)

var logsCmd = &cobra.Command{
	Use:   "logs [project]",
	Short: "View or stream logs for a project pipeline",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		project := args[0]

		_, running := getDaemonStatus()

		if running {
			// Stream or query from API
			lastLineCount := 0
			seenLines := make(map[string]bool)

			for {
				url := fmt.Sprintf("%s/logs/%s?limit=%d", env.GetAPIURL(daemonPort), project, limitFlag)
				resp, err := http.Get(url)
				if err != nil {
					fmt.Printf("Error: API disconnected: %v\n", err)
					break
				}

				var result struct {
					Logs []string `json:"logs"`
				}

				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				_ = json.Unmarshal(body, &result)

				for _, line := range result.Logs {
					var entry logger.LogEntry
					if err := json.Unmarshal([]byte(line), &entry); err == nil {
						key := entry.Timestamp + "_" + entry.Message
						if !seenLines[key] {
							seenLines[key] = true
							printColorLine(entry)
						}
					} else {
						// Raw line fallback
						fmt.Println(line)
					}
				}

				if !tailFlag {
					break
				}
				lastLineCount = len(result.Logs)
				_ = lastLineCount // prevent compiler warning
				time.Sleep(2 * time.Second)
			}
		} else {
			// Read locally from ~/.codeforge/logs/
			home, _ := os.UserHomeDir()
			logDir := filepath.Join(home, ".codeforge", "logs")
			l := logger.NewLogger(logDir)

			files, err := l.GetLogFilesForProject(project)
			if err != nil || len(files) == 0 {
				color.Red("No logs found for project %q.", project)
				return
			}

			// Read last file
			lastFile := files[len(files)-1]
			color.Cyan("Reading logs from filesystem: %s\n", filepath.Base(lastFile))

			data, err := os.ReadFile(lastFile)
			if err != nil {
				fmt.Println("Error reading file:", err)
				return
			}

			lines := strings.Split(string(data), "\n")
			startIdx := 0
			if len(lines) > limitFlag {
				startIdx = len(lines) - limitFlag - 1
			}

			for i := startIdx; i < len(lines); i++ {
				line := strings.TrimSpace(lines[i])
				if line == "" {
					continue
				}

				var entry logger.LogEntry
				if err := json.Unmarshal([]byte(line), &entry); err == nil {
					printColorLine(entry)
				} else {
					fmt.Println(line)
				}
			}
		}
	},
}

func printColorLine(e logger.LogEntry) {
	var colorFn func(format string, a ...interface{}) string
	level := strings.ToUpper(e.Level)
	switch level {
	case "SUCCESS":
		colorFn = color.New(color.FgGreen).SprintfFunc()
	case "FAILED", "ERROR":
		colorFn = color.New(color.FgRed).SprintfFunc()
	case "WARNING":
		colorFn = color.New(color.FgYellow).SprintfFunc()
	default:
		colorFn = color.New(color.FgCyan).SprintfFunc()
	}

	prefix := fmt.Sprintf("[%s] [%s]", e.Timestamp, e.Level)
	fmt.Printf("%s %s\n", color.New(color.FgHiBlack).Sprint(prefix), colorFn(e.Message))
}

func init() {
	logsCmd.Flags().BoolVarP(&tailFlag, "tail", "t", false, "Stream logs continuously")
	logsCmd.Flags().IntVarP(&limitFlag, "limit", "n", 50, "Number of lines to read")
	rootCmd.AddCommand(logsCmd)
}
