package wastesummary

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/utils/pricing"
	"github.com/stretchr/testify/assert"
)

func TestCompute_Empty(t *testing.T) {
	categories, total := Compute(model.RenderWasteInput{})

	assert.Empty(t, categories)
	assert.Equal(t, 0.0, total)
}

func TestCompute_ElasticIPs(t *testing.T) {
	input := model.RenderWasteInput{
		ElasticIPs: []types.Address{
			{PublicIp: aws.String("1.2.3.4"), AllocationId: aws.String("eipalloc-1")},
			{PublicIp: aws.String("5.6.7.8"), AllocationId: aws.String("eipalloc-2")},
		},
	}

	categories, total := Compute(input)

	assert.Len(t, categories, 1)
	assert.Equal(t, "Elastic IPs", categories[0].Name)
	assert.Equal(t, 2, categories[0].Count)
	assert.Equal(t, 2*pricing.EIPCostPerMonth, categories[0].Cost)
	assert.Equal(t, 2*pricing.EIPCostPerMonth, total)
}

func TestCompute_EBSVolumes(t *testing.T) {
	input := model.RenderWasteInput{
		UnusedVolumes: []types.Volume{
			{VolumeId: aws.String("vol-1"), Size: aws.Int32(100), VolumeType: types.VolumeTypeGp2},
		},
		StoppedVolumes: []types.Volume{
			{VolumeId: aws.String("vol-2"), Size: aws.Int32(50), VolumeType: types.VolumeTypeGp3},
		},
	}

	categories, total := Compute(input)

	assert.Len(t, categories, 2)
	assert.Equal(t, "EBS Volumes (Unattached)", categories[0].Name)
	assert.Equal(t, 100*pricing.EBSgp2CostPerGBMonth, categories[0].Cost)
	assert.Equal(t, "EBS Volumes (Stopped Inst.)", categories[1].Name)
	assert.Equal(t, 50*pricing.EBSgp3CostPerGBMonth, categories[1].Cost)
	assert.Equal(t, 100*pricing.EBSgp2CostPerGBMonth+50*pricing.EBSgp3CostPerGBMonth, total)
}

func TestCompute_LoadBalancers(t *testing.T) {
	input := model.RenderWasteInput{
		LoadBalancers: []elbtypes.LoadBalancer{
			{LoadBalancerName: aws.String("alb-1"), Type: elbtypes.LoadBalancerTypeEnumApplication},
			{LoadBalancerName: aws.String("clb-1"), Type: "classic"},
		},
	}

	categories, total := Compute(input)

	assert.Len(t, categories, 1)
	assert.Equal(t, 2, categories[0].Count)

	expectedCost := pricing.ALBCostPerMonth + pricing.CLBCostPerMonth
	assert.Equal(t, expectedCost, categories[0].Cost)
	assert.Equal(t, expectedCost, total)
}

func TestCompute_CloudWatchLogGroups(t *testing.T) {
	input := model.RenderWasteInput{
		CloudWatchLogGroups: []model.CloudWatchLogsWasteInfo{
			{LogGroupName: "lg-1", EstimatedMonthlyCost: 1.50},
			{LogGroupName: "lg-2", EstimatedMonthlyCost: 2.50},
		},
	}

	categories, total := Compute(input)

	assert.Len(t, categories, 1)
	assert.Equal(t, 4.0, categories[0].Cost)
	assert.Equal(t, 4.0, total)
}

func TestCompute_AMIs(t *testing.T) {
	input := model.RenderWasteInput{
		UnusedAMIs: []model.AMIWasteInfo{
			{ImageID: "ami-1", MaxPotentialSaving: 5.0},
			{ImageID: "ami-2", MaxPotentialSaving: 10.0},
		},
	}

	categories, total := Compute(input)

	assert.Len(t, categories, 1)
	assert.Equal(t, "Unused AMIs", categories[0].Name)
	assert.Equal(t, 2, categories[0].Count)
	assert.Equal(t, 15.0, categories[0].Cost)
	assert.Equal(t, 15.0, total)
}

