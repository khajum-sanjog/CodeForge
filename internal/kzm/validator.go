package kzm

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ValidationError represents a syntax or configuration error found in the KZM program.
type ValidationError struct {
	Line    int
	Message string
}

// ValidationWarning represents a code smell or non-critical issue in the KZM program.
type ValidationWarning struct {
	Line    int
	Message string
}

// ValidationResult accumulates errors and warnings generated during static analysis.
type ValidationResult struct {
	Errors   []ValidationError
	Warnings []ValidationWarning
}

var awsRegionRegex = regexp.MustCompile(`^[a-z]{2,4}-[a-z]{3,12}-\d$`)

// Validate checks the program AST for syntax issues, variable references, and configuration correctness.
func Validate(prog *Program) ValidationResult {
	res := ValidationResult{
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
	}

	if prog == nil {
		res.Errors = append(res.Errors, ValidationError{Line: 0, Message: "Program is empty"})
		return res
	}

	// 1. At least one trigger
	if len(prog.Triggers) == 0 {
		res.Errors = append(res.Errors, ValidationError{
			Line:    prog.Line,
			Message: "No triggers specified (must have at least one watch, every, or trigger block)",
		})
	}

	// Track defined variables
	definedVars := make(map[string]bool)
	for _, v := range prog.Variables {
		definedVars[v.Key] = true
	}

	// 2. Validate Phase Steps
	validatePhase := func(phase *Phase, phaseName string) {
		if phase == nil {
			return
		}
		for _, step := range phase.Steps {
			// Check for empty command
			cmdContent := strings.TrimPrefix(step.Command, "run: ")
			cmdContent = strings.TrimPrefix(cmdContent, "copy: ")
			cmdContent = strings.TrimPrefix(cmdContent, "set: ")
			cmdContent = strings.TrimPrefix(cmdContent, "plugin: ")
			if strings.TrimSpace(cmdContent) == "" {
				res.Errors = append(res.Errors, ValidationError{
					Line:    step.Line,
					Message: fmt.Sprintf("Empty command in %s phase", phaseName),
				})
			}

			// Check variable references like $VAR
			checkVarRefs(cmdContent, step.Line, definedVars, &res)
		}
	}

	validatePhase(prog.Before, "before deploy")
	validatePhase(prog.After, "after deploy")

	// 3. Validate Deploy Target
	if prog.Deploy != nil {
		validateDeployTarget(prog.Deploy, &res)
	}

	// Validate Environment Targets
	for _, env := range prog.Environments {
		if env.Target != nil {
			validateDeployTarget(env.Target, &res)
		}
	}

	return res
}

