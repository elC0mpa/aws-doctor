package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	cetypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/assert"
)

func TestFormatDateToMonthYear(t *testing.T) {
	s := &service{}

	tests := []struct {
		input    *string
		expected string
	}{
		{input: aws.String("2026-03-27"), expected: "March 2026"},
		{input: aws.String("2023-12-01"), expected: "December 2023"},
		{input: aws.String("2024-02-29"), expected: "February 2024"},
		{input: aws.String("invalid"), expected: "invalid"},
		{input: aws.String(""), expected: ""},
		{input: nil, expected: ""},
	}

	for _, tt := range tests {
		result := s.formatDateToMonthYear(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestGetDayWithSuffix(t *testing.T) {
	s := &service{}

	tests := []struct {
		day      int
		expected string
	}{
		{day: 1, expected: "1st"},
		{day: 2, expected: "2nd"},
		{day: 3, expected: "3rd"},
		{day: 4, expected: "4th"},
		{day: 10, expected: "10th"},
		{day: 11, expected: "11th"},
		{day: 12, expected: "12th"},
		{day: 13, expected: "13th"},
		{day: 14, expected: "14th"},
		{day: 20, expected: "20th"},
		{day: 21, expected: "21st"},
		{day: 22, expected: "22nd"},
		{day: 23, expected: "23rd"},
		{day: 24, expected: "24th"},
		{day: 30, expected: "30th"},
		{day: 31, expected: "31st"},
	}

	for _, tt := range tests {
		result := s.getDayWithSuffix(tt.day)
		assert.Equal(t, tt.expected, result)
	}
}

func TestGetReportPath(t *testing.T) {
	s := &service{}

	t.Run("CustomPath", func(t *testing.T) {
		path := "my-report.pdf"
		result := s.getReportPath(path, "cost")
		assert.Equal(t, path, result)
	})

	t.Run("DefaultPath", func(t *testing.T) {
		result := s.getReportPath("DEFAULT", "cost")
		assert.True(t, strings.HasPrefix(result, "aws-doctor-cost-"))
		assert.True(t, strings.HasSuffix(result, ".pdf"))
	})

	t.Run("EmptyPath", func(t *testing.T) {
		result := s.getReportPath("", "waste")
		assert.True(t, strings.HasPrefix(result, "aws-doctor-waste-"))
		assert.True(t, strings.HasSuffix(result, ".pdf"))
	})
}

func TestFormatDayRange(t *testing.T) {
	s := &service{}

	tests := []struct {
		start    *string
		end      *string
		expected string
	}{
		{start: aws.String("2026-03-01"), end: aws.String("2026-03-15"), expected: "(1st to 15th)"},
		{start: aws.String("2026-03-21"), end: aws.String("2026-03-22"), expected: "(21st to 22nd)"},
		{start: aws.String("2026-01-31"), end: aws.String("2026-02-01"), expected: "(31st to 1st)"},
		{start: aws.String("invalid"), end: aws.String("2026-03-15"), expected: "(invalid to 2026-03-15)"},
		{start: aws.String("2026-03-01"), end: aws.String("invalid"), expected: "(2026-03-01 to invalid)"},
		{start: nil, end: aws.String("2026-03-15"), expected: ""},
		{start: aws.String("2026-03-01"), end: nil, expected: ""},
	}

	for _, tt := range tests {
		result := s.formatDayRange(tt.start, tt.end)
		assert.Equal(t, tt.expected, result)
	}
}

func TestGenerateReports(t *testing.T) {
	s := NewService()
	tempDir, err := os.MkdirTemp("", "aws-doctor-tests")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("CostComparisonReport", func(t *testing.T) {
		path := filepath.Join(tempDir, "cost.pdf")
		input := model.RenderCostComparisonInput{
			AccountID:        "123456789012",
			CurrentTotalCost: "100.00 USD",
			LastTotalCost:    "80.00 USD",
			CurrentMonth: &model.CostInfo{
				DateInterval: cetypes.DateInterval{Start: aws.String("2026-03-01"), End: aws.String("2026-03-31")},
				CostGroup: model.CostGroup{
					"Total": {Amount: 100.0, Unit: "USD"},
					"S3":    {Amount: 20.0, Unit: "USD"},
				},
			},
			LastMonth: &model.CostInfo{
				DateInterval: cetypes.DateInterval{Start: aws.String("2026-02-01"), End: aws.String("2026-02-28")},
				CostGroup: model.CostGroup{
					"Total": {Amount: 80.0, Unit: "USD"},
					"S3":    {Amount: 15.0, Unit: "USD"},
				},
			},
		}

		err := s.GenerateCostComparisonReport(input, path)
		assert.NoError(t, err)
		_, err = os.Stat(path)
		assert.NoError(t, err)
	})

	t.Run("WasteReport_WithAllCategories", func(t *testing.T) {
		path := filepath.Join(tempDir, "waste_full.pdf")
		input := model.RenderWasteInput{
			AccountID: "123456789012",
			UnusedVolumes: []ec2types.Volume{
				{
					VolumeId:         aws.String("vol-123"),
					Size:             aws.Int32(10),
					AvailabilityZone: aws.String("us-east-1a"),
					CreateTime:       aws.Time(time.Now()),
					VolumeType:       ec2types.VolumeTypeGp3,
					State:            ec2types.VolumeStateAvailable,
				},
			},
			StoppedVolumes: []ec2types.Volume{
				{
					VolumeId:         aws.String("vol-456"),
					Size:             aws.Int32(20),
					AvailabilityZone: aws.String("us-east-1b"),
					CreateTime:       aws.Time(time.Now()),
					VolumeType:       ec2types.VolumeTypeGp2,
					State:            ec2types.VolumeStateInUse,
				},
			},
			ElasticIPs: []ec2types.Address{
				{PublicIp: aws.String("1.2.3.4"), AllocationId: aws.String("alloc-123")},
			},
			StoppedInstances: []ec2types.Instance{
				{InstanceId: aws.String("i-123"), InstanceType: ec2types.InstanceTypeT2Micro, StateTransitionReason: aws.String("User initiated (2026-02-25 12:00:00 GMT)")},
			},
			Ris: []model.RiExpirationInfo{
				{ReservedInstanceID: "ri-123", DaysUntilExpiry: 5, Status: "active"},
				{ReservedInstanceID: "ri-456", DaysUntilExpiry: -10, Status: "expired"},
			},
			LoadBalancers: []elbtypes.LoadBalancer{
				{LoadBalancerName: aws.String("lb-123"), Type: elbtypes.LoadBalancerTypeEnumApplication, LoadBalancerArn: aws.String("arn:lb-123"), CreatedTime: aws.Time(time.Now())},
			},
			S3Buckets: []model.S3BucketWasteInfo{
				{BucketName: "bucket-123", CreationDate: time.Now(), Reason: "no lifecycle"},
			},
			S3MultipartUploads: []model.S3MultipartUploadWasteInfo{
				{BucketName: "bucket-456", UploadCount: 5},
			},
			CloudWatchLogGroups: []model.CloudWatchLogsWasteInfo{
				{LogGroupName: "lg-123", StoredBytes: 1024 * 1024 * 1024, EstimatedMonthlyCost: 0.50},
			},
			UnusedAMIs: []model.AMIWasteInfo{
				{ImageID: "ami-123", MaxPotentialSaving: 10.0, DaysSinceCreate: 50},
			},
			OrphanedSnapshots: []model.SnapshotWasteInfo{
				{SnapshotID: "snap-123", Category: model.SnapshotCategoryOrphaned, MaxPotentialSavings: 5.0, SizeGB: 10},
			},
			UnusedKeyPairs: []model.KeyPairWasteInfo{
				{KeyName: "key-123", DaysSinceCreate: 100},
			},
			RDSInstances: []model.RDSInstanceWasteInfo{
				{DBInstanceID: "rds-123", Engine: "postgres", EstimatedMonthlyCost: 20.0},
			},
			RDSSnapshots: []model.RDSSnapshotWasteInfo{
				{DBSnapshotID: "rds-snap-123", Engine: "mysql", EstimatedMonthlyCost: 5.0, SnapshotCreateTime: time.Now()},
			},
			RDSIdleInstances: []model.RDSIdleInstanceInfo{
				{DBInstanceID: "rds-idle-123", Engine: "aurora", EstimatedMonthlyCost: 30.0, DaysChecked: 7},
			},
		}

		err := s.GenerateWasteReport(input, path)
		assert.NoError(t, err)
		_, err = os.Stat(path)
		assert.NoError(t, err)
	})

	t.Run("WasteReport_Empty", func(t *testing.T) {
		path := filepath.Join(tempDir, "waste_empty.pdf")
		input := model.RenderWasteInput{
			AccountID: "123456789012",
		}

		err := s.GenerateWasteReport(input, path)
		assert.NoError(t, err)
		_, err = os.Stat(path)
		assert.NoError(t, err)
	})

	t.Run("TrendReport_NoServices", func(t *testing.T) {
		path := filepath.Join(tempDir, "trend_no_svc.pdf")
		costInfo := []model.CostInfo{
			{
				DateInterval: cetypes.DateInterval{Start: aws.String("2026-01-01"), End: aws.String("2026-01-31")},
				CostGroup:    model.CostGroup{"Total": {Amount: 50.0, Unit: "USD"}},
			},
		}

		err := s.GenerateTrendReport("123456789012", costInfo, []string{}, path)
		assert.NoError(t, err)
		_, err = os.Stat(path)
		assert.NoError(t, err)
	})

	t.Run("TrendReport_WithServices", func(t *testing.T) {
		path := filepath.Join(tempDir, "trend_with_svc.pdf")
		costInfo := []model.CostInfo{
			{
				DateInterval: cetypes.DateInterval{Start: aws.String("2026-01-01"), End: aws.String("2026-01-31")},
				CostGroup:    model.CostGroup{"Total": {Amount: 50.0, Unit: "USD"}, "S3": {Amount: 10.0, Unit: "USD"}},
			},
		}

		err := s.GenerateTrendReport("123456789012", costInfo, []string{"s3"}, path)
		assert.NoError(t, err)
		_, err = os.Stat(path)
		assert.NoError(t, err)
	})

	t.Run("GenerateAndSave_InvalidPath", func(t *testing.T) {
		path := "/nonexistent/directory/report.pdf"
		input := model.RenderCostComparisonInput{
			AccountID:        "123456789012",
			CurrentTotalCost: "100.00 USD",
			LastTotalCost:    "80.00 USD",
			CurrentMonth: &model.CostInfo{
				DateInterval: cetypes.DateInterval{Start: aws.String("2026-03-01"), End: aws.String("2026-03-31")},
				CostGroup:    model.CostGroup{"Total": {Amount: 100.0, Unit: "USD"}},
			},
			LastMonth: &model.CostInfo{
				DateInterval: cetypes.DateInterval{Start: aws.String("2026-02-01"), End: aws.String("2026-02-28")},
				CostGroup:    model.CostGroup{"Total": {Amount: 80.0, Unit: "USD"}},
			},
		}

		err := s.GenerateCostComparisonReport(input, path)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save PDF")
	})
}