func TestCompute_Snapshots(t *testing.T) {
	input := model.RenderWasteInput{
		OrphanedSnapshots: []model.SnapshotWasteInfo{
			{SnapshotID: "snap-1", MaxPotentialSavings: 3.0},
			{SnapshotID: "snap-2", MaxPotentialSavings: 2.0},
			{SnapshotID: "snap-3", MaxPotentialSavings: 5.0},
		},
	}

	categories, total := Compute(input)

	assert.Len(t, categories, 1)
	assert.Equal(t, "EBS Snapshots", categories[0].Name)
	assert.Equal(t, 3, categories[0].Count)
	assert.Equal(t, 10.0, categories[0].Cost)
	assert.Equal(t, 10.0, total)
}

func TestCompute_RDS(t *testing.T) {
	input := model.RenderWasteInput{
		RDSInstances: []model.RDSInstanceWasteInfo{
			{DBInstanceID: "inst-1", EstimatedMonthlyCost: 10.0},
			{DBInstanceID: "inst-2", EstimatedMonthlyCost: 15.0},
		},
		RDSIdleInstances: []model.RDSIdleInstanceInfo{
			{DBInstanceID: "idle-1", EstimatedMonthlyCost: 20.0},
			{DBInstanceID: "idle-2", EstimatedMonthlyCost: 25.0},
		},
		RDSSnapshots: []model.RDSSnapshotWasteInfo{
			{DBSnapshotID: "snap-1", EstimatedMonthlyCost: 5.0},
			{DBSnapshotID: "snap-2", EstimatedMonthlyCost: 7.5},
		},
	}

	categories, total := Compute(input)

	assert.Len(t, categories, 3)
	assert.Equal(t, 82.5, total)

	assert.Equal(t, "RDS Instances (Stopped)", categories[0].Name)
	assert.Equal(t, 2, categories[0].Count)
	assert.Equal(t, 25.0, categories[0].Cost)

	assert.Equal(t, "RDS Instances (Idle)", categories[1].Name)
	assert.Equal(t, 2, categories[1].Count)
	assert.Equal(t, 45.0, categories[1].Cost)

	assert.Equal(t, "RDS Snapshots", categories[2].Name)
	assert.Equal(t, 2, categories[2].Count)
	assert.Equal(t, 12.5, categories[2].Cost)
}

func TestCompute_CountOnlyItems(t *testing.T) {
	input := model.RenderWasteInput{
		StoppedInstances:   []types.Instance{{InstanceId: aws.String("i-1")}},
		Ris:                []model.RiExpirationInfo{{ReservedInstanceID: "ri-1"}},
		S3Buckets:          []model.S3BucketWasteInfo{{BucketName: "bucket-1"}},
		S3MultipartUploads: []model.S3MultipartUploadWasteInfo{{BucketName: "bucket-2"}},
		UnusedKeyPairs:     []model.KeyPairWasteInfo{{KeyName: "key-1"}},
	}

	categories, total := Compute(input)

	assert.Len(t, categories, 5)
	assert.Equal(t, 0.0, total)

	for _, cat := range categories {
		assert.Equal(t, 1, cat.Count)
		assert.Equal(t, 0.0, cat.Cost)
	}
}

func TestCompute_MixedWaste(t *testing.T) {
	input := model.RenderWasteInput{
		ElasticIPs: []types.Address{
			{PublicIp: aws.String("1.2.3.4")},
		},
		UnusedKeyPairs: []model.KeyPairWasteInfo{
			{KeyName: "key-1"},
		},
	}

	categories, total := Compute(input)

	assert.Len(t, categories, 2)
	assert.Equal(t, pricing.EIPCostPerMonth, total)
	assert.Equal(t, "Elastic IPs", categories[0].Name)
	assert.Equal(t, "Unused Key Pairs", categories[1].Name)
	assert.Equal(t, 0.0, categories[1].Cost)
}

func TestCompute_NATGateways(t *testing.T) {
	input := model.RenderWasteInput{
		IdleNATGateways: []model.NATGatewayWasteInfo{
			{NATGatewayID: "nat-1", EstimatedMonthlyCost: pricing.NATGatewayCostPerMonth},
			{NATGatewayID: "nat-2", EstimatedMonthlyCost: pricing.NATGatewayCostPerMonth},
		},
	}

	categories, total := Compute(input)

	assert.Len(t, categories, 1)
	assert.Equal(t, "NAT Gateways (Idle)", categories[0].Name)
	assert.Equal(t, 2, categories[0].Count)
	assert.Equal(t, 2*pricing.NATGatewayCostPerMonth, categories[0].Cost)
	assert.Equal(t, 2*pricing.NATGatewayCostPerMonth, total)
}
