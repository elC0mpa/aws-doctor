package csvoutput

import (
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/pricing"
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

func mapElasticIPs(elasticIPs []ec2types.Address, pricingSvc pricing.Service) [][]string {
	result := make([][]string, 0, len(elasticIPs))
	for _, ip := range elasticIPs {
		result = append(result, outputshared.PresentElasticIP(ip, pricingSvc).ToSlice())
	}

	return result
}

func mapEBSVolumes(volumes []ec2types.Volume, status string, pricingSvc pricing.Service) [][]string {
	result := make([][]string, 0, len(volumes))
	for _, vol := range volumes {
		result = append(result, outputshared.PresentEBSVolume(vol, status, pricingSvc).ToSlice())
	}

	return result
}

func mapStoppedInstances(stoppedInstances []ec2types.Instance) [][]string {
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

func mapLoadBalancers(loadBalancers []elbtypes.LoadBalancer, pricingSvc pricing.Service) [][]string {
	result := make([][]string, 0, len(loadBalancers))
	for _, lb := range loadBalancers {
		result = append(result, outputshared.PresentLoadBalancer(lb, pricingSvc).ToSlice())
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

func mapIdleLoadBalancers(idleLBs []model.ELBIdleInfo) [][]string {
	result := make([][]string, 0, len(idleLBs))
	for _, lb := range idleLBs {
		result = append(result, outputshared.PresentIdleLoadBalancer(lb).ToSlice())
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

func mapNATGateways(natGateways []model.NATGatewayWasteInfo) [][]string {
	result := make([][]string, 0, len(natGateways))
	for _, ng := range natGateways {
		result = append(result, outputshared.PresentIdleNATGateway(ng).ToSlice())
	}

	return result
}

func mapLambdaOverProvisioned(lambdas []model.LambdaOverProvisionedInfo) [][]string {
	result := make([][]string, 0, len(lambdas))
	for _, fn := range lambdas {
		result = append(result, outputshared.PresentLambdaOverProvisioned(fn).ToSlice())
	}

	return result
}

func mapIdleSageMakerEndpoints(endpoints []model.IdleSageMakerEndpointInfo) [][]string {
	result := make([][]string, 0, len(endpoints))
	for _, ep := range endpoints {
		result = append(result, outputshared.PresentIdleSageMakerEndpoint(ep).ToSlice())
	}

	return result
}

func mapECRNoLifecyclePolicies(repos []model.ECRNoLifecyclePolicyInfo) [][]string {
	result := make([][]string, 0, len(repos))
	for _, r := range repos {
		result = append(result, outputshared.PresentECRNoLifecyclePolicy(r).ToSlice())
	}

	return result
}

func mapECREmptyRepositories(repos []model.ECREmptyRepositoryInfo) [][]string {
	result := make([][]string, 0, len(repos))
	for _, r := range repos {
		result = append(result, outputshared.PresentECREmptyRepository(r).ToSlice())
	}

	return result
}

func mapECRUntaggedImages(repos []model.ECRUntaggedImageInfo) [][]string {
	result := make([][]string, 0, len(repos))
	for _, r := range repos {
		result = append(result, outputshared.PresentECRUntaggedImages(r).ToSlice())
	}

	return result
}

func mapUnusedSecrets(secrets []model.UnusedSecretInfo, pricingSvc pricing.Service) [][]string {
	result := make([][]string, 0, len(secrets))
	for _, s := range secrets {
		result = append(result, outputshared.PresentUnusedSecret(s, pricingSvc).ToSlice())
	}

	return result
}
