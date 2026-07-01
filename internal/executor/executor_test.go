package executor

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"codeforge/internal/kzm"
	"codeforge/internal/logger"
)

func TestExecutorBasic(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "codeforge-executor-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logDir := filepath.Join(tempDir, "logs")
	l := logger.NewLogger(logDir)
	execInstance := NewExecutor(l, "production")

	// Create dummy files in workspace
	workspaceDir := filepath.Join(tempDir, "workspace")
	_ = os.MkdirAll(workspaceDir, 0755)
	_ = os.WriteFile(filepath.Join(workspaceDir, "test.txt"), []byte("hello world"), 0644)

	prog := &kzm.Program{
		Line: 1,
		Meta: &kzm.Meta{
			Name:    "TestProject",
			Version: "1.0",
		},
		Variables: []*kzm.Variable{
			{Key: "MY_VAR", Value: "hello", Line: 2},
		},
		Before: &kzm.Phase{
			Line: 4,
			Steps: []*kzm.Step{
				{
					Command:  `run: echo "$MY_VAR"`,
					MustPass: true,
					Line:     5,
				},
				{
					Command:  `run: echo "skipped step"`,
					MustPass: true,
					Condition: &kzm.Condition{
						EnvValue: "staging", // Should skip since active env is production
					},
					Line: 6,
				},
			},
		},
		Deploy: &kzm.DeployTarget{
			Type: "local",
			Name: "test-deploy",
			Options: map[string]string{
				"path": filepath.Join(tempDir, "deployed"),
			},
			Line: 10,
		},
	}

	res := execInstance.Execute(context.Background(), prog, workspaceDir)
	if !res.Success {
		t.Fatalf("pipeline execution failed: %v", res.Error)
	}

	if len(res.Steps) != 2 { // Before step 1 + Deploy target step (Before step 2 was skipped)
		t.Errorf("expected 2 steps completed in results, got %d", len(res.Steps))
		for i, s := range res.Steps {
			t.Logf("  step [%d]: %s (Success: %t, Output: %q)", i, s.Command, s.Success, s.Output)
		}
	}

	// Verify local deploy target created output
	deployedFile := filepath.Join(tempDir, "deployed", "test.txt")
	if _, err := os.Stat(deployedFile); err != nil {
		t.Errorf("expected local deploy target to copy test.txt, but was missing: %v", err)
	}
}
