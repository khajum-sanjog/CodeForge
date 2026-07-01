package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"codeforge/internal/executor"
	"codeforge/internal/kzm"
	"codeforge/internal/logger"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	envFlag string
)

var runCmd = &cobra.Command{
	Use:   "run [file.kzm]",
	Short: "Execute a KZM pipeline once",
	Long:  `Parses, validates, and runs a pipeline described in a .kzm file directly, printing stdout output to CLI.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error: failed to read file %q: %v\n", filePath, err)
			os.Exit(1)
		}

		sourceCode := string(data)

		// 1. Parse config
		lexer := kzm.NewLexer(sourceCode)
		tokens, err := lexer.Tokenize()
		if err != nil {
			fmt.Printf("Tokenize error: %v\n", err)
			os.Exit(1)
		}

		parser := kzm.NewParser(tokens)
		prog, err := parser.Parse()
		if err != nil {
			fmt.Printf("Parse error: %v\n", err)
			os.Exit(1)
		}

		// 2. Validate
		valRes := kzm.Validate(prog)
		if len(valRes.Errors) > 0 {
			for _, e := range valRes.Errors {
				fmt.Printf("Validation Error (line %d): %s\n", e.Line, e.Message)
			}
			os.Exit(1)
		}

		// 3. Setup workspace and logger
		sourceDir, _ := filepath.Abs(filepath.Dir(filePath))
		// Use standard ~/.codeforge/logs as destination for logger
		home, _ := os.UserHomeDir()
		logDir := filepath.Join(home, ".codeforge", "logs")

		l := logger.NewLogger(logDir)
		exec := executor.NewExecutor(l, envFlag)

		// 4. Run pipeline
		res := exec.Execute(context.Background(), prog, sourceDir)

		if !res.Success {
			color.Red("\nCodeForge ✗  Pipeline execution failed!")
			if res.Error != nil {
				fmt.Printf("Error: %v\n", res.Error)
			}
			os.Exit(1)
		}

		color.Green("\nCodeForge ✓  Pipeline executed successfully!")
	},
}

func init() {
	runCmd.Flags().StringVarP(&envFlag, "env", "e", "", "Target deployment environment override")
	rootCmd.AddCommand(runCmd)
}
