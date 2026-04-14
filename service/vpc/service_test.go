package vpc

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elC0mpa/aws-doctor/mocks/awsinterfaces"
	"github.com/elC0mpa/aws-doctor/mocks/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIdleNATGateways(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		natGateways   []types.NatGateway
		bytesOutMap   map[string]float64
		expectedLen   int
		expectedError bool
	}{
		{
			name: "idle NAT Gateway with 0 bytes should be detected",
			natGateways: []types.NatGateway{
				{
					NatGatewayId: aws.String("nat-1234567890abcdef0"),
					VpcId:        aws.String("vpc-12345678"),
					SubnetId:     aws.String("subnet-12345678"),
					State:        types.NatGatewayStateAvailable,
				},
			},
			bytesOutMap: map[string]float64{
				"nat-1234567890abcdef0": 0,
			},
			expectedLen:   1,
			expectedError: false,
		},
		{
			name: "active NAT Gateway with traffic should NOT be detected",
			natGateways: []types.NatGateway{
				{
					NatGatewayId: aws.String("nat-abcdef01234567890"),
					VpcId:        aws.String("vpc-87654321"),
					SubnetId:     aws.String("subnet-87654321"),
					State:        types.NatGatewayStateAvailable,
				},
			},
			bytesOutMap: map[string]float64{
				"nat-abcdef01234567890": 1024 * 1024, // 1MB
			},
			expectedLen:   0,
			expectedError: false,
		},
		{
			name: "mixed idle and active NAT Gateways",
			natGateways: []types.NatGateway{
				{
					NatGatewayId: aws.String("nat-idle-1"),
					VpcId:        aws.String("vpc-11111111"),
					SubnetId:     aws.String("subnet-11111111"),
					State:        types.NatGatewayStateAvailable,
				},
				{
					NatGatewayId: aws.String("nat-active-1"),
					VpcId:        aws.String("vpc-22222222"),
					SubnetId:     aws.String("subnet-22222222"),
					State:        types.NatGatewayStateAvailable,
				},
			},
			bytesOutMap: map[string]float64{
				"nat-idle-1":   0,
				"nat-active-1": 512 * 1024, // 512KB
			},
			expectedLen:   1,
			expectedError: false,
		},
		{
			name:          "empty NAT Gateway list should return empty slice",
			natGateways:   []types.NatGateway{},
			bytesOutMap:   map[string]float64{},
			expectedLen:   0,
			expectedError: false,
		},
		{
			name: "NAT Gateway with nil ID should be skipped",
			natGateways: []types.NatGateway{
				{
					NatGatewayId: nil, // nil ID
					VpcId:        aws.String("vpc-12345678"),
					SubnetId:     aws.String("subnet-12345678"),
					State:        types.NatGatewayStateAvailable,
				},
			},
			bytesOutMap:   map[string]float64{},
			expectedLen:   0,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock VPC client
			mockClient := new(awsinterfaces.MockVPCClient)
			mockClient.On("DescribeNatGateways", mock.Anything, mock.Anything, mock.Anything).Return(&ec2.DescribeNatGatewaysOutput{
				NatGateways: tt.natGateways,
			}, nil)

			// Setup mock CloudWatch service
			mockCW := new(services.MockCloudWatchMetricsService)
			for natID, bytesOut := range tt.bytesOutMap {
				mockCW.On("NatGatewayBytesOut", mock.Anything, natID, 7).Return(bytesOut, nil)
			}

			// Create service with mocks
			svc := &service{
				client:    mockClient,
				cwService: mockCW,
			}

			// Execute
			results, err := svc.IdleNATGateways(ctx, 7)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Len(t, results, tt.expectedLen)
		})
	}
}

func TestIdleNATGateways_Error(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		setupMocks    func(*awsinterfaces.MockVPCClient, *services.MockCloudWatchMetricsService)
		expectedError bool
	}{
		{
			name: "DescribeNATGateways error",
			setupMocks: func(mockClient *awsinterfaces.MockVPCClient, mockCW *services.MockCloudWatchMetricsService) {
				mockClient.On("DescribeNatGateways", mock.Anything, mock.Anything, mock.Anything).Return(
					(*ec2.DescribeNatGatewaysOutput)(nil), errors.New("API error"))
			},
			expectedError: true,
		},
		{
			name: "CloudWatch error continues processing",
			setupMocks: func(mockClient *awsinterfaces.MockVPCClient, mockCW *services.MockCloudWatchMetricsService) {
				mockClient.On("DescribeNatGateways", mock.Anything, mock.Anything, mock.Anything).Return(&ec2.DescribeNatGatewaysOutput{
					NatGateways: []types.NatGateway{
						{
							NatGatewayId: aws.String("nat-123"),
							VpcId:        aws.String("vpc-123"),
							SubnetId:     aws.String("subnet-123"),
							State:        types.NatGatewayStateAvailable,
						},
					},
				}, nil)
				// When CloudWatch errors, we skip the NAT gateway and continue
				mockCW.On("NatGatewayBytesOut", mock.Anything, "nat-123", 7).Return(0.0, errors.New("CW error"))
			},
			expectedError: false, // CloudWatch errors are logged and we continue
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(awsinterfaces.MockVPCClient)
			mockCW := new(services.MockCloudWatchMetricsService)
			tt.setupMocks(mockClient, mockCW)

			svc := &service{
				client:    mockClient,
				cwService: mockCW,
			}

			results, err := svc.IdleNATGateways(ctx, 7)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// results can be empty or nil depending on what was skipped
				_ = results // we just verify no error occurred
			}
		})
	}
}
