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

func (s *service) GetBucketsWithoutLifecyclePolicies(ctx context.Context) ([]model.S3BucketWasteInfo, error) {
	var results []model.S3BucketWasteInfo

	var mu sync.Mutex

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10) // Limit concurrency to avoid hitting rate limits

	paginator := s3.NewListBucketsPaginator(s.client, &s3.ListBucketsInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list buckets: %w", err)
		}

		for _, bucket := range output.Buckets {
			g.Go(func() error {
				_, err := s.client.GetBucketLifecycleConfiguration(ctx, &s3.GetBucketLifecycleConfigurationInput{
					Bucket: bucket.Name,
				})
				if err != nil {
					var apiErr smithy.APIError
					if errors.As(err, &apiErr) {
						if apiErr.ErrorCode() == "NoSuchLifecycleConfiguration" {
							mu.Lock()

							results = append(results, model.S3BucketWasteInfo{
								BucketName:   aws.ToString(bucket.Name),
								CreationDate: aws.ToTime(bucket.CreationDate),
								Reason:       "No lifecycle policy",
							})

							mu.Unlock()

							return nil
						}
					}
					// Other errors might be permissions or connectivity issues
					// We don't want to fail the whole process if one bucket fails.
					return nil
				}

				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}
