package csvoutput

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/utils/ec2"
	"github.com/elC0mpa/aws-doctor/utils/pricing"
)

const naValue = "-"

// OutputWasteCSV outputs waste detection data as CSV
func OutputWasteCSV(input model.RenderWasteInput) error {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	headers := []string{"Resource Category", "Resource Identifier", "Estimated Monthly Cost (USD)", "Metric / Size", "Age (Days)", "Additional Details"}
	if err := w.Write(headers); err != nil {
		return err
	}

	var rows [][]string
	rows = append(rows, mapEBSVolumes(input.StoppedVolumes, "stopped")...)
	rows = append(rows, mapEBSVolumes(input.UnusedVolumes, "unattached")...)
	rows = append(rows, mapElasticIPs(input.ElasticIPs)...)
	rows = append(rows, mapS3Buckets(input.S3Buckets)...)
	rows = append(rows, mapS3MultipartUploads(input.S3MultipartUploads)...)
	rows = append(rows, mapStoppedInstances(input.StoppedInstances)...)
	rows = append(rows, mapReservedInstances(input.Ris)...)
	rows = append(rows, mapLoadBalancers(input.LoadBalancers)...)
	rows = append(rows, mapAMIs(input.UnusedAMIs)...)

	orphaned, stale := mapSnapshots(input.OrphanedSnapshots)
	rows = append(rows, orphaned...)
	rows = append(rows, stale...)

	rows = append(rows, mapKeyPairs(input.UnusedKeyPairs)...)
	rows = append(rows, mapCloudWatchLogGroups(input.CloudWatchLogGroups)...)
	rows = append(rows, mapRDSInstances(input.RDSInstances)...)
	rows = append(rows, mapRDSSnapshots(input.RDSSnapshots)...)
	rows = append(rows, mapRDSIdleInstances(input.RDSIdleInstances)...)

	return w.WriteAll(rows)
}

func mapS3Buckets(buckets []model.S3BucketWasteInfo) [][]string {
	result := make([][]string, 0, len(buckets))

	for _, b := range buckets {
		category := fmt.Sprintf("S3 Bucket (%s)", b.Reason)
		identifier := b.BucketName
		estimatedCost := naValue
		metric := naValue
		age := naValue
		details := fmt.Sprintf("Created on %s", b.CreationDate.Format(time.RFC3339))
		result = append(result, []string{category, identifier, estimatedCost, metric, age, details})
	}

	return result
}

func mapS3MultipartUploads(buckets []model.S3MultipartUploadWasteInfo) [][]string {
	result := make([][]string, 0, len(buckets))

	for _, b := range buckets {
		category := "S3 Multipart Uploads"
		identifier := b.BucketName
		estimatedCost := naValue
		metric := fmt.Sprintf("%d incomplete uploads", b.UploadCount)
		age := naValue
		details := naValue
		result = append(result, []string{category, identifier, estimatedCost, metric, age, details})
	}

	return result
}

func mapElasticIPs(elasticIPs []types.Address) [][]string {
	result := make([][]string, 0, len(elasticIPs))

	for _, ip := range elasticIPs {
		category := "Elastic IP"
		identifier := aws.ToString(ip.PublicIp)
		estimatedCost := fmt.Sprintf("$%.2f", pricing.CalculateEIPMonthlyCost())
		metric := naValue
		age := naValue
		details := fmt.Sprintf("Allocation ID: %s", aws.ToString(ip.AllocationId))
		result = append(result, []string{category, identifier, estimatedCost, metric, age, details})
	}

	return result
}

func mapEBSVolumes(volumes []types.Volume, status string) [][]string {
	result := make([][]string, 0, len(volumes))

	for _, vol := range volumes {
		category := fmt.Sprintf("EBS Volume (%s)", status)
		identifier := aws.ToString(vol.VolumeId)
		estimatedCost := fmt.Sprintf("$%.2f", pricing.CalculateEBSMonthlyCost(aws.ToInt32(vol.Size), vol.VolumeType))
		metric := fmt.Sprintf("%d GiB", aws.ToInt32(vol.Size))
		age := naValue
		details := fmt.Sprintf("State: %s, Created on %s", vol.State, vol.CreateTime.Format(time.RFC3339))
		result = append(result, []string{category, identifier, estimatedCost, metric, age, details})
	}

	return result
}

func mapStoppedInstances(stoppedInstances []types.Instance) [][]string {
	result := make([][]string, 0, len(stoppedInstances))

	now := time.Now()

	for _, instance := range stoppedInstances {
		daysAgo := naValue
		stoppedAt := naValue

		if instance.StateTransitionReason != nil {
			if stopAt, err := ec2.ParseTransitionDate(*instance.StateTransitionReason); err == nil {
				stoppedAt = stopAt.Format(time.RFC3339)
				daysAgo = fmt.Sprintf("%d", int(now.Sub(stopAt).Hours()/24))
			}
		}

		category := "Stopped EC2 Instance"
		identifier := aws.ToString(instance.InstanceId)
		estimatedCost := naValue
		metric := string(instance.InstanceType)
		age := daysAgo
		details := fmt.Sprintf("Stopped on %s", stoppedAt)
		result = append(result, []string{category, identifier, estimatedCost, metric, age, details})
	}

	return result
}

func mapReservedInstances(ris []model.RiExpirationInfo) [][]string {
	result := make([][]string, 0, len(ris))

	for _, ri := range ris {
		category := fmt.Sprintf("Reserved Instances (%s)", ri.Status)
		identifier := ri.ReservedInstanceID
		estimatedCost := naValue
		metric := ri.InstanceType
		age := fmt.Sprintf("%d", ri.DaysUntilExpiry)
		details := fmt.Sprintf("State: %s", ri.State)
		result = append(result, []string{category, identifier, estimatedCost, metric, age, details})
	}

	return result
}

