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
		setupMocks        func(*awsinterfaces.MockRDSClient, *services.MockCloudWatchMetricsService, *services.MockPricingService)
		wantInstanceCount int
		wantSnapshotCount int
		wantIdleCount     int
		wantErr           bool
	}{
		{
			name: "stopped instance detected",
			setupMocks: func(m *awsinterfaces.MockRDSClient, cw *services.MockCloudWatchMetricsService, ps *services.MockPricingService) {
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

				ps.On("CalculateRDSInstanceMonthlyCost", int32(20), false).Return(2.3)

				recentTime := time.Now().AddDate(0, 0, -10)
				m.On("DescribeDBSnapshots", mock.Anything, mock.Anything, mock.Anything).Return(&rds.DescribeDBSnapshotsOutput{
					DBSnapshots: []rdstypes.DBSnapshot{
						{
							DBSnapshotIdentifier: aws.String("recent-snap"),
							DBInstanceIdentifier: aws.String("my-db"),
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
			name: "old snapshot detected",
			setupMocks: func(m *awsinterfaces.MockRDSClient, cw *services.MockCloudWatchMetricsService, ps *services.MockPricingService) {
				m.On("DescribeDBInstances", mock.Anything, mock.Anything, mock.Anything).Return(&rds.DescribeDBInstancesOutput{
					DBInstances: []rdstypes.DBInstance{},
				}, nil)

				oldTime := time.Now().AddDate(0, 0, -40)
				m.On("DescribeDBSnapshots", mock.Anything, mock.Anything, mock.Anything).Return(&rds.DescribeDBSnapshotsOutput{
					DBSnapshots: []rdstypes.DBSnapshot{
						{
							DBSnapshotIdentifier: aws.String("old-snap"),
							DBInstanceIdentifier: aws.String("my-db"),
							SnapshotCreateTime:   &oldTime,
							AllocatedStorage:     aws.Int32(100),
							Engine:               aws.String("mysql"),
						},
					},
				}, nil)

				ps.On("CalculateRDSSnapshotMonthlyCost", int32(100)).Return(9.5)
			},
			wantInstanceCount: 0,
			wantSnapshotCount: 1,
			wantIdleCount:     0,
			wantErr:           false,
		},
		{
			name: "idle instance detected",
			setupMocks: func(m *awsinterfaces.MockRDSClient, cw *services.MockCloudWatchMetricsService, ps *services.MockPricingService) {
				m.On("DescribeDBInstances", mock.Anything, mock.Anything, mock.Anything).Return(&rds.DescribeDBInstancesOutput{
					DBInstances: []rdstypes.DBInstance{
						{
							DBInstanceIdentifier: aws.String("my-idle-db"),
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

				cw.On("RDSHasZeroConnectionsInPeriod", mock.Anything, "my-idle-db", 7).Return(true, nil)

				ps.On("CalculateRDSIdleInstanceMonthlyCost", "db.t3.micro", int32(20), false).Return(15.0)
			},
			wantInstanceCount: 0,
			wantSnapshotCount: 0,
			wantIdleCount:     1,
			wantErr:           false,
		},
		{
			name: "aws api error",
			setupMocks: func(m *awsinterfaces.MockRDSClient, cw *services.MockCloudWatchMetricsService, ps *services.MockPricingService) {
				m.On("DescribeDBInstances", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("aws error"))
				m.On("DescribeDBSnapshots", mock.Anything, mock.Anything, mock.Anything).Return(&rds.DescribeDBSnapshotsOutput{}, nil)
			},
			wantInstanceCount: 0,
			wantSnapshotCount: 0,
			wantIdleCount:     0,
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRDS := new(awsinterfaces.MockRDSClient)
			mockCW := new(services.MockCloudWatchMetricsService)
			mockPricing := new(services.MockPricingService)
			tt.setupMocks(mockRDS, mockCW, mockPricing)

			svc := &service{
				client:         mockRDS,
				cwService:      mockCW,
				pricingService: mockPricing,
			}

			instances, snapshots, idle, err := svc.GetRDSWaste(ctx, 7, 30)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, instances, tt.wantInstanceCount)
			assert.Len(t, snapshots, tt.wantSnapshotCount)
			assert.Len(t, idle, tt.wantIdleCount)
			mockRDS.AssertExpectations(t)
			mockCW.AssertExpectations(t)
			mockPricing.AssertExpectations(t)
		})
	}
}

func TestNewService(t *testing.T) {
	cfg := aws.Config{}
	mockCW := new(services.MockCloudWatchMetricsService)
	svc := NewService(cfg, mockCW, nil)
	assert.NotNil(t, svc)
}

func TestAnalyzerMethods(t *testing.T) {
	svc := &service{}

	if svc.Name() == "" {
		t.Error("Name() should not be empty")
	}

	if svc.TabName() == "" {
		t.Error("TabName() should not be empty")
	}
}
