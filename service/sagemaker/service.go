// Package sagemaker provides a service for detecting idle SageMaker real-time inference endpoints.
package sagemaker

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	sm "github.com/aws/aws-sdk-go-v2/service/sagemaker"
	smtypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/cloudwatchmetrics"
	"github.com/elC0mpa/aws-doctor/utils/pricing"
	"golang.org/x/sync/errgroup"
)

const (
	defaultIdleDays = 14
	maxConcurrency  = 8
)

// NewService creates a new SageMaker service.
func NewService(awsconfig aws.Config, cwService cloudwatchmetrics.Service) Service {
	return &service{
		client:    sm.NewFromConfig(awsconfig),
		cwService: cwService,
	}
}

// GetIdleEndpoints returns InService SageMaker real-time inference endpoints whose production
// variants served zero invocations over the last idleDays days. When idleDays <= 0 a default of
// 14 days is used.
func (s *service) GetIdleEndpoints(ctx context.Context, idleDays int) ([]model.IdleSageMakerEndpointInfo, error) {
	if idleDays <= 0 {
		idleDays = defaultIdleDays
	}

	endpoints, err := s.listEndpoints(ctx)
	if err != nil {
		return nil, err
	}

	if len(endpoints) == 0 {
		return nil, nil
	}

	results := make([]endpointCheckResult, len(endpoints))

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(maxConcurrency)

	for i, ep := range endpoints {
		i, ep := i, ep

		g.Go(func() error {
			info, ok, err := s.checkEndpoint(ctx, ep, idleDays)
			if err != nil {
				return err
			}

			results[i] = endpointCheckResult{info: info, isIdle: ok}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	idle := make([]model.IdleSageMakerEndpointInfo, 0, len(results))

	for _, r := range results {
		if r.isIdle {
			idle = append(idle, r.info)
		}
	}

	return idle, nil
}

// endpointCheckResult is the per-endpoint output collected by the errgroup dispatch in
// GetIdleEndpoints. Using a named struct (per AGENTS.md) over an anonymous type keeps the result
// collection readable and avoids the mutex-protected append pattern.
type endpointCheckResult struct {
	info   model.IdleSageMakerEndpointInfo
	isIdle bool
}

func (s *service) listEndpoints(ctx context.Context) ([]smtypes.EndpointSummary, error) {
	var endpoints []smtypes.EndpointSummary

	var nextToken *string

	for {
		out, err := s.client.ListEndpoints(ctx, &sm.ListEndpointsInput{
			StatusEquals: smtypes.EndpointStatusInService,
			NextToken:    nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list SageMaker endpoints: %w", err)
		}

		endpoints = append(endpoints, out.Endpoints...)

		if out.NextToken == nil || *out.NextToken == "" {
			return endpoints, nil
		}

		nextToken = out.NextToken
	}
}

// checkEndpoint returns (info, true, nil) when the endpoint is idle, (_, false, nil) when it has
// traffic, and (_, _, err) on a fatal error. Per-endpoint CloudWatch errors are swallowed so a
// single bad metric call does not take down the whole scan.
func (s *service) checkEndpoint(ctx context.Context, summary smtypes.EndpointSummary, idleDays int) (model.IdleSageMakerEndpointInfo, bool, error) {
	name := aws.ToString(summary.EndpointName)

	describe, err := s.client.DescribeEndpoint(ctx, &sm.DescribeEndpointInput{
		EndpointName: summary.EndpointName,
	})
	if err != nil {
		return model.IdleSageMakerEndpointInfo{}, false, fmt.Errorf("describe endpoint %s: %w", name, err)
	}

	cfg, err := s.client.DescribeEndpointConfig(ctx, &sm.DescribeEndpointConfigInput{
		EndpointConfigName: describe.EndpointConfigName,
	})
	if err != nil {
		return model.IdleSageMakerEndpointInfo{}, false, fmt.Errorf("describe endpoint config %s: %w", aws.ToString(describe.EndpointConfigName), err)
	}

	variants := buildVariants(describe.ProductionVariants, cfg.ProductionVariants)
	if len(variants) == 0 {
		return model.IdleSageMakerEndpointInfo{}, false, nil
	}

	// Short-circuit as soon as any variant shows traffic to avoid extra CloudWatch calls on
	// endpoints that are obviously not idle.
	for _, v := range variants {
		n, err := s.cwService.SageMakerVariantInvocations(ctx, name, v.VariantName, idleDays)
		if err != nil {
			return model.IdleSageMakerEndpointInfo{}, false, nil
		}

		if n > 0 {
			return model.IdleSageMakerEndpointInfo{}, false, nil
		}
	}

	costInputs := make([]pricing.SageMakerVariantCost, 0, len(variants))
	for _, v := range variants {
		costInputs = append(costInputs, pricing.SageMakerVariantCost{
			InstanceType:  v.InstanceType,
			InstanceCount: v.InstanceCount,
		})
	}

	return model.IdleSageMakerEndpointInfo{
		EndpointName:         name,
		EndpointARN:          aws.ToString(summary.EndpointArn),
		Status:               string(summary.EndpointStatus),
		Variants:             variants,
		DaysChecked:          idleDays,
		EstimatedMonthlyCost: pricing.CalculateSageMakerEndpointMonthlyCost(costInputs),
	}, true, nil
}

// buildVariants merges the per-variant instance count from the endpoint summary with the
// instance type from the endpoint config, which is where SageMaker stores the provisioned type.
func buildVariants(summary []smtypes.ProductionVariantSummary, config []smtypes.ProductionVariant) []model.SageMakerVariant {
	typeByVariant := make(map[string]string, len(config))
	for _, v := range config {
		typeByVariant[aws.ToString(v.VariantName)] = string(v.InstanceType)
	}

	out := make([]model.SageMakerVariant, 0, len(summary))

	for _, v := range summary {
		name := aws.ToString(v.VariantName)
		out = append(out, model.SageMakerVariant{
			VariantName:   name,
			InstanceType:  typeByVariant[name],
			InstanceCount: aws.ToInt32(v.CurrentInstanceCount),
		})
	}

	return out
}
