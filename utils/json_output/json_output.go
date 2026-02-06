package jsonoutput

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/utils/cost"
	"github.com/elC0mpa/aws-doctor/utils/ec2"
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
func OutputWasteJSON(input model.RenderWasteInput) error {
	output := model.WasteReportJSON{
		AccountID:           input.AccountID,
		GeneratedAt:         time.Now().UTC().Format(time.RFC3339),
		UnusedElasticIPs:    mapElasticIPs(input.ElasticIPs),
		UnusedEBSVolumes:    mapEBSVolumes(input.UnusedVolumes, "available"),
		StoppedVolumes:      mapEBSVolumes(input.StoppedVolumes, "attached_to_stopped"),
		StoppedInstances:    mapStoppedInstances(input.StoppedInstances),
		ReservedInstances:   mapReservedInstances(input.Ris),
		UnusedLoadBalancers: mapLoadBalancers(input.LoadBalancers),
		UnusedAMIs:          mapAMIs(input.UnusedAMIs),
		UnusedKeyPairs:      mapKeyPairs(input.UnusedKeyPairs),
	}

	output.OrphanedSnapshots, output.StaleSnapshots = mapSnapshots(input.OrphanedSnapshots)

	output.HasWaste = len(output.UnusedElasticIPs) > 0 ||
		len(output.UnusedEBSVolumes) > 0 ||
		len(output.StoppedVolumes) > 0 ||
		len(output.StoppedInstances) > 0 ||
		len(output.ReservedInstances) > 0 ||
		len(output.UnusedLoadBalancers) > 0 ||
		len(output.UnusedAMIs) > 0 ||
		len(output.OrphanedSnapshots) > 0 ||
		len(output.StaleSnapshots) > 0 ||
		len(output.UnusedKeyPairs) > 0

	return printJSON(output)
}

func mapElasticIPs(elasticIPs []types.Address) []model.ElasticIPJSON {
	var result []model.ElasticIPJSON

	for _, ip := range elasticIPs {
		result = append(result, model.ElasticIPJSON{
			PublicIP:     aws.ToString(ip.PublicIp),
			AllocationID: aws.ToString(ip.AllocationId),
		})
	}

	return result
}

func mapEBSVolumes(volumes []types.Volume, status string) []model.EBSVolumeJSON {
	var result []model.EBSVolumeJSON

	for _, vol := range volumes {
		result = append(result, model.EBSVolumeJSON{
			VolumeID: aws.ToString(vol.VolumeId),
			Size:     aws.ToInt32(vol.Size),
			Status:   status,
		})
	}

	return result
}

func mapStoppedInstances(stoppedInstances []types.Instance) []model.StoppedInstanceJSON {
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

func mapLoadBalancers(loadBalancers []elbtypes.LoadBalancer) []model.LoadBalancerJSON {
	var result []model.LoadBalancerJSON

	for _, lb := range loadBalancers {
		result = append(result, model.LoadBalancerJSON{
			Name: aws.ToString(lb.LoadBalancerName),
			ARN:  aws.ToString(lb.LoadBalancerArn),
			Type: string(lb.Type),
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

func printJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))

	return nil
}
