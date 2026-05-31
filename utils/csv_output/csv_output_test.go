package csvoutput

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/mocks/services"
	"github.com/elC0mpa/aws-doctor/model"
)

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer

	_, _ = io.Copy(&buf, r)

	return buf.String()
}

func TestOutputCostComparisonCSV(t *testing.T) {
	input := model.RenderCostComparisonInput{
		AccountID:        "123456789012",
		LastTotalCost:    "100.00 USD",
		CurrentTotalCost: "120.00 USD",
		LastMonth: &model.CostInfo{
			CostGroup: model.CostGroup{
				"Amazon EC2": {Amount: 100.0, Unit: "USD"},
			},
		},
		CurrentMonth: &model.CostInfo{
			CostGroup: model.CostGroup{
				"Amazon EC2": {Amount: 120.0, Unit: "USD"},
			},
		},
	}

	output := captureStdout(func() {
		_ = OutputCostComparisonCSV(input)
	})

	if !strings.Contains(output, "Service,Last Month,Current Month,Difference") {
		t.Error("Output missing headers")
	}

	if !strings.Contains(output, "Total Costs,100.00 USD,120.00 USD,20.00 USD") {
		t.Error("Output missing total row")
	}

	if !strings.Contains(output, "Amazon EC2,100.00 USD,120.00 USD,20.00 USD") {
		t.Error("Output missing service row")
	}
}

func TestOutputCostComparisonCSV_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input model.RenderCostComparisonInput
	}{
		{
			name: "LastMonthNil",
			input: model.RenderCostComparisonInput{
				LastMonth:    nil,
				CurrentMonth: &model.CostInfo{},
			},
		},
		{
			name: "CurrentMonthNil",
			input: model.RenderCostComparisonInput{
				LastMonth:    &model.CostInfo{},
				CurrentMonth: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := OutputCostComparisonCSV(tt.input)
			if err == nil {
				t.Error("Expected error but got nil")
			}
		})
	}
}

func TestOutputTrendCSV(t *testing.T) {
	costs := []model.CostInfo{
		{
			CostGroup: model.CostGroup{
				"Total": {Amount: 100.50, Unit: "USD"},
			},
		},
	}
	costs[0].Start = aws.String("2024-01-01")
	costs[0].End = aws.String("2024-01-31")

	t.Run("WithoutServices", func(t *testing.T) {
		output := captureStdout(func() {
			_ = OutputTrendCSV(costs, []string{})
		})

		if !strings.Contains(output, "Period Start,Period End,Total Cost,Unit") {
			t.Error("Output missing headers")
		}

		if !strings.Contains(output, "2024-01-01,2024-01-31,100.50,USD") {
			t.Error("Output missing trend row")
		}
	})

	t.Run("WithServices", func(t *testing.T) {
		output := captureStdout(func() {
			_ = OutputTrendCSV(costs, []string{"ec2"})
		})

		if !strings.Contains(output, "Period Start,Period End,Total Cost,Unit,Services") {
			t.Error("Output missing services header")
		}

		if !strings.Contains(output, "2024-01-01,2024-01-31,100.50,USD,ec2") {
			t.Error("Output missing services value")
		}
	})
}

