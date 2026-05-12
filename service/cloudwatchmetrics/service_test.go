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

	tests := []struct {
		name       string
		setupMocks func(*awsinterfaces.MockCloudWatchClient)
		wantIdle   bool
		wantErr    bool
	}{
		{
			name: "zero connections returns true",
			setupMocks: func(cw *awsinterfaces.MockCloudWatchClient) {
				cw.On("GetMetricStatistics", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{
					Datapoints: []cwtypes.Datapoint{
						{Sum: aws.Float64(0)},
						{Sum: aws.Float64(0)},
					},
				}, nil)
			},
			wantIdle: true,
			wantErr:  false,
		},
		{
			name: "active connections returns false",
			setupMocks: func(cw *awsinterfaces.MockCloudWatchClient) {
				cw.On("GetMetricStatistics", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{
					Datapoints: []cwtypes.Datapoint{
						{Sum: aws.Float64(0)},
						{Sum: aws.Float64(150)},
					},
				}, nil)
			},
			wantIdle: false,
			wantErr:  false,
		},
		{
			name: "empty datapoints returns true",
			setupMocks: func(cw *awsinterfaces.MockCloudWatchClient) {
				cw.On("GetMetricStatistics", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{
					Datapoints: []cwtypes.Datapoint{},
				}, nil)
			},
			wantIdle: true,
			wantErr:  false,
		},
		{
			name: "cloudwatch error returns error",
			setupMocks: func(cw *awsinterfaces.MockCloudWatchClient) {
				cw.On("GetMetricStatistics", mock.Anything, mock.Anything, mock.Anything).Return((*cloudwatch.GetMetricStatisticsOutput)(nil), errors.New("cloudwatch error"))
			},
			wantIdle: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCW := new(awsinterfaces.MockCloudWatchClient)
			tt.setupMocks(mockCW)

			svc := &service{client: mockCW}
			idle, err := svc.RDSHasZeroConnectionsInPeriod(ctx, "test-db", 7)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantIdle, idle)
			}
		})
	}
}

func TestNATGatewayBytesOut(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		setupMocks func(*awsinterfaces.MockCloudWatchClient)
		wantBytes  float64
		wantErr    bool
	}{
		{
			name: "returns sum of all datapoints",
			setupMocks: func(cw *awsinterfaces.MockCloudWatchClient) {
				cw.On("GetMetricStatistics", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{
					Datapoints: []cwtypes.Datapoint{
						{Sum: aws.Float64(100)},
						{Sum: aws.Float64(200)},
						{Sum: aws.Float64(50)},
					},
				}, nil)
			},
			wantBytes: 350,
			wantErr:   false,
		},
		{
			name: "zero bytes when idle",
			setupMocks: func(cw *awsinterfaces.MockCloudWatchClient) {
				cw.On("GetMetricStatistics", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{
					Datapoints: []cwtypes.Datapoint{
						{Sum: aws.Float64(0)},
						{Sum: aws.Float64(0)},
					},
				}, nil)
			},
			wantBytes: 0,
			wantErr:   false,
		},
		{
			name: "empty datapoints returns zero",
			setupMocks: func(cw *awsinterfaces.MockCloudWatchClient) {
				cw.On("GetMetricStatistics", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{
					Datapoints: []cwtypes.Datapoint{},
				}, nil)
			},
			wantBytes: 0,
			wantErr:   false,
		},
		{
			name: "cloudwatch error returns error",
			setupMocks: func(cw *awsinterfaces.MockCloudWatchClient) {
				cw.On("GetMetricStatistics", mock.Anything, mock.Anything, mock.Anything).Return(
					(*cloudwatch.GetMetricStatisticsOutput)(nil), errors.New("cloudwatch error"))
			},
			wantBytes: 0,
			wantErr:   true,
		},
		{
			name: "nil sum values are skipped",
			setupMocks: func(cw *awsinterfaces.MockCloudWatchClient) {
				cw.On("GetMetricStatistics", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{
					Datapoints: []cwtypes.Datapoint{
						{Sum: aws.Float64(100)},
						{Sum: nil},
						{Sum: aws.Float64(50)},
					},
				}, nil)
			},
			wantBytes: 150,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCW := new(awsinterfaces.MockCloudWatchClient)
			tt.setupMocks(mockCW)

			svc := &service{client: mockCW}
			bytes, err := svc.NATGatewayBytesOut(ctx, "nat-123", 7)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantBytes, bytes)
			}
		})
	}
}

