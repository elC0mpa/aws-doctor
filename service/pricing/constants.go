package pricing

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// The Pricing API endpoint is only served from us-east-1 and ap-south-1, but it returns pricing
// data for every AWS region via the regionCode filter on each query. We always talk to us-east-1
// since it is available to every AWS account.
const pricingEndpointRegion = "us-east-1"

// maxPricingConcurrency caps parallel GetProducts calls during Load.
const maxPricingConcurrency = 20

// hoursPerMonth converts Pricing API hourly rates into an approximate monthly cost. 730 is the
// same figure AWS uses in Cost Explorer examples (365.25 * 24 / 12).
const hoursPerMonth = 730.0

// Cache keys are flat strings, usually "category" or "category:variant" (e.g. "nat",
// "ebs:gp3", "rds-instance:db.t3.medium").
const (
	categoryEBS              = "ebs"
	categoryEIP              = "eip"
	categoryNAT              = "nat"
	categoryLBApp            = "lb-app"
	categoryLBClassic        = "lb-classic"
	categoryCWLogs           = "cwlogs"
	categoryEC2Instance      = "ec2-instance"
	categoryRDSInstance      = "rds-instance"
	categoryRDSStorage       = "rds-storage"
	categoryRDSSnapshot      = "rds-snapshot"
	categorySageMakerHosting = "sagemaker-hosting"
	categoryECR              = "ecr"
	categorySecretsManager   = "secretsmanager"
)

const (
	// SecretsManagerCostPerSecretMonth is the default cost of a secret per month in Secrets Manager.
	SecretsManagerCostPerSecretMonth = 0.40

	// EIPCostPerMonth is the default cost of an unassociated Elastic IP (~$0.005/hour * 730 hours).
	EIPCostPerMonth = 3.65

	// EBSgp2CostPerGBMonth is the default cost of gp2 EBS storage per GB per month.
	EBSgp2CostPerGBMonth = 0.10
	// EBSgp3CostPerGBMonth is the default cost of gp3 EBS storage per GB per month.
	EBSgp3CostPerGBMonth = 0.08
	// EBSio1CostPerGBMonth is the default cost of io1 EBS storage per GB per month.
	EBSio1CostPerGBMonth = 0.125
	// EBSio2CostPerGBMonth is the default cost of io2 EBS storage per GB per month.
	EBSio2CostPerGBMonth = 0.125
	// EBSst1CostPerGBMonth is the default cost of st1 EBS storage per GB per month.
	EBSst1CostPerGBMonth = 0.045
	// EBSsc1CostPerGBMonth is the default cost of sc1 EBS storage per GB per month.
	EBSsc1CostPerGBMonth = 0.015

	// CloudWatchLogsCostPerGBMonth is the default cost of CloudWatch Logs storage per GB per month.
	CloudWatchLogsCostPerGBMonth = 0.03

	// NatGatewayCostPerMonth is the default cost of a NAT Gateway (~$0.045/hour * 730 hours in us-east-1).
	NatGatewayCostPerMonth = 32.85

	// ALBCostPerMonth is the default base cost of an ALB/NLB (~$0.0225/hour * 730 hours).
	ALBCostPerMonth = 16.43

	// CLBCostPerMonth is the default base cost of a CLB (~$0.025/hour * 730 hours).
	CLBCostPerMonth = 18.25

	// RDSStorageCostPerGBMonth is the default cost of RDS gp2 storage per GB per month.
	RDSStorageCostPerGBMonth = 0.115
	// RDSSnapshotCostPerGBMonth is the default cost of RDS snapshot storage per GB per month.
	RDSSnapshotCostPerGBMonth = 0.095

	// EBSSnapshotCostPerGBMonth is the default cost of EBS snapshot storage per GB per month.
	EBSSnapshotCostPerGBMonth = 0.05

	// ECRStorageCostPerGBMonth is the default cost of ECR storage per GB per month.
	ECRStorageCostPerGBMonth = 0.10
)

