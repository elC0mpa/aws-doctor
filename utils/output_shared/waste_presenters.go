package outputshared

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/pricing"
	"github.com/elC0mpa/aws-doctor/utils/ec2"
)

// PresentS3Bucket returns a ResourceRow for an S3 bucket
func PresentS3Bucket(b model.S3BucketWasteInfo) ResourceRow {
	return ResourceRow{
		Category:      fmt.Sprintf("S3 Bucket (%s)", b.Reason),
		Identifier:    b.BucketName,
		EstimatedCost: NAValue,
		Metric:        NAValue,
		Age:           NAValue,
		Details:       fmt.Sprintf("Created on %s", b.CreationDate.Format(time.RFC3339)),
	}
}

// PresentS3MultipartUpload returns a ResourceRow for an S3 bucket with incomplete multipart uploads
func PresentS3MultipartUpload(b model.S3MultipartUploadWasteInfo) ResourceRow {
	return ResourceRow{
		Category:      "S3 Multipart Uploads",
		Identifier:    b.BucketName,
		EstimatedCost: NAValue,
		Metric:        fmt.Sprintf("%d incomplete uploads", b.UploadCount),
		Age:           NAValue,
		Details:       NAValue,
	}
}

// PresentElasticIP returns a ResourceRow for an unused Elastic IP
func PresentElasticIP(ip types.Address, pricingSvc pricing.Service) ResourceRow {
	return ResourceRow{
		Category:      "Elastic IP",
		Identifier:    aws.ToString(ip.PublicIp),
		EstimatedCost: fmt.Sprintf("$%.2f", pricingSvc.CalculateEIPMonthlyCost()),
		Metric:        NAValue,
		Age:           NAValue,
		Details:       fmt.Sprintf("Allocation ID: %s", aws.ToString(ip.AllocationId)),
	}
}

// PresentEBSVolume returns a ResourceRow for an EBS volume
func PresentEBSVolume(vol types.Volume, status string, pricingSvc pricing.Service) ResourceRow {
	return ResourceRow{
		Category:      fmt.Sprintf("EBS Volume (%s)", status),
		Identifier:    aws.ToString(vol.VolumeId),
		EstimatedCost: fmt.Sprintf("$%.2f", pricingSvc.CalculateEBSMonthlyCost(aws.ToInt32(vol.Size), vol.VolumeType)),
		Metric:        fmt.Sprintf("%d GiB", aws.ToInt32(vol.Size)),
		Age:           NAValue,
		Details:       fmt.Sprintf("State: %s, Created on %s", vol.State, vol.CreateTime.Format(time.RFC3339)),
	}
}

// PresentStoppedInstance returns a ResourceRow for a stopped EC2 instance
func PresentStoppedInstance(instance types.Instance) ResourceRow {
	now := time.Now()
	daysAgo := NAValue
	stoppedAt := NAValue

	if instance.StateTransitionReason != nil {
		if stopAt, err := ec2.ParseTransitionDate(*instance.StateTransitionReason); err == nil {
			stoppedAt = stopAt.Format(time.RFC3339)
			daysAgo = fmt.Sprintf("%d", int(now.Sub(stopAt).Hours()/24))
		}
	}

	return ResourceRow{
		Category:      "Stopped EC2 Instance",
		Identifier:    aws.ToString(instance.InstanceId),
		EstimatedCost: NAValue,
		Metric:        string(instance.InstanceType),
		Age:           daysAgo,
		Details:       fmt.Sprintf("Stopped on %s", stoppedAt),
	}
}

// PresentReservedInstance returns a ResourceRow for a reserved instance
func PresentReservedInstance(ri model.RiExpirationInfo) ResourceRow {
	return ResourceRow{
		Category:      fmt.Sprintf("Reserved Instances (%s)", ri.Status),
		Identifier:    ri.ReservedInstanceID,
		EstimatedCost: NAValue,
		Metric:        ri.InstanceType,
		Age:           fmt.Sprintf("%d", ri.DaysUntilExpiry),
		Details:       fmt.Sprintf("State: %s", ri.State),
	}
}

// PresentLoadBalancer returns a ResourceRow for an unused load balancer
func PresentLoadBalancer(lb elbtypes.LoadBalancer, pricingSvc pricing.Service) ResourceRow {
	return ResourceRow{
		Category:      "Elastic Load Balancer",
		Identifier:    aws.ToString(lb.LoadBalancerArn),
		EstimatedCost: fmt.Sprintf("$%.2f", pricingSvc.CalculateLoadBalancerMonthlyCost(lb.Type)),
		Metric:        string(lb.Type),
		Age:           NAValue,
		Details:       fmt.Sprintf("Created on %s", lb.CreatedTime.Format(time.RFC3339)),
	}
}

