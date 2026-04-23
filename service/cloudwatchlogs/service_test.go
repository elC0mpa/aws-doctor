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
	"github.com/elC0mpa/aws-doctor/mocks/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCloudWatchLogsWaste(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		setupMocks func(*awsinterfaces.MockCloudWatchLogsClient, *services.MockPricingService)
		wantCount  int
		wantErr    bool
	}{
		{
			name: "log group without retention policy",
			setupMocks: func(m *awsinterfaces.MockCloudWatchLogsClient, ps *services.MockPricingService) {
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

				ps.On("CalculateCloudWatchLogsMonthlyCost", int64(1024)).Return(0.03)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "log group with retention policy",
			setupMocks: func(m *awsinterfaces.MockCloudWatchLogsClient, ps *services.MockPricingService) {
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
			setupMocks: func(m *awsinterfaces.MockCloudWatchLogsClient, ps *services.MockPricingService) {
				m.On("DescribeLogGroups", mock.Anything, mock.Anything, mock.Anything).Return((*cloudwatchlogs.DescribeLogGroupsOutput)(nil), errors.New("api error"))
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(awsinterfaces.MockCloudWatchLogsClient)
			mockPricing := new(services.MockPricingService)
			tt.setupMocks(mockClient, mockPricing)

			svc := &service{
				client:         mockClient,
				pricingService: mockPricing,
			}
			waste, err := svc.GetCloudWatchLogsWaste(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, waste, tt.wantCount)

				if tt.wantCount > 0 {
					assert.Equal(t, "waste-group", waste[0].LogGroupName)
					assert.Greater(t, waste[0].EstimatedMonthlyCost, 0.0)
				}
			}
			mockPricing.AssertExpectations(t)
		})
	}
}

