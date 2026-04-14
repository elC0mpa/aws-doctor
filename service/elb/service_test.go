package elb

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	elb "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/mocks/awsinterfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockCWMetricsService is a local mock for cloudwatchMetricsService used by GetLoadBalancerWaste.
type mockCWMetricsService struct {
	mock.Mock
}

func (m *mockCWMetricsService) ELBHasZeroRequestsInPeriod(ctx context.Context, loadBalancerArn string, lbType string, days int) (bool, error) {
	args := m.Called(ctx, loadBalancerArn, lbType, days)

	return args.Bool(0), args.Error(1)
}

func TestGetLoadBalancerWaste(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*awsinterfaces.MockELBClient, *mockCWMetricsService)
		wantUnused    int
		wantIdle      int
		wantErr       bool
		wantIdleNames []string
	}{
		{
			name: "separates unused and idle load balancers",
			setupMocks: func(mc *awsinterfaces.MockELBClient, cw *mockCWMetricsService) {
				mc.On("DescribeLoadBalancers", mock.Anything, mock.Anything, mock.Anything).Return(&elb.DescribeLoadBalancersOutput{
					LoadBalancers: []types.LoadBalancer{
						{
							LoadBalancerArn:  aws.String("arn:aws:elasticloadbalancing:us-east-1:123:loadbalancer/app/idle-alb/abc"),
							Type:             types.LoadBalancerTypeEnumApplication,
							LoadBalancerName: aws.String("idle-alb"),
						},
						{
							LoadBalancerArn:  aws.String("arn:aws:elasticloadbalancing:us-east-1:123:loadbalancer/net/active-nlb/def"),
							Type:             types.LoadBalancerTypeEnumNetwork,
							LoadBalancerName: aws.String("active-nlb"),
						},
						{
							LoadBalancerArn:  aws.String("arn:aws:elasticloadbalancing:us-east-1:123:loadbalancer/app/orphan-alb/ghi"),
							Type:             types.LoadBalancerTypeEnumApplication,
							LoadBalancerName: aws.String("orphan-alb"),
						},
						{
							LoadBalancerArn:  aws.String("arn:gwlb:unused"),
							Type:             types.LoadBalancerTypeEnumGateway,
							LoadBalancerName: aws.String("unused-gwlb"),
						},
					},
				}, nil)

				mc.On("DescribeTargetGroups", mock.Anything, mock.Anything, mock.Anything).Return(&elb.DescribeTargetGroupsOutput{
					TargetGroups: []types.TargetGroup{
						{LoadBalancerArns: []string{"arn:aws:elasticloadbalancing:us-east-1:123:loadbalancer/app/idle-alb/abc"}},
						{LoadBalancerArns: []string{"arn:aws:elasticloadbalancing:us-east-1:123:loadbalancer/net/active-nlb/def"}},
					},
				}, nil)

				// idle-alb: zero requests
				cw.On("ELBHasZeroRequestsInPeriod", mock.Anything,
					"arn:aws:elasticloadbalancing:us-east-1:123:loadbalancer/app/idle-alb/abc",
					"application", 7).Return(true, nil)
				// active-nlb: has traffic
				cw.On("ELBHasZeroRequestsInPeriod", mock.Anything,
					"arn:aws:elasticloadbalancing:us-east-1:123:loadbalancer/net/active-nlb/def",
					"network", 7).Return(false, nil)
			},
			wantUnused:    1, // orphan-alb
			wantIdle:      1, // idle-alb
			wantErr:       false,
			wantIdleNames: []string{"idle-alb"},
		},
		{
			name: "LB error returns error",
			setupMocks: func(mc *awsinterfaces.MockELBClient, cw *mockCWMetricsService) {
				mc.On("DescribeLoadBalancers", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("LB API error"))
			},
			wantUnused: 0,
			wantIdle:   0,
			wantErr:    true,
		},
		{
			name: "TG error returns error",
			setupMocks: func(mc *awsinterfaces.MockELBClient, cw *mockCWMetricsService) {
				mc.On("DescribeLoadBalancers", mock.Anything, mock.Anything, mock.Anything).Return(&elb.DescribeLoadBalancersOutput{}, nil)
				mc.On("DescribeTargetGroups", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("TG API error"))
			},
			wantUnused: 0,
			wantIdle:   0,
			wantErr:    true,
		},
		{
			name: "CloudWatch error returns error",
			setupMocks: func(mc *awsinterfaces.MockELBClient, cw *mockCWMetricsService) {
				mc.On("DescribeLoadBalancers", mock.Anything, mock.Anything, mock.Anything).Return(&elb.DescribeLoadBalancersOutput{
					LoadBalancers: []types.LoadBalancer{
						{
							LoadBalancerArn:  aws.String("arn:aws:elasticloadbalancing:us-east-1:123:loadbalancer/app/test-alb/abc"),
							Type:             types.LoadBalancerTypeEnumApplication,
							LoadBalancerName: aws.String("test-alb"),
						},
					},
				}, nil)

				mc.On("DescribeTargetGroups", mock.Anything, mock.Anything, mock.Anything).Return(&elb.DescribeTargetGroupsOutput{
					TargetGroups: []types.TargetGroup{
						{LoadBalancerArns: []string{"arn:aws:elasticloadbalancing:us-east-1:123:loadbalancer/app/test-alb/abc"}},
					},
				}, nil)

				cw.On("ELBHasZeroRequestsInPeriod", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(false, errors.New("CW error"))
			},
			wantUnused: 0,
			wantIdle:   0,
			wantErr:    true,
		},
		{
			name: "empty result with no load balancers",
			setupMocks: func(mc *awsinterfaces.MockELBClient, cw *mockCWMetricsService) {
				mc.On("DescribeLoadBalancers", mock.Anything, mock.Anything, mock.Anything).Return(&elb.DescribeLoadBalancersOutput{}, nil)
				mc.On("DescribeTargetGroups", mock.Anything, mock.Anything, mock.Anything).Return(&elb.DescribeTargetGroupsOutput{}, nil)
			},
			wantUnused: 0,
			wantIdle:   0,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(awsinterfaces.MockELBClient)
			mockCW := new(mockCWMetricsService)

			tt.setupMocks(mockClient, mockCW)

			s := &service{client: mockClient, cwService: mockCW}
			unused, idle, err := s.GetLoadBalancerWaste(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, unused, tt.wantUnused)
				assert.Len(t, idle, tt.wantIdle)

				for i, name := range tt.wantIdleNames {
					assert.Equal(t, name, idle[i].LoadBalancerName)
				}
			}
		})
	}
}
