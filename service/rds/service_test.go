package rds

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/elC0mpa/aws-doctor/mocks/awsinterfaces"
	"github.com/elC0mpa/aws-doctor/mocks/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRDSWaste(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name              string
		setupMocks        func(*awsinterfaces.MockRDSClient, *services.MockCloudWatchMetricsService)
		wantInstanceCount int
		wantSnapshotCount int
		wantIdleCount     int
		wantErr           bool
	}{
		{
			name: "stopped instance detected",
			setupMocks: func(m *awsinterfaces.MockRDSClient, cw *services.MockCloudWatchMetricsService) {
				m.On("DescribeDBInstances", mock.Anything, mock.Anything, mock.Anything).Return(&rds.DescribeDBInstancesOutput{
					DBInstances: []rdstypes.DBInstance{
						{
							DBInstanceIdentifier: aws.String("my-stopped-db"),
							DBInstanceClass:      aws.String("db.t3.micro"),
							Engine:               aws.String("mysql"),
							DBInstanceStatus:     aws.String("stopped"),
							MultiAZ:              aws.Bool(false),
							AllocatedStorage:     aws.Int32(20),
						},
					},
				}, nil)

				recentTime := time.Now().AddDate(0, 0, -10)
				m.On("DescribeDBSnapshots", mock.Anything, mock.Anything, mock.Anything).Return(&rds.DescribeDBSnapshotsOutput{
					DBSnapshots: []rdstypes.DBSnapshot{
						{
							DBSnapshotIdentifier: aws.String("recent-snap"),
							DBInstanceIdentifier: aws.String("my-db"),
							Engine:               aws.String("mysql"),
							AllocatedStorage:     aws.Int32(20),
							SnapshotCreateTime:   &recentTime,
						},
					},
				}, nil)
			},
			wantInstanceCount: 1,
			wantSnapshotCount: 0,
			wantIdleCount:     0,
			wantErr:           false,
		},
		{
			name: "old manual snapshot detected",
			setupMocks: func(m *awsinterfaces.MockRDSClient, cw *services.MockCloudWatchMetricsService) {
				m.On("DescribeDBInstances", mock.Anything, mock.Anything, mock.Anything).Return(&rds.DescribeDBInstancesOutput{
					DBInstances: []rdstypes.DBInstance{},
				}, nil)

				oldTime := time.Now().AddDate(0, -2, 0)
				m.On("DescribeDBSnapshots", mock.Anything, mock.Anything, mock.Anything).Return(&rds.DescribeDBSnapshotsOutput{
					DBSnapshots: []rdstypes.DBSnapshot{
						{
							DBSnapshotIdentifier: aws.String("old-snap"),
							DBInstanceIdentifier: aws.String("running-db"),
							Engine:               aws.String("postgres"),
							AllocatedStorage:     aws.Int32(50),
							SnapshotCreateTime:   &oldTime,
						},
					},
				}, nil)
			},
			wantInstanceCount: 0,
			wantSnapshotCount: 1,
			wantIdleCount:     0,
			wantErr:           false,
		},
		{
			name: "idle instance detected via CloudWatch",
			setupMocks: func(m *awsinterfaces.MockRDSClient, cw *services.MockCloudWatchMetricsService) {
				m.On("DescribeDBInstances", mock.Anything, mock.Anything, mock.Anything).Return(&rds.DescribeDBInstancesOutput{
					DBInstances: []rdstypes.DBInstance{
						{
							DBInstanceIdentifier: aws.String("idle-db"),
							DBInstanceClass:      aws.String("db.t3.micro"),
							Engine:               aws.String("mysql"),
							DBInstanceStatus:     aws.String("available"),
							MultiAZ:              aws.Bool(false),
							AllocatedStorage:     aws.Int32(20),
						},
					},
				}, nil)

				m.On("DescribeDBSnapshots", mock.Anything, mock.Anything, mock.Anything).Return(&rds.DescribeDBSnapshotsOutput{
					DBSnapshots: []rdstypes.DBSnapshot{},
				}, nil)

				cw.On("RDSHasZeroConnectionsInPeriod", mock.Anything, "idle-db", idleDaysThreshold).Return(true, nil)
			},
			wantInstanceCount: 0,
			wantSnapshotCount: 0,
			wantIdleCount:     1,
			wantErr:           false,
		},
		{
			name: "active instance not flagged as idle",
			setupMocks: func(m *awsinterfaces.MockRDSClient, cw *services.MockCloudWatchMetricsService) {
				m.On("DescribeDBInstances", mock.Anything, mock.Anything, mock.Anything).Return(&rds.DescribeDBInstancesOutput{
					DBInstances: []rdstypes.DBInstance{
						{
							DBInstanceIdentifier: aws.String("active-db"),
							DBInstanceClass:      aws.String("db.t3.micro"),
							Engine:               aws.String("mysql"),
							DBInstanceStatus:     aws.String("available"),
							MultiAZ:              aws.Bool(false),
							AllocatedStorage:     aws.Int32(20),
						},
					},
				}, nil)

				m.On("DescribeDBSnapshots", mock.Anything, mock.Anything, mock.Anything).Return(&rds.DescribeDBSnapshotsOutput{
					DBSnapshots: []rdstypes.DBSnapshot{},
				}, nil)

				cw.On("RDSHasZeroConnectionsInPeriod", mock.Anything, "active-db", idleDaysThreshold).Return(false, nil)
			},
			wantInstanceCount: 0,
			wantSnapshotCount: 0,
			wantIdleCount:     0,
			wantErr:           false,
		},
		{
			name: "no waste found",
			setupMocks: func(m *awsinterfaces.MockRDSClient, cw *services.MockCloudWatchMetricsService) {
				m.On("DescribeDBInstances", mock.Anything, mock.Anything, mock.Anything).Return(&rds.DescribeDBInstancesOutput{
					DBInstances: []rdstypes.DBInstance{},
				}, nil)
				m.On("DescribeDBSnapshots", mock.Anything, mock.Anything, mock.Anything).Return(&rds.DescribeDBSnapshotsOutput{
					DBSnapshots: []rdstypes.DBSnapshot{},
				}, nil)
			},
			wantInstanceCount: 0,
			wantSnapshotCount: 0,
			wantIdleCount:     0,
			wantErr:           false,
		},
		{
			name: "describe instances fails",
			setupMocks: func(m *awsinterfaces.MockRDSClient, cw *services.MockCloudWatchMetricsService) {
				m.On("DescribeDBInstances", mock.Anything, mock.Anything, mock.Anything).Return((*rds.DescribeDBInstancesOutput)(nil), errors.New("api error"))
				m.On("DescribeDBSnapshots", mock.Anything, mock.Anything, mock.Anything).Return(&rds.DescribeDBSnapshotsOutput{
					DBSnapshots: []rdstypes.DBSnapshot{},
				}, nil).Maybe()
			},
			wantErr: true,
		},
		{
			name: "describe snapshots fails",
			setupMocks: func(m *awsinterfaces.MockRDSClient, cw *services.MockCloudWatchMetricsService) {
				m.On("DescribeDBInstances", mock.Anything, mock.Anything, mock.Anything).Return(&rds.DescribeDBInstancesOutput{
					DBInstances: []rdstypes.DBInstance{},
				}, nil).Maybe()
				m.On("DescribeDBSnapshots", mock.Anything, mock.Anything, mock.Anything).Return((*rds.DescribeDBSnapshotsOutput)(nil), errors.New("snapshot api error"))
			},
			wantErr: true,
		},
		{
			name: "cloudwatch error skips instance gracefully",
			setupMocks: func(m *awsinterfaces.MockRDSClient, cw *services.MockCloudWatchMetricsService) {
				m.On("DescribeDBInstances", mock.Anything, mock.Anything, mock.Anything).Return(&rds.DescribeDBInstancesOutput{
					DBInstances: []rdstypes.DBInstance{
						{
							DBInstanceIdentifier: aws.String("cw-error-db"),
							DBInstanceClass:      aws.String("db.t3.micro"),
							Engine:               aws.String("mysql"),
							DBInstanceStatus:     aws.String("available"),
							MultiAZ:              aws.Bool(false),
							AllocatedStorage:     aws.Int32(20),
						},
					},
				}, nil)

				m.On("DescribeDBSnapshots", mock.Anything, mock.Anything, mock.Anything).Return(&rds.DescribeDBSnapshotsOutput{
					DBSnapshots: []rdstypes.DBSnapshot{},
				}, nil)

				cw.On("RDSHasZeroConnectionsInPeriod", mock.Anything, "cw-error-db", idleDaysThreshold).Return(false, errors.New("cloudwatch error"))
			},
			wantInstanceCount: 0,
			wantSnapshotCount: 0,
			wantIdleCount:     0,
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(awsinterfaces.MockRDSClient)
			mockCW := new(services.MockCloudWatchMetricsService)
			tt.setupMocks(mockClient, mockCW)

			svc := &service{client: mockClient, cwService: mockCW}
			instances, snapshots, idle, err := svc.GetRDSWaste(ctx)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, instances, tt.wantInstanceCount)
				assert.Len(t, snapshots, tt.wantSnapshotCount)
				assert.Len(t, idle, tt.wantIdleCount)

				if tt.wantInstanceCount > 0 {
					assert.Greater(t, instances[0].EstimatedMonthlyCost, 0.0)
				}

				if tt.wantSnapshotCount > 0 {
					assert.Greater(t, snapshots[0].EstimatedMonthlyCost, 0.0)
				}

				if tt.wantIdleCount > 0 {
					assert.Greater(t, idle[0].EstimatedMonthlyCost, 0.0)
					assert.Equal(t, idleDaysThreshold, idle[0].DaysChecked)
				}
			}
		})
	}
}

func TestNewService(t *testing.T) {
	cfg := aws.Config{}
	mockCW := new(services.MockCloudWatchMetricsService)
	svc := NewService(cfg, mockCW)
	assert.NotNil(t, svc)
}
