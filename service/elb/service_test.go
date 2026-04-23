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

func (m *mockCWMetricsService) ELBHasZeroRequestsInPeriod(ctx context.Context, loadBalancerArn string, lbType types.LoadBalancerTypeEnum, days int) (bool, error) {
	args := m.Called(ctx, loadBalancerArn, lbType, days)

	return args.Bool(0), args.Error(1)
}

type mockPricingService struct {
	mock.Mock
}

func (m *mockPricingService) CalculateLoadBalancerMonthlyCost(lbType types.LoadBalancerTypeEnum) float64 {
	args := m.Called(lbType)
	return args.Get(0).(float64)
}

func TestGetLoadBalancerWaste(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*awsinterfaces.MockELBClient, *mockCWMetricsService, *mockPricingService)
		wantUnused    int
		wantIdle      int
		wantErr       bool
		wantIdleNames []string
	}{
		{
			name: "separates unused and idle load balancers",
			setupMocks: func(mc *awsinterfaces.MockELBClient, cw *mockCWMetricsService, ps *mockPricingService) {
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
							LoadBalancerArn:  aws.String("arn:aws:elasticloadbalancing:us-east-1:123:loadbalancer/app/unused-alb/ghi"),
							Type:             types.LoadBalancerTypeEnumApplication,
							LoadBalancerName: aws.String("unused-alb"),
						},
					},
				}, nil)

				mc.On("DescribeTargetGroups", mock.Anything, mock.Anything, mock.Anything).Return(&elb.DescribeTargetGroupsOutput{
					TargetGroups: []types.TargetGroup{
						{LoadBalancerArns: []string{"arn:aws:elasticloadbalancing:us-east-1:123:loadbalancer/app/idle-alb/abc"}},
						{LoadBalancerArns: []string{"arn:aws:elasticloadbalancing:us-east-1:123:loadbalancer/net/active-nlb/def"}},
					},
				}, nil)

				cw.On("ELBHasZeroRequestsInPeriod", mock.Anything, "arn:aws:elasticloadbalancing:us-east-1:123:loadbalancer/app/idle-alb/abc", types.LoadBalancerTypeEnumApplication, 7).Return(true, nil)
				cw.On("ELBHasZeroRequestsInPeriod", mock.Anything, "arn:aws:elasticloadbalancing:us-east-1:123:loadbalancer/net/active-nlb/def", types.LoadBalancerTypeEnumNetwork, 7).Return(false, nil)

				ps.On("CalculateLoadBalancerMonthlyCost", types.LoadBalancerTypeEnumApplication).Return(16.43)
			},
			wantUnused:    1,
			wantIdle:      1,
			wantIdleNames: []string{"idle-alb"},
			wantErr:       false,
		},
		{
			name: "aws api error",
			setupMocks: func(mc *awsinterfaces.MockELBClient, cw *mockCWMetricsService, ps *mockPricingService) {
				mc.On("DescribeLoadBalancers", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("aws error"))
			},
			wantUnused: 0,
			wantIdle:   0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockELB := new(awsinterfaces.MockELBClient)
			mockCW := new(mockCWMetricsService)
			mockPricing := new(mockPricingService)
			tt.setupMocks(mockELB, mockCW, mockPricing)

			svc := &service{
				client:         mockELB,
				cwService:      mockCW,
				pricingService: mockPricing,
			}

			unused, idle, err := svc.GetLoadBalancerWaste(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, unused, tt.wantUnused)
			assert.Len(t, idle, tt.wantIdle)

			for i, name := range tt.wantIdleNames {
				assert.Equal(t, name, idle[i].Name)
			}

			mockELB.AssertExpectations(t)
			mockCW.AssertExpectations(t)
			mockPricing.AssertExpectations(t)
		})
	}
}

func TestNewService(t *testing.T) {
	cfg := aws.Config{}
	svc := NewService(cfg, nil, nil)
	assert.NotNil(t, svc)
}
