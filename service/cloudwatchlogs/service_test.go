package cloudwatchlogs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/elC0mpa/aws-doctor/mocks/awsinterfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCloudWatchLogsWaste(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		setupMocks func(*awsinterfaces.MockCloudWatchLogsClient)
		wantCount  int
		wantErr    bool
	}{
		{
			name: "log group without retention policy",
			setupMocks: func(m *awsinterfaces.MockCloudWatchLogsClient) {
				m.On("DescribeLogGroups", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatchlogs.DescribeLogGroupsOutput{
					LogGroups: []types.LogGroup{
						{
							LogGroupName:    aws.String("waste-group"),
							RetentionInDays: nil,
							CreationTime:    aws.Int64(1600000000000),
							StoredBytes:     aws.Int64(1024),
						},
					},
				}, nil)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "log group with retention policy",
			setupMocks: func(m *awsinterfaces.MockCloudWatchLogsClient) {
				m.On("DescribeLogGroups", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatchlogs.DescribeLogGroupsOutput{
					LogGroups: []types.LogGroup{
						{
							LogGroupName:    aws.String("clean-group"),
							RetentionInDays: aws.Int32(7),
							CreationTime:    aws.Int64(1600000000000),
							StoredBytes:     aws.Int64(1024),
						},
					},
				}, nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "describe log groups fails",
			setupMocks: func(m *awsinterfaces.MockCloudWatchLogsClient) {
				m.On("DescribeLogGroups", mock.Anything, mock.Anything, mock.Anything).Return((*cloudwatchlogs.DescribeLogGroupsOutput)(nil), errors.New("api error"))
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(awsinterfaces.MockCloudWatchLogsClient)
			tt.setupMocks(mockClient)

			svc := &service{client: mockClient}
			waste, err := svc.GetCloudWatchLogsWaste(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, waste, tt.wantCount)

				if tt.wantCount > 0 {
					assert.Equal(t, "waste-group", waste[0].LogGroupName)
					// 1024 bytes / (1024^3) * 0.03 = small value
					assert.Greater(t, waste[0].EstimatedMonthlyCost, 0.0)
				}
			}
		})
	}
}

func TestGetLambdaMaxMemoryUsed(t *testing.T) {
	ctx := context.Background()
	startTime := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 4, 14, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		setupMocks func(*awsinterfaces.MockCloudWatchLogsClient)
		wantMB     int32
		wantErr    bool
	}{
		{
			name: "successful query returns max memory",
			setupMocks: func(m *awsinterfaces.MockCloudWatchLogsClient) {
				m.On("StartQuery", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatchlogs.StartQueryOutput{
					QueryId: aws.String("query-123"),
				}, nil)
				m.On("GetQueryResults", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatchlogs.GetQueryResultsOutput{
					Status: types.QueryStatusComplete,
					Results: [][]types.ResultField{
						{
							{Field: aws.String("maxMemMB"), Value: aws.String("75.5")},
						},
					},
				}, nil)
			},
			wantMB:  76, // ceil(75.5)
			wantErr: false,
		},
		{
			name: "start query fails",
			setupMocks: func(m *awsinterfaces.MockCloudWatchLogsClient) {
				m.On("StartQuery", mock.Anything, mock.Anything, mock.Anything).Return((*cloudwatchlogs.StartQueryOutput)(nil), errors.New("access denied"))
			},
			wantMB:  0,
			wantErr: true,
		},
		{
			name: "query fails status",
			setupMocks: func(m *awsinterfaces.MockCloudWatchLogsClient) {
				m.On("StartQuery", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatchlogs.StartQueryOutput{
					QueryId: aws.String("query-456"),
				}, nil)
				m.On("GetQueryResults", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatchlogs.GetQueryResultsOutput{
					Status: types.QueryStatusFailed,
				}, nil)
			},
			wantMB:  0,
			wantErr: true,
		},
		{
			name: "empty results returns zero",
			setupMocks: func(m *awsinterfaces.MockCloudWatchLogsClient) {
				m.On("StartQuery", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatchlogs.StartQueryOutput{
					QueryId: aws.String("query-789"),
				}, nil)
				m.On("GetQueryResults", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatchlogs.GetQueryResultsOutput{
					Status:  types.QueryStatusComplete,
					Results: [][]types.ResultField{},
				}, nil)
			},
			wantMB:  0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(awsinterfaces.MockCloudWatchLogsClient)
			tt.setupMocks(mockClient)

			svc := &service{client: mockClient}
			mem, err := svc.GetLambdaMaxMemoryUsed(ctx, "/aws/lambda/test-fn", startTime, endTime)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantMB, mem)
			}
		})
	}
}

func TestNewService(t *testing.T) {
	cfg := aws.Config{}
	svc := NewService(cfg)
	assert.NotNil(t, svc)
}