func TestOutputWasteCSV(t *testing.T) {
	input := model.RenderWasteInput{
		AccountID: "123456789012",
		StoppedVolumes: []types.Volume{
			{VolumeId: aws.String("vol-123"), Size: aws.Int32(10), CreateTime: aws.Time(time.Now())},
		},
		ElasticIPs: []types.Address{
			{PublicIp: aws.String("1.2.3.4"), AllocationId: aws.String("eipalloc-123")},
		},
		S3Buckets: []model.S3BucketWasteInfo{
			{BucketName: "test-bucket", Reason: "No lifecycle policy", CreationDate: time.Now()},
		},
		S3MultipartUploads: []model.S3MultipartUploadWasteInfo{
			{BucketName: "multipart-bucket", UploadCount: 5},
		},
		StoppedInstances: []types.Instance{
			{InstanceId: aws.String("i-123"), StateTransitionReason: aws.String("User initiated (2024-01-01 00:00:00 UTC)")},
		},
		RDSInstances: []model.RDSInstanceWasteInfo{
			{DBInstanceID: "test-rds", MultiAZ: true, EstimatedMonthlyCost: 50.0},
		},
		LoadBalancers: []elbtypes.LoadBalancer{
			{LoadBalancerName: aws.String("test-lb"), Type: elbtypes.LoadBalancerTypeEnumApplication, CreatedTime: aws.Time(time.Now())},
		},
		OrphanedSnapshots: []model.SnapshotWasteInfo{
			{SnapshotID: "snap-orphaned", Category: model.SnapshotCategoryOrphaned, StartTime: time.Now()},
			{SnapshotID: "snap-stale", Category: model.SnapshotCategoryStale, StartTime: time.Now()},
		},
		UnusedKeyPairs: []model.KeyPairWasteInfo{
			{KeyName: "test-key", CreateTime: time.Now()},
		},
		UnusedIAMUsers: []model.IAMUserWasteInfo{
			{UserName: "test-csv-user", PasswordLastUsed: "Never", AccessKeysStatus: "Inactive"},
		},
		RootUserWaste: []model.IAMRootUserWasteInfo{
			{HasMFA: false},
		},
	}

	output := captureStdout(func() {
		_ = OutputWasteCSV(input, services.NewMockPricingService())
	})

	if !strings.Contains(output, "Resource Category,Resource Identifier,Estimated Monthly Cost (USD),Metric / Size,Age (Days),Additional Details") {
		t.Error("Output missing headers")
	}

	if !strings.Contains(output, "IAM User (Idle),test-csv-user") {
		t.Error("Output missing IAM user row")
	}

	if !strings.Contains(output, "IAM Root (No MFA),root") {
		t.Error("Output missing IAM root row")
	}

	if !strings.Contains(output, "EBS Volume (stopped),vol-123") {
		t.Error("Output missing EBS volume row")
	}

	if !strings.Contains(output, "Elastic IP,1.2.3.4") {
		t.Error("Output missing EIP row")
	}

	if !strings.Contains(output, "S3 Bucket (No lifecycle policy),test-bucket") {
		t.Error("Output missing S3 bucket row")
	}

	if !strings.Contains(output, "S3 Multipart Uploads,multipart-bucket") {
		t.Error("Output missing S3 multipart row")
	}

	if !strings.Contains(output, "Stopped RDS Instance,test-rds") {
		t.Error("Output missing RDS row")
	}

	if !strings.Contains(output, "Elastic Load Balancer") {
		t.Error("Output missing LB row")
	}

	if !strings.Contains(output, "EBS Snapshot (orphaned),snap-orphaned") {
		t.Error("Output missing orphaned snapshot row")
	}

	if !strings.Contains(output, "EBS Snapshot (stale),snap-stale") {
		t.Error("Output missing stale snapshot row")
	}

	if !strings.Contains(output, "Unused Key Pair,test-key") {
		t.Error("Output missing key pair row")
	}
}

