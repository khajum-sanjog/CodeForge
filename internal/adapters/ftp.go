package adapters

import (
	"context"

	"codeforge/internal/kzm"
)

// FTPAdapter deploys assets using the host FTP transmission mechanism.
type FTPAdapter struct {
	cpanel *CPanelAdapter
}

// NewFTPAdapter returns a new FTPAdapter instance.
func NewFTPAdapter() *FTPAdapter {
	return &FTPAdapter{cpanel: NewCPanelAdapter()}
}

// Deploy delegates execution to the cPanel SFTP/FTP engine.
func (a *FTPAdapter) Deploy(ctx context.Context, target *kzm.DeployTarget, sourceDir string) error {
	return a.cpanel.Deploy(ctx, target, sourceDir)
}

// Rollback delegates execution to the cPanel SFTP/FTP engine.
func (a *FTPAdapter) Rollback(ctx context.Context, target *kzm.DeployTarget, sourceDir string, snapshotPath string) error {
	return a.cpanel.Rollback(ctx, target, sourceDir, snapshotPath)
}

// Status delegates execution to the cPanel SFTP/FTP engine.
func (a *FTPAdapter) Status(ctx context.Context, target *kzm.DeployTarget) (string, error) {
	return a.cpanel.Status(ctx, target)
}
