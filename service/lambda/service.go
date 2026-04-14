// Package lambda provides a service for detecting over-provisioned Lambda functions.
package lambda

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awslambda "github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/elC0mpa/aws-doctor/model"
	"golang.org/x/sync/errgroup"
)

const (
	defaultMemoryThresholdPercent = 10
	lookbackDays                  = 14
	maxLogsConcurrency            = 10
)

var maxMemUsedRegex = regexp.MustCompile(`Max Memory Used:\s*(\d+)\s*MB`)

// NewService creates a new Lambda service.
func NewService(awsconfig aws.Config) Service {
	return &service{
		lambdaClient: awslambda.NewFromConfig(awsconfig),
		logsClient:   cloudwatchlogs.NewFromConfig(awsconfig),
	}
}

func (s *service) GetOverProvisionedFunctions(ctx context.Context, memoryThresholdPercent int) ([]model.LambdaOverProvisionedInfo, error) {
	if memoryThresholdPercent <= 0 {
		memoryThresholdPercent = defaultMemoryThresholdPercent
	}

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

			maxMemUsed, err := s.getMaxMemoryUsed(ctx, logGroupName, startTime, now)
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

// getMaxMemoryUsed paginates through REPORT lines from a Lambda function's log group
// and returns the highest "Max Memory Used" value in MB.
func (s *service) getMaxMemoryUsed(ctx context.Context, logGroupName string, startTime, endTime time.Time) (int32, error) {
	input := &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName:  aws.String(logGroupName),
		FilterPattern: aws.String("REPORT RequestId"),
		StartTime:     aws.Int64(startTime.UnixMilli()),
		EndTime:       aws.Int64(endTime.UnixMilli()),
		Interleaved:   aws.Bool(true),
	}

	var maxMem int32

	for {
		output, err := s.logsClient.FilterLogEvents(ctx, input)
		if err != nil {
			return 0, err
		}

		for _, event := range output.Events {
			if event.Message == nil {
				continue
			}

			matches := maxMemUsedRegex.FindStringSubmatch(*event.Message)
			if len(matches) >= 2 {
				mem, err := strconv.Atoi(matches[1])
				if err == nil && int32(mem) > maxMem {
					maxMem = int32(mem)
				}
			}
		}

		if output.NextToken == nil {
			break
		}

		input.NextToken = output.NextToken
	}

	return maxMem, nil
}