// TestMapTotalRow tests the mapTotalRow function for CSV output
func TestMapTotalRow(t *testing.T) {
	tests := []struct {
		name         string
		lastTotal    string
		currentTotal string
		want         []string // ordered
	}{
		{
			name:         "positive_difference",
			lastTotal:    "100 USD",
			currentTotal: "120 USD",
			want:         []string{"Total Costs", "100.00 USD", "120.00 USD", "20.00 USD"},
		},
		{
			name:         "negative_difference",
			lastTotal:    "150 USD",
			currentTotal: "100 USD",
			want:         []string{"Total Costs", "150.00 USD", "100.00 USD", "-50.00 USD"},
		},
		{
			name:         "no_difference",
			lastTotal:    "100 USD",
			currentTotal: "100 USD",
			want:         []string{"Total Costs", "100.00 USD", "100.00 USD", "0.00 USD"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapTotalRow(tt.lastTotal, tt.currentTotal)

			if len(got) != len(tt.want) {
				t.Fatalf("mapTotalRow() returned %d elements, want %d", len(got), len(tt.want))
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("mapTotalRow()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestMapS3Buckets tests the mapS3Buckets function
func TestMapS3Buckets(t *testing.T) {
	tests := []struct {
		name    string
		buckets []model.S3BucketWasteInfo
		wantLen int
	}{
		{
			name:    "empty_slice",
			buckets: []model.S3BucketWasteInfo{},
			wantLen: 0,
		},
		{
			name:    "nil_slice",
			buckets: nil,
			wantLen: 0,
		},
		{
			name: "single_bucket",
			buckets: []model.S3BucketWasteInfo{
				{BucketName: "test-bucket-1", Reason: "No lifecycle policy"},
			},
			wantLen: 1,
		},
		{
			name: "multiple_buckets",
			buckets: []model.S3BucketWasteInfo{
				{BucketName: "test-bucket-1", Reason: "No lifecycle policy"},
				{BucketName: "test-bucket-2", Reason: "No lifecycle policy"},
				{BucketName: "test-bucket-3", Reason: "No lifecycle policy"},
			},
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapS3Buckets(tt.buckets)
			if len(got) != tt.wantLen {
				t.Errorf("mapS3Buckets() returned %d elements, want %d", len(got), tt.wantLen)
			}

			// Verify bucket name and reason are correctly mapped
			// got[i][1] = Identifier (BucketName), got[i][0] = Category (contains Reason)
			for i, bucket := range tt.buckets {
				if got[i][1] != bucket.BucketName {
					t.Errorf("mapS3Buckets()[%d][1] = %q, want %q", i, got[i][1], bucket.BucketName)
				}

				if got[i][0] != fmt.Sprintf("S3 Bucket (%s)", bucket.Reason) {
					t.Errorf("mapS3Buckets()[%d][0] = %q, want %q", i, got[i][0], fmt.Sprintf("S3 Bucket (%s)", bucket.Reason))
				}
			}
		})
	}
}

// TestMapS3MultipartUploads tests the mapS3MultipartUploads function
func TestMapS3MultipartUploads(t *testing.T) {
	tests := []struct {
		name    string
		uploads []model.S3MultipartUploadWasteInfo
		wantLen int
	}{
		{
			name:    "empty_slice",
			uploads: []model.S3MultipartUploadWasteInfo{},
			wantLen: 0,
		},
		{
			name:    "nil_slice",
			uploads: nil,
			wantLen: 0,
		},
		{
			name: "single_upload",
			uploads: []model.S3MultipartUploadWasteInfo{
				{BucketName: "test-bucket", UploadCount: 5},
			},
			wantLen: 1,
		},
		{
			name: "multiple_uploads",
			uploads: []model.S3MultipartUploadWasteInfo{
				{BucketName: "bucket-1", UploadCount: 3},
				{BucketName: "bucket-2", UploadCount: 7},
			},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapS3MultipartUploads(tt.uploads)
			if len(got) != tt.wantLen {
				t.Errorf("mapS3MultipartUploads() returned %d elements, want %d", len(got), tt.wantLen)
			}
		})
	}
}

// TestMapTrendRows tests the mapTrendRows function
func TestMapTrendRows(t *testing.T) {
	tests := []struct {
		name         string
		monthlyCosts []model.CostInfo
		services     []string
		wantLen      int
	}{
		{
			name:         "empty_costs",
			monthlyCosts: []model.CostInfo{},
			services:     []string{},
			wantLen:      0,
		},
		{
			name:         "nil_costs",
			monthlyCosts: nil,
			services:     []string{},
			wantLen:      0,
		},
		{
			name: "single_month",
			monthlyCosts: []model.CostInfo{
				{CostGroup: model.CostGroup{"Total": {Amount: 100.0, Unit: "USD"}}},
			},
			services: []string{},
			wantLen:  1,
		},
		{
			name: "six_months",
			monthlyCosts: []model.CostInfo{
				{CostGroup: model.CostGroup{"Total": {Amount: 100.0, Unit: "USD"}}},
				{CostGroup: model.CostGroup{"Total": {Amount: 110.0, Unit: "USD"}}},
				{CostGroup: model.CostGroup{"Total": {Amount: 120.0, Unit: "USD"}}},
				{CostGroup: model.CostGroup{"Total": {Amount: 130.0, Unit: "USD"}}},
				{CostGroup: model.CostGroup{"Total": {Amount: 140.0, Unit: "USD"}}},
				{CostGroup: model.CostGroup{"Total": {Amount: 150.0, Unit: "USD"}}},
			},
			services: []string{},
			wantLen:  6,
		},
		{
			name: "with_services_column",
			monthlyCosts: []model.CostInfo{
				{CostGroup: model.CostGroup{"Total": {Amount: 100.0, Unit: "USD"}}},
			},
			services: []string{"ec2", "s3"},
			wantLen:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapTrendRows(tt.monthlyCosts, tt.services)
			if len(got) != tt.wantLen {
				t.Errorf("mapTrendRows() returned %d rows, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func captureStderr(f func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	f()

	_ = w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	return buf.String()
}

func TestOutputWasteCSV_WithErrors(t *testing.T) {
	input := model.RenderWasteInput{
		Errors: map[string]string{
			"IAM": "Access Denied",
		},
	}

	mockPricing := new(services.MockPricingService)

	var stdout, stderr string

	stdout = captureStdout(func() {
		stderr = captureStderr(func() {
			_ = OutputWasteCSV(input, mockPricing)
		})
	})

	if !strings.Contains(stderr, "Warning: Error in IAM: Access Denied") {
		t.Errorf("Expected stderr to contain error warning, got: %s", stderr)
	}

	if !strings.Contains(stdout, "Resource Category,Resource Identifier,Estimated Monthly Cost (USD)") {
		t.Errorf("Expected stdout to contain CSV headers, got: %s", stdout)
	}
}