// ec2InstancePricing maps common EC2 instance types to approximate monthly on-demand costs
// (us-east-1, Linux, Shared tenancy). Used as a fallback when the Pricing API does not return
// a rate for the instance type in the caller's region.
var ec2InstancePricing = map[string]float64{
	"t2.nano": 4.25, "t2.micro": 8.47, "t2.small": 16.94, "t2.medium": 33.87, "t2.large": 67.74, "t2.xlarge": 135.49, "t2.2xlarge": 270.98,
	"t3.nano": 3.80, "t3.micro": 7.59, "t3.small": 15.18, "t3.medium": 30.37, "t3.large": 60.74, "t3.xlarge": 121.47, "t3.2xlarge": 242.94,
	"t3a.nano": 3.43, "t3a.micro": 6.86, "t3a.small": 13.73, "t3a.medium": 27.45, "t3a.large": 54.90, "t3a.xlarge": 109.81, "t3a.2xlarge": 219.62,
	"t4g.nano": 3.07, "t4g.micro": 6.13, "t4g.small": 12.26, "t4g.medium": 24.53, "t4g.large": 49.06, "t4g.xlarge": 98.11, "t4g.2xlarge": 196.22,
	"m5.large": 70.08, "m5.xlarge": 140.16, "m5.2xlarge": 280.32, "m5.4xlarge": 560.64, "m5.8xlarge": 1121.28, "m5.12xlarge": 1681.92, "m5.16xlarge": 2242.56, "m5.24xlarge": 3363.84,
	"m6i.large": 70.08, "m6i.xlarge": 140.16, "m6i.2xlarge": 280.32, "m6i.4xlarge": 560.64, "m6i.8xlarge": 1121.28,
	"m6g.large": 56.06, "m6g.xlarge": 112.13, "m6g.2xlarge": 224.26, "m6g.4xlarge": 448.51, "m6g.8xlarge": 897.02,
	"m7g.large": 59.57, "m7g.xlarge": 119.14, "m7g.2xlarge": 238.27, "m7g.4xlarge": 476.55, "m7g.8xlarge": 953.10,
	"c5.large": 62.05, "c5.xlarge": 124.10, "c5.2xlarge": 248.20, "c5.4xlarge": 496.40, "c5.9xlarge": 1116.90, "c5.12xlarge": 1489.20, "c5.18xlarge": 2233.80, "c5.24xlarge": 2978.40,
	"c6i.large": 62.05, "c6i.xlarge": 124.10, "c6i.2xlarge": 248.20, "c6i.4xlarge": 496.40,
	"c6g.large": 49.64, "c6g.xlarge": 99.28, "c6g.2xlarge": 198.56, "c6g.4xlarge": 397.12,
	"c7g.large": 52.71, "c7g.xlarge": 105.41, "c7g.2xlarge": 210.82, "c7g.4xlarge": 421.65,
	"r5.large": 91.98, "r5.xlarge": 183.96, "r5.2xlarge": 367.92, "r5.4xlarge": 735.84, "r5.8xlarge": 1471.68, "r5.12xlarge": 2207.52,
	"r6i.large": 91.98, "r6i.xlarge": 183.96, "r6i.2xlarge": 367.92, "r6i.4xlarge": 735.84,
	"r6g.large": 73.58, "r6g.xlarge": 147.17, "r6g.2xlarge": 294.34, "r6g.4xlarge": 588.67,
	"r7g.large": 78.18, "r7g.xlarge": 156.37, "r7g.2xlarge": 312.73, "r7g.4xlarge": 625.47,
}

// ebsSpec maps each supported EBS volume type to its Pricing API variant name and default rate.
var ebsSpec = map[types.VolumeType]struct {
	variant     string
	defaultRate float64
}{
	types.VolumeTypeGp2: {"gp2", EBSgp2CostPerGBMonth},
	types.VolumeTypeGp3: {"gp3", EBSgp3CostPerGBMonth},
	types.VolumeTypeIo1: {"io1", EBSio1CostPerGBMonth},
	types.VolumeTypeIo2: {"io2", EBSio2CostPerGBMonth},
	types.VolumeTypeSt1: {"st1", EBSst1CostPerGBMonth},
	types.VolumeTypeSc1: {"sc1", EBSsc1CostPerGBMonth},
}

// rdsInstancePricing maps RDS instance classes to approximate monthly compute costs (us-east-1,
// Single-AZ, On-Demand MySQL).
var rdsInstancePricing = map[string]float64{
	"db.t3.micro": 12.41, "db.t3.small": 24.82, "db.t3.medium": 49.64, "db.t3.large": 99.28, "db.t3.xlarge": 198.56,
	"db.t4g.micro": 11.52, "db.t4g.small": 23.04, "db.t4g.medium": 46.08, "db.t4g.large": 92.16, "db.t4g.xlarge": 184.32,
	"db.m5.large": 124.10, "db.m5.xlarge": 248.20, "db.m5.2xlarge": 496.40, "db.m5.4xlarge": 992.80, "db.m5.8xlarge": 1985.60,
	"db.m6g.large": 118.26, "db.m6g.xlarge": 236.52, "db.m6g.2xlarge": 473.04, "db.m6g.4xlarge": 946.08, "db.m6g.8xlarge": 1892.16,
	"db.m7g.large": 127.02, "db.m7g.xlarge": 254.04, "db.m7g.2xlarge": 508.08, "db.m7g.4xlarge": 1016.16, "db.m7g.8xlarge": 2032.32,
	"db.r5.large": 175.20, "db.r5.xlarge": 350.40, "db.r5.2xlarge": 700.80, "db.r5.4xlarge": 1401.60, "db.r5.8xlarge": 2803.20,
	"db.r6g.large": 166.44, "db.r6g.xlarge": 332.88, "db.r6g.2xlarge": 665.76, "db.r6g.4xlarge": 1331.52, "db.r6g.8xlarge": 2663.04,
	"db.r7g.large": 178.85, "db.r7g.xlarge": 357.70, "db.r7g.2xlarge": 715.40, "db.r7g.4xlarge": 1430.80, "db.r7g.8xlarge": 2861.60,
}

var sagemakerInstancePricing = map[string]float64{
	"ml.t2.medium": 46.72, "ml.t2.large": 93.44, "ml.t2.xlarge": 186.15, "ml.t2.2xlarge": 372.30,
	"ml.t3.medium": 41.61, "ml.t3.large": 83.22, "ml.t3.xlarge": 166.44, "ml.t3.2xlarge": 332.88,
	"ml.m4.xlarge": 204.40, "ml.m4.2xlarge": 408.80, "ml.m4.4xlarge": 817.60,
	"ml.m5.large": 101.91, "ml.m5.xlarge": 203.82, "ml.m5.2xlarge": 407.64, "ml.m5.4xlarge": 815.28,
	"ml.c4.xlarge": 176.66, "ml.c4.2xlarge": 353.32, "ml.c4.4xlarge": 706.64,
	"ml.c5.large": 87.60, "ml.c5.xlarge": 175.20, "ml.c5.2xlarge": 350.40, "ml.c5.4xlarge": 700.80,
	"ml.p2.xlarge": 1144.80, "ml.g4dn.xlarge": 538.72, "ml.g5.xlarge": 802.82,
	"ml.inf1.xlarge": 218.40,
}
