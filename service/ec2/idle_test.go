package ec2

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

func TestGetIdleInstances(t *testing.T) {
	ctx := context.Background()

	const (
		idleDays           = 14
		cpuThreshold       = 5.0
		networkBytesPerDay = 5 * 1024 * 1024
	)

	tests := []struct {
		name      string
		setup     func(*awsinterfaces.MockEC2Client, *services.MockCloudWatchMetricsService, *services.MockPricingService)
		wantCount int
		wantErr   bool
	}{
		{
			name: "flags an idle instance and skips an active one",
			setup: func(c *awsinterfaces.MockEC2Client, cw *services.MockCloudWatchMetricsService, ps *services.MockPricingService) {
				c.On("DescribeInstances", mock.Anything, mock.Anything, mock.Anything).Return(&ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{{
						Instances: []types.Instance{
							{
								InstanceId:   aws.String("i-idle"),
								InstanceType: types.InstanceType("t3.medium"),
								Tags: []types.Tag{
									{Key: aws.String("Name"), Value: aws.String("dev-box")},
								},
							},
							{
								InstanceId:   aws.String("i-active"),
								InstanceType: types.InstanceType("m5.large"),
							},
						},
					}},
				}, nil)

				cw.On("EC2InstanceIdleStats", mock.Anything, "i-idle", idleDays).Return(1.5, float64(1024), nil)
				cw.On("EC2InstanceIdleStats", mock.Anything, "i-active", idleDays).Return(60.0, float64(100*1024*1024), nil)

				ps.On("CalculateEC2InstanceMonthlyCost", "t3.medium").Return(30.37)
			},
			wantCount: 1,
		},
		{
			name: "skips instance when only network exceeds threshold",
			setup: func(c *awsinterfaces.MockEC2Client, cw *services.MockCloudWatchMetricsService, _ *services.MockPricingService) {
				c.On("DescribeInstances", mock.Anything, mock.Anything, mock.Anything).Return(&ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{{
						Instances: []types.Instance{
							{InstanceId: aws.String("i-low-cpu"), InstanceType: types.InstanceType("t3.small")},
						},
					}},
				}, nil)

				cw.On("EC2InstanceIdleStats", mock.Anything, "i-low-cpu", idleDays).Return(1.0, float64(50*1024*1024), nil)
			},
			wantCount: 0,
		},
		{
			name: "cloudwatch errors are skipped per instance",
			setup: func(c *awsinterfaces.MockEC2Client, cw *services.MockCloudWatchMetricsService, _ *services.MockPricingService) {
				c.On("DescribeInstances", mock.Anything, mock.Anything, mock.Anything).Return(&ec2.DescribeInstancesOutput{
					Reservations: []types.Reservation{{
						Instances: []types.Instance{
							{InstanceId: aws.String("i-broken"), InstanceType: types.InstanceType("t3.micro")},
						},
					}},
				}, nil)

				cw.On("EC2InstanceIdleStats", mock.Anything, "i-broken", idleDays).Return(0.0, 0.0, errors.New("cw err"))
			},
			wantCount: 0,
		},
		{
			name: "describe error is returned",
			setup: func(c *awsinterfaces.MockEC2Client, _ *services.MockCloudWatchMetricsService, _ *services.MockPricingService) {
				c.On("DescribeInstances", mock.Anything, mock.Anything, mock.Anything).Return((*ec2.DescribeInstancesOutput)(nil), errors.New("api"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(awsinterfaces.MockEC2Client)
			mockCW := new(services.MockCloudWatchMetricsService)
			mockPricing := new(services.MockPricingService)
			tt.setup(mockClient, mockCW, mockPricing)

			svc := &service{
				client:         mockClient,
				cwService:      mockCW,
				pricingService: mockPricing,
			}

			result, err := svc.GetIdleInstances(ctx, idleDays, cpuThreshold, networkBytesPerDay)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, result, tt.wantCount)

			if tt.wantCount > 0 {
				assert.Equal(t, "dev-box", result[0].Name)
				assert.Equal(t, "t3.medium", result[0].InstanceType)
				assert.Equal(t, idleDays, result[0].DaysChecked)
				assert.Equal(t, 30.37, result[0].EstimatedMonthlyCost)
			}

			mockClient.AssertExpectations(t)
			mockCW.AssertExpectations(t)
			mockPricing.AssertExpectations(t)
		})
	}
}

func TestNameTag(t *testing.T) {
	t.Run("matches Name", func(t *testing.T) {
		tags := []types.Tag{
			{Key: aws.String("env"), Value: aws.String("dev")},
			{Key: aws.String("Name"), Value: aws.String("primary-web")},
		}

		assert.Equal(t, "primary-web", nameTag(tags))
	})

	t.Run("matches lowercase name", func(t *testing.T) {
		tags := []types.Tag{
			{Key: aws.String("name"), Value: aws.String("worker-1")},
		}

		assert.Equal(t, "worker-1", nameTag(tags))
	})

	t.Run("empty tags returns empty", func(t *testing.T) {
		assert.Equal(t, "", nameTag([]types.Tag{}))
	})
}
