package outputshared

import (
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
)

func TestPresentS3Bucket(t *testing.T) {
	b := model.S3BucketWasteInfo{
		BucketName:   "test-bucket",
		Reason:       "No lifecycle policy",
		CreationDate: time.Now(),
	}

	p := PresentS3Bucket(b)

	if p.Identifier != "test-bucket" {
		t.Errorf("Identifier = %v, want 'test-bucket'", p.Identifier)
	}

	if !strings.Contains(p.Category, "No lifecycle policy") {
		t.Errorf("Category %q missing reason", p.Category)
	}
}

func TestPresentS3MultipartUpload(t *testing.T) {
	b := model.S3MultipartUploadWasteInfo{
		BucketName:  "test-bucket",
		UploadCount: 10,
	}

	p := PresentS3MultipartUpload(b)

	if p.Identifier != "test-bucket" {
		t.Errorf("Identifier = %v, want 'test-bucket'", p.Identifier)
	}

	if !strings.Contains(p.Metric, "10") {
		t.Errorf("Metric %q missing count", p.Metric)
	}
}

func TestPresentEBSVolume(t *testing.T) {
	tests := []struct {
		name           string
		vol            types.Volume
		status         string
		wantIdentifier string
	}{
		{
			name: "unattached_volume",
			vol: types.Volume{
				VolumeId:   aws.String("vol-12345"),
				Size:       aws.Int32(100),
				State:      types.VolumeStateAvailable,
				CreateTime: aws.Time(time.Now()),
			},
			status:         "unattached",
			wantIdentifier: "vol-12345",
		},
		{
			name: "attached_to_stopped_instance",
			vol: types.Volume{
				VolumeId:   aws.String("vol-67890"),
				Size:       aws.Int32(50),
				State:      types.VolumeStateInUse,
				CreateTime: aws.Time(time.Now()),
				Attachments: []types.VolumeAttachment{
					{InstanceId: aws.String("i-abcdef")},
				},
			},
			status:         "stopped",
			wantIdentifier: "vol-67890 (i-abcdef)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PresentEBSVolume(tt.vol, tt.status)

			if p.Identifier != tt.wantIdentifier {
				t.Errorf("Identifier = %v, want %v", p.Identifier, tt.wantIdentifier)
			}

			if !strings.HasPrefix(p.EstimatedCost, "$") {
				t.Errorf("EstimatedCost %q does not start with $", p.EstimatedCost)
			}
		})
	}
}

