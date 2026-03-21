package outputshared

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/utils/ec2"
	"github.com/elC0mpa/aws-doctor/utils/pricing"
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
func PresentElasticIP(ip types.Address) ResourceRow {
	return ResourceRow{
		Category:      "Elastic IP",
		Identifier:    aws.ToString(ip.PublicIp),
		EstimatedCost: fmt.Sprintf("$%.2f", pricing.CalculateEIPMonthlyCost()),
		Metric:        NAValue,
		Age:           NAValue,
		Details:       fmt.Sprintf("Allocation ID: %s", aws.ToString(ip.AllocationId)),
	}
}

// PresentEBSVolume returns a ResourceRow for an EBS volume
func PresentEBSVolume(vol types.Volume, status string) ResourceRow {
	return ResourceRow{
		Category:      fmt.Sprintf("EBS Volume (%s)", status),
		Identifier:    aws.ToString(vol.VolumeId),
		EstimatedCost: fmt.Sprintf("$%.2f", pricing.CalculateEBSMonthlyCost(aws.ToInt32(vol.Size), vol.VolumeType)),
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
func PresentLoadBalancer(lb elbtypes.LoadBalancer) ResourceRow {
	return ResourceRow{
		Category:      "Elastic Load Balancer",
		Identifier:    aws.ToString(lb.LoadBalancerArn),
		EstimatedCost: fmt.Sprintf("$%.2f", pricing.CalculateLoadBalancerMonthlyCost(lb.Type)),
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
		EstimatedCost: NAValue,
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
