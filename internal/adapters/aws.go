package adapters

import (
	"context"

	"codeforge/internal/kzm"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

// getAWSConfig loads the AWS configuration, using static credentials if provided in target options,
// or falling back to the standard AWS credential chain (e.g., IAM Instance/Task roles).
func getAWSConfig(ctx context.Context, target *kzm.DeployTarget, region string) (aws.Config, error) {
	var optFns []func(*config.LoadOptions) error
	if region != "" {
		optFns = append(optFns, config.WithRegion(region))
	}

	accessKey := target.Options["access_key"]
	if accessKey == "" {
		accessKey = target.Options["aws_access_key"]
	}
	secretKey := target.Options["secret_key"]
	if secretKey == "" {
		secretKey = target.Options["aws_secret_key"]
	}

	if accessKey != "" && secretKey != "" {
		optFns = append(optFns, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		))
	}

	return config.LoadDefaultConfig(ctx, optFns...)
}