func TestAttachedInstanceID(t *testing.T) {
	tests := []struct {
		name string
		vol  types.Volume
		want string
	}{
		{
			name: "no_attachments",
			vol:  types.Volume{},
			want: NAValue,
		},
		{
			name: "with_attachment",
			vol: types.Volume{
				Attachments: []types.VolumeAttachment{
					{InstanceId: aws.String("i-12345")},
				},
			},
			want: "i-12345",
		},
		{
			name: "nil_instance_id",
			vol: types.Volume{
				Attachments: []types.VolumeAttachment{
					{InstanceId: nil},
				},
			},
			want: NAValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AttachedInstanceID(tt.vol)
			if got != tt.want {
				t.Errorf("AttachedInstanceID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPresentElasticIP(t *testing.T) {
	ip := types.Address{
		PublicIp:     aws.String("1.2.3.4"),
		AllocationId: aws.String("eipalloc-12345"),
	}

	p := PresentElasticIP(ip)

	if p.Identifier != "1.2.3.4" {
		t.Errorf("Identifier = %v, want '1.2.3.4'", p.Identifier)
	}

	if !strings.HasPrefix(p.EstimatedCost, "$") {
		t.Errorf("EstimatedCost %q does not start with $", p.EstimatedCost)
	}

	if !strings.Contains(p.Details, "eipalloc-12345") {
		t.Errorf("Details %q does not contain allocation ID", p.Details)
	}
}

func TestPresentStoppedInstance(t *testing.T) {
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30).Format("2006-01-02 15:04:05") + " UTC"

	instance := types.Instance{
		InstanceId:            aws.String("i-12345"),
		InstanceType:          types.InstanceTypeT3Micro,
		StateTransitionReason: aws.String("User initiated (" + thirtyDaysAgo + ")"),
	}

	p := PresentStoppedInstance(instance)

	if p.Identifier != "i-12345" {
		t.Errorf("Identifier = %v, want 'i-12345'", p.Identifier)
	}

	if p.Age == NAValue {
		t.Error("Age was not calculated")
	}

	if p.Metric != "t3.micro" {
		t.Errorf("Metric = %v, want 't3.micro'", p.Metric)
	}
}

func TestPresentReservedInstance(t *testing.T) {
	ri := model.RiExpirationInfo{
		ReservedInstanceID: "ri-12345",
		InstanceType:       "t3.micro",
		DaysUntilExpiry:    15,
		Status:             "EXPIRING SOON",
		State:              "active",
	}

	p := PresentReservedInstance(ri)

	if p.Identifier != "ri-12345" {
		t.Errorf("Identifier = %v, want 'ri-12345'", p.Identifier)
	}

	if p.Age != "15" {
		t.Errorf("Age = %v, want '15'", p.Age)
	}

	if p.Metric != "t3.micro" {
		t.Errorf("Metric = %v, want 't3.micro'", p.Metric)
	}
}

func TestPresentLoadBalancer(t *testing.T) {
	lb := elbtypes.LoadBalancer{
		LoadBalancerArn: aws.String("arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-load-balancer/50dc6c495c0c9188"),
		Type:            elbtypes.LoadBalancerTypeEnumApplication,
		CreatedTime:     aws.Time(time.Now()),
	}

	p := PresentLoadBalancer(lb)

	if p.Identifier != *lb.LoadBalancerArn {
		t.Errorf("Identifier = %v, want %v", p.Identifier, *lb.LoadBalancerArn)
	}

	if p.Metric != "application" {
		t.Errorf("Metric = %v, want 'application'", p.Metric)
	}

	if !strings.HasPrefix(p.EstimatedCost, "$") {
		t.Errorf("EstimatedCost %q does not start with $", p.EstimatedCost)
	}
}

func TestPresentAMI(t *testing.T) {
	ami := model.AMIWasteInfo{
		ImageID:         "ami-12345",
		Name:            "my-ami",
		DaysSinceCreate: 90,
		CreationDate:    time.Now(),
	}

	p := PresentAMI(ami)

	if p.Identifier != "ami-12345" {
		t.Errorf("Identifier = %v, want 'ami-12345'", p.Identifier)
	}

	if p.Age != "90" {
		t.Errorf("Age = %v, want '90'", p.Age)
	}
}

func TestPresentSnapshot(t *testing.T) {
	snap := model.SnapshotWasteInfo{
		SnapshotID:          "snap-12345",
		Category:            model.SnapshotCategoryOrphaned,
		SizeGB:              10,
		DaysSinceCreate:     30,
		MaxPotentialSavings: 5.00,
		StartTime:           time.Now(),
	}

	p := PresentSnapshot(snap)

	if p.Identifier != "snap-12345" {
		t.Errorf("Identifier = %v, want 'snap-12345'", p.Identifier)
	}

	if p.Metric != "10 GiB" {
		t.Errorf("Metric = %v, want '10 GiB'", p.Metric)
	}

	if !strings.HasPrefix(p.EstimatedCost, "$") {
		t.Errorf("EstimatedCost %q does not start with $", p.EstimatedCost)
	}
}

func TestPresentKeyPair(t *testing.T) {
	kp := model.KeyPairWasteInfo{
		KeyName:         "test-key",
		DaysSinceCreate: 60,
		CreateTime:      time.Now(),
	}

	p := PresentKeyPair(kp)

	if p.Identifier != "test-key" {
		t.Errorf("Identifier = %v, want 'test-key'", p.Identifier)
	}

	if p.Age != "60" {
		t.Errorf("Age = %v, want '60'", p.Age)
	}
}

func TestPresentCloudWatchLogGroup(t *testing.T) {
	lg := model.CloudWatchLogsWasteInfo{
		LogGroupName:         "test-loggroup",
		StoredBytes:          1024 * 1024 * 1024, // 1 GB
		EstimatedMonthlyCost: 0.50,
		CreationTime:         time.Now(),
	}

	p := PresentCloudWatchLogGroup(lg)

	if p.Identifier != "test-loggroup" {
		t.Errorf("Identifier = %v, want 'test-loggroup'", p.Identifier)
	}

	if p.Metric != "1.00 GB stored" {
		t.Errorf("Metric = %v, want '1.00 GB stored'", p.Metric)
	}

	if p.EstimatedCost != "$0.50" {
		t.Errorf("EstimatedCost = %v, want '$0.50'", p.EstimatedCost)
	}
}

func TestPresentRDSInstance(t *testing.T) {
	inst := model.RDSInstanceWasteInfo{
		DBInstanceID:         "test-db",
		Engine:               "mysql",
		Status:               "stopped",
		MultiAZ:              true,
		EstimatedMonthlyCost: 10.00,
	}

	p := PresentRDSInstance(inst)

	if p.Identifier != "test-db" {
		t.Errorf("Identifier = %v, want 'test-db'", p.Identifier)
	}

	if p.Metric != "Is Multi AZ: true" {
		t.Errorf("Metric = %v, want 'Is Multi AZ: true'", p.Metric)
	}

	if p.EstimatedCost != "$10.00" {
		t.Errorf("EstimatedCost = %v, want '$10.00'", p.EstimatedCost)
	}
}

func TestPresentRDSSnapshot(t *testing.T) {
	snap := model.RDSSnapshotWasteInfo{
		DBSnapshotID:         "test-snap",
		AllocatedStorage:     20,
		DaysSinceCreate:      45,
		Engine:               "postgres",
		EstimatedMonthlyCost: 2.00,
		SnapshotCreateTime:   time.Now(),
	}

	p := PresentRDSSnapshot(snap)

	if p.Identifier != "test-snap" {
		t.Errorf("Identifier = %v, want 'test-snap'", p.Identifier)
	}

	if p.Age != "45" {
		t.Errorf("Age = %v, want '45'", p.Age)
	}

	if p.Metric != "20 GiB" {
		t.Errorf("Metric = %v, want '20 GiB'", p.Metric)
	}
}

func TestPresentRDSIdleInstance(t *testing.T) {
	inst := model.RDSIdleInstanceInfo{
		DBInstanceID:         "idle-db",
		Engine:               "aurora",
		DaysChecked:          7,
		EstimatedMonthlyCost: 50.00,
	}

	p := PresentRDSIdleInstance(inst)

	if p.Identifier != "idle-db" {
		t.Errorf("Identifier = %v, want 'idle-db'", p.Identifier)
	}

	if !strings.Contains(p.Metric, "7 days") {
		t.Errorf("Metric %q does not contain '7 days'", p.Metric)
	}

	if p.EstimatedCost != "$50.00" {
		t.Errorf("EstimatedCost = %v, want '$50.00'", p.EstimatedCost)
	}
}
