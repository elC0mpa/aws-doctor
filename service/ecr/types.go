package ecr

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/analyzer"
	"github.com/elC0mpa/aws-doctor/service/pricing"
)

// ClientAPI is the interface for the AWS ECR client methods used by the service.
type ClientAPI interface {
	DescribeRepositories(ctx context.Context, params *ecr.DescribeRepositoriesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeRepositoriesOutput, error)
	DescribeImages(ctx context.Context, params *ecr.DescribeImagesInput, optFns ...func(*ecr.Options)) (*ecr.DescribeImagesOutput, error)
	GetLifecyclePolicy(ctx context.Context, params *ecr.GetLifecyclePolicyInput, optFns ...func(*ecr.Options)) (*ecr.GetLifecyclePolicyOutput, error)
}

type service struct {
	client         ClientAPI
	pricingService pricing.Service
}

// Service is the interface for AWS ECR service.
type Service interface {
	analyzer.WasteAnalyzer
	GetECRWaste(ctx context.Context) ([]model.ECRNoLifecyclePolicyInfo, []model.ECREmptyRepositoryInfo, []model.ECRUntaggedImageInfo, error)
}
