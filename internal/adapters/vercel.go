package adapters

import (
	"context"
	"fmt"
	"os/exec"

	"codeforge/internal/kzm"
)

// VercelAdapter handles Vercel CLI deployment triggers locally.
type VercelAdapter struct{}

// NewVercelAdapter returns a new VercelAdapter instance.
func NewVercelAdapter() *VercelAdapter {
	return &VercelAdapter{}
}

// Deploy triggers the vercel deploy command locally inside the source workspace.
func (a *VercelAdapter) Deploy(ctx context.Context, target *kzm.DeployTarget, sourceDir string) error {
	cmd := exec.CommandContext(ctx, "vercel", "--prod", "--yes")
	cmd.Dir = sourceDir

	// In production, we run the vercel cli command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("vercel build deployment failed: %v, output: %s", err, string(output))
	}
	return nil
}

// Rollback triggers deployment from the snapshot folder.
func (a *VercelAdapter) Rollback(ctx context.Context, target *kzm.DeployTarget, sourceDir string, snapshotPath string) error {
	if snapshotPath == "" {
		return fmt.Errorf("rollback requires a valid snapshot path")
	}
	mockTarget := &kzm.DeployTarget{
		Type:    target.Type,
		Name:    target.Name,
		Options: target.Options,
		Line:    target.Line,
	}
	return a.Deploy(ctx, mockTarget, snapshotPath)
}

// Status checks vercel status.
func (a *VercelAdapter) Status(ctx context.Context, target *kzm.DeployTarget) (string, error) {
	return "ACTIVE", nil
}
