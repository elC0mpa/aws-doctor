// Package vpc provides a service for interacting with AWS VPC resources.
package vpc

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elC0mpa/aws-doctor/model"
	awscloudwatchmetrics "github.com/elC0mpa/aws-doctor/service/cloudwatchmetrics"
	"github.com/elC0mpa/aws-doctor/utils/pricing"
)

// NewService creates a new VPC service instance.
func NewService(awsCfg aws.Config, cwService awscloudwatchmetrics.Service) Service {
	client := ec2.NewFromConfig(awsCfg)

	return &service{
		client:    client,
		cwService: cwService,
	}
}

// GetIdleNatGateways returns NAT Gateways that have processed 0 bytes over the idleDays period.
func (s *service) GetIdleNatGateways(ctx context.Context, idleDays int) ([]model.NatGatewayWasteInfo, error) {
	var idleNatGateways []model.NatGatewayWasteInfo

	paginator := ec2.NewDescribeNatGatewaysPaginator(s.client, &ec2.DescribeNatGatewaysInput{
		Filter: []types.Filter{
			{
				Name:   aws.String("state"),
				Values: []string{"available"},
			},
		},
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe NAT gateways: %w", err)
		}

		for _, natGateway := range output.NatGateways {
			natGatewayID := aws.ToString(natGateway.NatGatewayId)
			if natGatewayID == "" {
				continue // Skip NAT Gateways without ID
			}

			bytesOut, err := s.cwService.NatGatewayBytesOut(ctx, natGatewayID, idleDays)
			if err != nil {
				log.Printf("Warning: failed to get CloudWatch metrics for NAT Gateway %s: %v", natGatewayID, err)
				continue
			}

			if bytesOut == 0 {
				idleNatGateways = append(idleNatGateways, model.NatGatewayWasteInfo{
					NatGatewayID:          natGatewayID,
					VPCID:                 aws.ToString(natGateway.VpcId),
					SubnetID:              aws.ToString(natGateway.SubnetId),
					State:                 string(natGateway.State),
					BytesOutToDestination: bytesOut,
					EstimatedMonthlyCost:  pricing.NatGatewayCostPerMonth,
				})
			}
		}
	}

	return idleNatGateways, nil
}