func mapLoadBalancers(loadBalancers []elbtypes.LoadBalancer) [][]string {
	result := make([][]string, 0, len(loadBalancers))

	for _, lb := range loadBalancers {
		category := "Elastic Load Balancer"
		identifier := aws.ToString(lb.LoadBalancerArn)
		estimatedCost := fmt.Sprintf("$%.2f", pricing.CalculateLoadBalancerMonthlyCost(lb.Type))
		metric := string(lb.Type)
		age := naValue
		details := fmt.Sprintf("Created on %s", lb.CreatedTime.Format(time.RFC3339))
		result = append(result, []string{category, identifier, estimatedCost, metric, age, details})
	}

	return result
}

func mapAMIs(unusedAMIs []model.AMIWasteInfo) [][]string {
	result := make([][]string, 0, len(unusedAMIs))

	for _, ami := range unusedAMIs {
		category := "Unused AMI"
		identifier := ami.ImageID
		estimatedCost := naValue
		metric := naValue
		age := fmt.Sprintf("%d", ami.DaysSinceCreate)
		details := fmt.Sprintf("Created on %s", ami.CreationDate.Format(time.RFC3339))
		result = append(result, []string{category, identifier, estimatedCost, metric, age, details})
	}

	return result
}

func mapSnapshots(snapshots []model.SnapshotWasteInfo) ([][]string, [][]string) {
	orphaned := make([][]string, 0, len(snapshots))
	stale := make([][]string, 0, len(snapshots))

	for _, snap := range snapshots {
		category := fmt.Sprintf("EBS Snapshot (%s)", snap.Category)
		identifier := snap.SnapshotID
		estimatedCost := fmt.Sprintf("$%.2f", snap.MaxPotentialSavings)
		metric := fmt.Sprintf("%d GiB", snap.SizeGB)
		age := fmt.Sprintf("%d", snap.DaysSinceCreate)
		details := fmt.Sprintf("Created on %s", snap.StartTime.Format(time.RFC3339))

		row := []string{category, identifier, estimatedCost, metric, age, details}

		if snap.Category == model.SnapshotCategoryOrphaned {
			orphaned = append(orphaned, row)
		} else {
			stale = append(stale, row)
		}
	}

	return orphaned, stale
}

func mapKeyPairs(unusedKeyPairs []model.KeyPairWasteInfo) [][]string {
	result := make([][]string, 0, len(unusedKeyPairs))

	for _, kp := range unusedKeyPairs {
		category := "Unused Key Pair"
		identifier := kp.KeyName
		estimatedCost := naValue
		metric := naValue
		age := fmt.Sprintf("%d", kp.DaysSinceCreate)
		details := fmt.Sprintf("Created on %s", kp.CreateTime.Format(time.RFC3339))
		result = append(result, []string{category, identifier, estimatedCost, metric, age, details})
	}

	return result
}

func mapCloudWatchLogGroups(logGroups []model.CloudWatchLogsWasteInfo) [][]string {
	result := make([][]string, 0, len(logGroups))

	for _, lg := range logGroups {
		sizeGB := float64(lg.StoredBytes) / (1024 * 1024 * 1024)
		category := "CloudWatch Log Group"
		identifier := lg.LogGroupName
		estimatedCost := fmt.Sprintf("$%.2f", lg.EstimatedMonthlyCost)
		metric := fmt.Sprintf("%.2f GB stored", sizeGB)
		age := naValue
		details := fmt.Sprintf("Creation time: %s", lg.CreationTime.Format(time.RFC3339))
		result = append(result, []string{category, identifier, estimatedCost, metric, age, details})
	}

	return result
}

func mapRDSInstances(instances []model.RDSInstanceWasteInfo) [][]string {
	result := make([][]string, 0, len(instances))

	for _, inst := range instances {
		category := "Stopped RDS Instance"
		identifier := inst.DBInstanceID
		estimatedCost := fmt.Sprintf("$%.2f", inst.EstimatedMonthlyCost)
		metric := fmt.Sprintf("Is Multi AZ: %t", inst.MultiAZ)
		age := naValue
		details := fmt.Sprintf("Engine: %s / Status: %s", inst.Engine, inst.Status)
		result = append(result, []string{category, identifier, estimatedCost, metric, age, details})
	}

	return result
}

func mapRDSSnapshots(snapshots []model.RDSSnapshotWasteInfo) [][]string {
	result := make([][]string, 0, len(snapshots))

	for _, snap := range snapshots {
		category := "Old Manual RDS Snapshot"
		identifier := snap.DBSnapshotID
		estimatedCost := fmt.Sprintf("$%.2f", snap.EstimatedMonthlyCost)
		metric := fmt.Sprintf("%d GiB", snap.AllocatedStorage)
		age := fmt.Sprintf("%d", snap.DaysSinceCreate)
		details := fmt.Sprintf("Created on %s / Engine: %s", snap.SnapshotCreateTime.Format(time.RFC3339), snap.Engine)
		result = append(result, []string{category, identifier, estimatedCost, metric, age, details})
	}

	return result
}

func mapRDSIdleInstances(instances []model.RDSIdleInstanceInfo) [][]string {
	result := make([][]string, 0, len(instances))

	for _, inst := range instances {
		category := "Idle RDS Instance"
		identifier := inst.DBInstanceID
		estimatedCost := fmt.Sprintf("$%.2f", inst.EstimatedMonthlyCost)
		metric := fmt.Sprintf("No active connections in last %d days", inst.DaysChecked)
		age := naValue
		details := fmt.Sprintf("Engine: %s", inst.Engine)
		result = append(result, []string{category, identifier, estimatedCost, metric, age, details})
	}

	return result
}