func TestELBHasZeroRequestsInPeriod(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		arn        string
		lbType     elbtypes.LoadBalancerTypeEnum
		setupMocks func(*awsinterfaces.MockCloudWatchClient)
		wantIdle   bool
		wantErr    bool
	}{
		{
			name:   "ALB with zero requests returns true",
			arn:    "arn:aws:elasticloadbalancing:us-east-1:123456789:loadbalancer/app/my-alb/abc123",
			lbType: elbtypes.LoadBalancerTypeEnumApplication,
			setupMocks: func(cw *awsinterfaces.MockCloudWatchClient) {
				cw.On("GetMetricStatistics", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{
					Datapoints: []cwtypes.Datapoint{
						{Sum: aws.Float64(0)},
						{Sum: aws.Float64(0)},
					},
				}, nil)
			},
			wantIdle: true,
			wantErr:  false,
		},
		{
			name:   "NLB with active flows returns false",
			arn:    "arn:aws:elasticloadbalancing:us-east-1:123456789:loadbalancer/net/my-nlb/abc123",
			lbType: elbtypes.LoadBalancerTypeEnumNetwork,
			setupMocks: func(cw *awsinterfaces.MockCloudWatchClient) {
				cw.On("GetMetricStatistics", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{
					Datapoints: []cwtypes.Datapoint{
						{Sum: aws.Float64(0)},
						{Sum: aws.Float64(42)},
					},
				}, nil)
			},
			wantIdle: false,
			wantErr:  false,
		},
		{
			name:   "empty datapoints returns true",
			arn:    "arn:aws:elasticloadbalancing:us-east-1:123456789:loadbalancer/app/my-alb/abc123",
			lbType: elbtypes.LoadBalancerTypeEnumApplication,
			setupMocks: func(cw *awsinterfaces.MockCloudWatchClient) {
				cw.On("GetMetricStatistics", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{
					Datapoints: []cwtypes.Datapoint{},
				}, nil)
			},
			wantIdle: true,
			wantErr:  false,
		},
		{
			name:   "cloudwatch error returns error",
			arn:    "arn:aws:elasticloadbalancing:us-east-1:123456789:loadbalancer/app/my-alb/abc123",
			lbType: elbtypes.LoadBalancerTypeEnumApplication,
			setupMocks: func(cw *awsinterfaces.MockCloudWatchClient) {
				cw.On("GetMetricStatistics", mock.Anything, mock.Anything, mock.Anything).Return((*cloudwatch.GetMetricStatisticsOutput)(nil), errors.New("cloudwatch error"))
			},
			wantIdle: false,
			wantErr:  true,
		},
		{
			name:       "invalid ARN returns error",
			arn:        "invalid-arn",
			lbType:     "application",
			setupMocks: func(cw *awsinterfaces.MockCloudWatchClient) {},
			wantIdle:   false,
			wantErr:    true,
		},
		{
			name:       "unsupported LB type returns error",
			arn:        "arn:aws:elasticloadbalancing:us-east-1:123456789:loadbalancer/gw/my-gwlb/abc123",
			lbType:     elbtypes.LoadBalancerTypeEnumGateway,
			setupMocks: func(cw *awsinterfaces.MockCloudWatchClient) {},
			wantIdle:   false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCW := new(awsinterfaces.MockCloudWatchClient)
			tt.setupMocks(mockCW)

			svc := &service{client: mockCW}
			idle, err := svc.ELBHasZeroRequestsInPeriod(ctx, tt.arn, tt.lbType, 7)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantIdle, idle)
			}
		})
	}
}

