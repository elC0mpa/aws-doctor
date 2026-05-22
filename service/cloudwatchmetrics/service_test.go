// Package cloudwatchmetrics provides a service for querying CloudWatch metrics.
package cloudwatchmetrics

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/mocks/awsinterfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRDSHasZeroConnectionsInPeriod(t *testing.T) {
	ctx := context.Background()

	t.Run("returns true when all datapoints are zero", func(t *testing.T) {
		cw := new(awsinterfaces.MockCloudWatchClient)
		svc := &service{client: cw}

		cw.On("GetMetricStatistics", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{
			Datapoints: []cwtypes.Datapoint{
				{Sum: aws.Float64(0)},
				{Sum: aws.Float64(0)},
			},
		}, nil).Once()

		got, err := svc.RDSHasZeroConnectionsInPeriod(ctx, "db-123", 7)
		assert.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("returns false when any datapoint is positive", func(t *testing.T) {
		cw := new(awsinterfaces.MockCloudWatchClient)
		svc := &service{client: cw}

		cw.On("GetMetricStatistics", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{
			Datapoints: []cwtypes.Datapoint{
				{Sum: aws.Float64(0)},
				{Sum: aws.Float64(1)},
			},
		}, nil).Once()

		got, err := svc.RDSHasZeroConnectionsInPeriod(ctx, "db-123", 7)
		assert.NoError(t, err)
		assert.False(t, got)
	})

	t.Run("error is surfaced", func(t *testing.T) {
		cw := new(awsinterfaces.MockCloudWatchClient)
		svc := &service{client: cw}

		cw.On("GetMetricStatistics", mock.Anything, mock.Anything, mock.Anything).Return((*cloudwatch.GetMetricStatisticsOutput)(nil), errors.New("boom")).Once()

		_, err := svc.RDSHasZeroConnectionsInPeriod(ctx, "db-123", 7)
		assert.Error(t, err)
	})
}

func TestNATGatewayBytesOut(t *testing.T) {
	ctx := context.Background()

	t.Run("returns total bytes across datapoints", func(t *testing.T) {
		cw := new(awsinterfaces.MockCloudWatchClient)
		svc := &service{client: cw}

		cw.On("GetMetricStatistics", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{
			Datapoints: []cwtypes.Datapoint{
				{Sum: aws.Float64(100.5)},
				{Sum: aws.Float64(200.5)},
			},
		}, nil).Once()

		got, err := svc.NATGatewayBytesOut(ctx, "nat-123", 7)
		assert.NoError(t, err)
		assert.Equal(t, 301.0, got)
	})

	t.Run("error is surfaced", func(t *testing.T) {
		cw := new(awsinterfaces.MockCloudWatchClient)
		svc := &service{client: cw}

		cw.On("GetMetricStatistics", mock.Anything, mock.Anything, mock.Anything).Return((*cloudwatch.GetMetricStatisticsOutput)(nil), errors.New("boom")).Once()

		_, err := svc.NATGatewayBytesOut(ctx, "nat-123", 7)
		assert.Error(t, err)
	})
}

func TestExtractLoadBalancerID(t *testing.T) {
	tests := []struct {
		arn     string
		want    string
		wantErr bool
	}{
		{
			arn:  "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-load-balancer/50dc6c495c0c9188",
			want: "app/my-load-balancer/50dc6c495c0c9188",
		},
		{
			arn:  "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/net/my-network-lb/c4f2e519c288f6a9",
			want: "net/my-network-lb/c4f2e519c288f6a9",
		},
		{
			arn:     "invalid-arn",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.arn, func(t *testing.T) {
			got, err := ExtractLoadBalancerID(tt.arn)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestELBHasZeroRequestsInPeriod(t *testing.T) {
	ctx := context.Background()

	t.Run("ALB with requests returns false", func(t *testing.T) {
		cw := new(awsinterfaces.MockCloudWatchClient)
		svc := &service{client: cw}

		cw.On("GetMetricStatistics", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{
			Datapoints: []cwtypes.Datapoint{{Sum: aws.Float64(1)}},
		}, nil).Once()

		got, err := svc.ELBHasZeroRequestsInPeriod(ctx, "arn:aws:elb:us-east-1:123:loadbalancer/app/foo/bar", elbtypes.LoadBalancerTypeEnumApplication, 7)
		assert.NoError(t, err)
		assert.False(t, got)
	})

	t.Run("NLB with zero flows returns true", func(t *testing.T) {
		cw := new(awsinterfaces.MockCloudWatchClient)
		svc := &service{client: cw}

		cw.On("GetMetricStatistics", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{
			Datapoints: []cwtypes.Datapoint{{Sum: aws.Float64(0)}},
		}, nil).Once()

		got, err := svc.ELBHasZeroRequestsInPeriod(ctx, "arn:aws:elb:us-east-1:123:loadbalancer/net/foo/bar", elbtypes.LoadBalancerTypeEnumNetwork, 7)
		assert.NoError(t, err)
		assert.True(t, got)
	})

	t.Run("unsupported type returns error", func(t *testing.T) {
		svc := &service{}
		_, err := svc.ELBHasZeroRequestsInPeriod(ctx, "arn:aws:elb:us-east-1:123:loadbalancer/gateway/foo/bar", elbtypes.LoadBalancerTypeEnumGateway, 7)
		assert.Error(t, err)
	})
}

func TestEC2InstanceIdleStats(t *testing.T) {
	ctx := context.Background()

	t.Run("returns averaged cpu and per-day network bytes", func(t *testing.T) {
		cw := new(awsinterfaces.MockCloudWatchClient)
		svc := &service{client: cw}

		cw.On("GetMetricData", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatch.GetMetricDataOutput{
			MetricDataResults: []cwtypes.MetricDataResult{
				{
					Id:     aws.String("cpu"),
					Values: []float64{2.0, 4.0},
				},
				{
					Id:     aws.String("net_in"),
					Values: []float64{3 * 1024 * 1024, 2 * 1024 * 1024},
				},
				{
					Id:     aws.String("net_out"),
					Values: []float64{2 * 1024 * 1024, 1 * 1024 * 1024},
				},
			},
		}, nil).Once()

		cpu, network, err := svc.EC2InstanceIdleStats(ctx, "i-123", 2)
		assert.NoError(t, err)
		assert.InDelta(t, 3.0, cpu, 0.001)
		// Total bytes = 8 MB, divided by 2 days = 4 MB/day.
		assert.InDelta(t, 4*1024*1024, network, 0.001)
	})

	t.Run("error is surfaced", func(t *testing.T) {
		cw := new(awsinterfaces.MockCloudWatchClient)
		svc := &service{client: cw}

		cw.On("GetMetricData", mock.Anything, mock.Anything, mock.Anything).Return((*cloudwatch.GetMetricDataOutput)(nil), errors.New("boom")).Once()

		_, _, err := svc.EC2InstanceIdleStats(ctx, "i-123", 7)
		assert.Error(t, err)
	})

	t.Run("zero days does not divide", func(t *testing.T) {
		cw := new(awsinterfaces.MockCloudWatchClient)
		svc := &service{client: cw}

		cw.On("GetMetricData", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatch.GetMetricDataOutput{
			MetricDataResults: []cwtypes.MetricDataResult{
				{
					Id:     aws.String("cpu"),
					Values: []float64{1.0},
				},
				{
					Id:     aws.String("net_in"),
					Values: []float64{100},
				},
			},
		}, nil).Once()

		cpu, network, err := svc.EC2InstanceIdleStats(ctx, "i-123", 0)
		assert.NoError(t, err)
		assert.Equal(t, 1.0, cpu)
		assert.Equal(t, 100.0, network)
	})
}

func TestSageMakerVariantInvocations(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		setupMocks func(*awsinterfaces.MockCloudWatchClient)
		want       float64
		wantErr    bool
	}{
		{
			name: "returns total sum of invocations",
			setupMocks: func(m *awsinterfaces.MockCloudWatchClient) {
				m.On("GetMetricStatistics", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{
					Datapoints: []cwtypes.Datapoint{
						{Sum: aws.Float64(10.0)},
						{Sum: aws.Float64(25.5)},
					},
				}, nil).Once()
			},
			want: 35.5,
		},
		{
			name: "error is surfaced",
			setupMocks: func(m *awsinterfaces.MockCloudWatchClient) {
				m.On("GetMetricStatistics", mock.Anything, mock.Anything, mock.Anything).Return((*cloudwatch.GetMetricStatisticsOutput)(nil), errors.New("api error")).Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCW := new(awsinterfaces.MockCloudWatchClient)
			tt.setupMocks(mockCW)

			svc := &service{client: mockCW}
			got, err := svc.SageMakerVariantInvocations(ctx, "ep-1", "AllTraffic", 14)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestNewService(t *testing.T) {
	s := NewService(aws.Config{})
	if s == nil {
		t.Error("NewService returned nil")
	}
}
