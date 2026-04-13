// Package wastesummary computes waste cost totals from RenderWasteInput.
package wastesummary

import (
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/utils/pricing"
)

// Compute returns per-category summaries and the total estimated monthly cost.
func Compute(input model.RenderWasteInput) ([]model.CategorySummary, float64) {
	categories := costCategories(input)
	categories = append(categories, countOnlyCategories(input)...)

	var total float64
	for _, cat := range categories {
		total += cat.Cost
	}

	return categories, total
}

func costCategories(input model.RenderWasteInput) []model.CategorySummary {
	var categories []model.CategorySummary

	if n := len(input.ElasticIPs); n > 0 {
		categories = append(categories, model.CategorySummary{Name: "Elastic IPs", Count: n, Cost: float64(n) * pricing.EIPCostPerMonth})
	}

	if n := len(input.UnusedVolumes); n > 0 {
		categories = append(categories, model.CategorySummary{Name: "EBS Volumes (Unattached)", Count: n, Cost: ebsVolumeCost(input.UnusedVolumes)})
	}

	if n := len(input.StoppedVolumes); n > 0 {
		categories = append(categories, model.CategorySummary{Name: "EBS Volumes (Stopped Inst.)", Count: n, Cost: ebsVolumeCost(input.StoppedVolumes)})
	}

	if n := len(input.LoadBalancers); n > 0 {
		categories = append(categories, model.CategorySummary{Name: "Load Balancers", Count: n, Cost: lbCost(input.LoadBalancers)})
	}

	if n := len(input.CloudWatchLogGroups); n > 0 {
		var cost float64
		for _, lg := range input.CloudWatchLogGroups {
			cost += lg.EstimatedMonthlyCost
		}

		categories = append(categories, model.CategorySummary{Name: "CloudWatch Log Groups", Count: n, Cost: cost})
	}

	if n := len(input.UnusedAMIs); n > 0 {
		var cost float64
		for _, ami := range input.UnusedAMIs {
			cost += ami.MaxPotentialSaving
		}

		categories = append(categories, model.CategorySummary{Name: "Unused AMIs", Count: n, Cost: cost})
	}

	if n := len(input.OrphanedSnapshots); n > 0 {
		var cost float64
		for _, snap := range input.OrphanedSnapshots {
			cost += snap.MaxPotentialSavings
		}

		categories = append(categories, model.CategorySummary{Name: "EBS Snapshots", Count: n, Cost: cost})
	}

	if n := len(input.RDSInstances); n > 0 {
		var cost float64
		for _, inst := range input.RDSInstances {
			cost += inst.EstimatedMonthlyCost
		}

		categories = append(categories, model.CategorySummary{Name: "RDS Instances (Stopped)", Count: n, Cost: cost})
	}

	if n := len(input.RDSIdleInstances); n > 0 {
		var cost float64
		for _, inst := range input.RDSIdleInstances {
			cost += inst.EstimatedMonthlyCost
		}

		categories = append(categories, model.CategorySummary{Name: "RDS Instances (Idle)", Count: n, Cost: cost})
	}

	if n := len(input.RDSSnapshots); n > 0 {
		var cost float64
		for _, snap := range input.RDSSnapshots {
			cost += snap.EstimatedMonthlyCost
		}

		categories = append(categories, model.CategorySummary{Name: "RDS Snapshots", Count: n, Cost: cost})
	}

	if n := len(input.IdleNATGateways); n > 0 {
		var cost float64
		for _, gw := range input.IdleNATGateways {
			cost += gw.EstimatedMonthlyCost
		}

		categories = append(categories, model.CategorySummary{Name: "NAT Gateways (Idle)", Count: n, Cost: cost})
	}

	return categories
}

func countOnlyCategories(input model.RenderWasteInput) []model.CategorySummary {
	var categories []model.CategorySummary

	if n := len(input.StoppedInstances); n > 0 {
		categories = append(categories, model.CategorySummary{Name: "Stopped EC2 Instances", Count: n})
	}

	if n := len(input.Ris); n > 0 {
		categories = append(categories, model.CategorySummary{Name: "Reserved Instances", Count: n})
	}

	if n := len(input.S3Buckets); n > 0 {
		categories = append(categories, model.CategorySummary{Name: "S3 Buckets (No Lifecycle)", Count: n})
	}

	if n := len(input.S3MultipartUploads); n > 0 {
		categories = append(categories, model.CategorySummary{Name: "S3 Incomplete Multipart", Count: n})
	}

	if n := len(input.UnusedKeyPairs); n > 0 {
		categories = append(categories, model.CategorySummary{Name: "Unused Key Pairs", Count: n})
	}

	return categories
}

func ebsVolumeCost(volumes []ec2types.Volume) float64 {
	var cost float64
	for _, vol := range volumes {
		cost += float64(*vol.Size) * pricing.EBSCostPerGBMonth(vol.VolumeType)
	}

	return cost
}

func lbCost(loadBalancers []elbtypes.LoadBalancer) float64 {
	var cost float64

	for _, lb := range loadBalancers {
		if lb.Type == "classic" {
			cost += pricing.CLBCostPerMonth
		} else {
			cost += pricing.ALBCostPerMonth
		}
	}

	return cost
}
