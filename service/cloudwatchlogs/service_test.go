package cloudwatchlogs

import (
	"context"
	"errors"
	"testing"

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
				}
			}
		})
	}
}
