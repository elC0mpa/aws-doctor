package model

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

// CategorySummary holds the name, count, and estimated cost for a waste category.
type CategorySummary struct {
	Name  string
	Count int
	Cost  float64
}

// RenderWasteInput represents the input data for rendering the waste report
type RenderWasteInput struct {
	AccountID              string
	ElasticIPs             []types.Address
	UnusedVolumes          []types.Volume
	StoppedVolumes         []types.Volume
	Ris                    []RiExpirationInfo
	StoppedInstances       []types.Instance
	IdleEC2Instances       []EC2IdleInstanceInfo
	LoadBalancers          []elbtypes.LoadBalancer
	UnusedAMIs             []AMIWasteInfo
	OrphanedSnapshots      []SnapshotWasteInfo
	UnusedKeyPairs         []KeyPairWasteInfo
	S3Buckets              []S3BucketWasteInfo
	S3MultipartUploads     []S3MultipartUploadWasteInfo
	CloudWatchLogGroups    []CloudWatchLogsWasteInfo
	RDSInstances           []RDSInstanceWasteInfo
	RDSSnapshots           []RDSSnapshotWasteInfo
	RDSIdleInstances       []RDSIdleInstanceInfo
	IdleNATGateways        []NATGatewayWasteInfo
	IdleLoadBalancers      []ELBIdleInfo
	OverProvisionedLambdas []LambdaOverProvisionedInfo
	IdleSageMakerEndpoints []IdleSageMakerEndpointInfo
	ECRNoLifecyclePolicies []ECRNoLifecyclePolicyInfo
	ECREmptyRepositories   []ECREmptyRepositoryInfo
	ECRUntaggedImages      []ECRUntaggedImageInfo
	UnusedSecrets          []UnusedSecretInfo
}

// RenderCostComparisonInput represents the input data for rendering the cost comparison report
type RenderCostComparisonInput struct {
	AccountID        string
	LastTotalCost    string
	CurrentTotalCost string
	LastMonth        *CostInfo
	CurrentMonth     *CostInfo
}
