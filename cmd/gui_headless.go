//go:build headless

package cmd

import (
	"fmt"
	"os"
)

func runGUI() {
	fmt.Println("Error: GUI dashboard is disabled in headless server builds.")
	fmt.Println("Please run codeforge as a daemon (e.g. 'codeforge daemon start') or run a pipeline directly (e.g. 'codeforge run <file.kzm>').")
	os.Exit(1)
}
