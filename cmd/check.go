package cmd

import (
	"fmt"
	"os"

	"codeforge/internal/kzm"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check [file.kzm]",
	Short: "Validate a KZM pipeline configuration file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error: failed to read file %q: %v\n", filePath, err)
			os.Exit(1)
		}

		sourceCode := string(data)

		// 1. Tokenize
		lexer := kzm.NewLexer(sourceCode)
		tokens, err := lexer.Tokenize()
		if err != nil {
			kzmErr := &kzm.KzmError{
				Line:    1, // approximate line
				Message: err.Error(),
			}
			fmt.Println(kzmErr.Format(sourceCode))
			os.Exit(1)
		}

		// 2. Parse
		parser := kzm.NewParser(tokens)
		prog, err := parser.Parse()
		if err != nil {
			kzmErr := &kzm.KzmError{
				Line:    1, // Parse errors can contain line info if we parsed it, or approximate
				Message: err.Error(),
			}
			// Extract line number if present in parse error message
			var parsedLine int
			_, _ = fmt.Sscanf(err.Error(), "expected %*s on line %d", &parsedLine)
			if parsedLine > 0 {
				kzmErr.Line = parsedLine
			}
			fmt.Println(kzmErr.Format(sourceCode))
			os.Exit(1)
		}

		// 3. Validate AST
		res := kzm.Validate(prog)

		if len(res.Errors) > 0 {
			for _, e := range res.Errors {
				kzmErr := &kzm.KzmError{
					Line:    e.Line,
					Message: e.Message,
				}
				fmt.Println(kzmErr.Format(sourceCode))
			}
			os.Exit(1)
		}

		if len(res.Warnings) > 0 {
			color.Yellow("Warnings:")
			for _, w := range res.Warnings {
				fmt.Printf("  Line %d: %s\n", w.Line, w.Message)
			}
		}

		color.Green("CodeForge ✓  Configuration %q is valid!", filePath)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
