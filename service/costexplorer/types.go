package awscostexplorer

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/elC0mpa/aws-doctor/model"
)

// CostExplorerClientAPI is the interface for the AWS Cost Explorer client methods used by the service.
type CostExplorerClientAPI interface {
	GetCostAndUsage(ctx context.Context, params *costexplorer.GetCostAndUsageInput, optFns ...func(*costexplorer.Options)) (*costexplorer.GetCostAndUsageOutput, error)
}

type service struct {
	client CostExplorerClientAPI
}

// Service is the interface for AWS Cost Explorer service.
type Service interface {
	GetCurrentMonthCostsByService(ctx context.Context) (*model.CostInfo, error)
	GetLastMonthCostsByService(ctx context.Context) (*model.CostInfo, error)
	GetMonthCostsByService(ctx context.Context, endDate time.Time) (*model.CostInfo, error)
	GetCurrentMonthTotalCosts(ctx context.Context) (*string, error)
	GetLastMonthTotalCosts(ctx context.Context) (*string, error)
	GetLastSixMonthsCosts(ctx context.Context) ([]model.CostInfo, error)
}
