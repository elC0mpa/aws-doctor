package awscostexplorer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/utils"
)

func NewService(awsconfig aws.Config) *service {
	client := ec2.NewFromConfig(awsconfig)
	return &service{
		client: client,
	}
}

func (s *service) GetElasticIpAddressesInfo(ctx context.Context) (*model.ElasticIpInfo, error) {
	output, err := s.client.DescribeAddresses(ctx, nil)
	if err != nil {
		return nil, err
	}

	var unusedEips []string
	var attachedEips []model.AttachedIpInfo
	for _, address := range output.Addresses {
		if address.AssociationId == nil {
			unusedEips = append(unusedEips, aws.ToString(address.AllocationId))
		}

		attachedIp := model.AttachedIpInfo{
			IpAddress:    aws.ToString(address.PublicIp),
			AllocationId: aws.ToString(address.AllocationId),
			ResourceType: "ec2",
		}

		if address.InstanceId == nil {
			networkInterface, err := s.client.DescribeNetworkInterfaces(ctx, &ec2.DescribeNetworkInterfacesInput{
				NetworkInterfaceIds: []string{aws.ToString(address.NetworkInterfaceId)},
			})
			if err != nil {
				return nil, err
			}

			interfaceType := networkInterface.NetworkInterfaces[0].InterfaceType
			if interfaceType == types.NetworkInterfaceTypeInterface {
				interfaceType = s.getResourceTypeFromDescription(aws.ToString(networkInterface.NetworkInterfaces[0].Description))
			}

			attachedIp.ResourceType = string(interfaceType)
		}

		attachedEips = append(attachedEips, attachedIp)
	}

	return &model.ElasticIpInfo{
		UnusedElasticIpAddresses: unusedEips,
		UsedElasticIpAddresses:   attachedEips,
	}, nil
}

func (s *service) GetUnusedElasticIpAddressesInfo(ctx context.Context) ([]types.Address, error) {
	output, err := s.client.DescribeAddresses(ctx, nil)
	if err != nil {
		return nil, err
	}

	var unusedEips []types.Address

	for _, address := range output.Addresses {
		if address.AssociationId == nil {
			unusedEips = append(unusedEips, address)
		}
	}

	return unusedEips, nil
}

func (s *service) GetUnusedEBSVolumes(ctx context.Context) ([]types.Volume, error) {
	var allVolumes []types.Volume

	paginator := ec2.NewDescribeVolumesPaginator(s.client, &ec2.DescribeVolumesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("status"),
				Values: []string{"available"},
			},
		},
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		allVolumes = append(allVolumes, output.Volumes...)
	}

	return allVolumes, nil
}

func (s *service) GetStoppedInstancesInfo(ctx context.Context) ([]types.Instance, []types.Volume, error) {
	input := &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"stopped"},
			},
		},
	}

	var stoppedInstanceVolumeIDs []string
	var stoppedInstanceForMoreThan30Days []types.Instance

	thresholdTime := time.Now().Add(-30 * 24 * time.Hour)

	// Use paginator to handle large numbers of instances
	paginator := ec2.NewDescribeInstancesPaginator(s.client, input)

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, nil, err
		}

		for _, reservation := range output.Reservations {
			for _, instance := range reservation.Instances {
				for _, mapping := range instance.BlockDeviceMappings {
					if mapping.Ebs != nil {
						stoppedInstanceVolumeIDs = append(stoppedInstanceVolumeIDs, aws.ToString(mapping.Ebs.VolumeId))
					}
				}
				reason := aws.ToString(instance.StateTransitionReason)
				stoppedAt, err := utils.ParseTransitionDate(reason)
				if err != nil {
					continue
				}

				if stoppedAt.Before(thresholdTime) {
					stoppedInstanceForMoreThan30Days = append(stoppedInstanceForMoreThan30Days, instance)
				}
			}
		}
	}

	var stoppedInstanceVolumes []types.Volume

	if len(stoppedInstanceVolumeIDs) > 0 {
		// Use paginator for volumes as well (in case of many volumes)
		volumePaginator := ec2.NewDescribeVolumesPaginator(s.client, &ec2.DescribeVolumesInput{
			VolumeIds: stoppedInstanceVolumeIDs,
		})

		for volumePaginator.HasMorePages() {
			outputEBS, err := volumePaginator.NextPage(ctx)
			if err != nil {
				return nil, nil, err
			}
			stoppedInstanceVolumes = append(stoppedInstanceVolumes, outputEBS.Volumes...)
		}
	}

	return stoppedInstanceForMoreThan30Days, stoppedInstanceVolumes, nil
}

func (s *service) GetReservedInstanceExpiringOrExpired30DaysWaste(ctx context.Context) ([]model.RiExpirationInfo, error) {
	input := &ec2.DescribeReservedInstancesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("state"),
				Values: []string{"active", "retired"},
			},
		},
	}

	output, err := s.client.DescribeReservedInstances(ctx, input)
	if err != nil {
		return nil, err
	}

	var results []model.RiExpirationInfo

	now := time.Now()
	next30Days := now.Add(30 * 24 * time.Hour)
	prev30Days := now.Add(-30 * 24 * time.Hour)

	for _, ri := range output.ReservedInstances {
		if ri.End == nil {
			continue
		}

		endTime := *ri.End
		daysDiff := int(endTime.Sub(now).Hours() / 24)

		if ri.State == types.ReservedInstanceStateActive && endTime.Before(next30Days) {
			results = append(results, model.RiExpirationInfo{
				ReservedInstanceId: aws.ToString(ri.ReservedInstancesId),
				InstanceType:       string(ri.InstanceType),
				ExpirationDate:     endTime,
				DaysUntilExpiry:    daysDiff,
				State:              string(ri.State),
				Status:             "EXPIRING SOON",
			})
		}

		if endTime.After(prev30Days) && endTime.Before(now) {
			results = append(results, model.RiExpirationInfo{
				ReservedInstanceId: aws.ToString(ri.ReservedInstancesId),
				InstanceType:       string(ri.InstanceType),
				ExpirationDate:     endTime,
				DaysUntilExpiry:    daysDiff,
				State:              string(ri.State),
				Status:             "RECENTLY EXPIRED",
			})
		}
	}

	return results, nil
}

