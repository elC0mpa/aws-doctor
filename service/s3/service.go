// Package s3 provides a service for interacting with AWS S3.
package s3

import (
	"context"
	"errors"
	"fmt"
	"sync"

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

func (s *service) GetS3Waste(ctx context.Context) ([]model.S3BucketWasteInfo, []model.S3MultipartUploadWasteInfo, error) {
	var bucketsWithoutPolicy []model.S3BucketWasteInfo

	var bucketsWithMultipart []model.S3MultipartUploadWasteInfo

	var mu sync.Mutex

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10) // Limit concurrency to avoid hitting rate limits

	paginator := s3.NewListBucketsPaginator(s.client, &s3.ListBucketsInput{})

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
					if errors.As(err, &apiErr) {
						if apiErr.ErrorCode() == "NoSuchLifecycleConfiguration" {
							mu.Lock()

							bucketsWithoutPolicy = append(bucketsWithoutPolicy, model.S3BucketWasteInfo{
								BucketName:   aws.ToString(bucketName),
								CreationDate: aws.ToTime(creationDate),
								Reason:       "No lifecycle policy",
							})

							mu.Unlock()
						}
					}
				}

				// Check Incomplete Multipart Uploads
				uploadCount, err := s.countMultipartUploads(ctx, bucketName)
				if err == nil && uploadCount > 0 {
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
