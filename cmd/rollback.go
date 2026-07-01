package cmd

import (
	"fmt"
	"net/http"
	"os"

	"codeforge/internal/env"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback [project]",
	Short: "Rollback a pipeline to its last successful deployment snapshot",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		project := args[0]

		url := fmt.Sprintf("%s/rollback/%s", env.GetAPIURL(daemonPort), project)
		resp, err := http.Post(url, "application/json", nil)
		if err != nil {
			color.Red("Error: failed to connect to daemon API: %v", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			color.Green("CodeForge ✓  Rollback triggered for project %q.", project)
		} else {
			color.Red("Error: failed to trigger rollback for %q (Status %d)", project, resp.StatusCode)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
}
