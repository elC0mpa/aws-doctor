package csvoutput

import (
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
	outputshared "github.com/elC0mpa/aws-doctor/utils/output_shared"
)

func mapS3Buckets(buckets []model.S3BucketWasteInfo) [][]string {
	result := make([][]string, 0, len(buckets))
	for _, b := range buckets {
		result = append(result, outputshared.PresentS3Bucket(b).ToSlice())
	}

	return result
}

func mapS3MultipartUploads(buckets []model.S3MultipartUploadWasteInfo) [][]string {
	result := make([][]string, 0, len(buckets))
	for _, b := range buckets {
		result = append(result, outputshared.PresentS3MultipartUpload(b).ToSlice())
	}

	return result
}

func mapElasticIPs(elasticIPs []types.Address) [][]string {
	result := make([][]string, 0, len(elasticIPs))
	for _, ip := range elasticIPs {
		result = append(result, outputshared.PresentElasticIP(ip).ToSlice())
	}

	return result
}

func mapEBSVolumes(volumes []types.Volume, status string) [][]string {
	result := make([][]string, 0, len(volumes))
	for _, vol := range volumes {
		result = append(result, outputshared.PresentEBSVolume(vol, status).ToSlice())
	}

	return result
}

func mapStoppedInstances(stoppedInstances []types.Instance) [][]string {
	result := make([][]string, 0, len(stoppedInstances))
	for _, instance := range stoppedInstances {
		result = append(result, outputshared.PresentStoppedInstance(instance).ToSlice())
	}

	return result
}

func mapReservedInstances(ris []model.RiExpirationInfo) [][]string {
	result := make([][]string, 0, len(ris))
	for _, ri := range ris {
		result = append(result, outputshared.PresentReservedInstance(ri).ToSlice())
	}

	return result
}

func mapLoadBalancers(loadBalancers []elbtypes.LoadBalancer) [][]string {
	result := make([][]string, 0, len(loadBalancers))
	for _, lb := range loadBalancers {
		result = append(result, outputshared.PresentLoadBalancer(lb).ToSlice())
	}

	return result
}

func mapAMIs(unusedAMIs []model.AMIWasteInfo) [][]string {
	result := make([][]string, 0, len(unusedAMIs))
	for _, ami := range unusedAMIs {
		result = append(result, outputshared.PresentAMI(ami).ToSlice())
	}

	return result
}

func mapSnapshots(snapshots []model.SnapshotWasteInfo) ([][]string, [][]string) {
	orphaned := make([][]string, 0, len(snapshots))

	stale := make([][]string, 0, len(snapshots))
	for _, snap := range snapshots {
		row := outputshared.PresentSnapshot(snap).ToSlice()
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
		result = append(result, outputshared.PresentKeyPair(kp).ToSlice())
	}

	return result
}

func mapCloudWatchLogGroups(logGroups []model.CloudWatchLogsWasteInfo) [][]string {
	result := make([][]string, 0, len(logGroups))
	for _, lg := range logGroups {
		result = append(result, outputshared.PresentCloudWatchLogGroup(lg).ToSlice())
	}

	return result
}

func mapRDSInstances(instances []model.RDSInstanceWasteInfo) [][]string {
	result := make([][]string, 0, len(instances))
	for _, inst := range instances {
		result = append(result, outputshared.PresentRDSInstance(inst).ToSlice())
	}

	return result
}

func mapRDSSnapshots(snapshots []model.RDSSnapshotWasteInfo) [][]string {
	result := make([][]string, 0, len(snapshots))
	for _, snap := range snapshots {
		result = append(result, outputshared.PresentRDSSnapshot(snap).ToSlice())
	}

	return result
}

func mapRDSIdleInstances(instances []model.RDSIdleInstanceInfo) [][]string {
	result := make([][]string, 0, len(instances))
	for _, inst := range instances {
		result = append(result, outputshared.PresentRDSIdleInstance(inst).ToSlice())
	}

	return result
}

func mapNATGateways(gateways []model.NATGatewayWasteInfo) [][]string {
	result := make([][]string, 0, len(gateways))
	for _, gw := range gateways {
		result = append(result, outputshared.PresentNATGateway(gw).ToSlice())
	}

	return result
}
