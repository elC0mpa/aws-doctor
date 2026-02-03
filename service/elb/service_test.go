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
