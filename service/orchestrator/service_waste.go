package orchestrator

import (
	"context"

	"github.com/elC0mpa/aws-doctor/model"
	"golang.org/x/sync/errgroup"
)

func (s *service) queueEC2Checks(ctx context.Context, g *errgroup.Group, input *model.RenderWasteInput) {
	g.Go(func() error {
		var err error

		input.ElasticIPs, err = s.ec2Service.GetUnusedElasticIPAddressesInfo(ctx)

		return err
	})

	g.Go(func() error {
		var err error

		input.UnusedVolumes, err = s.ec2Service.GetUnusedEBSVolumes(ctx)

		return err
	})

	g.Go(func() error {
		var err error

		input.StoppedInstances, input.StoppedVolumes, err = s.ec2Service.GetStoppedInstancesInfo(ctx)

		return err
	})

	g.Go(func() error {
		var err error

		input.Ris, err = s.ec2Service.GetReservedInstanceExpiringOrExpired30DaysWaste(ctx)

		return err
	})

	g.Go(func() error {
		var err error

		input.UnusedAMIs, err = s.ec2Service.GetUnusedAMIs(ctx, 90)

		return err
	})

	g.Go(func() error {
		var err error

		input.OrphanedSnapshots, err = s.ec2Service.GetOrphanedSnapshots(ctx, 90)

		return err
	})

	g.Go(func() error {
		var err error

		input.UnusedKeyPairs, err = s.ec2Service.GetUnusedKeyPairs(ctx)

		return err
	})
}

func (s *service) queueVPCChecks(ctx context.Context, g *errgroup.Group, input *model.RenderWasteInput) {
	g.Go(func() error {
		var err error

		input.IdleNATGateways, err = s.vpcService.GetIdleNATGateways(ctx, 7)

		return err
	})
}

func (s *service) queueELBChecks(ctx context.Context, g *errgroup.Group, input *model.RenderWasteInput) {
	g.Go(func() error {
		var err error

		input.LoadBalancers, input.IdleLoadBalancers, err = s.elbService.GetLoadBalancerWaste(ctx)

		return err
	})
}

func (s *service) queueS3Checks(ctx context.Context, g *errgroup.Group, input *model.RenderWasteInput) {
	g.Go(func() error {
		var err error

		input.S3Buckets, input.S3MultipartUploads, err = s.s3Service.GetS3Waste(ctx)

		return err
	})
}

func (s *service) queueCloudWatchLogsChecks(ctx context.Context, g *errgroup.Group, input *model.RenderWasteInput) {
	g.Go(func() error {
		var err error

		input.CloudWatchLogGroups, err = s.cloudwatchlogsService.GetCloudWatchLogsWaste(ctx)

		return err
	})
}

func (s *service) queueRDSChecks(ctx context.Context, g *errgroup.Group, input *model.RenderWasteInput) {
	g.Go(func() error {
		var err error

		input.RDSInstances, input.RDSSnapshots, input.RDSIdleInstances, err = s.rdsService.GetRDSWaste(ctx)

		return err
	})
}

func (s *service) queueLambdaChecks(ctx context.Context, g *errgroup.Group, input *model.RenderWasteInput, lambdaMemoryThreshold int) {
	g.Go(func() error {
		var err error

		input.OverProvisionedLambdas, err = s.lambdaService.GetOverProvisionedFunctions(ctx, lambdaMemoryThreshold)

		return err
	})
}

func (s *service) queueSagemakerChecks(ctx context.Context, g *errgroup.Group, input *model.RenderWasteInput) {
	g.Go(func() error {
		var err error

		input.IdleSageMakerEndpoints, err = s.sagemakerService.GetIdleEndpoints(ctx, sagemakerIdleDays)

		return err
	})
}
