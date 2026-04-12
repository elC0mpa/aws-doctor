package cloudwatchmetrics

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
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

func TestGetNatGatewayBytesOut(t *testing.T) {
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
			bytes, err := svc.GetNatGatewayBytesOut(ctx, "nat-123", 7)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantBytes, bytes)
			}
		})
	}
}

func TestNewService(t *testing.T) {
	cfg := aws.Config{}
	svc := NewService(cfg)
	assert.NotNil(t, svc)
}
