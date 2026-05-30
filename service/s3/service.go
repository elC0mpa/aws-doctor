// Package s3 provides a service for interacting with AWS S3.
package s3

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/elC0mpa/aws-doctor/model"
	"golang.org/x/sync/errgroup"
)

// NewService creates a new S3 service.
func NewService(awsconfig aws.Config) Service {
	client := s3.NewFromConfig(awsconfig)

	return &service{
		client: client,
	}
}

func (s *service) Name() string {
	return "s3"
}

func (s *service) TabName() string {
	return "S3"
}

func (s *service) Analyze(ctx context.Context, flags model.Flags) (model.ScopeResult, error) {
	start := time.Now()
	input := model.RenderWasteInput{}

	var errs []error

	unusedBuckets, uploads, err := s.GetS3Waste(ctx)
	if err != nil {
		errs = append(errs, err)
	} else {
		input.S3Buckets = unusedBuckets
		input.S3MultipartUploads = uploads
	}

	var finalErr error
	if len(errs) > 0 {
		finalErr = fmt.Errorf("s3 analyze errors: %w", errors.Join(errs...))
	}

	return model.ScopeResult{
		Scope:    s.Name(),
		Input:    input,
		Duration: time.Since(start),
		Err:      finalErr,
	}, nil
}

func (s *service) GetS3Waste(ctx context.Context) ([]model.S3BucketWasteInfo, []model.S3MultipartUploadWasteInfo, error) {
	var bucketsWithoutPolicy []model.S3BucketWasteInfo

	var bucketsWithMultipart []model.S3MultipartUploadWasteInfo

	var mu sync.Mutex

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10) // Limit concurrency to avoid hitting rate limits

	input := &s3.ListBucketsInput{}
	if region := s.client.Options().Region; region != "" {
		input.BucketRegion = &region
	}

	paginator := s3.NewListBucketsPaginator(s.client, input)

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list buckets: %w", err)
		}

		for _, bucket := range output.Buckets {
			bucketName := bucket.Name
			creationDate := bucket.CreationDate

			g.Go(func() error {
				// Check Lifecycle Policy
				_, err := s.client.GetBucketLifecycleConfiguration(ctx, &s3.GetBucketLifecycleConfigurationInput{
					Bucket: bucketName,
				})
				if err != nil {
					var apiErr smithy.APIError
					if errors.As(err, &apiErr) && apiErr.ErrorCode() == "NoSuchLifecycleConfiguration" {
						mu.Lock()

						bucketsWithoutPolicy = append(bucketsWithoutPolicy, model.S3BucketWasteInfo{
							BucketName:   aws.ToString(bucketName),
							CreationDate: aws.ToTime(creationDate),
							Reason:       "No lifecycle policy",
						})

						mu.Unlock()
					} else {
						return fmt.Errorf("failed to get lifecycle configuration for bucket %s: %w", aws.ToString(bucketName), err)
					}
				}

				// Check Incomplete Multipart Uploads
				uploadCount, err := s.countMultipartUploads(ctx, bucketName)
				if err != nil {
					return fmt.Errorf("failed to count multipart uploads for bucket %s: %w", aws.ToString(bucketName), err)
				}

				if uploadCount > 0 {
					mu.Lock()

					bucketsWithMultipart = append(bucketsWithMultipart, model.S3MultipartUploadWasteInfo{
						BucketName:  aws.ToString(bucketName),
						UploadCount: uploadCount,
					})

					mu.Unlock()
				}

				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		return nil, nil, err
	}

	return bucketsWithoutPolicy, bucketsWithMultipart, nil
}

func (s *service) countMultipartUploads(ctx context.Context, bucketName *string) (int, error) {
	paginator := s3.NewListMultipartUploadsPaginator(s.client, &s3.ListMultipartUploadsInput{
		Bucket: bucketName,
	})

	uploadCount := 0

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return 0, err
		}

		uploadCount += len(output.Uploads)
	}

	return uploadCount, nil
}
