package adapters

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"codeforge/internal/kzm"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaTypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// LambdaAdapter packages files as a ZIP and deploys them to AWS Lambda.
type LambdaAdapter struct{}

// NewLambdaAdapter returns a new LambdaAdapter instance.
func NewLambdaAdapter() *LambdaAdapter {
	return &LambdaAdapter{}
}

// Deploy zips the source directory, uploads to Lambda, updates configs, and publishes a version.
func (a *LambdaAdapter) Deploy(ctx context.Context, target *kzm.DeployTarget, sourceDir string) error {
	functionName := target.Name
	region := target.Options["region"]
	if region == "" {
		region = "us-east-1"
	}

	zipBytes, err := buildZipInMemory(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to build function zip archive: %w", err)
	}

	cfg, err := getAWSConfig(ctx, target, region)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	lambdaClient := lambda.NewFromConfig(cfg)

	// Update Lambda Function Code
	_, err = lambdaClient.UpdateFunctionCode(ctx, &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(functionName),
		ZipFile:      zipBytes,
	})
	if err != nil {
		return fmt.Errorf("failed to upload function code: %w", err)
	}

	// Construct configuration updates
	cfgInput := &lambda.UpdateFunctionConfigurationInput{
		FunctionName: aws.String(functionName),
	}
	hasCfgUpdates := false

	if runtimeVal, ok := target.Options["runtime"]; ok && runtimeVal != "" {
		cfgInput.Runtime = lambdaTypes.Runtime(runtimeVal)
		hasCfgUpdates = true
	}

	if memoryVal, ok := target.Options["memory"]; ok && memoryVal != "" {
		if memInt, err := strconv.Atoi(memoryVal); err == nil {
			cfgInput.MemorySize = aws.Int32(int32(memInt))
			hasCfgUpdates = true
		}
	}

	if timeoutVal, ok := target.Options["timeout"]; ok && timeoutVal != "" {
		if timeoutInt, err := strconv.Atoi(timeoutVal); err == nil {
			cfgInput.Timeout = aws.Int32(int32(timeoutInt))
			hasCfgUpdates = true
		}
	}

	if len(target.EnvVars) > 0 {
		envVars := make(map[string]string)
		for k, v := range target.EnvVars {
			envVars[k] = v
		}
		cfgInput.Environment = &lambdaTypes.Environment{
			Variables: envVars,
		}
		hasCfgUpdates = true
	}

	if hasCfgUpdates {
		_, err = lambdaClient.UpdateFunctionConfiguration(ctx, cfgInput)
		if err != nil {
			return fmt.Errorf("failed to update function configurations: %w", err)
		}
	}

	// Publish new Lambda version
	_, err = lambdaClient.PublishVersion(ctx, &lambda.PublishVersionInput{
		FunctionName: aws.String(functionName),
	})
	if err != nil {
		return fmt.Errorf("failed to publish function version: %w", err)
	}

	return nil
}

// Rollback deploys from a snapshot zip path.
func (a *LambdaAdapter) Rollback(ctx context.Context, target *kzm.DeployTarget, sourceDir string, snapshotPath string) error {
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

// Status checks lambda details.
func (a *LambdaAdapter) Status(ctx context.Context, target *kzm.DeployTarget) (string, error) {
	region := target.Options["region"]
	if region == "" {
		region = "us-east-1"
	}
	cfg, err := getAWSConfig(ctx, target, region)
	if err != nil {
		return "ERROR", err
	}
	lambdaClient := lambda.NewFromConfig(cfg)
	res, err := lambdaClient.GetFunction(ctx, &lambda.GetFunctionInput{
		FunctionName: aws.String(target.Name),
	})
	if err != nil {
		return "INACTIVE/NOT FOUND", nil
	}
	return string(res.Configuration.State), nil
}

func buildZipInMemory(sourceDir string) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		if shouldSkip(rel) {
			return nil
		}

		w, err := zw.Create(filepath.ToSlash(rel))
		if err != nil {
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(w, f)
		return err
	})
	if err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
