package model

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

// CategorySummary holds the name, count, and estimated cost for a waste category.
type CategorySummary struct {
	Name  string
	Count int
	Cost  float64
}

// ScopeResult holds the result of a waste detection scope, for partial stream updates.
type ScopeResult struct {
	Scope    string
	Input    RenderWasteInput // Populated only with the fields related to this scope
	Duration time.Duration    // Time taken to retrieve the data for this scope
	Err      error
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
	UnusedIAMUsers         []IAMUserWasteInfo
	RootUserWaste          []IAMRootUserWasteInfo
	PublicIPv4Summary      *PublicIPv4Summary
	Errors                 map[string]string
	Flags                  Flags
}

// RenderCostComparisonInput represents the input data for rendering the cost comparison report
type RenderCostComparisonInput struct {
	AccountID        string
	LastTotalCost    string
	CurrentTotalCost string
	LastMonth        *CostInfo
	CurrentMonth     *CostInfo
}

// Merge copies all elements from the source struct slices into the destination struct.
//
//nolint:gocyclo
func (dest *RenderWasteInput) Merge(src RenderWasteInput) {
	if len(src.ElasticIPs) > 0 {
		dest.ElasticIPs = append(dest.ElasticIPs, src.ElasticIPs...)
	}

	if len(src.UnusedVolumes) > 0 {
		dest.UnusedVolumes = append(dest.UnusedVolumes, src.UnusedVolumes...)
	}

	if len(src.StoppedVolumes) > 0 {
		dest.StoppedVolumes = append(dest.StoppedVolumes, src.StoppedVolumes...)
	}

	if len(src.Ris) > 0 {
		dest.Ris = append(dest.Ris, src.Ris...)
	}

	if len(src.StoppedInstances) > 0 {
		dest.StoppedInstances = append(dest.StoppedInstances, src.StoppedInstances...)
	}

	if len(src.IdleEC2Instances) > 0 {
		dest.IdleEC2Instances = append(dest.IdleEC2Instances, src.IdleEC2Instances...)
	}

	if len(src.LoadBalancers) > 0 {
		dest.LoadBalancers = append(dest.LoadBalancers, src.LoadBalancers...)
	}

	if len(src.UnusedAMIs) > 0 {
		dest.UnusedAMIs = append(dest.UnusedAMIs, src.UnusedAMIs...)
	}

	if len(src.OrphanedSnapshots) > 0 {
		dest.OrphanedSnapshots = append(dest.OrphanedSnapshots, src.OrphanedSnapshots...)
	}

	if len(src.UnusedKeyPairs) > 0 {
		dest.UnusedKeyPairs = append(dest.UnusedKeyPairs, src.UnusedKeyPairs...)
	}

	if len(src.S3Buckets) > 0 {
		dest.S3Buckets = append(dest.S3Buckets, src.S3Buckets...)
	}

	if len(src.S3MultipartUploads) > 0 {
		dest.S3MultipartUploads = append(dest.S3MultipartUploads, src.S3MultipartUploads...)
	}

	if len(src.CloudWatchLogGroups) > 0 {
		dest.CloudWatchLogGroups = append(dest.CloudWatchLogGroups, src.CloudWatchLogGroups...)
	}

	if len(src.RDSInstances) > 0 {
		dest.RDSInstances = append(dest.RDSInstances, src.RDSInstances...)
	}

	if len(src.RDSSnapshots) > 0 {
		dest.RDSSnapshots = append(dest.RDSSnapshots, src.RDSSnapshots...)
	}

	if len(src.RDSIdleInstances) > 0 {
		dest.RDSIdleInstances = append(dest.RDSIdleInstances, src.RDSIdleInstances...)
	}

	if len(src.IdleNATGateways) > 0 {
		dest.IdleNATGateways = append(dest.IdleNATGateways, src.IdleNATGateways...)
	}

	if len(src.IdleLoadBalancers) > 0 {
		dest.IdleLoadBalancers = append(dest.IdleLoadBalancers, src.IdleLoadBalancers...)
	}

	if len(src.OverProvisionedLambdas) > 0 {
		dest.OverProvisionedLambdas = append(dest.OverProvisionedLambdas, src.OverProvisionedLambdas...)
	}

	if len(src.IdleSageMakerEndpoints) > 0 {
		dest.IdleSageMakerEndpoints = append(dest.IdleSageMakerEndpoints, src.IdleSageMakerEndpoints...)
	}

	if len(src.ECRNoLifecyclePolicies) > 0 {
		dest.ECRNoLifecyclePolicies = append(dest.ECRNoLifecyclePolicies, src.ECRNoLifecyclePolicies...)
	}

	if len(src.ECREmptyRepositories) > 0 {
		dest.ECREmptyRepositories = append(dest.ECREmptyRepositories, src.ECREmptyRepositories...)
	}

	if len(src.ECRUntaggedImages) > 0 {
		dest.ECRUntaggedImages = append(dest.ECRUntaggedImages, src.ECRUntaggedImages...)
	}

	if len(src.UnusedSecrets) > 0 {
		dest.UnusedSecrets = append(dest.UnusedSecrets, src.UnusedSecrets...)
	}

	if len(src.UnusedIAMUsers) > 0 {
		dest.UnusedIAMUsers = append(dest.UnusedIAMUsers, src.UnusedIAMUsers...)
	}

	if len(src.RootUserWaste) > 0 {
		dest.RootUserWaste = append(dest.RootUserWaste, src.RootUserWaste...)
	}

	if len(src.Errors) > 0 {
		if dest.Errors == nil {
			dest.Errors = make(map[string]string)
		}

		for k, v := range src.Errors {
			dest.Errors[k] = v
		}
	}

	dest.Flags = src.Flags
}
