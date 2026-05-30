// Package vpc provides a service for interacting with AWS VPC resources.
package vpc

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"golang.org/x/sync/errgroup"
)

// NewService creates a new VPC service.
func NewService(awsconfig aws.Config, cwService cloudwatchMetricsService, pricingSvc pricingService) Service {
	client := ec2.NewFromConfig(awsconfig)

	return &service{
		client:         client,
		cwService:      cwService,
		pricingService: pricingSvc,
	}
}

func (s *service) Name() string {
	return "vpc"
}

func (s *service) Analyze(ctx context.Context, flags model.Flags) (model.ScopeResult, error) {
	start := time.Now()
	input := model.RenderWasteInput{}

	var errs []error

	idleNATs, err := s.GetIdleNATGateways(ctx, 7) // Hardcoding threshold
	if err != nil {
		errs = append(errs, err)
	} else {
		input.IdleNATGateways = idleNATs
	}

	var finalErr error
	if len(errs) > 0 {
		finalErr = fmt.Errorf("vpc analyze errors: %v", errs)
	}

	return model.ScopeResult{
		Scope:    s.Name(),
		Input:    input,
		Duration: time.Since(start),
		Err:      finalErr,
	}, nil
}

// GetIdleNATGateways returns NAT Gateways that have processed 0 bytes over the idleDays period.
func (s *service) GetIdleNATGateways(ctx context.Context, idleDays int) ([]model.NATGatewayWasteInfo, error) {
	paginator := ec2.NewDescribeNatGatewaysPaginator(s.client, &ec2.DescribeNatGatewaysInput{
		Filter: []types.Filter{
			{
				Name:   aws.String("state"),
				Values: []string{"available"},
			},
		},
	})

	var idleNATGateways []model.NATGatewayWasteInfo

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe NAT gateways: %w", err)
		}

		validGateways := s.collectValidGateways(output.NatGateways)

		results, err := s.fetchBytesOutConcurrently(ctx, validGateways, idleDays)
		if err != nil {
			return nil, err
		}

		idleNATGateways = append(idleNATGateways, s.buildIdleNATGateways(validGateways, results)...)
	}

	return idleNATGateways, nil
}

// collectValidGateways filters NAT Gateways with non-nil IDs.
func (s *service) collectValidGateways(natGateways []types.NatGateway) []types.NatGateway {
	valid := make([]types.NatGateway, 0, len(natGateways))

	for _, natGateway := range natGateways {
		natGatewayID := aws.ToString(natGateway.NatGatewayId)
		if natGatewayID == "" {
			continue
		}

		valid = append(valid, natGateway)
	}

	return valid
}

// fetchBytesOutConcurrently fetches CloudWatch bytesOut metrics for NAT Gateways in parallel.
func (s *service) fetchBytesOutConcurrently(ctx context.Context, gateways []types.NatGateway, idleDays int) ([]natGatewayMetricResult, error) {
	g, ctx := errgroup.WithContext(ctx)

	results := make([]natGatewayMetricResult, len(gateways))

	for i, natGateway := range gateways {
		natGatewayID := aws.ToString(natGateway.NatGatewayId)

		g.Go(func() error {
			bytesOut, err := s.cwService.NATGatewayBytesOut(ctx, natGatewayID, idleDays)

			results[i] = natGatewayMetricResult{
				bytesOut: bytesOut,
				err:      err,
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

// buildIdleNATGateways constructs NATGatewayWasteInfo slice from gateways with zero bytesOut.
func (s *service) buildIdleNATGateways(gateways []types.NatGateway, results []natGatewayMetricResult) []model.NATGatewayWasteInfo {
	var idle []model.NATGatewayWasteInfo

	for i, natGateway := range gateways {
		result := results[i]

		if result.err != nil {
			continue
		}

		if natGateway.State != types.NatGatewayStateAvailable {
			continue
		}

		if result.bytesOut == 0 {
			idle = append(idle, s.natGatewayToWasteInfo(natGateway))
		}
	}

	return idle
}

// natGatewayToWasteInfo converts a NAT Gateway to NATGatewayWasteInfo.
func (s *service) natGatewayToWasteInfo(natGateway types.NatGateway) model.NATGatewayWasteInfo {
	daysSinceCreate := 0
	if natGateway.CreateTime != nil {
		daysSinceCreate = int(time.Since(*natGateway.CreateTime).Hours() / 24)
	}

	return model.NATGatewayWasteInfo{
		NATGatewayID:          aws.ToString(natGateway.NatGatewayId),
		VPCID:                 aws.ToString(natGateway.VpcId),
		SubnetID:              aws.ToString(natGateway.SubnetId),
		State:                 string(natGateway.State),
		BytesOutToDestination: 0,
		EstimatedMonthlyCost:  s.pricingService.CalculateNATGatewayMonthlyCost(),
		DaysSinceCreate:       daysSinceCreate,
	}
}
