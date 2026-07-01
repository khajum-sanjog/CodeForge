package adapters

import (
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"codeforge/internal/kzm"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	cfTypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Adapter deploys files and folders to an Amazon S3 Bucket.
type S3Adapter struct{}

// NewS3Adapter returns a new S3Adapter instance.
func NewS3Adapter() *S3Adapter {
	return &S3Adapter{}
}

// Deploy uploads local build artifacts to the S3 bucket and triggers optional CloudFront invalidations.
func (a *S3Adapter) Deploy(ctx context.Context, target *kzm.DeployTarget, sourceDir string) error {
	bucketName := target.Name
	folderOption := target.Options["folder"]
	if folderOption == "" {
		folderOption = "."
	}

	uploadDir := folderOption
	if !filepath.IsAbs(uploadDir) {
		uploadDir = filepath.Join(sourceDir, folderOption)
	}

	region := target.Options["region"]
	if region == "" {
		region = "us-east-1"
	}

	cfg, err := getAWSConfig(ctx, target, region)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	// Walk directory and upload files
	err = filepath.Walk(uploadDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}

		rel, err := filepath.Rel(uploadDir, path)
		if err != nil {
			return err
		}
		if shouldSkip(rel) {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		keyName := filepath.ToSlash(rel)
		contentType := mime.TypeByExtension(filepath.Ext(path))
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		acl := s3Types.ObjectCannedACLPrivate
		if val, ok := target.Options["public"]; ok && (val == "yes" || val == "true") {
			acl = s3Types.ObjectCannedACLPublicRead
		}

		_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(bucketName),
			Key:         aws.String(keyName),
			Body:        file,
			ContentType: aws.String(contentType),
			ACL:         acl,
		})
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to upload files to S3: %w", err)
	}

	// Optional CloudFront Invalidation
	invalidateCloudFront := false
	if val, ok := target.Options["invalidate"]; ok && strings.ToLower(val) == "cloudfront" {
		invalidateCloudFront = true
	}
	if val, ok := target.Options["cloudfront"]; ok && (val == "yes" || val == "true") {
		invalidateCloudFront = true
	}

	if invalidateCloudFront {
		distID := target.Options["distribution"]
		if distID == "" {
			// Fail or issue a warning if distribution is missing
			return fmt.Errorf("CloudFront invalidation requested but no 'distribution' ID was provided")
		}

		cfClient := cloudfront.NewFromConfig(cfg)
		_, err = cfClient.CreateInvalidation(ctx, &cloudfront.CreateInvalidationInput{
			DistributionId: aws.String(distID),
			InvalidationBatch: &cfTypes.InvalidationBatch{
				CallerReference: aws.String(fmt.Sprintf("codeforge-%d", time.Now().UnixNano())),
				Paths: &cfTypes.Paths{
					Quantity: aws.Int32(1),
					Items:    []string{"/*"},
				},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create CloudFront invalidation: %w", err)
		}
	}

	return nil
}

// Rollback uploads from the snapshot directory to S3.
func (a *S3Adapter) Rollback(ctx context.Context, target *kzm.DeployTarget, sourceDir string, snapshotPath string) error {
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

// Status checks bucket accessibility.
func (a *S3Adapter) Status(ctx context.Context, target *kzm.DeployTarget) (string, error) {
	region := target.Options["region"]
	if region == "" {
		region = "us-east-1"
	}
	cfg, err := getAWSConfig(ctx, target, region)
	if err != nil {
		return "ERROR", err
	}
	s3Client := s3.NewFromConfig(cfg)
	_, err = s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(target.Name),
	})
	if err != nil {
		return "INACCESSIBLE", nil
	}
	return "ACTIVE", nil
}
