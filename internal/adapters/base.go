package adapters

import (
	"context"
	"fmt"
	"strings"

	"codeforge/internal/kzm"
)

// Adapter specifies the actions required to deploy, rollback, and check health on target systems.
type Adapter interface {
	Deploy(ctx context.Context, target *kzm.DeployTarget, sourceDir string) error
	Rollback(ctx context.Context, target *kzm.DeployTarget, sourceDir string, snapshotPath string) error
	Status(ctx context.Context, target *kzm.DeployTarget) (string, error)
}

// GetAdapter instantiates the corresponding Adapter by target type.
func GetAdapter(targetType string) (Adapter, error) {
	switch strings.ToLower(targetType) {
	case "local":
		return NewLocalAdapter(), nil
	case "ssh":
		return NewSSHAdapter(), nil
	case "s3":
		return NewS3Adapter(), nil
	case "lambda":
		return NewLambdaAdapter(), nil
	case "cpanel":
		return NewCPanelAdapter(), nil
	case "docker":
		return NewDockerAdapter(), nil
	case "vps":
		return NewVPSAdapter(), nil
	case "ftp":
		return NewFTPAdapter(), nil
	case "vercel":
		return NewVercelAdapter(), nil
	default:
		return nil, fmt.Errorf("unsupported deploy target type %q", targetType)
	}
}
