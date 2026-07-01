package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version is injected at build time using ldflags
	Version = "1.0.0"
)

var rootCmd = &cobra.Command{
	Use:   "codeforge",
	Short: "CodeForge CI/CD Daemon and Dashboard Client",
	Long: `CodeForge is a complete, production-ready CI/CD daemon with a custom DSL (.kzm files), 
a secure encrypted credential vault, and a beautiful Fyne desktop GUI dashboard.

If run without arguments, it will launch the GUI automatically.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Launch GUI by default if no subcommand is executed
		runGUI()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("CodeForge CLI Error: %v\n", err)
		os.Exit(1)
	}
}