// PresentAMI returns a ResourceRow for an unused AMI
func PresentAMI(ami model.AMIWasteInfo) ResourceRow {
	return ResourceRow{
		Category:      "Unused AMI",
		Identifier:    ami.ImageID,
		EstimatedCost: fmt.Sprintf("$%.2f", ami.MaxPotentialSaving),
		Metric:        NAValue,
		Age:           fmt.Sprintf("%d", ami.DaysSinceCreate),
		Details:       fmt.Sprintf("Created on %s", ami.CreationDate.Format(time.RFC3339)),
	}
}

// PresentSnapshot returns a ResourceRow for a snapshot
func PresentSnapshot(snap model.SnapshotWasteInfo) ResourceRow {
	return ResourceRow{
		Category:      fmt.Sprintf("EBS Snapshot (%s)", snap.Category),
		Identifier:    snap.SnapshotID,
		EstimatedCost: fmt.Sprintf("$%.2f", snap.MaxPotentialSavings),
		Metric:        fmt.Sprintf("%d GiB", snap.SizeGB),
		Age:           fmt.Sprintf("%d", snap.DaysSinceCreate),
		Details:       fmt.Sprintf("Created on %s", snap.StartTime.Format(time.RFC3339)),
	}
}

// PresentKeyPair returns a ResourceRow for an unused key pair
func PresentKeyPair(kp model.KeyPairWasteInfo) ResourceRow {
	return ResourceRow{
		Category:      "Unused Key Pair",
		Identifier:    kp.KeyName,
		EstimatedCost: NAValue,
		Metric:        NAValue,
		Age:           fmt.Sprintf("%d", kp.DaysSinceCreate),
		Details:       fmt.Sprintf("Created on %s", kp.CreateTime.Format(time.RFC3339)),
	}
}

// PresentCloudWatchLogGroup returns a ResourceRow for a CloudWatch Log Group
func PresentCloudWatchLogGroup(lg model.CloudWatchLogsWasteInfo) ResourceRow {
	sizeGB := float64(lg.StoredBytes) / (1024 * 1024 * 1024)

	return ResourceRow{
		Category:      "CloudWatch Log Group",
		Identifier:    lg.LogGroupName,
		EstimatedCost: fmt.Sprintf("$%.2f", lg.EstimatedMonthlyCost),
		Metric:        fmt.Sprintf("%.2f GB stored", sizeGB),
		Age:           NAValue,
		Details:       fmt.Sprintf("Creation time: %s", lg.CreationTime.Format(time.RFC3339)),
	}
}

// PresentRDSInstance returns a ResourceRow for a stopped RDS instance
func PresentRDSInstance(inst model.RDSInstanceWasteInfo) ResourceRow {
	return ResourceRow{
		Category:      "Stopped RDS Instance",
		Identifier:    inst.DBInstanceID,
		EstimatedCost: fmt.Sprintf("$%.2f", inst.EstimatedMonthlyCost),
		Metric:        fmt.Sprintf("Is Multi AZ: %t", inst.MultiAZ),
		Age:           NAValue,
		Details:       fmt.Sprintf("Engine: %s / Status: %s", inst.Engine, inst.Status),
	}
}

// PresentRDSSnapshot returns a ResourceRow for an old RDS manual snapshot
func PresentRDSSnapshot(snap model.RDSSnapshotWasteInfo) ResourceRow {
	return ResourceRow{
		Category:      "Old Manual RDS Snapshot",
		Identifier:    snap.DBSnapshotID,
		EstimatedCost: fmt.Sprintf("$%.2f", snap.EstimatedMonthlyCost),
		Metric:        fmt.Sprintf("%d GiB", snap.AllocatedStorage),
		Age:           fmt.Sprintf("%d", snap.DaysSinceCreate),
		Details:       fmt.Sprintf("Created on %s / Engine: %s", snap.SnapshotCreateTime.Format(time.RFC3339), snap.Engine),
	}
}

// PresentIdleLoadBalancer returns a ResourceRow for an idle load balancer with zero connections.
func PresentIdleLoadBalancer(lb model.ELBIdleInfo) ResourceRow {
	return ResourceRow{
		Category:      "Idle Load Balancer",
		Identifier:    lb.ARN,
		EstimatedCost: fmt.Sprintf("$%.2f", lb.EstimatedMonthlyCost),
		Metric:        lb.Type,
		Age:           NAValue,
		Details:       fmt.Sprintf("0 connections over %d days", lb.DaysChecked),
	}
}

// PresentRDSIdleInstance returns a ResourceRow for an idle RDS instance
func PresentRDSIdleInstance(inst model.RDSIdleInstanceInfo) ResourceRow {
	return ResourceRow{
		Category:      "Idle RDS Instance",
		Identifier:    inst.DBInstanceID,
		EstimatedCost: fmt.Sprintf("$%.2f", inst.EstimatedMonthlyCost),
		Metric:        fmt.Sprintf("No active connections in last %d days", inst.DaysChecked),
		Age:           NAValue,
		Details:       fmt.Sprintf("Engine: %s", inst.Engine),
	}
}

