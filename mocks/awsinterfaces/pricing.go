package awsinterfaces

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/pricing"
	"github.com/stretchr/testify/mock"
)

// MockPricingClient is a mock of the AWS Pricing client
type MockPricingClient struct {
	mock.Mock
}

// GetProducts mocks the GetProducts method
func (m *MockPricingClient) GetProducts(ctx context.Context, params *pricing.GetProductsInput, optFns ...func(*pricing.Options)) (*pricing.GetProductsOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*pricing.GetProductsOutput), args.Error(1)
}
