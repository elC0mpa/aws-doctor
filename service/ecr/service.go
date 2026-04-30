// Package ecr provides a service for interacting with AWS ECR.
package ecr

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/smithy-go"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/pricing"
	"golang.org/x/sync/errgroup"
)

// NewService creates a new ECR service.
func NewService(awsconfig aws.Config, pricingService pricing.Service) Service {
	client := ecr.NewFromConfig(awsconfig)

	return &service{
		client:         client,
		pricingService: pricingService,
	}
}

func (s *service) GetECRWaste(ctx context.Context) ([]model.ECRNoLifecyclePolicyInfo, []model.ECREmptyRepositoryInfo, []model.ECRUntaggedImageInfo, error) {
	var (
		noPolicy       []model.ECRNoLifecyclePolicyInfo
		emptyRepos     []model.ECREmptyRepositoryInfo
		untaggedImages []model.ECRUntaggedImageInfo
		mu             sync.Mutex
	)

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10) // Limit concurrency to avoid hitting rate limits

	paginator := ecr.NewDescribeRepositoriesPaginator(s.client, &ecr.DescribeRepositoriesInput{})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to describe repositories: %w", err)
		}

		for _, repo := range output.Repositories {
			repoName := repo.RepositoryName

			g.Go(func() error {
				// Check Lifecycle Policy
				_, err := s.client.GetLifecyclePolicy(ctx, &ecr.GetLifecyclePolicyInput{
					RepositoryName: repoName,
				})
				if err != nil {
					var apiErr smithy.APIError
					if errors.As(err, &apiErr) && apiErr.ErrorCode() == "LifecyclePolicyNotFoundException" {
						mu.Lock()

						noPolicy = append(noPolicy, model.ECRNoLifecyclePolicyInfo{RepositoryName: aws.ToString(repoName)})
						mu.Unlock()
					} else {
						return fmt.Errorf("failed to get lifecycle policy for repo %s: %w", aws.ToString(repoName), err)
					}
				}

				// Check Images (Untagged and Empty)
				imagePaginator := ecr.NewDescribeImagesPaginator(s.client, &ecr.DescribeImagesInput{
					RepositoryName: repoName,
				})

				var (
					totalImages   int
					untaggedCount int
					untaggedSize  int64
				)

				for imagePaginator.HasMorePages() {
					imageOutput, err := imagePaginator.NextPage(ctx)
					if err != nil {
						return fmt.Errorf("failed to describe images for repo %s: %w", aws.ToString(repoName), err)
					}

					totalImages += len(imageOutput.ImageDetails)

					for _, image := range imageOutput.ImageDetails {
						if len(image.ImageTags) == 0 {
							untaggedCount++
							untaggedSize += aws.ToInt64(image.ImageSizeInBytes)
						}
					}
				}

				mu.Lock()
				defer mu.Unlock()

				if totalImages == 0 {
					emptyRepos = append(emptyRepos, model.ECREmptyRepositoryInfo{RepositoryName: aws.ToString(repoName)})
				}

				if untaggedCount > 0 {
					untaggedSizeGB := untaggedSize / (1024 * 1024 * 1024)
					untaggedImages = append(untaggedImages, model.ECRUntaggedImageInfo{
						RepositoryName:       aws.ToString(repoName),
						UntaggedImageCount:   untaggedCount,
						UntaggedSizeBytes:    untaggedSize,
						EstimatedMonthlyCost: s.pricingService.CalculateECRStorageMonthlyCost(untaggedSizeGB),
					})
				}

				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		return nil, nil, nil, err
	}

	return noPolicy, emptyRepos, untaggedImages, nil
}
