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
	lookbackDays              = 14
	maxLogsConcurrency        = 10
	logGroupsPerInsightsQuery = 50
	minRecommendedMemoryMB    = 128
	lambdaLogGroupPrefix      = "/aws/lambda/"
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

	if len(functions) == 0 {
		return nil, nil
	}

	existingLogGroups, err := s.logsService.ListExistingLogGroups(ctx, lambdaLogGroupPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list Lambda log groups: %w", err)
	}

	logGroupToFunction := make(map[string]lambdatypes.FunctionConfiguration, len(functions))
	logGroupNames := make([]string, 0, len(functions))

	for _, fn := range functions {
		logGroupName := lambdaLogGroupPrefix + aws.ToString(fn.FunctionName)
		if _, ok := existingLogGroups[logGroupName]; !ok {
			continue
		}

		logGroupToFunction[logGroupName] = fn
		logGroupNames = append(logGroupNames, logGroupName)
	}

	if len(logGroupNames) == 0 {
		return nil, nil
	}

	now := time.Now()
	startTime := now.AddDate(0, 0, -lookbackDays)

	maxMemByLogGroup, err := s.queryMaxMemoryInBatches(ctx, logGroupNames, startTime, now)
	if err != nil {
		return nil, err
	}

	return buildOverProvisionedResults(logGroupToFunction, maxMemByLogGroup, memoryThresholdPercent), nil
}

// queryMaxMemoryInBatches runs batched CloudWatch Logs Insights queries (up to
// logGroupsPerInsightsQuery per query) concurrently and merges the per-log-group results.
func (s *service) queryMaxMemoryInBatches(ctx context.Context, logGroupNames []string, startTime, endTime time.Time) (map[string]int32, error) {
	var (
		mu     sync.Mutex
		merged = make(map[string]int32, len(logGroupNames))
	)

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(maxLogsConcurrency)

	for start := 0; start < len(logGroupNames); start += logGroupsPerInsightsQuery {
		end := start + logGroupsPerInsightsQuery
		if end > len(logGroupNames) {
			end = len(logGroupNames)
		}

		batch := logGroupNames[start:end]

		g.Go(func() error {
			results, err := s.logsService.GetLambdaMaxMemoryUsedBatch(ctx, batch, startTime, endTime)
			if err != nil {
				return fmt.Errorf("batch Insights query failed for %d log groups: %w", len(batch), err)
			}

			mu.Lock()
			for k, v := range results {
				merged[k] = v
			}
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return merged, nil
}

// buildOverProvisionedResults returns recommendations for functions whose observed max memory
// falls below memoryThresholdPercent of configured memory. Recommendation is max(observed * 2,
// minRecommendedMemoryMB) to leave headroom while respecting Lambda's 128 MB floor.
func buildOverProvisionedResults(
	logGroupToFunction map[string]lambdatypes.FunctionConfiguration,
	maxMemByLogGroup map[string]int32,
	memoryThresholdPercent int,
) []model.LambdaOverProvisionedInfo {
	result := make([]model.LambdaOverProvisionedInfo, 0, len(maxMemByLogGroup))

	for logGroup, maxMemUsed := range maxMemByLogGroup {
		fn, ok := logGroupToFunction[logGroup]
		if !ok {
			continue
		}

		configuredMemoryMB := aws.ToInt32(fn.MemorySize)
		if maxMemUsed <= 0 || configuredMemoryMB <= 0 {
			continue
		}

		utilizationPercent := (float64(maxMemUsed) / float64(configuredMemoryMB)) * 100
		if utilizationPercent >= float64(memoryThresholdPercent) {
			continue
		}

		recommendedMB := maxMemUsed * 2
		if recommendedMB < minRecommendedMemoryMB {
			recommendedMB = minRecommendedMemoryMB
		}

		result = append(result, model.LambdaOverProvisionedInfo{
			FunctionName:        aws.ToString(fn.FunctionName),
			Runtime:             string(fn.Runtime),
			ConfiguredMemoryMB:  configuredMemoryMB,
			MaxMemoryUsedMB:     maxMemUsed,
			MemoryUtilization:   utilizationPercent,
			RecommendedMemoryMB: recommendedMB,
		})
	}

	return result
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
