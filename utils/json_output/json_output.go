package jsonoutput

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/pricing"
	"github.com/elC0mpa/aws-doctor/utils/cost"
	"github.com/elC0mpa/aws-doctor/utils/ec2"
	wastesummary "github.com/elC0mpa/aws-doctor/utils/waste_summary"
)

// OutputCostComparisonJSON outputs cost comparison data as JSON
func OutputCostComparisonJSON(input model.RenderCostComparisonInput) error {
	lastTotalCost := cost.ParseCostString(input.LastTotalCost)
	currentTotalCost := cost.ParseCostString(input.CurrentTotalCost)

	output := model.CostComparisonJSON{
		AccountID:   input.AccountID,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		CurrentMonth: model.CostPeriodJSON{
			Start: aws.ToString(input.CurrentMonth.Start),
			End:   aws.ToString(input.CurrentMonth.End),
			Total: currentTotalCost,
			Unit:  "USD",
		},
		LastMonth: model.CostPeriodJSON{
			Start: aws.ToString(input.LastMonth.Start),
			End:   aws.ToString(input.LastMonth.End),
			Total: lastTotalCost,
			Unit:  "USD",
		},
		ServiceBreakdown: []model.ServiceCostCompareJSON{},
	}

	// Add service breakdown
	for serviceName, currentCost := range input.CurrentMonth.CostGroup {
		lastCost := input.LastMonth.CostGroup[serviceName]
		output.ServiceBreakdown = append(output.ServiceBreakdown, model.ServiceCostCompareJSON{
			Service:     serviceName,
			CurrentCost: currentCost.Amount,
			LastCost:    lastCost.Amount,
			Difference:  currentCost.Amount - lastCost.Amount,
			Unit:        currentCost.Unit,
		})
	}

	return printJSON(output)
}

// OutputTrendJSON outputs trend data as JSON
func OutputTrendJSON(accountID string, costInfo []model.CostInfo, services []string) error {
	output := model.TrendJSON{
		AccountID:   accountID,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Services:    services,
		Months:      []model.MonthCostJSON{},
	}

	for _, info := range costInfo {
		if total, ok := info.CostGroup["Total"]; ok {
			output.Months = append(output.Months, model.MonthCostJSON{
				Start: aws.ToString(info.Start),
				End:   aws.ToString(info.End),
				Total: total.Amount,
				Unit:  total.Unit,
			})
		}
	}

	return printJSON(output)
}

// OutputWasteJSON outputs waste detection data as JSON
func OutputWasteJSON(input model.RenderWasteInput, pricingSvc pricing.Service) error {
	output := model.WasteReportJSON{
		AccountID:              input.AccountID,
		GeneratedAt:            time.Now().UTC().Format(time.RFC3339),
		UnusedElasticIPs:       mapElasticIPs(input.ElasticIPs, pricingSvc),
		UnusedEBSVolumes:       mapEBSVolumes(input.UnusedVolumes, "available", pricingSvc),
		StoppedVolumes:         mapEBSVolumes(input.StoppedVolumes, "attached_to_stopped", pricingSvc),
		StoppedInstances:       mapStoppedInstances(input.StoppedInstances),
		ReservedInstances:      mapReservedInstances(input.Ris),
		UnusedLoadBalancers:    mapLoadBalancers(input.LoadBalancers, pricingSvc),
		UnusedAMIs:             mapAMIs(input.UnusedAMIs),
		UnusedKeyPairs:         mapKeyPairs(input.UnusedKeyPairs),
		S3Buckets:              mapS3Buckets(input.S3Buckets),
		S3MultipartUploads:     mapS3MultipartUploads(input.S3MultipartUploads),
		CloudWatchLogGroups:    mapCloudWatchLogGroups(input.CloudWatchLogGroups),
		StoppedRDSInstances:    mapRDSInstances(input.RDSInstances),
		OldRDSSnapshots:        mapRDSSnapshots(input.RDSSnapshots),
		IdleRDSInstances:       mapRDSIdleInstances(input.RDSIdleInstances),
		IdleNATGateways:        mapNATGateways(input.IdleNATGateways),
		IdleLoadBalancers:      mapIdleLoadBalancers(input.IdleLoadBalancers),
		OverProvisionedLambdas: mapLambdaOverProvisioned(input.OverProvisionedLambdas),
		IdleSageMakerEndpoints: mapIdleSageMakerEndpoints(input.IdleSageMakerEndpoints),
		ECRNoLifecyclePolicies: mapECRNoLifecyclePolicies(input.ECRNoLifecyclePolicies),
		ECREmptyRepositories:   mapECREmptyRepositories(input.ECREmptyRepositories),
		ECRUntaggedImages:      mapECRUntaggedImages(input.ECRUntaggedImages),
		UnusedSecrets:          mapUnusedSecrets(input.UnusedSecrets, pricingSvc),
	}

	output.OrphanedSnapshots, output.StaleSnapshots = mapSnapshots(input.OrphanedSnapshots)

	output.HasWaste = hasAnyWasteJSON(output)

	categories, total := wastesummary.Compute(input, pricingSvc)
	output.TotalEstimatedMonthlyCost = total

	for _, cat := range categories {
		output.Summary = append(output.Summary, model.WasteSummaryJSON{
			Category:             cat.Name,
			Count:                cat.Count,
			EstimatedMonthlyCost: cat.Cost,
		})
	}

	return printJSON(output)
}

