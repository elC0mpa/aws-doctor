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

const ebsSnapshotCostPerGBMonth = 0.05

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

// GetUnusedAMIs returns AMIs that are not used by any running or stopped instances.
// It uses pagination for DescribeImages and adds safety warnings about potential
// ASG/Launch Template usage.
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

	// Get all owned AMIs using pagination
	amiPaginator := ec2.NewDescribeImagesPaginator(s.client, &ec2.DescribeImagesInput{
		Owners: []string{"self"},
	})

	cutoffTime := time.Now().AddDate(0, 0, -staleDays)
	now := time.Now()

	for amiPaginator.HasMorePages() {
		page, err := amiPaginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe images: %w", err)
		}

		for _, image := range page.Images {
			imageId := aws.ToString(image.ImageId)
			usageCount := amiUsage[imageId]

			// Parse creation date with proper error handling
			var creationDate time.Time
			if image.CreationDate != nil {
				parsedDate, err := time.Parse(time.RFC3339, *image.CreationDate)
				if err != nil {
					// Log warning but continue processing - use zero time as fallback
					// This handles gracefully any unexpected date formats
					creationDate = time.Time{}
				} else {
					creationDate = parsedDate
				}
			}

			daysSinceCreate := 0
			if !creationDate.IsZero() {
				daysSinceCreate = int(now.Sub(creationDate).Hours() / 24)
			}
			isStale := !creationDate.IsZero() && creationDate.Before(cutoffTime)

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
				// Note: This is max potential savings - actual snapshot billing is incremental
				maxPotentialSaving := float64(totalSnapshotSize) * 0.05

				// Safety warning: AMI may be used by ASGs or Launch Templates
				safetyWarning := "Verify before deleting: AMI may be used by Auto Scaling Groups or Launch Templates not currently running instances"

				results = append(results, model.AMIWasteInfo{
					ImageId:            imageId,
					Name:               aws.ToString(image.Name),
					Description:        aws.ToString(image.Description),
					CreationDate:       creationDate,
					DaysSinceCreate:    daysSinceCreate,
					IsPublic:           aws.ToBool(image.Public),
					SnapshotIds:        snapshotIds,
					SnapshotSizeGB:     totalSnapshotSize,
					UsedByInstances:    usageCount,
					MaxPotentialSaving: maxPotentialSaving,
					SafetyWarning:      safetyWarning,
				})
			}
		}
	}

	return results, nil
}

// GetOrphanedSnapshots returns EBS snapshots that are potentially orphaned
// (source volume deleted, not used by any AMI, or older than staleDays)
func (s *service) GetOrphanedSnapshots(ctx context.Context, staleDays int) ([]model.SnapshotWasteInfo, error) {
	var results []model.SnapshotWasteInfo

	// Collect all snapshots owned by this account using pagination
	var allSnapshots []types.Snapshot
	snapshotPaginator := ec2.NewDescribeSnapshotsPaginator(s.client, &ec2.DescribeSnapshotsInput{
		OwnerIds: []string{"self"},
	})
	for snapshotPaginator.HasMorePages() {
		page, err := snapshotPaginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe snapshots: %w", err)
		}
		allSnapshots = append(allSnapshots, page.Snapshots...)
	}

	// Build a set of existing volume IDs using pagination
	existingVolumes := make(map[string]bool)
	volumePaginator := ec2.NewDescribeVolumesPaginator(s.client, &ec2.DescribeVolumesInput{})
	for volumePaginator.HasMorePages() {
		page, err := volumePaginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe volumes: %w", err)
		}
		for _, vol := range page.Volumes {
			existingVolumes[aws.ToString(vol.VolumeId)] = true
		}
	}

	// Build a map of snapshot IDs used by AMIs (with pagination)
	snapshotToAMI := make(map[string]string)
	imagePaginator := ec2.NewDescribeImagesPaginator(s.client, &ec2.DescribeImagesInput{
		Owners: []string{"self"},
	})
	for imagePaginator.HasMorePages() {
		page, err := imagePaginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to describe images: %w", err)
		}
		for _, image := range page.Images {
			for _, bdm := range image.BlockDeviceMappings {
				if bdm.Ebs != nil && bdm.Ebs.SnapshotId != nil {
					snapshotToAMI[*bdm.Ebs.SnapshotId] = aws.ToString(image.ImageId)
				}
			}
		}
	}

	cutoffTime := time.Now().AddDate(0, 0, -staleDays)
	now := time.Now()

	for _, snapshot := range allSnapshots {
		volumeId := aws.ToString(snapshot.VolumeId)
		snapshotId := aws.ToString(snapshot.SnapshotId)
		volumeExists := existingVolumes[volumeId]
		amiId, usedByAMI := snapshotToAMI[snapshotId]

		startTime := time.Time{}
		if snapshot.StartTime != nil {
			startTime = *snapshot.StartTime
		}

		daysSinceCreate := int(now.Sub(startTime).Hours() / 24)

		// Skip snapshots used by AMIs - they are not waste
		if usedByAMI {
			continue
		}

		sizeGB := int32(0)
		if snapshot.VolumeSize != nil {
			sizeGB = *snapshot.VolumeSize
		}

		// EBS Snapshot pricing: ~$0.05 per GB per month
		// Note: Actual savings may be lower due to incremental storage
		maxPotentialSavings := float64(sizeGB) * ebsSnapshotCostPerGBMonth

		// Categorize based on whether source volume exists
		if !volumeExists {
			// Orphaned: Volume no longer exists - safe to delete (high confidence)
			results = append(results, model.SnapshotWasteInfo{
				SnapshotId:          snapshotId,
				VolumeId:            volumeId,
				VolumeExists:        volumeExists,
				UsedByAMI:           usedByAMI,
				AMIId:               amiId,
				SizeGB:              sizeGB,
				StartTime:           startTime,
				DaysSinceCreate:     daysSinceCreate,
				Description:         aws.ToString(snapshot.Description),
				Category:            model.SnapshotCategoryOrphaned,
				Reason:              "Volume Deleted",
				MaxPotentialSavings: maxPotentialSavings,
			})
		} else if startTime.Before(cutoffTime) {
			// Stale: Volume exists but snapshot is old - needs review (low confidence)
			results = append(results, model.SnapshotWasteInfo{
				SnapshotId:          snapshotId,
				VolumeId:            volumeId,
				VolumeExists:        volumeExists,
				UsedByAMI:           usedByAMI,
				AMIId:               amiId,
				SizeGB:              sizeGB,
				StartTime:           startTime,
				DaysSinceCreate:     daysSinceCreate,
				Description:         aws.ToString(snapshot.Description),
				Category:            model.SnapshotCategoryStale,
				Reason:              "Old Backup",
				MaxPotentialSavings: maxPotentialSavings,
			})
		}
	}

	return results, nil
}

func (s *service) GetNatGateways(ctx context.Context) ([]types.NatGateway, error) {
	output, err := s.client.DescribeNatGateways(ctx, &ec2.DescribeNatGatewaysInput{
		Filter: []types.Filter{
			{
				Name:   aws.String("state"),
				Values: []string{"available"},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return output.NatGateways, nil
}

func (s *service) GetVpcEndpoints(ctx context.Context) ([]types.VpcEndpoint, error) {
	output, err := s.client.DescribeVpcEndpoints(ctx, &ec2.DescribeVpcEndpointsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-endpoint-type"),
				Values: []string{"Interface"},
			},
			{
				Name:   aws.String("vpc-endpoint-state"),
				Values: []string{"available"},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return output.VpcEndpoints, nil
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
