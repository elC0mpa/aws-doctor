package awsinterfaces

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/stretchr/testify/mock"
)

// MockCostExplorerClient is a mock of CostExplorerClientAPI
type MockCostExplorerClient struct {
	mock.Mock
}

// GetCostAndUsage provides a mock implementation for the corresponding AWS Cost Explorer API call.
func (m *MockCostExplorerClient) GetCostAndUsage(ctx context.Context, params *costexplorer.GetCostAndUsageInput, optFns ...func(*costexplorer.Options)) (*costexplorer.GetCostAndUsageOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*costexplorer.GetCostAndUsageOutput), args.Error(1)
}