func mapS3Buckets(buckets []model.S3BucketWasteInfo) []model.S3BucketJSON {
	var result []model.S3BucketJSON

	for _, b := range buckets {
		result = append(result, model.S3BucketJSON{
			BucketName:   b.BucketName,
			CreationDate: b.CreationDate.Format(time.RFC3339),
		})
	}

	return result
}

func mapS3MultipartUploads(buckets []model.S3MultipartUploadWasteInfo) []model.S3MultipartJSON {
	var result []model.S3MultipartJSON

	for _, b := range buckets {
		result = append(result, model.S3MultipartJSON(b))
	}

	return result
}

func mapElasticIPs(elasticIPs []ec2types.Address, pricingSvc pricing.Service) []model.ElasticIPJSON {
	var result []model.ElasticIPJSON

	for _, ip := range elasticIPs {
		result = append(result, model.ElasticIPJSON{
			PublicIP:             aws.ToString(ip.PublicIp),
			AllocationID:         aws.ToString(ip.AllocationId),
			EstimatedMonthlyCost: pricingSvc.CalculateEIPMonthlyCost(),
		})
	}

	return result
}

func mapEBSVolumes(volumes []ec2types.Volume, status string, pricingSvc pricing.Service) []model.EBSVolumeJSON {
	var result []model.EBSVolumeJSON

	for _, vol := range volumes {
		size := aws.ToInt32(vol.Size)

		result = append(result, model.EBSVolumeJSON{
			VolumeID:             aws.ToString(vol.VolumeId),
			Size:                 size,
			Status:               status,
			EstimatedMonthlyCost: pricingSvc.CalculateEBSMonthlyCost(size, vol.VolumeType),
		})
	}

	return result
}

func mapStoppedInstances(stoppedInstances []ec2types.Instance) []model.StoppedInstanceJSON {
	var result []model.StoppedInstanceJSON

	now := time.Now()

	for _, instance := range stoppedInstances {
		si := model.StoppedInstanceJSON{
			InstanceID: aws.ToString(instance.InstanceId),
		}

		if instance.StateTransitionReason != nil {
			if stoppedAt, err := ec2.ParseTransitionDate(*instance.StateTransitionReason); err == nil {
				si.StoppedAt = stoppedAt.Format(time.RFC3339)
				si.DaysAgo = int(now.Sub(stoppedAt).Hours() / 24)
			}
		}

		result = append(result, si)
	}

	return result
}

func mapReservedInstances(ris []model.RiExpirationInfo) []model.ReservedInstanceJSON {
	var result []model.ReservedInstanceJSON

	for _, ri := range ris {
		result = append(result, model.ReservedInstanceJSON{
			ReservedInstanceID: ri.ReservedInstanceID,
			InstanceType:       ri.InstanceType,
			ExpirationDate:     ri.ExpirationDate.Format(time.RFC3339),
			DaysUntilExpiry:    ri.DaysUntilExpiry,
			State:              ri.State,
			Status:             ri.Status,
		})
	}

	return result
}

func mapLoadBalancers(loadBalancers []elbtypes.LoadBalancer, pricingSvc pricing.Service) []model.LoadBalancerJSON {
	var result []model.LoadBalancerJSON

	for _, lb := range loadBalancers {
		result = append(result, model.LoadBalancerJSON{
			Name:                 aws.ToString(lb.LoadBalancerName),
			ARN:                  aws.ToString(lb.LoadBalancerArn),
			Type:                 string(lb.Type),
			EstimatedMonthlyCost: pricingSvc.CalculateLoadBalancerMonthlyCost(lb.Type),
		})
	}

	return result
}