// PresentIdleNATGateway returns a ResourceRow for an idle NAT Gateway
func PresentIdleNATGateway(ng model.NATGatewayWasteInfo) ResourceRow {
	return ResourceRow{
		Category:      "Idle NAT Gateway",
		Identifier:    ng.NATGatewayID,
		EstimatedCost: fmt.Sprintf("$%.2f", ng.EstimatedMonthlyCost),
		Metric:        fmt.Sprintf("%.2f bytes transferred", ng.BytesOutToDestination),
		Age:           fmt.Sprintf("%d", ng.DaysSinceCreate),
		Details:       fmt.Sprintf("VPC: %s / Subnet: %s / State: %s", ng.VPCID, ng.SubnetID, ng.State),
	}
}

// PresentLambdaOverProvisioned returns a ResourceRow for an over-provisioned Lambda function
func PresentLambdaOverProvisioned(fn model.LambdaOverProvisionedInfo) ResourceRow {
	return ResourceRow{
		Category:      "Lambda (Over-Provisioned)",
		Identifier:    fn.FunctionName,
		EstimatedCost: NAValue,
		Metric:        fmt.Sprintf("%.1f%% utilization", fn.MemoryUtilization),
		Age:           NAValue,
		Details:       fmt.Sprintf("Runtime: %s / Configured: %d MB / Used: %d MB / Recommended: %d MB", fn.Runtime, fn.ConfiguredMemoryMB, fn.MaxMemoryUsedMB, fn.RecommendedMemoryMB),
	}
}

// PresentIdleSageMakerEndpoint returns a ResourceRow for an idle SageMaker real-time inference
// endpoint. The details column summarizes variants in the form "variant(instance_type x count)".
func PresentIdleSageMakerEndpoint(ep model.IdleSageMakerEndpointInfo) ResourceRow {
	variantParts := make([]string, 0, len(ep.Variants))
	for _, v := range ep.Variants {
		variantParts = append(variantParts, fmt.Sprintf("%s(%s x%d)", v.VariantName, v.InstanceType, v.InstanceCount))
	}

	return ResourceRow{
		Category:      "SageMaker Endpoints (Idle)",
		Identifier:    ep.EndpointName,
		EstimatedCost: fmt.Sprintf("$%.2f", ep.EstimatedMonthlyCost),
		Metric:        fmt.Sprintf("0 invocations over %d days", ep.DaysChecked),
		Age:           NAValue,
		Details:       strings.Join(variantParts, ", "),
	}
}

// PresentECRNoLifecyclePolicy returns a ResourceRow for an ECR repository missing a lifecycle policy
func PresentECRNoLifecyclePolicy(repo model.ECRNoLifecyclePolicyInfo) ResourceRow {
	return ResourceRow{
		Category:      "ECR (No Lifecycle Policy)",
		Identifier:    repo.RepositoryName,
		EstimatedCost: NAValue,
		Metric:        NAValue,
		Age:           NAValue,
		Details:       "Lifecycle policy is recommended",
	}
}

// PresentECREmptyRepository returns a ResourceRow for an empty ECR repository
func PresentECREmptyRepository(repo model.ECREmptyRepositoryInfo) ResourceRow {
	return ResourceRow{
		Category:      "ECR (Empty Repository)",
		Identifier:    repo.RepositoryName,
		EstimatedCost: NAValue,
		Metric:        "0 images",
		Age:           NAValue,
		Details:       "Consider deleting empty repositories",
	}
}

// PresentECRUntaggedImages returns a ResourceRow for an ECR repository with untagged images
func PresentECRUntaggedImages(repo model.ECRUntaggedImageInfo) ResourceRow {
	sizeGB := float64(repo.UntaggedSizeBytes) / (1024 * 1024 * 1024)

	return ResourceRow{
		Category:      "ECR (Untagged Images)",
		Identifier:    repo.RepositoryName,
		EstimatedCost: fmt.Sprintf("$%.2f", repo.EstimatedMonthlyCost),
		Metric:        fmt.Sprintf("%d untagged images", repo.UntaggedImageCount),
		Age:           NAValue,
		Details:       fmt.Sprintf("Consuming %.2f GB of storage", sizeGB),
	}
}

// PresentUnusedSecret returns a ResourceRow for an unused secret
func PresentUnusedSecret(secret model.UnusedSecretInfo, pricingSvc pricing.Service) ResourceRow {
	lastAccessed := NAValue
	daysAgo := NAValue

	if secret.LastAccessedDate != nil {
		lastAccessed = secret.LastAccessedDate.Format(time.RFC3339)
		daysAgo = fmt.Sprintf("%d", int(time.Since(*secret.LastAccessedDate).Hours()/24))
	}

	return ResourceRow{
		Category:      "Unused Secret",
		Identifier:    secret.Name,
		EstimatedCost: fmt.Sprintf("$%.2f", pricingSvc.CalculateSecretsManagerMonthlyCost(1)),
		Metric:        NAValue,
		Age:           daysAgo,
		Details:       fmt.Sprintf("Last accessed on %s", lastAccessed),
	}
}