func TestExtractLoadBalancerID(t *testing.T) {
	tests := []struct {
		name    string
		arn     string
		want    string
		wantErr bool
	}{
		{
			name: "valid ALB ARN",
			arn:  "arn:aws:elasticloadbalancing:us-east-1:123456789:loadbalancer/app/my-alb/abc123",
			want: "app/my-alb/abc123",
		},
		{
			name: "valid NLB ARN",
			arn:  "arn:aws:elasticloadbalancing:us-east-1:123456789:loadbalancer/net/my-nlb/def456",
			want: "net/my-nlb/def456",
		},
		{
			name:    "invalid ARN",
			arn:     "invalid-arn",
			wantErr: true,
		},
		{
			name:    "empty string",
			arn:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func TestSageMakerVariantInvocations(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		setupMocks func(*awsinterfaces.MockCloudWatchClient)
		want       float64
		wantErr    bool
	}{
		{
			name: "sums Sum datapoints",
			setupMocks: func(cw *awsinterfaces.MockCloudWatchClient) {
				cw.On("GetMetricStatistics", mock.Anything, mock.MatchedBy(func(in *cloudwatch.GetMetricStatisticsInput) bool {
					if aws.ToString(in.Namespace) != "AWS/SageMaker" || aws.ToString(in.MetricName) != "Invocations" {
						return false
					}

					names := make(map[string]string, len(in.Dimensions))
					for _, d := range in.Dimensions {
						names[aws.ToString(d.Name)] = aws.ToString(d.Value)
					}

					return names["EndpointName"] == "ep-1" && names["VariantName"] == "AllTraffic"
				}), mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{
					Datapoints: []cwtypes.Datapoint{
						{Sum: aws.Float64(10)},
						{Sum: aws.Float64(5)},
					},
				}, nil)
			},
			want:    15,
			wantErr: false,
		},
		{
			name: "zero datapoints returns zero",
			setupMocks: func(cw *awsinterfaces.MockCloudWatchClient) {
				cw.On("GetMetricStatistics", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{
					Datapoints: []cwtypes.Datapoint{},
				}, nil)
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "cloudwatch error propagates",
			setupMocks: func(cw *awsinterfaces.MockCloudWatchClient) {
				cw.On("GetMetricStatistics", mock.Anything, mock.Anything, mock.Anything).Return((*cloudwatch.GetMetricStatisticsOutput)(nil), errors.New("access denied"))
			},
			want:    0,
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

func TestEC2InstanceIdleStats(t *testing.T) {
	ctx := context.Background()

	t.Run("returns averaged cpu and per-day network bytes", func(t *testing.T) {
		cw := new(awsinterfaces.MockCloudWatchClient)
		svc := &service{client: cw}

		// CPUUtilization call (Average statistic).
		cw.On("GetMetricStatistics", mock.Anything, mock.MatchedBy(func(in *cloudwatch.GetMetricStatisticsInput) bool {
			return aws.ToString(in.MetricName) == "CPUUtilization"
		}), mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{
			Datapoints: []cwtypes.Datapoint{
				{Average: aws.Float64(2.0)},
				{Average: aws.Float64(4.0)},
			},
		}, nil).Once()

		// NetworkIn (Sum) returns 5 MB across 2 days, NetworkOut returns 3 MB.
		cw.On("GetMetricStatistics", mock.Anything, mock.MatchedBy(func(in *cloudwatch.GetMetricStatisticsInput) bool {
			return aws.ToString(in.MetricName) == "NetworkIn"
		}), mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{
			Datapoints: []cwtypes.Datapoint{
				{Sum: aws.Float64(3 * 1024 * 1024)},
				{Sum: aws.Float64(2 * 1024 * 1024)},
			},
		}, nil).Once()

		cw.On("GetMetricStatistics", mock.Anything, mock.MatchedBy(func(in *cloudwatch.GetMetricStatisticsInput) bool {
			return aws.ToString(in.MetricName) == "NetworkOut"
		}), mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{
			Datapoints: []cwtypes.Datapoint{
				{Sum: aws.Float64(2 * 1024 * 1024)},
				{Sum: aws.Float64(1 * 1024 * 1024)},
			},
		}, nil).Once()

		cpu, network, err := svc.EC2InstanceIdleStats(ctx, "i-123", 2)
		assert.NoError(t, err)
		assert.InDelta(t, 3.0, cpu, 0.001)
		// Total bytes = 8 MB, divided by 2 days = 4 MB/day.
		assert.InDelta(t, 4*1024*1024, network, 0.001)
	})

	t.Run("cpu error is surfaced", func(t *testing.T) {
		cw := new(awsinterfaces.MockCloudWatchClient)
		svc := &service{client: cw}

		cw.On("GetMetricStatistics", mock.Anything, mock.MatchedBy(func(in *cloudwatch.GetMetricStatisticsInput) bool {
			return aws.ToString(in.MetricName) == "CPUUtilization"
		}), mock.Anything).Return((*cloudwatch.GetMetricStatisticsOutput)(nil), errors.New("boom")).Once()

		_, _, err := svc.EC2InstanceIdleStats(ctx, "i-123", 7)
		assert.Error(t, err)
	})

	t.Run("network error is surfaced", func(t *testing.T) {
		cw := new(awsinterfaces.MockCloudWatchClient)
		svc := &service{client: cw}

		cw.On("GetMetricStatistics", mock.Anything, mock.MatchedBy(func(in *cloudwatch.GetMetricStatisticsInput) bool {
			return aws.ToString(in.MetricName) == "CPUUtilization"
		}), mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{}, nil).Once()

		cw.On("GetMetricStatistics", mock.Anything, mock.MatchedBy(func(in *cloudwatch.GetMetricStatisticsInput) bool {
			return aws.ToString(in.MetricName) == "NetworkIn"
		}), mock.Anything).Return((*cloudwatch.GetMetricStatisticsOutput)(nil), errors.New("net err")).Once()

		_, _, err := svc.EC2InstanceIdleStats(ctx, "i-123", 7)
		assert.Error(t, err)
	})

	t.Run("zero days does not divide", func(t *testing.T) {
		cw := new(awsinterfaces.MockCloudWatchClient)
		svc := &service{client: cw}

		cw.On("GetMetricStatistics", mock.Anything, mock.MatchedBy(func(in *cloudwatch.GetMetricStatisticsInput) bool {
			return aws.ToString(in.MetricName) == "CPUUtilization"
		}), mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{}, nil).Once()

		cw.On("GetMetricStatistics", mock.Anything, mock.MatchedBy(func(in *cloudwatch.GetMetricStatisticsInput) bool {
			return aws.ToString(in.MetricName) == "NetworkIn"
		}), mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{
			Datapoints: []cwtypes.Datapoint{{Sum: aws.Float64(100)}},
		}, nil).Once()

		cw.On("GetMetricStatistics", mock.Anything, mock.MatchedBy(func(in *cloudwatch.GetMetricStatisticsInput) bool {
			return aws.ToString(in.MetricName) == "NetworkOut"
		}), mock.Anything).Return(&cloudwatch.GetMetricStatisticsOutput{
			Datapoints: []cwtypes.Datapoint{{Sum: aws.Float64(50)}},
		}, nil).Once()

		cpu, network, err := svc.EC2InstanceIdleStats(ctx, "i-123", 0)
		assert.NoError(t, err)
		assert.Equal(t, 0.0, cpu)
		assert.Equal(t, 150.0, network)
	})
}

func TestNewService(t *testing.T) {
	cfg := aws.Config{}
	svc := NewService(cfg)
	assert.NotNil(t, svc)
}