func mapAMIs(unusedAMIs []model.AMIWasteInfo) []model.AMIJSON {
	var result []model.AMIJSON

	for _, ami := range unusedAMIs {
		result = append(result, model.AMIJSON{
			ImageID:            ami.ImageID,
			Name:               ami.Name,
			Description:        ami.Description,
			CreationDate:       ami.CreationDate.Format(time.RFC3339),
			DaysSinceCreate:    ami.DaysSinceCreate,
			IsPublic:           ami.IsPublic,
			SnapshotIDs:        ami.SnapshotIDs,
			SnapshotSizeGB:     ami.SnapshotSizeGB,
			MaxPotentialSaving: ami.MaxPotentialSaving,
			SafetyWarning:      ami.SafetyWarning,
		})
	}

	return result
}

func mapSnapshots(orphanedSnapshots []model.SnapshotWasteInfo) ([]model.SnapshotJSON, []model.SnapshotJSON) {
	var orphaned, stale []model.SnapshotJSON

	for _, snap := range orphanedSnapshots {
		snapshotJSON := model.SnapshotJSON{
			SnapshotID:          snap.SnapshotID,
			VolumeID:            snap.VolumeID,
			VolumeExists:        snap.VolumeExists,
			UsedByAMI:           snap.UsedByAMI,
			AMIID:               snap.AMIID,
			SizeGB:              snap.SizeGB,
			StartTime:           snap.StartTime.Format(time.RFC3339),
			DaysSinceCreate:     snap.DaysSinceCreate,
			Description:         snap.Description,
			Category:            string(snap.Category),
			Reason:              snap.Reason,
			MaxPotentialSavings: snap.MaxPotentialSavings,
		}

		if snap.Category == model.SnapshotCategoryOrphaned {
			orphaned = append(orphaned, snapshotJSON)
		} else {
			stale = append(stale, snapshotJSON)
		}
	}

	return orphaned, stale
}

func mapKeyPairs(unusedKeyPairs []model.KeyPairWasteInfo) []model.KeyPairJSON {
	var result []model.KeyPairJSON

	for _, kp := range unusedKeyPairs {
		result = append(result, model.KeyPairJSON{
			KeyName:         kp.KeyName,
			KeyPairID:       kp.KeyPairID,
			CreationDate:    kp.CreateTime.Format(time.RFC3339),
			DaysSinceCreate: kp.DaysSinceCreate,
		})
	}

	return result
}

func mapCloudWatchLogGroups(logGroups []model.CloudWatchLogsWasteInfo) []model.CloudWatchLogGroupJSON {
	var result []model.CloudWatchLogGroupJSON

	for _, lg := range logGroups {
		result = append(result, model.CloudWatchLogGroupJSON{
			LogGroupName:         lg.LogGroupName,
			CreationTime:         lg.CreationTime.Format(time.RFC3339),
			StoredBytes:          lg.StoredBytes,
			EstimatedMonthlyCost: lg.EstimatedMonthlyCost,
		})
	}

	return result
}

func mapRDSInstances(instances []model.RDSInstanceWasteInfo) []model.RDSInstanceJSON {
	var result []model.RDSInstanceJSON

	for _, inst := range instances {
		result = append(result, model.RDSInstanceJSON(inst))
	}

	return result
}

func mapRDSSnapshots(snapshots []model.RDSSnapshotWasteInfo) []model.RDSSnapshotJSON {
	var result []model.RDSSnapshotJSON

	for _, snap := range snapshots {
		result = append(result, model.RDSSnapshotJSON{
			DBSnapshotID:         snap.DBSnapshotID,
			DBInstanceID:         snap.DBInstanceID,
			Engine:               snap.Engine,
			AllocatedStorage:     snap.AllocatedStorage,
			SnapshotCreateTime:   snap.SnapshotCreateTime.Format(time.RFC3339),
			DaysSinceCreate:      snap.DaysSinceCreate,
			EstimatedMonthlyCost: snap.EstimatedMonthlyCost,
		})
	}

	return result
}

func mapIdleLoadBalancers(idleLBs []model.ELBIdleInfo) []model.ELBIdleJSON {
	var result []model.ELBIdleJSON

	for _, lb := range idleLBs {
		result = append(result, model.ELBIdleJSON(lb))
	}

	return result
}

func mapRDSIdleInstances(instances []model.RDSIdleInstanceInfo) []model.RDSIdleInstanceJSON {
	var result []model.RDSIdleInstanceJSON

	for _, inst := range instances {
		result = append(result, model.RDSIdleInstanceJSON(inst))
	}

	return result
}

func mapNATGateways(natGateways []model.NATGatewayWasteInfo) []model.NATGatewayJSON {
	var result []model.NATGatewayJSON

	for _, ng := range natGateways {
		result = append(result, model.NATGatewayJSON(ng))
	}

	return result
}

