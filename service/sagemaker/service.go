// Package sagemaker provides a service for detecting idle SageMaker real-time inference endpoints.
//
// Flow of GetIdleEndpoints:
//
//  1. ListEndpoints (paginated): pull every endpoint in the region whose status is InService.
//     Non-InService endpoints (Creating, Updating, Failed, Deleting) are skipped because we cannot
//     meaningfully assess traffic on them.
//  2. For each endpoint, concurrently run checkEndpoint:
//     a. DescribeEndpoint to read the ProductionVariants currently attached (name + instance count).
//     b. DescribeEndpointConfig to read the instance type per variant (the summary only has
//     counts, not types, so both calls are needed).
//     c. For each variant, query CloudWatch Invocations (AWS/SageMaker namespace, dimensions
//     EndpointName + VariantName) over the lookback window. The first variant with any
//     invocation short-circuits the endpoint as active.
//  3. An endpoint is idle only when every variant reports zero invocations for the whole window.
//     Idle endpoints are costed via pricing.CalculateSageMakerEndpointMonthlyCost, which sums the
//     hourly rate for each variant's instance type times its instance count.
//
// Per-endpoint CloudWatch errors are swallowed (endpoint dropped from results) so one bad metric
// call does not abort the scan. DescribeEndpoint / DescribeEndpointConfig errors propagate.
package sagemaker

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	sm "github.com/aws/aws-sdk-go-v2/service/sagemaker"
	smtypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/elC0mpa/aws-doctor/model"
	"golang.org/x/sync/errgroup"
)

const maxConcurrency = 8

// NewService creates a new SageMaker service.
func NewService(awsconfig aws.Config, cwService cloudWatchMetricsService, pricingSvc pricingService) Service {
	return &service{
		client:         sm.NewFromConfig(awsconfig),
		cwService:      cwService,
		pricingService: pricingSvc,
	}
}

func (s *service) Name() string {
	return "sagemaker"
}

func (s *service) Analyze(ctx context.Context, flags model.Flags) (model.ScopeResult, error) {
	start := time.Now()
	input := model.RenderWasteInput{}

	var errs []error

	idleEndpoints, err := s.GetIdleEndpoints(ctx, flags.SageMakerIdleDays)
	if err != nil {
		errs = append(errs, err)
	} else {
		input.IdleSageMakerEndpoints = idleEndpoints
	}

	var finalErr error
	if len(errs) > 0 {
		finalErr = fmt.Errorf("sagemaker analyze errors: %v", errs)
	}

	return model.ScopeResult{
		Scope:    s.Name(),
		Input:    input,
		Duration: time.Since(start),
		Err:      finalErr,
	}, nil
}

// GetIdleEndpoints returns InService SageMaker real-time inference endpoints whose production
// variants served zero invocations over the last idleDays days.
func (s *service) GetIdleEndpoints(ctx context.Context, idleDays int) ([]model.IdleSageMakerEndpointInfo, error) {
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
	paginator := sm.NewListEndpointsPaginator(s.client, &sm.ListEndpointsInput{
		StatusEquals: smtypes.EndpointStatusInService,
	})

	var endpoints []smtypes.EndpointSummary

	for paginator.HasMorePages() {
		out, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list SageMaker endpoints: %w", err)
		}

		endpoints = append(endpoints, out.Endpoints...)
	}

	return endpoints, nil
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

	return model.IdleSageMakerEndpointInfo{
		EndpointName:         name,
		EndpointARN:          aws.ToString(summary.EndpointArn),
		Status:               string(summary.EndpointStatus),
		Variants:             variants,
		DaysChecked:          idleDays,
		EstimatedMonthlyCost: s.pricingService.CalculateSageMakerEndpointMonthlyCost(variants),
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
