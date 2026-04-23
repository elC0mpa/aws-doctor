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

func TestGetIdleNATGateways(t *testing.T) {
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
				"nat-abcdef01234567890": 1024,
			},
			expectedLen:   0,
			expectedError: false,
		},
		{
			name: "should ignore deleted or pending NAT Gateways",
			natGateways: []types.NatGateway{
				{
					NatGatewayId: aws.String("nat-deleted"),
					State:        types.NatGatewayStateDeleted,
				},
				{
					NatGatewayId: aws.String("nat-pending"),
					State:        types.NatGatewayStatePending,
				},
			},
			bytesOutMap: map[string]float64{
				"nat-deleted": 0,
				"nat-pending": 0,
			},
			expectedLen:   0,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			idleDays := 7

			// Setup mock EC2 client
			mockEC2 := new(awsinterfaces.MockEC2Client)
			mockEC2.On("DescribeNatGateways", mock.Anything, mock.Anything, mock.Anything).Return(&ec2.DescribeNatGatewaysOutput{
				NatGateways: tt.natGateways,
			}, nil)

			// Setup mock CloudWatch service
			mockCW := new(services.MockCloudWatchMetricsService)
			for natGatewayID, bytesOut := range tt.bytesOutMap {
				mockCW.On("NATGatewayBytesOut", mock.Anything, natGatewayID, idleDays).Return(bytesOut, nil)
			}

			// Setup mock pricing service
			mockPricing := new(services.MockPricingService)
			if tt.expectedLen > 0 {
				mockPricing.On("CalculateNATGatewayMonthlyCost").Return(32.85)
			}

			svc := &service{
				client:         mockEC2,
				cwService:      mockCW,
				pricingService: mockPricing,
			}

			result, err := svc.GetIdleNATGateways(ctx, idleDays)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLen, len(result))
			}

			mockEC2.AssertExpectations(t)
			mockCW.AssertExpectations(t)
			mockPricing.AssertExpectations(t)
		})
	}
}

func TestGetIdleNATGateways_Error(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		setupMocks    func(*awsinterfaces.MockEC2Client, *services.MockCloudWatchMetricsService)
		expectedError bool
	}{
		{
			name: "EC2 API error",
			setupMocks: func(ec2Mock *awsinterfaces.MockEC2Client, cwMock *services.MockCloudWatchMetricsService) {
				ec2Mock.On("DescribeNatGateways", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("EC2 error"))
			},
			expectedError: true,
		},
		{
			name: "CloudWatch API error should log the error for the NAT gateway and continue",
			setupMocks: func(ec2Mock *awsinterfaces.MockEC2Client, cwMock *services.MockCloudWatchMetricsService) {
				ec2Mock.On("DescribeNatGateways", mock.Anything, mock.Anything, mock.Anything).Return(&ec2.DescribeNatGatewaysOutput{
					NatGateways: []types.NatGateway{
						{
							NatGatewayId: aws.String("nat-123"),
							State:        types.NatGatewayStateAvailable,
						},
					},
				}, nil)

				cwMock.On("NATGatewayBytesOut", mock.Anything, "nat-123", 7).Return(0.0, errors.New("CW error"))
			},
			expectedError: false, // CloudWatch errors are logged and we continue
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEC2 := new(awsinterfaces.MockEC2Client)
			mockCW := new(services.MockCloudWatchMetricsService)
			tt.setupMocks(mockEC2, mockCW)

			svc := &service{
				client:         mockEC2,
				cwService:      mockCW,
				pricingService: new(services.MockPricingService),
			}

			_, err := svc.GetIdleNATGateways(ctx, 7)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewService(t *testing.T) {
	cfg := aws.Config{}
	svc := NewService(cfg, nil, nil)
	assert.NotNil(t, svc)
}