func TestGetLambdaMaxMemoryUsedBatch(t *testing.T) {
	ctx := context.Background()
	startTime := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 4, 14, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		logGroups  []string
		setupMocks func(*awsinterfaces.MockCloudWatchLogsClient)
		want       map[string]int32
		wantErr    bool
	}{
		{
			name:      "successful query returns max memory per log group",
			logGroups: []string{"/aws/lambda/fn-a", "/aws/lambda/fn-b"},
			setupMocks: func(m *awsinterfaces.MockCloudWatchLogsClient) {
				m.On("StartQuery", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatchlogs.StartQueryOutput{
					QueryId: aws.String("query-123"),
				}, nil)
				m.On("GetQueryResults", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatchlogs.GetQueryResultsOutput{
					Status: types.QueryStatusComplete,
					Results: [][]types.ResultField{
						{
							{Field: aws.String("@log"), Value: aws.String("123456789012:/aws/lambda/fn-a")},
							{Field: aws.String("maxMemMB"), Value: aws.String("75.5")},
						},
						{
							{Field: aws.String("@log"), Value: aws.String("123456789012:/aws/lambda/fn-b")},
							{Field: aws.String("maxMemMB"), Value: aws.String("200")},
						},
					},
				}, nil)
			},
			want: map[string]int32{
				"/aws/lambda/fn-a": 76,
				"/aws/lambda/fn-b": 200,
			},
			wantErr: false,
		},
		{
			name:      "empty log groups returns empty map without calling AWS",
			logGroups: []string{},
			setupMocks: func(_ *awsinterfaces.MockCloudWatchLogsClient) {
			},
			want:    map[string]int32{},
			wantErr: false,
		},
		{
			name:      "start query fails",
			logGroups: []string{"/aws/lambda/fn-a"},
			setupMocks: func(m *awsinterfaces.MockCloudWatchLogsClient) {
				m.On("StartQuery", mock.Anything, mock.Anything, mock.Anything).Return((*cloudwatchlogs.StartQueryOutput)(nil), errors.New("access denied"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:      "query fails status",
			logGroups: []string{"/aws/lambda/fn-a"},
			setupMocks: func(m *awsinterfaces.MockCloudWatchLogsClient) {
				m.On("StartQuery", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatchlogs.StartQueryOutput{
					QueryId: aws.String("query-456"),
				}, nil)
				m.On("GetQueryResults", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatchlogs.GetQueryResultsOutput{
					Status: types.QueryStatusFailed,
				}, nil)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:      "empty results returns empty map",
			logGroups: []string{"/aws/lambda/fn-a"},
			setupMocks: func(m *awsinterfaces.MockCloudWatchLogsClient) {
				m.On("StartQuery", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatchlogs.StartQueryOutput{
					QueryId: aws.String("query-789"),
				}, nil)
				m.On("GetQueryResults", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatchlogs.GetQueryResultsOutput{
					Status:  types.QueryStatusComplete,
					Results: [][]types.ResultField{},
				}, nil)
			},
			want:    map[string]int32{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(awsinterfaces.MockCloudWatchLogsClient)
			tt.setupMocks(mockClient)

			svc := &service{client: mockClient}
			got, err := svc.GetLambdaMaxMemoryUsedBatch(ctx, tt.logGroups, startTime, endTime)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetLambdaMaxMemoryUsedBatch_IssuesGroupByLogQuery(t *testing.T) {
	mockClient := new(awsinterfaces.MockCloudWatchLogsClient)

	mockClient.On("StartQuery", mock.Anything, mock.MatchedBy(func(input *cloudwatchlogs.StartQueryInput) bool {
		return input != nil && input.QueryString != nil &&
			aws.ToString(input.QueryString) == `filter @type = "REPORT" | stats max(@maxMemoryUsed / 1048576) as maxMemMB by @log` &&
			len(input.LogGroupNames) == 2
	}), mock.Anything).Return(&cloudwatchlogs.StartQueryOutput{
		QueryId: aws.String("query-assert"),
	}, nil)

	mockClient.On("GetQueryResults", mock.Anything, mock.Anything, mock.Anything).Return(&cloudwatchlogs.GetQueryResultsOutput{
		Status:  types.QueryStatusComplete,
		Results: [][]types.ResultField{},
	}, nil)

	svc := &service{client: mockClient}
	_, err := svc.GetLambdaMaxMemoryUsedBatch(context.Background(), []string{"/aws/lambda/a", "/aws/lambda/b"}, time.Now().AddDate(0, 0, -1), time.Now())

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestParseMaxMemMBByGroup_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		results [][]types.ResultField
		want    map[string]int32
	}{
		{
			name: "strips accountId prefix from @log",
			results: [][]types.ResultField{
				{
					{Field: aws.String("@log"), Value: aws.String("999999999999:/aws/lambda/fn")},
					{Field: aws.String("maxMemMB"), Value: aws.String("10")},
				},
			},
			want: map[string]int32{"/aws/lambda/fn": 10},
		},
		{
			name: "handles @log without colon by using the raw value",
			results: [][]types.ResultField{
				{
					{Field: aws.String("@log"), Value: aws.String("/aws/lambda/fn")},
					{Field: aws.String("maxMemMB"), Value: aws.String("20")},
				},
			},
			want: map[string]int32{"/aws/lambda/fn": 20},
		},
		{
			name: "drops rows missing @log",
			results: [][]types.ResultField{
				{
					{Field: aws.String("maxMemMB"), Value: aws.String("30")},
				},
			},
			want: map[string]int32{},
		},
		{
			name: "drops rows missing maxMemMB",
			results: [][]types.ResultField{
				{
					{Field: aws.String("@log"), Value: aws.String("1:/aws/lambda/fn")},
				},
			},
			want: map[string]int32{},
		},
		{
			name: "drops rows with unparseable maxMemMB",
			results: [][]types.ResultField{
				{
					{Field: aws.String("@log"), Value: aws.String("1:/aws/lambda/fn")},
					{Field: aws.String("maxMemMB"), Value: aws.String("not-a-number")},
				},
			},
			want: map[string]int32{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseMaxMemMBByGroup(tt.results)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestListExistingLogGroups(t *testing.T) {
	ctx := context.Background()

	mockClient := new(awsinterfaces.MockCloudWatchLogsClient)

	mockClient.On("DescribeLogGroups", mock.Anything, mock.MatchedBy(func(input *cloudwatchlogs.DescribeLogGroupsInput) bool {
		return input != nil && aws.ToString(input.LogGroupNamePrefix) == "/aws/lambda/"
	}), mock.Anything).Return(&cloudwatchlogs.DescribeLogGroupsOutput{
		LogGroups: []types.LogGroup{
			{LogGroupName: aws.String("/aws/lambda/fn-a")},
			{LogGroupName: aws.String("/aws/lambda/fn-b")},
			{LogGroupName: nil},
		},
	}, nil)

	svc := &service{client: mockClient}
	got, err := svc.ListExistingLogGroups(ctx, "/aws/lambda/")

	assert.NoError(t, err)
	assert.Equal(t, map[string]struct{}{
		"/aws/lambda/fn-a": {},
		"/aws/lambda/fn-b": {},
	}, got)
}

func TestListExistingLogGroups_DescribeError(t *testing.T) {
	mockClient := new(awsinterfaces.MockCloudWatchLogsClient)
	mockClient.On("DescribeLogGroups", mock.Anything, mock.Anything, mock.Anything).Return((*cloudwatchlogs.DescribeLogGroupsOutput)(nil), errors.New("throttled"))

	svc := &service{client: mockClient}
	_, err := svc.ListExistingLogGroups(context.Background(), "/aws/lambda/")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to describe log groups")
}

func TestNewService(t *testing.T) {
	cfg := aws.Config{}
	svc := NewService(cfg, nil)
	assert.NotNil(t, svc)
}
