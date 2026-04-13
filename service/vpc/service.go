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
	awscloudwatchmetrics "github.com/elC0mpa/aws-doctor/service/cloudwatchmetrics"
	"github.com/elC0mpa/aws-doctor/utils/pricing"
	"golang.org/x/sync/errgroup"
)

// NewService creates a new VPC service instance.
func NewService(awsCfg aws.Config, cwService awscloudwatchmetrics.Service) Service {
	client := ec2.NewFromConfig(awsCfg)

	return &service{
		client:    client,
		cwService: cwService,
	}
}

// IdleNATGateways returns NAT Gateways that have processed 0 bytes over the idleDays period.
func (s *service) IdleNATGateways(ctx context.Context, idleDays int) ([]model.NATGatewayWasteInfo, error) {
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
func (s *service) collectValidGateways(natGateways []types.NatGateway) []natGatewayWithIndex {
	valid := make([]natGatewayWithIndex, 0, len(natGateways))

	for i, natGateway := range natGateways {
		natGatewayID := aws.ToString(natGateway.NatGatewayId)
		if natGatewayID == "" {
			continue
		}

		valid = append(valid, natGatewayWithIndex{i, natGateway})
	}

	return valid
}

// fetchBytesOutConcurrently fetches CloudWatch bytesOut metrics for NAT Gateways in parallel.
func (s *service) fetchBytesOutConcurrently(ctx context.Context, gateways []natGatewayWithIndex, idleDays int) ([]struct {
	bytesOut float64
	err      error
}, error,
) {
	g, ctx := errgroup.WithContext(ctx)

	results := make([]struct {
		bytesOut float64
		err      error
	}, len(gateways))

	for _, vg := range gateways {
		i := vg.index
		natGateway := vg.natGateway
		natGatewayID := aws.ToString(natGateway.NatGatewayId)

		g.Go(func() error {
			bytesOut, err := s.cwService.NatGatewayBytesOut(ctx, natGatewayID, idleDays)

			results[i] = struct {
				bytesOut float64
				err      error
			}{bytesOut: bytesOut, err: err}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

// buildIdleNATGateways constructs NATGatewayWasteInfo slice from gateways with zero bytesOut.
func (s *service) buildIdleNATGateways(gateways []natGatewayWithIndex, results []struct {
	bytesOut float64
	err      error
},
) []model.NATGatewayWasteInfo {
	var idle []model.NATGatewayWasteInfo

	for _, vg := range gateways {
		natGateway := vg.natGateway
		result := results[vg.index]

		if result.err != nil {
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
		EstimatedMonthlyCost:  pricing.NatGatewayCostPerMonth,
		DaysSinceCreate:       daysSinceCreate,
	}
}
