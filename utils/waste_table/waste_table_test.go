package wastetable

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

// captureWasteOutput captures stdout during function execution
func captureWasteOutput(f func()) string {
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

func TestDrawWasteTable_NoWaste(t *testing.T) {
	output := captureWasteOutput(func() {
		DrawWasteTable(model.RenderWasteInput{AccountID: "123456789012"})
	})

	if !strings.Contains(output, "AWS DOCTOR CHECKUP") {
		t.Error("DrawWasteTable() missing header")
	}

	if !strings.Contains(output, "123456789012") {
		t.Error("DrawWasteTable() missing account ID")
	}

	if !strings.Contains(output, "healthy") || !strings.Contains(output, "No waste found") {
		t.Error("DrawWasteTable() with no waste should show healthy message")
	}
}

func TestDrawWasteTable_WithElasticIPs(t *testing.T) {
	elasticIPs := []types.Address{
		{PublicIp: aws.String("1.2.3.4"), AllocationId: aws.String("eipalloc-123")},
	}

	output := captureWasteOutput(func() {
		DrawWasteTable(model.RenderWasteInput{
			AccountID:  "123456789012",
			ElasticIPs: elasticIPs,
		})
	})

	if !strings.Contains(output, "Elastic IP") {
		t.Error("DrawWasteTable() with elastic IPs missing Elastic IP section")
	}
}

func TestDrawWasteTable_WithEBSVolumes(t *testing.T) {
	unusedVolumes := []types.Volume{
		{
			VolumeId:   aws.String("vol-123"),
			Size:       aws.Int32(100),
			CreateTime: aws.Time(time.Now()),
		},
	}

	output := captureWasteOutput(func() {
		DrawWasteTable(model.RenderWasteInput{
			AccountID:     "123456789012",
			UnusedVolumes: unusedVolumes,
		})
	})

	if !strings.Contains(output, "EBS") {
		t.Error("DrawWasteTable() with EBS volumes missing EBS section")
	}
}

func TestDrawWasteTable_WithStoppedInstances(t *testing.T) {
	stoppedInstances := []types.Instance{
		{
			InstanceId:            aws.String("i-123"),
			StateTransitionReason: aws.String("User initiated (2024-01-01 00:00:00 UTC)"),
		},
	}

	output := captureWasteOutput(func() {
		DrawWasteTable(model.RenderWasteInput{
			AccountID:        "123456789012",
			StoppedInstances: stoppedInstances,
		})
	})

	if !strings.Contains(output, "EC2") || !strings.Contains(output, "Reserved Instance") {
		t.Error("DrawWasteTable() with stopped instances missing EC2 section")
	}
}

func TestDrawWasteTable_WithReservedInstances(t *testing.T) {
	ris := []model.RiExpirationInfo{
		{
			ReservedInstanceID: "ri-123",
			DaysUntilExpiry:    15,
			Status:             "EXPIRING SOON",
		},
	}

	output := captureWasteOutput(func() {
		DrawWasteTable(model.RenderWasteInput{
			AccountID: "123456789012",
			Ris:       ris,
		})
	})

	if !strings.Contains(output, "Reserved Instance") {
		t.Error("DrawWasteTable() with reserved instances missing RI section")
	}
}

func TestDrawWasteTable_WithLoadBalancers(t *testing.T) {
	loadBalancers := []elbtypes.LoadBalancer{
		{
			LoadBalancerName: aws.String("my-alb"),
			Type:             elbtypes.LoadBalancerTypeEnumApplication,
			CreatedTime:      aws.Time(time.Now()),
		},
	}

	output := captureWasteOutput(func() {
		DrawWasteTable(model.RenderWasteInput{
			AccountID:     "123456789012",
			LoadBalancers: loadBalancers,
		})
	})

	if !strings.Contains(output, "Load Balancer") {
		t.Error("DrawWasteTable() with load balancers missing LB section")
	}
}

func TestDrawWasteTable_AllWasteTypes(t *testing.T) {
	elasticIPs := []types.Address{
		{PublicIp: aws.String("1.2.3.4"), AllocationId: aws.String("eipalloc-123")},
	}
	unusedVolumes := []types.Volume{
		{
			VolumeId:   aws.String("vol-123"),
			Size:       aws.Int32(100),
			CreateTime: aws.Time(time.Now()),
		},
	}
	stoppedVolumes := []types.Volume{
		{
			VolumeId:   aws.String("vol-456"),
			Size:       aws.Int32(200),
			CreateTime: aws.Time(time.Now()),
		},
	}
	ris := []model.RiExpirationInfo{
		{ReservedInstanceID: "ri-123", DaysUntilExpiry: 15, Status: "EXPIRING SOON"},
	}
	stoppedInstances := []types.Instance{
		{InstanceId: aws.String("i-123"), StateTransitionReason: aws.String("User initiated (2024-01-01 00:00:00 UTC)")},
	}
	loadBalancers := []elbtypes.LoadBalancer{
		{
			LoadBalancerName: aws.String("my-alb"),
			Type:             elbtypes.LoadBalancerTypeEnumApplication,
			CreatedTime:      aws.Time(time.Now()),
		},
	}

	output := captureWasteOutput(func() {
		DrawWasteTable(model.RenderWasteInput{
			AccountID:        "123456789012",
			ElasticIPs:       elasticIPs,
			UnusedVolumes:    unusedVolumes,
			StoppedVolumes:   stoppedVolumes,
			Ris:              ris,
			StoppedInstances: stoppedInstances,
			LoadBalancers:    loadBalancers,
		})
	})

	// Should have all sections
	if !strings.Contains(output, "EBS") {
		t.Error("Missing EBS section")
	}

	if !strings.Contains(output, "Elastic IP") {
		t.Error("Missing Elastic IP section")
	}

	if !strings.Contains(output, "EC2") {
		t.Error("Missing EC2 section")
	}

	if !strings.Contains(output, "Load Balancer") {
		t.Error("Missing Load Balancer section")
	}
}

func TestDrawEBSTable(t *testing.T) {
	unusedVolumes := []types.Volume{
		{
			VolumeId:   aws.String("vol-111"),
			Size:       aws.Int32(100),
			CreateTime: aws.Time(time.Now()),
		},
		{
			VolumeId:   aws.String("vol-222"),
			Size:       aws.Int32(200),
			CreateTime: aws.Time(time.Now()),
		},
	}
	stoppedVolumes := []types.Volume{
		{
			VolumeId:   aws.String("vol-333"),
			Size:       aws.Int32(300),
			CreateTime: aws.Time(time.Now()),
		},
	}

	output := captureWasteOutput(func() {
		drawEBSTable(unusedVolumes, stoppedVolumes)
	})

	if !strings.Contains(output, "EBS Volume Waste") {
		t.Error("drawEBSTable() missing title")
	}

	if !strings.Contains(output, "vol-111") {
		t.Error("drawEBSTable() missing unused volume ID")
	}

	if !strings.Contains(output, "vol-333") {
		t.Error("drawEBSTable() missing stopped volume ID")
	}
}

func TestDrawEBSTable_OnlyUnused(t *testing.T) {
	unusedVolumes := []types.Volume{
		{
			VolumeId:   aws.String("vol-111"),
			Size:       aws.Int32(100),
			CreateTime: aws.Time(time.Now()),
		},
	}

	output := captureWasteOutput(func() {
		drawEBSTable(unusedVolumes, nil)
	})

	if !strings.Contains(output, "Available") {
		t.Error("drawEBSTable() with only unused volumes missing Available status")
	}
}

func TestDrawEBSTable_OnlyStopped(t *testing.T) {
	stoppedVolumes := []types.Volume{
		{
			VolumeId:   aws.String("vol-333"),
			Size:       aws.Int32(300),
			CreateTime: aws.Time(time.Now()),
		},
	}

	output := captureWasteOutput(func() {
		drawEBSTable(nil, stoppedVolumes)
	})

	if !strings.Contains(output, "Stopped Instance") {
		t.Error("drawEBSTable() with only stopped volumes missing Stopped Instance status")
	}
}

func TestDrawEC2Table(t *testing.T) {
	instances := []types.Instance{
		{
			InstanceId:            aws.String("i-123"),
			StateTransitionReason: aws.String("User initiated (2024-01-01 00:00:00 UTC)"),
		},
	}
	ris := []model.RiExpirationInfo{
		{ReservedInstanceID: "ri-123", DaysUntilExpiry: 15, Status: "EXPIRING SOON"},
		{ReservedInstanceID: "ri-456", DaysUntilExpiry: -5, Status: "EXPIRED"},
	}

	output := captureWasteOutput(func() {
		drawEC2Table(instances, ris)
	})

	if !strings.Contains(output, "EC2 & Reserved Instance Waste") {
		t.Error("drawEC2Table() missing title")
	}

	if !strings.Contains(output, "i-123") {
		t.Error("drawEC2Table() missing instance ID")
	}

	if !strings.Contains(output, "ri-123") {
		t.Error("drawEC2Table() missing RI ID")
	}
}

func TestDrawEC2Table_OnlyInstances(t *testing.T) {
	instances := []types.Instance{
		{InstanceId: aws.String("i-123"), StateTransitionReason: aws.String("User initiated (2024-01-01 00:00:00 UTC)")},
	}

	output := captureWasteOutput(func() {
		drawEC2Table(instances, nil)
	})

	if !strings.Contains(output, "Stopped Instance") {
		t.Error("drawEC2Table() with only instances missing Stopped Instance status")
	}
}

func TestDrawEC2Table_OnlyRIs(t *testing.T) {
	ris := []model.RiExpirationInfo{
		{ReservedInstanceID: "ri-123", DaysUntilExpiry: 15, Status: "EXPIRING SOON"},
	}

	output := captureWasteOutput(func() {
		drawEC2Table(nil, ris)
	})

	if !strings.Contains(output, "Expiring Soon") {
		t.Error("drawEC2Table() with only expiring RIs missing Expiring Soon status")
	}
}

func TestDrawElasticIPTable(t *testing.T) {
	elasticIPs := []types.Address{
		{PublicIp: aws.String("1.2.3.4"), AllocationId: aws.String("eipalloc-123")},
		{PublicIp: aws.String("5.6.7.8"), AllocationId: aws.String("eipalloc-456")},
	}

	output := captureWasteOutput(func() {
		drawElasticIPTable(elasticIPs)
	})

	if !strings.Contains(output, "Elastic IP Waste") {
		t.Error("drawElasticIPTable() missing title")
	}

	if !strings.Contains(output, "1.2.3.4") {
		t.Error("drawElasticIPTable() missing IP address")
	}

	if !strings.Contains(output, "eipalloc-123") {
		t.Error("drawElasticIPTable() missing allocation ID")
	}
}

func TestDrawLoadBalancerTable(t *testing.T) {
	loadBalancers := []elbtypes.LoadBalancer{
		{
			LoadBalancerName: aws.String("my-alb"),
			Type:             elbtypes.LoadBalancerTypeEnumApplication,
			CreatedTime:      aws.Time(time.Now()),
		},
		{
			LoadBalancerName: aws.String("my-nlb"),
			Type:             elbtypes.LoadBalancerTypeEnumNetwork,
			CreatedTime:      aws.Time(time.Now()),
		},
	}

	output := captureWasteOutput(func() {
		drawLoadBalancerTable(loadBalancers, nil)
	})

	if !strings.Contains(output, "Load Balancer Waste") {
		t.Error("drawLoadBalancerTable() missing title")
	}

	if !strings.Contains(output, "my-alb") {
		t.Error("drawLoadBalancerTable() missing ALB name")
	}

	if !strings.Contains(output, "application") {
		t.Error("drawLoadBalancerTable() missing ALB type")
	}
}

func TestDrawSnapshotTable(t *testing.T) {
	snapshots := []model.SnapshotWasteInfo{
		{
			SnapshotID: "snap-123",
			Category:   model.SnapshotCategoryOrphaned,
			Reason:     "Volume deleted",
			SizeGB:     10,
			StartTime:  time.Now(),
		},
		{
			SnapshotID: "snap-456",
			Category:   model.SnapshotCategoryStale,
			Reason:     "Old backup",
			SizeGB:     20,
			StartTime:  time.Now(),
		},
	}

	output := captureWasteOutput(func() {
		drawSnapshotTable(snapshots)
	})

	if !strings.Contains(output, "EBS Snapshot Waste") {
		t.Error("drawSnapshotTable() missing title")
	}

	if !strings.Contains(output, "snap-123") {
		t.Error("drawSnapshotTable() missing orphaned snapshot ID")
	}

	if !strings.Contains(output, "snap-456") {
		t.Error("drawSnapshotTable() missing stale snapshot ID")
	}
}

func TestDrawWasteTable_IndividualResources(t *testing.T) {
	accountID := "123456789012"

	// Test EBS only
	unusedVolumes := []types.Volume{
		{
			VolumeId:   aws.String("vol-1"),
			Size:       aws.Int32(10),
			CreateTime: aws.Time(time.Now()),
		},
	}

	output := captureWasteOutput(func() {
		DrawWasteTable(model.RenderWasteInput{
			AccountID:     accountID,
			UnusedVolumes: unusedVolumes,
		})
	})

	if !strings.Contains(output, "EBS Volume Waste") {
		t.Error("DrawWasteTable with EBS only missing title")
	}

	// Test LoadBalancer only
	lbs := []elbtypes.LoadBalancer{
		{
			LoadBalancerName: aws.String("lb-1"),
			Type:             elbtypes.LoadBalancerTypeEnumApplication,
			CreatedTime:      aws.Time(time.Now()),
		},
	}

	output = captureWasteOutput(func() {
		DrawWasteTable(model.RenderWasteInput{
			AccountID:     accountID,
			LoadBalancers: lbs,
		})
	})

	if !strings.Contains(output, "Load Balancer Waste") {
		t.Error("DrawWasteTable with LB only missing title")
	}

	// Test AMIs only
	amis := []model.AMIWasteInfo{{ImageID: "ami-1", Name: "ami-1", DaysSinceCreate: 10, CreationDate: time.Now()}}

	output = captureWasteOutput(func() {
		DrawWasteTable(model.RenderWasteInput{
			AccountID:  accountID,
			UnusedAMIs: amis,
		})
	})

	if !strings.Contains(output, "Unused AMI Waste") {
		t.Error("DrawWasteTable with AMIs only missing title")
	}

	// Test Snapshots only
	snaps := []model.SnapshotWasteInfo{{SnapshotID: "snap-1", Category: model.SnapshotCategoryOrphaned, StartTime: time.Now()}}

	output = captureWasteOutput(func() {
		DrawWasteTable(model.RenderWasteInput{
			AccountID:         accountID,
			OrphanedSnapshots: snaps,
		})
	})

	if !strings.Contains(output, "EBS Snapshot Waste") {
		t.Error("DrawWasteTable with Snapshots only missing title")
	}
}

func TestDrawAMITable(t *testing.T) {
	amis := []model.AMIWasteInfo{
		{
			ImageID:            "ami-12345",
			Name:               "my-test-ami",
			DaysSinceCreate:    60,
			MaxPotentialSaving: 5.00,
			SafetyWarning:      "Verify before deleting",
			CreationDate:       time.Now(),
		},
		{
			ImageID:            "ami-67890",
			Name:               "another-ami",
			DaysSinceCreate:    90,
			MaxPotentialSaving: 7.50,
			SafetyWarning:      "Verify before deleting",
			CreationDate:       time.Now(),
		},
	}

	output := captureWasteOutput(func() {
		drawAMITable(amis)
	})

	// Check for table title
	if !strings.Contains(output, "Unused AMI Waste") {
		t.Error("drawAMITable() missing title")
	}

	// Check for AMI IDs
	if !strings.Contains(output, "ami-12345") {
		t.Error("drawAMITable() missing first AMI ID")
	}

	if !strings.Contains(output, "ami-67890") {
		t.Error("drawAMITable() missing second AMI ID")
	}

	// Check for warning message
	if !strings.Contains(output, "Warning") || !strings.Contains(output, "Auto Scaling") {
		t.Error("drawAMITable() missing safety warning footer")
	}
}

func TestDrawWasteTable_WithUnusedAMIs(t *testing.T) {
	unusedAMIs := []model.AMIWasteInfo{
		{
			ImageID:            "ami-waste123",
			Name:               "unused-ami",
			DaysSinceCreate:    120,
			MaxPotentialSaving: 10.00,
			SafetyWarning:      "Verify before deleting",
			CreationDate:       time.Now(),
		},
	}

	output := captureWasteOutput(func() {
		DrawWasteTable(model.RenderWasteInput{
			AccountID:  "123456789012",
			UnusedAMIs: unusedAMIs,
		})
	})

	if !strings.Contains(output, "Unused AMI") {
		t.Error("DrawWasteTable() with unused AMIs missing AMI section")
	}

	if !strings.Contains(output, "ami-waste123") {
		t.Error("DrawWasteTable() with unused AMIs missing AMI ID")
	}
}

func TestDrawKeyPairTable(t *testing.T) {
	keyPairs := []model.KeyPairWasteInfo{
		{
			KeyName:         "unused-key",
			KeyPairID:       "key-abcde",
			DaysSinceCreate: 60,
			CreateTime:      time.Now(),
		},
	}

	output := captureWasteOutput(func() {
		drawKeyPairTable(keyPairs)
	})

	if !strings.Contains(output, "Unused EC2 Key Pair Waste") {
		t.Error("drawKeyPairTable() missing title")
	}

	if !strings.Contains(output, "unused-key") {
		t.Error("drawKeyPairTable() missing key name")
	}

	if !strings.Contains(output, "key-abcde") {
		t.Error("drawKeyPairTable() missing key ID")
	}
}

func TestDrawWasteTable_WithKeyPairs(t *testing.T) {
	keyPairs := []model.KeyPairWasteInfo{
		{
			KeyName:         "waste-key",
			KeyPairID:       "key-waste",
			DaysSinceCreate: 45,
			CreateTime:      time.Now(),
		},
	}

	output := captureWasteOutput(func() {
		DrawWasteTable(model.RenderWasteInput{
			AccountID:      "123456789012",
			UnusedKeyPairs: keyPairs,
		})
	})

	if !strings.Contains(output, "Key Pair Waste") {
		t.Error("DrawWasteTable() with key pairs missing Key Pair section")
	}

	if !strings.Contains(output, "waste-key") {
		t.Error("DrawWasteTable() with key pairs missing key name")
	}
}

func TestDrawS3Table(t *testing.T) {
	buckets := []model.S3BucketWasteInfo{
		{
			BucketName:   "waste-bucket",
			Reason:       "No lifecycle policy",
			CreationDate: time.Now(),
		},
	}

	multipart := []model.S3MultipartUploadWasteInfo{
		{
			BucketName:  "multipart-bucket",
			UploadCount: 5,
		},
	}

	output := captureWasteOutput(func() {
		drawS3Table(buckets, multipart)
	})

	if !strings.Contains(output, "S3 Bucket Waste") {
		t.Error("drawS3Table() missing title")
	}

	if !strings.Contains(output, "No Lifecycle Policy") {
		t.Error("drawS3Table() missing lifecycle status")
	}

	if !strings.Contains(output, "Incomplete Multipart") {
		t.Error("drawS3Table() missing multipart status")
	}

	if !strings.Contains(output, "waste-bucket") {
		t.Error("drawS3Table() missing bucket name")
	}

	if !strings.Contains(output, "multipart-bucket") {
		t.Error("drawS3Table() missing multipart bucket name")
	}
}

func TestDrawWasteTable_WithS3Buckets(t *testing.T) {
	buckets := []model.S3BucketWasteInfo{
		{
			BucketName:   "s3-waste-bucket",
			Reason:       "No lifecycle policy",
			CreationDate: time.Now(),
		},
	}

	output := captureWasteOutput(func() {
		DrawWasteTable(model.RenderWasteInput{
			AccountID: "123456789012",
			S3Buckets: buckets,
		})
	})

	if !strings.Contains(output, "S3 Bucket Waste") {
		t.Error("DrawWasteTable() with S3 buckets missing S3 section")
	}

	if !strings.Contains(output, "s3-waste-bucket") {
		t.Error("DrawWasteTable() with S3 buckets missing bucket name")
	}
}

func TestDrawWasteTable_WithS3Multipart(t *testing.T) {
	buckets := []model.S3MultipartUploadWasteInfo{
		{
			BucketName:  "s3-multipart-waste-bucket",
			UploadCount: 3,
		},
	}

	output := captureWasteOutput(func() {
		DrawWasteTable(model.RenderWasteInput{
			AccountID:          "123456789012",
			S3MultipartUploads: buckets,
		})
	})

	if !strings.Contains(output, "S3 Bucket Waste") {
		t.Error("DrawWasteTable() with S3 multipart missing S3 section")
	}

	if !strings.Contains(output, "Incomplete Multipart") {
		t.Error("DrawWasteTable() with S3 multipart missing multipart status")
	}

	if !strings.Contains(output, "s3-multipart-waste-bucket") {
		t.Error("DrawWasteTable() with S3 multipart missing bucket name")
	}
}

func TestDrawCloudWatchLogsTable(t *testing.T) {
	logGroups := []model.CloudWatchLogsWasteInfo{
		{
			LogGroupName: "waste-loggroup",
			StoredBytes:  2048,
			CreationTime: time.Now(),
		},
	}

	output := captureWasteOutput(func() {
		drawCloudWatchLogsTable(logGroups)
	})

	if !strings.Contains(output, "CloudWatch Log Group Waste") {
		t.Error("drawCloudWatchLogsTable() missing title")
	}

	if !strings.Contains(output, "No Retention Policy") {
		t.Error("drawCloudWatchLogsTable() missing retention status")
	}

	if !strings.Contains(output, "waste-loggroup") {
		t.Error("drawCloudWatchLogsTable() missing log group name")
	}
}

func TestDrawWasteTable_WithCloudWatchLogs(t *testing.T) {
	logGroups := []model.CloudWatchLogsWasteInfo{
		{
			LogGroupName: "cw-waste-loggroup",
			StoredBytes:  512,
			CreationTime: time.Now(),
		},
	}

	output := captureWasteOutput(func() {
		DrawWasteTable(model.RenderWasteInput{
			AccountID:           "123456789012",
			CloudWatchLogGroups: logGroups,
		})
	})

	if !strings.Contains(output, "CloudWatch Log Group Waste") {
		t.Error("DrawWasteTable() with CloudWatch logs missing CloudWatch section")
	}

	if !strings.Contains(output, "cw-waste-loggroup") {
		t.Error("DrawWasteTable() with CloudWatch logs missing log group name")
	}
}

func TestDrawRDSTable(t *testing.T) {
	instances := []model.RDSInstanceWasteInfo{
		{DBInstanceID: "stopped-rds", Engine: "mysql", Status: "stopped", MultiAZ: true, EstimatedMonthlyCost: 10.0},
	}
	snapshots := []model.RDSSnapshotWasteInfo{
		{DBSnapshotID: "old-snap", Engine: "postgres", AllocatedStorage: 20, SnapshotCreateTime: time.Now()},
	}
	idleInstances := []model.RDSIdleInstanceInfo{
		{DBInstanceID: "idle-rds", Engine: "aurora", DaysChecked: 7, EstimatedMonthlyCost: 50.0},
	}

	output := captureWasteOutput(func() {
		drawRDSTable(instances, snapshots, idleInstances)
	})

	if !strings.Contains(output, "RDS Waste") {
		t.Error("drawRDSTable() missing title")
	}

	if !strings.Contains(output, "stopped-rds") {
		t.Error("drawRDSTable() missing stopped instance ID")
	}

	if !strings.Contains(output, "old-snap") {
		t.Error("drawRDSTable() missing snapshot ID")
	}

	if !strings.Contains(output, "idle-rds") {
		t.Error("drawRDSTable() missing idle instance ID")
	}
}

// TestHasAnyWaste tests the hasAnyWaste function that checks if any waste data exists
func TestHasAnyWaste(t *testing.T) {
	tests := []struct {
		name  string
		input model.RenderWasteInput
		want  bool
	}{
		{
			name:  "empty_input_returns_false",
			input: model.RenderWasteInput{},
			want:  false,
		},
		{
			name: "only_elastic_ips_set_returns_true",
			input: model.RenderWasteInput{
				ElasticIPs: []types.Address{
					{PublicIp: aws.String("1.2.3.4"), AllocationId: aws.String("eipalloc-123")},
				},
			},
			want: true,
		},
		{
			name: "only_unused_volumes_set_returns_true",
			input: model.RenderWasteInput{
				UnusedVolumes: []types.Volume{
					{VolumeId: aws.String("vol-123")},
				},
			},
			want: true,
		},
		{
			name: "only_stopped_volumes_set_returns_true",
			input: model.RenderWasteInput{
				StoppedVolumes: []types.Volume{
					{VolumeId: aws.String("vol-456")},
				},
			},
			want: true,
		},
		{
			name: "only_stopped_instances_set_returns_true",
			input: model.RenderWasteInput{
				StoppedInstances: []types.Instance{
					{InstanceId: aws.String("i-123")},
				},
			},
			want: true,
		},
		{
			name: "only_reserved_instances_set_returns_true",
			input: model.RenderWasteInput{
				Ris: []model.RiExpirationInfo{
					{ReservedInstanceID: "ri-123"},
				},
			},
			want: true,
		},
		{
			name: "only_unused_am_is_set_returns_true",
			input: model.RenderWasteInput{
				UnusedAMIs: []model.AMIWasteInfo{
					{ImageID: "ami-123"},
				},
			},
			want: true,
		},
		{
			name: "only_orphaned_snapshots_set_returns_true",
			input: model.RenderWasteInput{
				OrphanedSnapshots: []model.SnapshotWasteInfo{
					{SnapshotID: "snap-123", Category: model.SnapshotCategoryOrphaned},
				},
			},
			want: true,
		},
		{
			name: "only_unused_key_pairs_set_returns_true",
			input: model.RenderWasteInput{
				UnusedKeyPairs: []model.KeyPairWasteInfo{
					{KeyName: "test-key"},
				},
			},
			want: true,
		},
		{
			name: "only_s3_buckets_set_returns_true",
			input: model.RenderWasteInput{
				S3Buckets: []model.S3BucketWasteInfo{
					{BucketName: "test-bucket"},
				},
			},
			want: true,
		},
		{
			name: "only_s3_multipart_uploads_set_returns_true",
			input: model.RenderWasteInput{
				S3MultipartUploads: []model.S3MultipartUploadWasteInfo{
					{BucketName: "test-bucket", UploadCount: 5},
				},
			},
			want: true,
		},
		{
			name: "only_cloudwatch_log_groups_set_returns_true",
			input: model.RenderWasteInput{
				CloudWatchLogGroups: []model.CloudWatchLogsWasteInfo{
					{LogGroupName: "/aws/lambda/test"},
				},
			},
			want: true,
		},
		{
			name: "only_rds_instances_set_returns_true",
			input: model.RenderWasteInput{
				RDSInstances: []model.RDSInstanceWasteInfo{
					{DBInstanceID: "test-rds"},
				},
			},
			want: true,
		},
		{
			name: "only_rds_snapshots_set_returns_true",
			input: model.RenderWasteInput{
				RDSSnapshots: []model.RDSSnapshotWasteInfo{
					{DBSnapshotID: "rds:snap-123"},
				},
			},
			want: true,
		},
		{
			name: "only_rds_idle_instances_set_returns_true",
			input: model.RenderWasteInput{
				RDSIdleInstances: []model.RDSIdleInstanceInfo{
					{DBInstanceID: "test-rds-idle"},
				},
			},
			want: true,
		},
		{
			name: "only_idle_nat_gateways_set_returns_true",
			input: model.RenderWasteInput{
				IdleNATGateways: []model.NATGatewayWasteInfo{
					{NATGatewayID: "nat-123"},
				},
			},
			want: true,
		},
		{
			name: "only_idle_load_balancers_set_returns_true",
			input: model.RenderWasteInput{
				IdleLoadBalancers: []model.ELBIdleInfo{
					{Name: "lb-123"},
				},
			},
			want: true,
		},
		{
			name: "only_over_provisioned_lambdas_set_returns_true",
			input: model.RenderWasteInput{
				OverProvisionedLambdas: []model.LambdaOverProvisionedInfo{
					{FunctionName: "test-lambda"},
				},
			},
			want: true,
		},
		{
			name: "mix_of_fields_returns_true",
			input: model.RenderWasteInput{
				ElasticIPs: []types.Address{
					{PublicIp: aws.String("1.2.3.4")},
				},
				StoppedInstances: []types.Instance{
					{InstanceId: aws.String("i-123")},
				},
				S3Buckets: []model.S3BucketWasteInfo{
					{BucketName: "test-bucket"},
				},
			},
			want: true,
		},
		{
			name: "all_fields_empty_returns_false",
			input: model.RenderWasteInput{
				ElasticIPs:             []types.Address{},
				UnusedVolumes:          []types.Volume{},
				StoppedVolumes:         []types.Volume{},
				StoppedInstances:       []types.Instance{},
				Ris:                    []model.RiExpirationInfo{},
				LoadBalancers:          []elbtypes.LoadBalancer{},
				UnusedAMIs:             []model.AMIWasteInfo{},
				OrphanedSnapshots:      []model.SnapshotWasteInfo{},
				UnusedKeyPairs:         []model.KeyPairWasteInfo{},
				S3Buckets:              []model.S3BucketWasteInfo{},
				S3MultipartUploads:     []model.S3MultipartUploadWasteInfo{},
				CloudWatchLogGroups:    []model.CloudWatchLogsWasteInfo{},
				RDSInstances:           []model.RDSInstanceWasteInfo{},
				RDSSnapshots:           []model.RDSSnapshotWasteInfo{},
				RDSIdleInstances:       []model.RDSIdleInstanceInfo{},
				IdleNATGateways:        []model.NATGatewayWasteInfo{},
				IdleLoadBalancers:      []model.ELBIdleInfo{},
				OverProvisionedLambdas: []model.LambdaOverProvisionedInfo{},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasAnyWaste(tt.input)
			if got != tt.want {
				t.Errorf("hasAnyWaste() = %v, want %v", got, tt.want)
			}
		})
	}
}
