package orchestrator

import (
	"context"
	"time"

	"github.com/elC0mpa/aws-doctor/model"
	"golang.org/x/sync/errgroup"
)

func (s *service) queueEC2Checks(ctx context.Context, g *errgroup.Group, resultCh chan<- model.ScopeResult) {
	g.Go(func() error {
		start := time.Now()
		var input model.RenderWasteInput
		ec2Group, ec2Ctx := errgroup.WithContext(ctx)

		ec2Group.Go(func() error {
			var err error
			input.ElasticIPs, err = s.ec2Service.GetUnusedElasticIPAddressesInfo(ec2Ctx)
			return err
		})

		ec2Group.Go(func() error {
			var err error
			input.UnusedVolumes, err = s.ec2Service.GetUnusedEBSVolumes(ec2Ctx)
			return err
		})

		ec2Group.Go(func() error {
			var err error
			input.StoppedInstances, input.StoppedVolumes, err = s.ec2Service.GetStoppedInstancesInfo(ec2Ctx, ec2StoppedDays)
			return err
		})

		ec2Group.Go(func() error {
			var err error
			input.Ris, err = s.ec2Service.GetReservedInstanceExpiringOrExpiredWaste(ec2Ctx, ec2RiExpiringDays)
			return err
		})

		ec2Group.Go(func() error {
			var err error
			input.UnusedAMIs, err = s.ec2Service.GetUnusedAMIs(ec2Ctx, ec2AmiStaleDays)
			return err
		})

		ec2Group.Go(func() error {
			var err error
			input.OrphanedSnapshots, err = s.ec2Service.GetOrphanedSnapshots(ec2Ctx, ec2SnapshotStaleDays)
			return err
		})

		ec2Group.Go(func() error {
			var err error
			input.UnusedKeyPairs, err = s.ec2Service.GetUnusedKeyPairs(ec2Ctx)
			return err
		})

		ec2Group.Go(func() error {
			var err error
			input.IdleEC2Instances, err = s.ec2Service.GetIdleInstances(ec2Ctx, ec2IdleDays, ec2IdleCPUPercent, ec2IdleNetworkBytesPerDay)
			return err
		})

		err := ec2Group.Wait()
		resultCh <- model.ScopeResult{Scope: "EC2", Input: input, Duration: time.Since(start), Err: err}
		return err
	})
}

func (s *service) queueVPCChecks(ctx context.Context, g *errgroup.Group, resultCh chan<- model.ScopeResult) {
	g.Go(func() error {
		start := time.Now()
		var input model.RenderWasteInput
		var err error
		input.IdleNATGateways, err = s.vpcService.GetIdleNATGateways(ctx, vpcNatIdleDays)
		resultCh <- model.ScopeResult{Scope: "VPC", Input: input, Duration: time.Since(start), Err: err}
		return err
	})
}

func (s *service) queueELBChecks(ctx context.Context, g *errgroup.Group, resultCh chan<- model.ScopeResult) {
	g.Go(func() error {
		start := time.Now()
		var input model.RenderWasteInput
		var err error
		input.LoadBalancers, input.IdleLoadBalancers, err = s.elbService.GetLoadBalancerWaste(ctx, elbIdleDays)
		resultCh <- model.ScopeResult{Scope: "ELB", Input: input, Duration: time.Since(start), Err: err}
		return err
	})
}

func (s *service) queueS3Checks(ctx context.Context, g *errgroup.Group, resultCh chan<- model.ScopeResult) {
	g.Go(func() error {
		start := time.Now()
		var input model.RenderWasteInput
		var err error
		input.S3Buckets, input.S3MultipartUploads, err = s.s3Service.GetS3Waste(ctx)
		resultCh <- model.ScopeResult{Scope: "S3", Input: input, Duration: time.Since(start), Err: err}
		return err
	})
}

func (s *service) queueCloudWatchLogsChecks(ctx context.Context, g *errgroup.Group, resultCh chan<- model.ScopeResult) {
	g.Go(func() error {
		start := time.Now()
		var input model.RenderWasteInput
		var err error
		input.CloudWatchLogGroups, err = s.cloudwatchlogsService.GetCloudWatchLogsWaste(ctx)
		resultCh <- model.ScopeResult{Scope: "CloudWatch", Input: input, Duration: time.Since(start), Err: err}
		return err
	})
}

func (s *service) queueRDSChecks(ctx context.Context, g *errgroup.Group, resultCh chan<- model.ScopeResult) {
	g.Go(func() error {
		start := time.Now()
		var input model.RenderWasteInput
		var err error
		input.RDSInstances, input.RDSSnapshots, input.RDSIdleInstances, err = s.rdsService.GetRDSWaste(ctx, rdsIdleDays, rdsSnapshotDays)
		resultCh <- model.ScopeResult{Scope: "RDS", Input: input, Duration: time.Since(start), Err: err}
		return err
	})
}

func (s *service) queueLambdaChecks(ctx context.Context, g *errgroup.Group, resultCh chan<- model.ScopeResult, lambdaMemoryThreshold int) {
	g.Go(func() error {
		start := time.Now()
		var input model.RenderWasteInput
		var err error
		input.OverProvisionedLambdas, err = s.lambdaService.GetOverProvisionedFunctions(ctx, lambdaMemoryThreshold, lambdaLookbackDays)
		resultCh <- model.ScopeResult{Scope: "Lambda", Input: input, Duration: time.Since(start), Err: err}
		return err
	})
}

func (s *service) queueSagemakerChecks(ctx context.Context, g *errgroup.Group, resultCh chan<- model.ScopeResult) {
	g.Go(func() error {
		start := time.Now()
		var input model.RenderWasteInput
		var err error
		input.IdleSageMakerEndpoints, err = s.sagemakerService.GetIdleEndpoints(ctx, sagemakerIdleDays)
		resultCh <- model.ScopeResult{Scope: "SageMaker", Input: input, Duration: time.Since(start), Err: err}
		return err
	})
}

func (s *service) queueECRChecks(ctx context.Context, g *errgroup.Group, resultCh chan<- model.ScopeResult) {
	g.Go(func() error {
		start := time.Now()
		var input model.RenderWasteInput
		var err error
		input.ECRNoLifecyclePolicies, input.ECREmptyRepositories, input.ECRUntaggedImages, err = s.ecrService.GetECRWaste(ctx)
		resultCh <- model.ScopeResult{Scope: "ECR", Input: input, Duration: time.Since(start), Err: err}
		return err
	})
}

func (s *service) queueSecretsManagerChecks(ctx context.Context, g *errgroup.Group, resultCh chan<- model.ScopeResult, secretsIdleDays int) {
	g.Go(func() error {
		start := time.Now()
		var input model.RenderWasteInput
		var err error
		input.UnusedSecrets, err = s.secretsmanagerService.GetUnusedSecrets(ctx, secretsIdleDays)
		resultCh <- model.ScopeResult{Scope: "SecretsManager", Input: input, Duration: time.Since(start), Err: err}
		return err
	})
}