// GetUnusedAMIs returns AMIs that are not used by any running or stopped instances
func (s *service) GetUnusedAMIs(ctx context.Context, staleDays int) ([]model.AMIWasteInfo, error) {
	var results []model.AMIWasteInfo

	// Get all instances to find which AMIs are in use using pagination
	amiUsage := make(map[string]int)
	instancePaginator := ec2.NewDescribeInstancesPaginator(s.client, &ec2.DescribeInstancesInput{})
	for instancePaginator.HasMorePages() {
		page, err := instancePaginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe instances: %w", err)
		}
		for _, reservation := range page.Reservations {
			for _, instance := range reservation.Instances {
				if instance.ImageId != nil {
					amiUsage[*instance.ImageId]++
				}
			}
		}
	}

	// Get all owned AMIs
	// Note: DescribeImages doesn't have a paginator, but typically returns all at once
	amiOutput, err := s.client.DescribeImages(ctx, &ec2.DescribeImagesInput{
		Owners: []string{"self"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe images: %w", err)
	}

	cutoffTime := time.Now().AddDate(0, 0, -staleDays)
	now := time.Now()

	for _, image := range amiOutput.Images {
		imageId := aws.ToString(image.ImageId)
		usageCount := amiUsage[imageId]

		// Parse creation date
		creationDate := time.Time{}
		if image.CreationDate != nil {
			creationDate, _ = time.Parse(time.RFC3339, *image.CreationDate)
		}

		daysSinceCreate := int(now.Sub(creationDate).Hours() / 24)
		isStale := creationDate.Before(cutoffTime)

		// Consider unused if not used by any instance AND is stale
		if usageCount == 0 && isStale {
			// Collect snapshot IDs and sizes
			var snapshotIds []string
			var totalSnapshotSize int64

			for _, bdm := range image.BlockDeviceMappings {
				if bdm.Ebs != nil && bdm.Ebs.SnapshotId != nil {
					snapshotIds = append(snapshotIds, *bdm.Ebs.SnapshotId)
					if bdm.Ebs.VolumeSize != nil {
						totalSnapshotSize += int64(*bdm.Ebs.VolumeSize)
					}
				}
			}

			// EBS Snapshot pricing: ~$0.05 per GB per month
			estimatedCost := float64(totalSnapshotSize) * 0.05

			results = append(results, model.AMIWasteInfo{
				ImageId:         imageId,
				Name:            aws.ToString(image.Name),
				Description:     aws.ToString(image.Description),
				CreationDate:    creationDate,
				DaysSinceCreate: daysSinceCreate,
				IsPublic:        aws.ToBool(image.Public),
				SnapshotIds:     snapshotIds,
				SnapshotSizeGB:  totalSnapshotSize,
				UsedByInstances: usageCount,
				EstimatedCost:   estimatedCost,
			})
		}
	}

	return results, nil
}

func (s *service) getResourceTypeFromDescription(description string) types.NetworkInterfaceType {
	desc := strings.ToLower(description)

	if strings.Contains(desc, "elb app/") {
		return types.NetworkInterfaceTypeLoadBalancer
	}

	if strings.Contains(desc, "elb net/") {
		return types.NetworkInterfaceTypeNetworkLoadBalancer
	}

	if strings.Contains(desc, "nat gateway") || strings.Contains(desc, "nat-gateway") {
		return types.NetworkInterfaceTypeNatGateway
	}

	if strings.Contains(desc, "globalaccelerator") {
		return types.NetworkInterfaceTypeGlobalAcceleratorManaged
	}

	if strings.Contains(desc, "vpc endpoint") || strings.Contains(desc, "vpce-") {
		return types.NetworkInterfaceTypeVpcEndpoint
	}

	if strings.Contains(desc, "transit gateway") || strings.Contains(desc, "tgw-") {
		return types.NetworkInterfaceTypeTransitGateway
	}

	if strings.Contains(desc, "aws lambda") {
		return types.NetworkInterfaceTypeLambda
	}

	if strings.Contains(desc, "api gateway") {
		return types.NetworkInterfaceTypeApiGatewayManaged
	}

	if strings.Contains(desc, "iot rules") {
		return types.NetworkInterfaceTypeIotRulesManaged
	}

	if strings.Contains(desc, "gateway load balancer") {
		return types.NetworkInterfaceTypeGatewayLoadBalancer
	}

	if strings.Contains(desc, "redshift") {
		return types.NetworkInterfaceType("redshift_cluster")
	}

	if strings.Contains(desc, "rds") {
		return types.NetworkInterfaceType("rds_database")
	}

	if strings.Contains(desc, "directory service") {
		return types.NetworkInterfaceType("directory_service")
	}

	if strings.Contains(desc, "fsx") {
		return types.NetworkInterfaceType("fsx")
	}

	return types.NetworkInterfaceType("interface")
}
