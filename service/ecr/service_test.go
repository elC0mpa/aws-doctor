package ecr

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/elC0mpa/aws-doctor/mocks/awsinterfaces"
	"github.com/elC0mpa/aws-doctor/mocks/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewService(t *testing.T) {
	mockPricing := new(services.MockPricingService)
	svc := NewService(aws.Config{}, mockPricing)
	assert.NotNil(t, svc)
}

func TestGetECRWaste(t *testing.T) {
	mockClient := new(awsinterfaces.MockECRClient)
	mockPricing := new(services.MockPricingService)
	svc := &service{
		client:         mockClient,
		pricingService: mockPricing,
	}

	// Mock DescribeRepositories
	mockClient.On("DescribeRepositories", mock.Anything, mock.Anything, mock.Anything).Return(&ecr.DescribeRepositoriesOutput{
		Repositories: []types.Repository{
			{RepositoryName: aws.String("repo-clean")},
			{RepositoryName: aws.String("repo-no-policy")},
			{RepositoryName: aws.String("repo-untagged")},
			{RepositoryName: aws.String("repo-empty")},
		},
	}, nil)

	// Lifecycle Policies
	mockClient.On("GetLifecyclePolicy", mock.Anything, &ecr.GetLifecyclePolicyInput{RepositoryName: aws.String("repo-clean")}, mock.Anything).Return(&ecr.GetLifecyclePolicyOutput{}, nil)
	mockClient.On("GetLifecyclePolicy", mock.Anything, &ecr.GetLifecyclePolicyInput{RepositoryName: aws.String("repo-no-policy")}, mock.Anything).Return(nil, &types.LifecyclePolicyNotFoundException{})
	mockClient.On("GetLifecyclePolicy", mock.Anything, &ecr.GetLifecyclePolicyInput{RepositoryName: aws.String("repo-untagged")}, mock.Anything).Return(&ecr.GetLifecyclePolicyOutput{}, nil)
	mockClient.On("GetLifecyclePolicy", mock.Anything, &ecr.GetLifecyclePolicyInput{RepositoryName: aws.String("repo-empty")}, mock.Anything).Return(&ecr.GetLifecyclePolicyOutput{}, nil)

	// Images
	// Clean repo: 1 tagged image
	mockClient.On("DescribeImages", mock.Anything, &ecr.DescribeImagesInput{RepositoryName: aws.String("repo-clean")}, mock.Anything).Return(&ecr.DescribeImagesOutput{
		ImageDetails: []types.ImageDetail{
			{ImageTags: []string{"v1"}, ImageSizeInBytes: aws.Int64(100)},
		},
	}, nil)

	// No policy repo: 1 tagged image
	mockClient.On("DescribeImages", mock.Anything, &ecr.DescribeImagesInput{RepositoryName: aws.String("repo-no-policy")}, mock.Anything).Return(&ecr.DescribeImagesOutput{
		ImageDetails: []types.ImageDetail{
			{ImageTags: []string{"v1"}, ImageSizeInBytes: aws.Int64(100)},
		},
	}, nil)

	// Untagged repo: 1 tagged, 1 untagged (2 GiB = 2147483648 bytes → 2 GB after conversion)
	mockClient.On("DescribeImages", mock.Anything, &ecr.DescribeImagesInput{RepositoryName: aws.String("repo-untagged")}, mock.Anything).Return(&ecr.DescribeImagesOutput{
		ImageDetails: []types.ImageDetail{
			{ImageTags: []string{"v1"}, ImageSizeInBytes: aws.Int64(100)},
			{ImageTags: []string{}, ImageSizeInBytes: aws.Int64(2 * 1024 * 1024 * 1024)},
		},
	}, nil)

	// Empty repo: 0 images
	mockClient.On("DescribeImages", mock.Anything, &ecr.DescribeImagesInput{RepositoryName: aws.String("repo-empty")}, mock.Anything).Return(&ecr.DescribeImagesOutput{
		ImageDetails: []types.ImageDetail{},
	}, nil)

	// Pricing mock: 2 GiB untagged → 2 GB passed to pricing
	mockPricing.On("CalculateECRStorageMonthlyCost", int64(2)).Return(0.20)

	noPolicy, empty, untagged, err := svc.GetECRWaste(context.Background())

	assert.NoError(t, err)
	assert.Len(t, noPolicy, 1)
	assert.Equal(t, "repo-no-policy", noPolicy[0].RepositoryName)

	assert.Len(t, empty, 1)
	assert.Equal(t, "repo-empty", empty[0].RepositoryName)

	assert.Len(t, untagged, 1)
	assert.Equal(t, "repo-untagged", untagged[0].RepositoryName)
	assert.Equal(t, 1, untagged[0].UntaggedImageCount)
	assert.Equal(t, 0.20, untagged[0].EstimatedMonthlyCost)
}

func TestGetECRWaste_Errors(t *testing.T) {
	t.Run("DescribeRepositories_Error", func(t *testing.T) {
		mockClient := new(awsinterfaces.MockECRClient)
		svc := &service{client: mockClient}

		mockClient.On("DescribeRepositories", mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)

		_, _, _, err := svc.GetECRWaste(context.Background())
		assert.Error(t, err)
	})

	t.Run("GetLifecyclePolicy_Error", func(t *testing.T) {
		mockClient := new(awsinterfaces.MockECRClient)
		svc := &service{client: mockClient}

		mockClient.On("DescribeRepositories", mock.Anything, mock.Anything, mock.Anything).Return(&ecr.DescribeRepositoriesOutput{
			Repositories: []types.Repository{{RepositoryName: aws.String("repo")}},
		}, nil)

		mockClient.On("GetLifecyclePolicy", mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)

		_, _, _, err := svc.GetECRWaste(context.Background())
		assert.Error(t, err)
	})

	t.Run("DescribeImages_Error", func(t *testing.T) {
		mockClient := new(awsinterfaces.MockECRClient)
		svc := &service{client: mockClient}

		mockClient.On("DescribeRepositories", mock.Anything, mock.Anything, mock.Anything).Return(&ecr.DescribeRepositoriesOutput{
			Repositories: []types.Repository{{RepositoryName: aws.String("repo")}},
		}, nil)

		mockClient.On("GetLifecyclePolicy", mock.Anything, mock.Anything, mock.Anything).Return(&ecr.GetLifecyclePolicyOutput{}, nil)
		mockClient.On("DescribeImages", mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)

		_, _, _, err := svc.GetECRWaste(context.Background())
		assert.Error(t, err)
	})
}

func TestAnalyzerMethods(t *testing.T) {
	svc := &service{}

	if svc.Name() == "" {
		t.Error("Name() should not be empty")
	}

	if svc.TabName() == "" {
		t.Error("TabName() should not be empty")
	}
}
