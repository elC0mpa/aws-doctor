package csvoutput

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
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
	}

	output := captureStdout(func() {
		_ = OutputWasteCSV(input)
	})

	if !strings.Contains(output, "Resource Category,Resource Identifier,Estimated Monthly Cost (USD),Metric / Size,Age (Days),Additional Details") {
		t.Error("Output missing headers")
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
