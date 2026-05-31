package model

import (
	"testing"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

func assertLenOne(t *testing.T, field string, length int) {
	t.Helper()

	if length != 1 {
		t.Errorf("Expected %s length to be 1, got %d", field, length)
	}
}

func TestMerge(t *testing.T) {
	dest := RenderWasteInput{
		AccountID: "123",
		Errors: map[string]string{
			"EC2": "timeout",
		},
	}

	src := RenderWasteInput{
		AccountID: "123",
		Errors: map[string]string{
			"IAM": "access denied",
		},
		ElasticIPs:             []ec2types.Address{{}},
		UnusedVolumes:          []ec2types.Volume{{}},
		StoppedVolumes:         []ec2types.Volume{{}},
		Ris:                    []RiExpirationInfo{{}},
		StoppedInstances:       []ec2types.Instance{{}},
		IdleEC2Instances:       []EC2IdleInstanceInfo{{}},
		LoadBalancers:          []elbtypes.LoadBalancer{{}},
		UnusedAMIs:             []AMIWasteInfo{{}},
		OrphanedSnapshots:      []SnapshotWasteInfo{{}},
		UnusedKeyPairs:         []KeyPairWasteInfo{{}},
		S3Buckets:              []S3BucketWasteInfo{{}},
		S3MultipartUploads:     []S3MultipartUploadWasteInfo{{}},
		CloudWatchLogGroups:    []CloudWatchLogsWasteInfo{{}},
		RDSInstances:           []RDSInstanceWasteInfo{{}},
		RDSSnapshots:           []RDSSnapshotWasteInfo{{}},
		RDSIdleInstances:       []RDSIdleInstanceInfo{{}},
		IdleNATGateways:        []NATGatewayWasteInfo{{}},
		IdleLoadBalancers:      []ELBIdleInfo{{}},
		OverProvisionedLambdas: []LambdaOverProvisionedInfo{{}},
		IdleSageMakerEndpoints: []IdleSageMakerEndpointInfo{{}},
		ECRNoLifecyclePolicies: []ECRNoLifecyclePolicyInfo{{}},
		ECREmptyRepositories:   []ECREmptyRepositoryInfo{{}},
		ECRUntaggedImages:      []ECRUntaggedImageInfo{{}},
		UnusedSecrets:          []UnusedSecretInfo{{}},
		UnusedIAMUsers:         []IAMUserWasteInfo{{}},
		RootUserWaste:          []IAMRootUserWasteInfo{{}},
	}

	dest.Merge(src)

	if dest.Errors["IAM"] != "access denied" {
		t.Error("Failed")
	}

	assertLenOne(t, "ElasticIPs", len(dest.ElasticIPs))
	assertLenOne(t, "UnusedVolumes", len(dest.UnusedVolumes))
	assertLenOne(t, "StoppedVolumes", len(dest.StoppedVolumes))
	assertLenOne(t, "Ris", len(dest.Ris))
	assertLenOne(t, "StoppedInstances", len(dest.StoppedInstances))
	assertLenOne(t, "IdleEC2Instances", len(dest.IdleEC2Instances))
	assertLenOne(t, "LoadBalancers", len(dest.LoadBalancers))
	assertLenOne(t, "UnusedAMIs", len(dest.UnusedAMIs))
	assertLenOne(t, "OrphanedSnapshots", len(dest.OrphanedSnapshots))
	assertLenOne(t, "UnusedKeyPairs", len(dest.UnusedKeyPairs))
	assertLenOne(t, "S3Buckets", len(dest.S3Buckets))
	assertLenOne(t, "S3MultipartUploads", len(dest.S3MultipartUploads))
	assertLenOne(t, "CloudWatchLogGroups", len(dest.CloudWatchLogGroups))
	assertLenOne(t, "RDSInstances", len(dest.RDSInstances))
	assertLenOne(t, "RDSSnapshots", len(dest.RDSSnapshots))
	assertLenOne(t, "RDSIdleInstances", len(dest.RDSIdleInstances))
	assertLenOne(t, "IdleNATGateways", len(dest.IdleNATGateways))
	assertLenOne(t, "IdleLoadBalancers", len(dest.IdleLoadBalancers))
	assertLenOne(t, "OverProvisionedLambdas", len(dest.OverProvisionedLambdas))
	assertLenOne(t, "IdleSageMakerEndpoints", len(dest.IdleSageMakerEndpoints))
	assertLenOne(t, "ECRNoLifecyclePolicies", len(dest.ECRNoLifecyclePolicies))
	assertLenOne(t, "ECREmptyRepositories", len(dest.ECREmptyRepositories))
	assertLenOne(t, "ECRUntaggedImages", len(dest.ECRUntaggedImages))
	assertLenOne(t, "UnusedSecrets", len(dest.UnusedSecrets))
	assertLenOne(t, "UnusedIAMUsers", len(dest.UnusedIAMUsers))
	assertLenOne(t, "RootUserWaste", len(dest.RootUserWaste))
}
