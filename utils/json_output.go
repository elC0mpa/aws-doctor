package utils

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
)

// OutputCostComparisonJSON outputs cost comparison data as JSON
func OutputCostComparisonJSON(accountID string, lastTotalCost, currentTotalCost float64, lastMonth, currentMonth *model.CostInfo) error {
	output := model.CostComparisonJSON{
		AccountID:   accountID,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		CurrentMonth: model.CostPeriodJSON{
			Start: aws.ToString(currentMonth.Start),
			End:   aws.ToString(currentMonth.End),
			Total: currentTotalCost,
			Unit:  "USD",
		},
		LastMonth: model.CostPeriodJSON{
			Start: aws.ToString(lastMonth.Start),
			End:   aws.ToString(lastMonth.End),
			Total: lastTotalCost,
			Unit:  "USD",
		},
		ServiceBreakdown: []model.ServiceCostCompareJSON{},
	}

	// Add service breakdown
	for serviceName, currentCost := range currentMonth.CostGroup {
		lastCost := lastMonth.CostGroup[serviceName]
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
func OutputTrendJSON(accountID string, costInfo []model.CostInfo) error {
	output := model.TrendJSON{
		AccountID:   accountID,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
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
func OutputWasteJSON(accountID string, elasticIPs []types.Address, unusedVolumes []types.Volume, stoppedVolumes []types.Volume, ris []model.RiExpirationInfo, stoppedInstances []types.Instance, loadBalancers []elbtypes.LoadBalancer, unusedAMIs []model.AMIWasteInfo, orphanedSnapshots []model.SnapshotWasteInfo) error {
	output := model.WasteReportJSON{
		AccountID:           accountID,
		GeneratedAt:         time.Now().UTC().Format(time.RFC3339),
		UnusedElasticIPs:    []model.ElasticIPJSON{},
		UnusedEBSVolumes:    []model.EBSVolumeJSON{},
		StoppedVolumes:      []model.EBSVolumeJSON{},
		StoppedInstances:    []model.StoppedInstanceJSON{},
		ReservedInstances:   []model.ReservedInstanceJSON{},
		UnusedLoadBalancers: []model.LoadBalancerJSON{},
		UnusedAMIs:          []model.AMIJSON{},
		OrphanedSnapshots:   []model.SnapshotJSON{},
		StaleSnapshots:      []model.SnapshotJSON{},
	}

	// Unused Elastic IPs
	for _, ip := range elasticIPs {
		output.UnusedElasticIPs = append(output.UnusedElasticIPs, model.ElasticIPJSON{
			PublicIP:     aws.ToString(ip.PublicIp),
			AllocationID: aws.ToString(ip.AllocationId),
		})
	}

	// Unused EBS Volumes
	for _, vol := range unusedVolumes {
		output.UnusedEBSVolumes = append(output.UnusedEBSVolumes, model.EBSVolumeJSON{
			VolumeID: aws.ToString(vol.VolumeId),
			Size:     aws.ToInt32(vol.Size),
			Status:   "available",
		})
	}

	// EBS Volumes attached to stopped instances
	for _, vol := range stoppedVolumes {
		output.StoppedVolumes = append(output.StoppedVolumes, model.EBSVolumeJSON{
			VolumeID: aws.ToString(vol.VolumeId),
			Size:     aws.ToInt32(vol.Size),
			Status:   "attached_to_stopped",
		})
	}

	// Stopped instances
	now := time.Now()
	for _, instance := range stoppedInstances {
		si := model.StoppedInstanceJSON{
			InstanceID: aws.ToString(instance.InstanceId),
		}
		if instance.StateTransitionReason != nil {
			if stoppedAt, err := ParseTransitionDate(*instance.StateTransitionReason); err == nil {
				si.StoppedAt = stoppedAt.Format(time.RFC3339)
				si.DaysAgo = int(now.Sub(stoppedAt).Hours() / 24)
			}
		}
		output.StoppedInstances = append(output.StoppedInstances, si)
	}

	// Reserved instances
	for _, ri := range ris {
		output.ReservedInstances = append(output.ReservedInstances, model.ReservedInstanceJSON{
			ReservedInstanceID: ri.ReservedInstanceID,
			InstanceType:       ri.InstanceType,
			ExpirationDate:     ri.ExpirationDate.Format(time.RFC3339),
			DaysUntilExpiry:    ri.DaysUntilExpiry,
			State:              ri.State,
			Status:             ri.Status,
		})
	}

	// Unused load balancers
	for _, lb := range loadBalancers {
		output.UnusedLoadBalancers = append(output.UnusedLoadBalancers, model.LoadBalancerJSON{
			Name: aws.ToString(lb.LoadBalancerName),
			ARN:  aws.ToString(lb.LoadBalancerArn),
			Type: string(lb.Type),
		})
	}

	// Unused AMIs
	for _, ami := range unusedAMIs {
		output.UnusedAMIs = append(output.UnusedAMIs, model.AMIJSON{
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

	// Snapshots (split by category: orphaned vs stale)
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
			output.OrphanedSnapshots = append(output.OrphanedSnapshots, snapshotJSON)
		} else {
			output.StaleSnapshots = append(output.StaleSnapshots, snapshotJSON)
		}
	}

	output.HasWaste = len(output.UnusedElasticIPs) > 0 ||
		len(output.UnusedEBSVolumes) > 0 ||
		len(output.StoppedVolumes) > 0 ||
		len(output.StoppedInstances) > 0 ||
		len(output.ReservedInstances) > 0 ||
		len(output.UnusedLoadBalancers) > 0 ||
		len(output.UnusedAMIs) > 0 ||
		len(output.OrphanedSnapshots) > 0 ||
		len(output.StaleSnapshots) > 0

	return printJSON(output)
}

func printJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
