package kzm

import (
	"fmt"
	"strings"
)

// KzmError holds details about a parsing or validation error, including line, message, and suggestion.
type KzmError struct {
	Line       int
	Message    string
	Suggestion string
}

// Error returns the basic string representation of the error.
func (e *KzmError) Error() string {
	return fmt.Sprintf("KZM Error on line %d: %s", e.Line, e.Message)
}

// Format returns a beautiful, formatted CLI error representation including the source code line and an underline.
func (e *KzmError) Format(sourceCode string) string {
	lines := strings.Split(sourceCode, "\n")
	sourceLine := ""
	if e.Line > 0 && e.Line <= len(lines) {
		sourceLine = lines[e.Line-1]
	}

	// Create a squiggly underline matching the length of the source line.
	// Trim leading space for the underline but keep it aligned with the code.
	trimmedLine := strings.TrimLeft(sourceLine, " \t")
	leadingSpaces := len(sourceLine) - len(trimmedLine)
	
	underline := strings.Repeat(" ", leadingSpaces) + strings.Repeat("~", max(len(trimmedLine), 3))

	formatted := fmt.Sprintf("CodeForge ✗  KZM Error on line %d:\n  %s\n\n  %d | %s\n    %s\n", e.Line, e.Message, e.Line, sourceLine, underline)
	if e.Suggestion != "" {
		formatted += fmt.Sprintf("\n  Suggestion: %s\n", e.Suggestion)
	}
	return formatted
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