func validateDeployTarget(target *DeployTarget, res *ValidationResult) {
	validTargets := []string{"ssh", "lambda", "cpanel", "s3", "docker", "vps", "local", "ftp", "vercel"}

	// Check target type
	isTargetValid := false
	lowerType := strings.ToLower(target.Type)
	for _, vt := range validTargets {
		if lowerType == vt {
			isTargetValid = true
			break
		}
	}

	if !isTargetValid {
		// Suggest typo using Levenshtein distance
		bestMatch := ""
		minDist := 999
		for _, vt := range validTargets {
			dist := levenshtein(lowerType, vt)
			if dist < minDist {
				minDist = dist
				bestMatch = vt
			}
		}
		msg := fmt.Sprintf("Unknown deploy target type %q", target.Type)
		if minDist <= 2 {
			msg += fmt.Sprintf(". Did you mean %q?", bestMatch)
		}
		res.Errors = append(res.Errors, ValidationError{
			Line:    target.Line,
			Message: msg,
		})
		return
	}

	// Target-specific checks
	switch lowerType {
	case "ssh":
		// Option validation
		if _, ok := target.Options["key"]; !ok {
			res.Warnings = append(res.Warnings, ValidationWarning{
				Line:    target.Line,
				Message: "SSH target should specify a private 'key' path; defaulting to system default SSH keys",
			})
		}
	case "lambda":
		// Requires region, runtime, memory, timeout
		if reg, ok := target.Options["region"]; !ok || reg == "" {
			res.Errors = append(res.Errors, ValidationError{
				Line:    target.Line,
				Message: "Lambda target requires 'region' specification",
			})
		} else if !awsRegionRegex.MatchString(reg) {
			res.Errors = append(res.Errors, ValidationError{
				Line:    target.Line,
				Message: fmt.Sprintf("Invalid AWS region format: %q", reg),
			})
		}

		if _, ok := target.Options["runtime"]; !ok {
			res.Errors = append(res.Errors, ValidationError{
				Line:    target.Line,
				Message: "Lambda target requires 'runtime' specification (e.g. 'nodejs20.x')",
			})
		}

		if memStr, ok := target.Options["memory"]; ok {
			if mem, err := strconv.Atoi(memStr); err != nil || mem < 128 || mem > 10240 {
				res.Errors = append(res.Errors, ValidationError{
					Line:    target.Line,
					Message: "Lambda memory must be a number between 128 and 10240",
				})
			}
		}
		if toStr, ok := target.Options["timeout"]; ok {
			if to, err := strconv.Atoi(toStr); err != nil || to < 1 || to > 900 {
				res.Errors = append(res.Errors, ValidationError{
					Line:    target.Line,
					Message: "Lambda timeout must be a number between 1 and 900 seconds",
				})
			}
		}
	case "s3":
		if reg, ok := target.Options["region"]; ok && reg != "" {
			if !awsRegionRegex.MatchString(reg) {
				res.Errors = append(res.Errors, ValidationError{
					Line:    target.Line,
					Message: fmt.Sprintf("Invalid S3 region format: %q", reg),
				})
			}
		}
		if _, ok := target.Options["folder"]; !ok {
			res.Errors = append(res.Errors, ValidationError{
				Line:    target.Line,
				Message: "S3 target requires 'folder' specification for files to upload",
			})
		}
	case "cpanel":
		if _, ok := target.Options["user"]; !ok {
			res.Errors = append(res.Errors, ValidationError{
				Line:    target.Line,
				Message: "cPanel target requires 'user' specification",
			})
		}
	case "docker":
		if _, ok := target.Options["image"]; !ok {
			res.Errors = append(res.Errors, ValidationError{
				Line:    target.Line,
				Message: "Docker target requires 'image' specification",
			})
		}
	}
}

func checkVarRefs(content string, line int, definedVars map[string]bool, res *ValidationResult) {
	// Simple regex to find $VAR patterns, ignoring $secret.VAR patterns
	varRegex := regexp.MustCompile(`\$([A-Za-z0-9_]+)`)
	matches := varRegex.FindAllStringSubmatch(content, -1)
	for _, m := range matches {
		varName := m[1]
		// Skip special secrets prefix or system env vars like $PATH
		if varName == "secret" || strings.HasPrefix(varName, "PATH") || varName == "HOME" || varName == "USER" {
			continue
		}
		// If it's a dotted secret $secret.VAR, it matches $secret, which we skipped.
		// If it's a regular variable, verify it is defined.
		if !definedVars[varName] {
			res.Warnings = append(res.Warnings, ValidationWarning{
				Line:    line,
				Message: fmt.Sprintf("Reference to undefined variable $%s. If this is a system environment variable, ignore this warning.", varName),
			})
		}
	}
}

// levenshtein calculates the Levenshtein distance between two strings.
func levenshtein(a, b string) int {
	la := len(a)
	lb := len(b)
	d := make([][]int, la+1)
	for i := range d {
		d[i] = make([]int, lb+1)
	}

	for i := 0; i <= la; i++ {
		d[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		d[0][j] = j
	}

	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			d[i][j] = min(
				d[i-1][j]+1, // deletion
				min(
					d[i][j-1]+1, // insertion
					d[i-1][j-1]+cost, // substitution
				),
			)
		}
	}
	return d[la][lb]
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
