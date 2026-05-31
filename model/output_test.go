package model

import (
	"testing"
	
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

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
		ElasticIPs: []ec2types.Address{{}},
		UnusedVolumes: []ec2types.Volume{{}},
		StoppedVolumes: []ec2types.Volume{{}},
		Ris: []RiExpirationInfo{{}},
		StoppedInstances: []ec2types.Instance{{}},
		IdleEC2Instances: []EC2IdleInstanceInfo{{}},
		LoadBalancers: []elbtypes.LoadBalancer{{}},
		UnusedAMIs: []AMIWasteInfo{{}},
		OrphanedSnapshots: []SnapshotWasteInfo{{}},
		UnusedKeyPairs: []KeyPairWasteInfo{{}},
		S3Buckets: []S3BucketWasteInfo{{}},
		S3MultipartUploads: []S3MultipartUploadWasteInfo{{}},
		CloudWatchLogGroups: []CloudWatchLogsWasteInfo{{}},
		RDSInstances: []RDSInstanceWasteInfo{{}},
		RDSSnapshots: []RDSSnapshotWasteInfo{{}},
		RDSIdleInstances: []RDSIdleInstanceInfo{{}},
		IdleNATGateways: []NATGatewayWasteInfo{{}},
		IdleLoadBalancers: []ELBIdleInfo{{}},
		OverProvisionedLambdas: []LambdaOverProvisionedInfo{{}},
		IdleSageMakerEndpoints: []IdleSageMakerEndpointInfo{{}},
		ECRNoLifecyclePolicies: []ECRNoLifecyclePolicyInfo{{}},
		ECREmptyRepositories: []ECREmptyRepositoryInfo{{}},
		ECRUntaggedImages: []ECRUntaggedImageInfo{{}},
		UnusedSecrets: []UnusedSecretInfo{{}},
		UnusedIAMUsers: []IAMUserWasteInfo{{}},
		RootUserWaste: []IAMRootUserWasteInfo{{}},
	}

	dest.Merge(src)

	if dest.Errors["IAM"] != "access denied" { t.Error("Failed") }
	if len(dest.ElasticIPs) != 1 { t.Error("Failed") }
	if len(dest.UnusedVolumes) != 1 { t.Error("Failed") }
	if len(dest.StoppedVolumes) != 1 { t.Error("Failed") }
	if len(dest.Ris) != 1 { t.Error("Failed") }
	if len(dest.StoppedInstances) != 1 { t.Error("Failed") }
	if len(dest.IdleEC2Instances) != 1 { t.Error("Failed") }
	if len(dest.LoadBalancers) != 1 { t.Error("Failed") }
	if len(dest.UnusedAMIs) != 1 { t.Error("Failed") }
	if len(dest.OrphanedSnapshots) != 1 { t.Error("Failed") }
	if len(dest.UnusedKeyPairs) != 1 { t.Error("Failed") }
	if len(dest.S3Buckets) != 1 { t.Error("Failed") }
	if len(dest.S3MultipartUploads) != 1 { t.Error("Failed") }
	if len(dest.CloudWatchLogGroups) != 1 { t.Error("Failed") }
	if len(dest.RDSInstances) != 1 { t.Error("Failed") }
	if len(dest.RDSSnapshots) != 1 { t.Error("Failed") }
	if len(dest.RDSIdleInstances) != 1 { t.Error("Failed") }
	if len(dest.IdleNATGateways) != 1 { t.Error("Failed") }
	if len(dest.IdleLoadBalancers) != 1 { t.Error("Failed") }
	if len(dest.OverProvisionedLambdas) != 1 { t.Error("Failed") }
	if len(dest.IdleSageMakerEndpoints) != 1 { t.Error("Failed") }
	if len(dest.ECRNoLifecyclePolicies) != 1 { t.Error("Failed") }
	if len(dest.ECREmptyRepositories) != 1 { t.Error("Failed") }
	if len(dest.ECRUntaggedImages) != 1 { t.Error("Failed") }
	if len(dest.UnusedSecrets) != 1 { t.Error("Failed") }
	if len(dest.UnusedIAMUsers) != 1 { t.Error("Failed") }
	if len(dest.RootUserWaste) != 1 { t.Error("Failed") }
}
