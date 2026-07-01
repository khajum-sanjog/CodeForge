package cmd

import (
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of pipelines in the running daemon (alias for daemon status)",
	Run: func(cmd *cobra.Command, args []string) {
		daemonStatusCmd.Run(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
