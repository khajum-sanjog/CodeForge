//go:build !headless

package cmd

import (
	"github.com/spf13/cobra"

	"codeforge/gui"
)

var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Launch the graphical desktop dashboard",
	Run: func(cmd *cobra.Command, args []string) {
		runGUI()
	},
}

func init() {
	rootCmd.AddCommand(guiCmd)
}

func runGUI() {
	app := gui.NewApp()
	app.Run()
}
