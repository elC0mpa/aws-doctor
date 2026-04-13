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

// mockCWMetricsService is a local mock for cloudwatchMetricsService used by GetIdleLoadBalancers.
type mockCWMetricsService struct {
	mock.Mock
}

func (m *mockCWMetricsService) ELBHasZeroRequestsInPeriod(ctx context.Context, loadBalancerArn string, lbType string, days int) (bool, error) {
	args := m.Called(ctx, loadBalancerArn, lbType, days)

	return args.Bool(0), args.Error(1)
}

func TestGetUnusedLoadBalancers(t *testing.T) {
	mockClient := new(awsinterfaces.MockELBClient)
	s := &service{client: mockClient}

	// Mock DescribeLoadBalancers
	mockClient.On("DescribeLoadBalancers", mock.Anything, mock.Anything, mock.Anything).Return(&elb.DescribeLoadBalancersOutput{
		LoadBalancers: []types.LoadBalancer{
			// Used ALB
			{
				LoadBalancerArn:  aws.String("arn:alb:used"),
				Type:             types.LoadBalancerTypeEnumApplication,
				LoadBalancerName: aws.String("used-alb"),
			},
			// Unused ALB
			{
				LoadBalancerArn:  aws.String("arn:alb:unused"),
				Type:             types.LoadBalancerTypeEnumApplication,
				LoadBalancerName: aws.String("unused-alb"),
			},
			// Used NLB
			{
				LoadBalancerArn:  aws.String("arn:nlb:used"),
				Type:             types.LoadBalancerTypeEnumNetwork,
				LoadBalancerName: aws.String("used-nlb"),
			},
			// Unused NLB
			{
				LoadBalancerArn:  aws.String("arn:nlb:unused"),
				Type:             types.LoadBalancerTypeEnumNetwork,
				LoadBalancerName: aws.String("unused-nlb"),
			},
			// Other Type (e.g. Gateway - should be skipped by logic, but for safety)
			{
				LoadBalancerArn:  aws.String("arn:gwlb:unused"),
				Type:             types.LoadBalancerTypeEnumGateway,
				LoadBalancerName: aws.String("unused-gwlb"),
			},
		},
	}, nil)

	// Mock DescribeTargetGroups (defines which LBs are used)
	mockClient.On("DescribeTargetGroups", mock.Anything, mock.Anything, mock.Anything).Return(&elb.DescribeTargetGroupsOutput{
		TargetGroups: []types.TargetGroup{
			{
				LoadBalancerArns: []string{"arn:alb:used"},
			},
			{
				LoadBalancerArns: []string{"arn:nlb:used"},
			},
		},
	}, nil)

	result, err := s.GetUnusedLoadBalancers(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Verify we got the unused ones
	foundUnusedALB := false
	foundUnusedNLB := false

	for _, lb := range result {
		if *lb.LoadBalancerName == "unused-alb" {
			foundUnusedALB = true
		}

		if *lb.LoadBalancerName == "unused-nlb" {
			foundUnusedNLB = true
		}
	}

	assert.True(t, foundUnusedALB)
	assert.True(t, foundUnusedNLB)

	mockClient.AssertExpectations(t)
}

func TestGetUnusedLoadBalancers_LBError(t *testing.T) {
	mockClient := new(awsinterfaces.MockELBClient)
	s := &service{client: mockClient}

	mockClient.On("DescribeLoadBalancers", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("LB API error"))

	_, err := s.GetUnusedLoadBalancers(context.Background())
	assert.Error(t, err)
	mockClient.AssertExpectations(t)
}

func TestGetUnusedLoadBalancers_TGError(t *testing.T) {
	mockClient := new(awsinterfaces.MockELBClient)
	s := &service{client: mockClient}

	mockClient.On("DescribeLoadBalancers", mock.Anything, mock.Anything, mock.Anything).Return(&elb.DescribeLoadBalancersOutput{}, nil)
	mockClient.On("DescribeTargetGroups", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("TG API error"))

	_, err := s.GetUnusedLoadBalancers(context.Background())
	assert.Error(t, err)
	mockClient.AssertExpectations(t)
}

func TestGetIdleLoadBalancers(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*awsinterfaces.MockELBClient, *mockCWMetricsService)
		wantCount  int
		wantErr    bool
		wantNames  []string
	}{
		{
			name: "returns idle ALB with target groups but zero traffic",
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
			wantCount: 1,
			wantErr:   false,
			wantNames: []string{"idle-alb"},
		},
		{
			name: "LB error returns error",
			setupMocks: func(mc *awsinterfaces.MockELBClient, cw *mockCWMetricsService) {
				mc.On("DescribeLoadBalancers", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("LB error"))
			},
			wantCount: 0,
			wantErr:   true,
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
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(awsinterfaces.MockELBClient)
			mockCW := new(mockCWMetricsService)

			tt.setupMocks(mockClient, mockCW)

			s := &service{client: mockClient, cwService: mockCW}
			result, err := s.GetIdleLoadBalancers(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.wantCount)

				for i, name := range tt.wantNames {
					assert.Equal(t, name, result[i].LoadBalancerName)
				}
			}
		})
	}
}
