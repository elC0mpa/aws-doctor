// Package lambda provides a service for detecting over-provisioned Lambda functions.
package lambda

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awslambda "github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/cloudwatchlogs"
	"golang.org/x/sync/errgroup"
)

const (
	lookbackDays       = 14
	maxLogsConcurrency = 10
)

// NewService creates a new Lambda service.
func NewService(awsconfig aws.Config, cwLogsService cloudwatchlogs.Service) Service {
	return &service{
		lambdaClient: awslambda.NewFromConfig(awsconfig),
		logsService:  cwLogsService,
	}
}

func (s *service) GetOverProvisionedFunctions(ctx context.Context, memoryThresholdPercent int) ([]model.LambdaOverProvisionedInfo, error) {
	functions, err := s.listAllFunctions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list Lambda functions: %w", err)
	}

	now := time.Now()
	startTime := now.AddDate(0, 0, -lookbackDays)

	var (
		mu     sync.Mutex
		result []model.LambdaOverProvisionedInfo
	)

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(maxLogsConcurrency)

	for _, fn := range functions {
		fn := fn

		g.Go(func() error {
			functionName := aws.ToString(fn.FunctionName)
			configuredMemoryMB := aws.ToInt32(fn.MemorySize)
			logGroupName := fmt.Sprintf("/aws/lambda/%s", functionName)

			maxMemUsed, err := s.logsService.GetMaxMemoryUsed(ctx, logGroupName, startTime, now)
			if err != nil || maxMemUsed <= 0 || configuredMemoryMB <= 0 {
				return nil
			}

			utilizationPercent := (float64(maxMemUsed) / float64(configuredMemoryMB)) * 100

			if utilizationPercent < float64(memoryThresholdPercent) {
				recommendedMB := maxMemUsed * 2
				if recommendedMB < 128 {
					recommendedMB = 128
				}

				mu.Lock()

				result = append(result, model.LambdaOverProvisionedInfo{
					FunctionName:        functionName,
					Runtime:             string(fn.Runtime),
					ConfiguredMemoryMB:  configuredMemoryMB,
					MaxMemoryUsedMB:     maxMemUsed,
					MemoryUtilization:   utilizationPercent,
					RecommendedMemoryMB: recommendedMB,
				})
				mu.Unlock()
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *service) listAllFunctions(ctx context.Context) ([]lambdatypes.FunctionConfiguration, error) {
	var functions []lambdatypes.FunctionConfiguration

	paginator := awslambda.NewListFunctionsPaginator(s.lambdaClient, &awslambda.ListFunctionsInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		functions = append(functions, page.Functions...)
	}

	return functions, nil
}
