// Package wastesummary computes waste cost totals from RenderWasteInput.
package wastesummary

import (
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/utils/pricing"
)

// CategorySummary holds the name, count, and estimated cost for a waste category.
type CategorySummary struct {
	Name  string
	Count int
	Cost  float64
}

// Compute returns per-category summaries and the total estimated monthly cost.
func Compute(input model.RenderWasteInput) ([]CategorySummary, float64) {
	categories := costCategories(input)
	categories = append(categories, countOnlyCategories(input)...)

	var total float64
	for _, cat := range categories {
		total += cat.Cost
	}

	return categories, total
}

func costCategories(input model.RenderWasteInput) []CategorySummary {
	var categories []CategorySummary

	if n := len(input.ElasticIPs); n > 0 {
		categories = append(categories, CategorySummary{"Elastic IPs", n, float64(n) * pricing.EIPCostPerMonth})
	}

	if n := len(input.UnusedVolumes); n > 0 {
		categories = append(categories, CategorySummary{"EBS Volumes (Unattached)", n, ebsVolumeCost(input.UnusedVolumes)})
	}

	if n := len(input.StoppedVolumes); n > 0 {
		categories = append(categories, CategorySummary{"EBS Volumes (Stopped Inst.)", n, ebsVolumeCost(input.StoppedVolumes)})
	}

	if n := len(input.LoadBalancers); n > 0 {
		categories = append(categories, CategorySummary{"Load Balancers", n, lbCost(input.LoadBalancers)})
	}

	if n := len(input.CloudWatchLogGroups); n > 0 {
		var cost float64
		for _, lg := range input.CloudWatchLogGroups {
			cost += lg.EstimatedMonthlyCost
		}

		categories = append(categories, CategorySummary{"CloudWatch Log Groups", n, cost})
	}

	if n := len(input.UnusedAMIs); n > 0 {
		var cost float64
		for _, ami := range input.UnusedAMIs {
			cost += ami.MaxPotentialSaving
		}

		categories = append(categories, CategorySummary{"Unused AMIs", n, cost})
	}

	if n := len(input.OrphanedSnapshots); n > 0 {
		var cost float64
		for _, snap := range input.OrphanedSnapshots {
			cost += snap.MaxPotentialSavings
		}

		categories = append(categories, CategorySummary{"EBS Snapshots", n, cost})
	}

	return categories
}

func countOnlyCategories(input model.RenderWasteInput) []CategorySummary {
	var categories []CategorySummary

	if n := len(input.StoppedInstances); n > 0 {
		categories = append(categories, CategorySummary{"Stopped EC2 Instances", n, 0})
	}

	if n := len(input.Ris); n > 0 {
		categories = append(categories, CategorySummary{"Reserved Instances", n, 0})
	}

	if n := len(input.S3Buckets); n > 0 {
		categories = append(categories, CategorySummary{"S3 Buckets (No Lifecycle)", n, 0})
	}

	if n := len(input.S3MultipartUploads); n > 0 {
		categories = append(categories, CategorySummary{"S3 Incomplete Multipart", n, 0})
	}

	if n := len(input.UnusedKeyPairs); n > 0 {
		categories = append(categories, CategorySummary{"Unused Key Pairs", n, 0})
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
