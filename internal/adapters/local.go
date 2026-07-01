package adapters

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"codeforge/internal/kzm"
)

// LocalAdapter manages deployments onto the local filesystem of the CI/CD server.
type LocalAdapter struct{}

// NewLocalAdapter returns a new LocalAdapter instance.
func NewLocalAdapter() *LocalAdapter {
	return &LocalAdapter{}
}

// Deploy copies files recursively from the source directory to the destination path specified in targets options.
func (a *LocalAdapter) Deploy(ctx context.Context, target *kzm.DeployTarget, sourceDir string) error {
	dest := target.Options["path"]
	if dest == "" {
		dest = target.Name
	}
	if dest == "" {
		return fmt.Errorf("local deploy target requires a destination path")
	}

	dest = resolvePath(dest)
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("failed to create destination dir: %w", err)
	}

	// Read options
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Skip VCS directories and dependencies
		rel, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if shouldSkip(rel) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		targetPath := filepath.Join(dest, rel)
		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		return copyFile(path, targetPath)
	})
}

// Rollback restores the files using a snapshot path.
func (a *LocalAdapter) Rollback(ctx context.Context, target *kzm.DeployTarget, sourceDir string, snapshotPath string) error {
	// For local rollback, we can simply Deploy from the snapshot directory
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

// Status returns the status of the local destination folder.
func (a *LocalAdapter) Status(ctx context.Context, target *kzm.DeployTarget) (string, error) {
	dest := target.Options["path"]
	if dest == "" {
		dest = target.Name
	}
	dest = resolvePath(dest)
	if _, err := os.Stat(dest); err != nil {
		return "INACTIVE (Path not found)", nil
	}
	return "ACTIVE", nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func shouldSkip(relPath string) bool {
	parts := strings.Split(relPath, string(filepath.Separator))
	for _, p := range parts {
		if p == ".git" || p == "node_modules" || p == "vendor" || p == ".kzm" || p == ".codeforge" {
			return true
		}
	}
	return false
}

func resolvePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