func mapLambdaOverProvisioned(lambdas []model.LambdaOverProvisionedInfo) []model.LambdaOverProvisionedJSON {
	var result []model.LambdaOverProvisionedJSON

	for _, fn := range lambdas {
		result = append(result, model.LambdaOverProvisionedJSON(fn))
	}

	return result
}

func mapECRNoLifecyclePolicies(repos []model.ECRNoLifecyclePolicyInfo) []model.ECRNoLifecyclePolicyJSON {
	var result []model.ECRNoLifecyclePolicyJSON

	for _, r := range repos {
		result = append(result, model.ECRNoLifecyclePolicyJSON(r))
	}

	return result
}

func mapECREmptyRepositories(repos []model.ECREmptyRepositoryInfo) []model.ECREmptyRepositoryJSON {
	var result []model.ECREmptyRepositoryJSON

	for _, r := range repos {
		result = append(result, model.ECREmptyRepositoryJSON(r))
	}

	return result
}

func mapECRUntaggedImages(repos []model.ECRUntaggedImageInfo) []model.ECRUntaggedImageJSON {
	var result []model.ECRUntaggedImageJSON

	for _, r := range repos {
		result = append(result, model.ECRUntaggedImageJSON(r))
	}

	return result
}

func mapIdleSageMakerEndpoints(eps []model.IdleSageMakerEndpointInfo) []model.IdleSageMakerEndpointJSON {
	var result []model.IdleSageMakerEndpointJSON

	for _, ep := range eps {
		result = append(result, model.IdleSageMakerEndpointJSON(ep))
	}

	return result
}

func printJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))

	return nil
}

// hasAnyWasteJSON returns true when any category in the JSON report contains entries.
// Extracted from OutputWasteJSON to keep that function below the gocyclo threshold.
func hasAnyWasteJSON(o model.WasteReportJSON) bool {
	return hasComputeWasteJSON(o) ||
		hasStorageWasteJSON(o) ||
		hasDatabaseWasteJSON(o) ||
		hasNetworkWasteJSON(o) ||
		hasServerlessWasteJSON(o) ||
		hasContainerWasteJSON(o) ||
		len(o.UnusedSecrets) > 0
}

func hasComputeWasteJSON(o model.WasteReportJSON) bool {
	return len(o.StoppedInstances) > 0 ||
		len(o.ReservedInstances) > 0 ||
		len(o.UnusedAMIs) > 0 ||
		len(o.UnusedKeyPairs) > 0 ||
		len(o.CloudWatchLogGroups) > 0
}

func hasStorageWasteJSON(o model.WasteReportJSON) bool {
	return len(o.UnusedEBSVolumes) > 0 ||
		len(o.StoppedVolumes) > 0 ||
		len(o.OrphanedSnapshots) > 0 ||
		len(o.StaleSnapshots) > 0 ||
		len(o.S3Buckets) > 0 ||
		len(o.S3MultipartUploads) > 0
}

func hasDatabaseWasteJSON(o model.WasteReportJSON) bool {
	return len(o.StoppedRDSInstances) > 0 ||
		len(o.OldRDSSnapshots) > 0 ||
		len(o.IdleRDSInstances) > 0
}

func hasNetworkWasteJSON(o model.WasteReportJSON) bool {
	return len(o.UnusedElasticIPs) > 0 ||
		len(o.UnusedLoadBalancers) > 0 ||
		len(o.IdleNATGateways) > 0 ||
		len(o.IdleLoadBalancers) > 0
}

func hasServerlessWasteJSON(o model.WasteReportJSON) bool {
	return len(o.OverProvisionedLambdas) > 0 ||
		len(o.IdleSageMakerEndpoints) > 0
}

func hasContainerWasteJSON(o model.WasteReportJSON) bool {
	return len(o.ECRNoLifecyclePolicies) > 0 ||
		len(o.ECREmptyRepositories) > 0 ||
		len(o.ECRUntaggedImages) > 0
}

func mapUnusedSecrets(secrets []model.UnusedSecretInfo, pricingSvc pricing.Service) []model.UnusedSecretJSON {
	var result []model.UnusedSecretJSON

	for _, s := range secrets {
		json := model.UnusedSecretJSON{
			Name:                 s.Name,
			EstimatedMonthlyCost: pricingSvc.CalculateSecretsManagerMonthlyCost(1),
		}

		if s.LastAccessedDate != nil {
			json.LastAccessedDate = s.LastAccessedDate.Format(time.RFC3339)
		}

		result = append(result, json)
	}

	return result
}
