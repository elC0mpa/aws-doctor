package outputshared

import (
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/mocks/services"
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
		name      string
		vol       types.Volume
		status    string
		wantCost  string
		mockValue float64
	}{
		{
			name: "gp2 volume",
			vol: types.Volume{
				VolumeId:   aws.String("vol-1"),
				Size:       aws.Int32(100),
				VolumeType: types.VolumeTypeGp2,
				State:      types.VolumeStateAvailable,
				CreateTime: aws.Time(time.Now()),
			},
			status:    "unattached",
			wantCost:  "$10.00",
			mockValue: 10.0,
		},
		{
			name: "gp3 volume",
			vol: types.Volume{
				VolumeId:   aws.String("vol-2"),
				Size:       aws.Int32(50),
				VolumeType: types.VolumeTypeGp3,
				State:      types.VolumeStateInUse,
				CreateTime: aws.Time(time.Now()),
			},
			status:    "stopped",
			wantCost:  "$4.00",
			mockValue: 4.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := new(services.MockPricingService)
			m.On("CalculateEBSMonthlyCost", *tt.vol.Size, tt.vol.VolumeType).Return(tt.mockValue)

			p := PresentEBSVolume(tt.vol, tt.status, m)

			if p.EstimatedCost != tt.wantCost {
				t.Errorf("EstimatedCost = %v, want %v", p.EstimatedCost, tt.wantCost)
			}

			m.AssertExpectations(t)
		})
	}
}

func TestPresentElasticIP(t *testing.T) {
	ip := types.Address{
		PublicIp:     aws.String("1.2.3.4"),
		AllocationId: aws.String("eipalloc-1"),
	}

	m := new(services.MockPricingService)
	m.On("CalculateEIPMonthlyCost").Return(3.65)

	p := PresentElasticIP(ip, m)

	if p.Identifier != "1.2.3.4" {
		t.Errorf("Identifier = %v, want '1.2.3.4'", p.Identifier)
	}

	if p.EstimatedCost != "$3.65" {
		t.Errorf("EstimatedCost = %v, want '$3.65'", p.EstimatedCost)
	}

	m.AssertExpectations(t)
}

func TestPresentStoppedInstance(t *testing.T) {
	inst := types.Instance{
		InstanceId:   aws.String("i-123"),
		InstanceType: types.InstanceTypeT3Micro,
		LaunchTime:   aws.Time(time.Now().AddDate(0, 0, -10)),
	}

	p := PresentStoppedInstance(inst)

	if p.Identifier != "i-123" {
		t.Errorf("Identifier = %v, want 'i-123'", p.Identifier)
	}
}

func TestPresentReservedInstance(t *testing.T) {
	ri := model.RiExpirationInfo{
		ReservedInstanceID: "ri-123",
		DaysUntilExpiry:    15,
	}

	p := PresentReservedInstance(ri)

	if p.Identifier != "ri-123" {
		t.Errorf("Identifier = %v, want 'ri-123'", p.Identifier)
	}
}

func TestPresentLoadBalancer(t *testing.T) {
	lb := elbtypes.LoadBalancer{
		LoadBalancerArn:  aws.String("arn:aws:elb:us-east-1:123:lb/app/my-lb/123"),
		LoadBalancerName: aws.String("my-lb"),
		Type:             elbtypes.LoadBalancerTypeEnumApplication,
		CreatedTime:      aws.Time(time.Now()),
	}

	m := new(services.MockPricingService)
	m.On("CalculateLoadBalancerMonthlyCost", lb.Type).Return(16.43)

	p := PresentLoadBalancer(lb, m)

	if p.Identifier != "arn:aws:elb:us-east-1:123:lb/app/my-lb/123" {
		t.Errorf("Identifier = %v, want 'arn:aws:elb:us-east-1:123:lb/app/my-lb/123'", p.Identifier)
	}

	if p.EstimatedCost != "$16.43" {
		t.Errorf("EstimatedCost = %v, want '$16.43'", p.EstimatedCost)
	}

	m.AssertExpectations(t)
}

func TestPresentCloudWatchLogGroup(t *testing.T) {
	lg := model.CloudWatchLogsWasteInfo{
		LogGroupName:         "test-lg",
		StoredBytes:          1024 * 1024 * 1024,
		EstimatedMonthlyCost: 0.50,
	}

	p := PresentCloudWatchLogGroup(lg)

	if p.Identifier != "test-lg" {
		t.Errorf("Identifier = %v, want 'test-lg'", p.Identifier)
	}

	if p.EstimatedCost != "$0.50" {
		t.Errorf("EstimatedCost = %v, want '$0.50'", p.EstimatedCost)
	}
}

func TestPresentAMI(t *testing.T) {
	ami := model.AMIWasteInfo{
		ImageID:            "ami-123",
		Name:               "test-ami",
		CreationDate:       time.Now(),
		DaysSinceCreate:    100,
		MaxPotentialSaving: 12.50,
	}

	p := PresentAMI(ami)

	assert := func(got, want string) {
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	}

	assert(p.Identifier, "ami-123")
	assert(p.EstimatedCost, "$12.50")
	assert(p.Age, "100")
}

func TestPresentSnapshot(t *testing.T) {
	snap := model.SnapshotWasteInfo{
		SnapshotID:          "snap-123",
		SizeGB:              100,
		VolumeExists:        false,
		DaysSinceCreate:     50,
		MaxPotentialSavings: 5.00,
		Category:            model.SnapshotCategoryOrphaned,
	}

	p := PresentSnapshot(snap)

	if p.Identifier != "snap-123" {
		t.Errorf("Identifier = %v, want 'snap-123'", p.Identifier)
	}

	if p.EstimatedCost != "$5.00" {
		t.Errorf("EstimatedCost = %v, want '$5.00'", p.EstimatedCost)
	}
}

func TestPresentKeyPair(t *testing.T) {
	kp := model.KeyPairWasteInfo{
		KeyName:         "test-key",
		DaysSinceCreate: 30,
	}

	p := PresentKeyPair(kp)

	if p.Identifier != "test-key" {
		t.Errorf("Identifier = %v, want 'test-key'", p.Identifier)
	}

	if p.Age != "30" {
		t.Errorf("Age = %v, want '30'", p.Age)
	}
}

func TestPresentRDS(t *testing.T) {
	t.Run("stopped instance", func(t *testing.T) {
		inst := model.RDSInstanceWasteInfo{
			DBInstanceID:         "db-1",
			EstimatedMonthlyCost: 20.0,
		}

		p := PresentRDSInstance(inst)

		if p.EstimatedCost != "$20.00" {
			t.Errorf("got %v, want $20.00", p.EstimatedCost)
		}
	})

	t.Run("old snapshot", func(t *testing.T) {
		snap := model.RDSSnapshotWasteInfo{
			DBSnapshotID:         "snap-1",
			EstimatedMonthlyCost: 5.0,
		}

		p := PresentRDSSnapshot(snap)

		if p.EstimatedCost != "$5.00" {
			t.Errorf("got %v, want $5.00", p.EstimatedCost)
		}
	})

	t.Run("idle instance", func(t *testing.T) {
		inst := model.RDSIdleInstanceInfo{
			DBInstanceID:         "idle-1",
			EstimatedMonthlyCost: 45.0,
		}

		p := PresentRDSIdleInstance(inst)

		if p.EstimatedCost != "$45.00" {
			t.Errorf("got %v, want $45.00", p.EstimatedCost)
		}
	})
}

func TestPresentIdleNATGateway(t *testing.T) {
	ng := model.NATGatewayWasteInfo{
		NATGatewayID:         "nat-123",
		EstimatedMonthlyCost: 32.85,
	}

	p := PresentIdleNATGateway(ng)

	if p.Identifier != "nat-123" {
		t.Errorf("Identifier = %v, want 'nat-123'", p.Identifier)
	}

	if p.EstimatedCost != "$32.85" {
		t.Errorf("EstimatedCost = %v, want '$32.85'", p.EstimatedCost)
	}
}

func TestPresentIdleLoadBalancer(t *testing.T) {
	lb := model.ELBIdleInfo{
		ARN:                  "idle-lb",
		EstimatedMonthlyCost: 16.43,
	}

	p := PresentIdleLoadBalancer(lb)

	if p.Identifier != "idle-lb" {
		t.Errorf("Identifier = %v, want 'idle-lb'", p.Identifier)
	}

	if p.EstimatedCost != "$16.43" {
		t.Errorf("EstimatedCost = %v, want '$16.43'", p.EstimatedCost)
	}
}

func TestPresentLambdaOverProvisioned(t *testing.T) {
	fn := model.LambdaOverProvisionedInfo{
		FunctionName: "test-fn",
	}

	p := PresentLambdaOverProvisioned(fn)

	if p.Identifier != "test-fn" {
		t.Errorf("Identifier = %v, want 'test-fn'", p.Identifier)
	}
}

func TestPresentECRNoLifecyclePolicy(t *testing.T) {
	repo := model.ECRNoLifecyclePolicyInfo{RepositoryName: "my-repo"}

	p := PresentECRNoLifecyclePolicy(repo)

	if p.Identifier != "my-repo" {
		t.Errorf("Identifier = %v, want 'my-repo'", p.Identifier)
	}

	if p.EstimatedCost != NAValue {
		t.Errorf("EstimatedCost = %v, want %v", p.EstimatedCost, NAValue)
	}

	if p.Details != "Lifecycle policy is recommended" {
		t.Errorf("Details = %v, want 'Lifecycle policy is recommended'", p.Details)
	}
}

func TestPresentECREmptyRepository(t *testing.T) {
	repo := model.ECREmptyRepositoryInfo{RepositoryName: "empty-repo"}

	p := PresentECREmptyRepository(repo)

	if p.Identifier != "empty-repo" {
		t.Errorf("Identifier = %v, want 'empty-repo'", p.Identifier)
	}

	if p.Metric != "0 images" {
		t.Errorf("Metric = %v, want '0 images'", p.Metric)
	}

	if p.Details != "Consider deleting empty repositories" {
		t.Errorf("Details = %v, want 'Consider deleting empty repositories'", p.Details)
	}
}

func TestPresentECRUntaggedImages(t *testing.T) {
	repo := model.ECRUntaggedImageInfo{
		RepositoryName:       "untagged-repo",
		UntaggedImageCount:   3,
		UntaggedSizeBytes:    2 * 1024 * 1024 * 1024,
		EstimatedMonthlyCost: 0.20,
	}

	p := PresentECRUntaggedImages(repo)

	if p.Identifier != "untagged-repo" {
		t.Errorf("Identifier = %v, want 'untagged-repo'", p.Identifier)
	}

	if p.EstimatedCost != "$0.20" {
		t.Errorf("EstimatedCost = %v, want '$0.20'", p.EstimatedCost)
	}

	if p.Metric != "3 untagged images" {
		t.Errorf("Metric = %v, want '3 untagged images'", p.Metric)
	}

	if p.Details != "Consuming 2.00 GB of storage" {
		t.Errorf("Details = %v, want 'Consuming 2.00 GB of storage'", p.Details)
	}
}

func TestPresentIdleSageMakerEndpoint(t *testing.T) {
	ep := model.IdleSageMakerEndpointInfo{
		EndpointName:         "test-ep",
		EstimatedMonthlyCost: 46.72,
		Variants: []model.SageMakerVariant{
			{InstanceType: "ml.t2.medium", InstanceCount: 1},
		},
	}

	p := PresentIdleSageMakerEndpoint(ep)

	if p.Identifier != "test-ep" {
		t.Errorf("Identifier = %v, want 'test-ep'", p.Identifier)
	}

	if p.EstimatedCost != "$46.72" {
		t.Errorf("EstimatedCost = %v, want '$46.72'", p.EstimatedCost)
	}
}

func TestPresentIdleEC2Instance(t *testing.T) {
	t.Run("with name tag", func(t *testing.T) {
		inst := model.EC2IdleInstanceInfo{
			InstanceID:           "i-1234",
			InstanceType:         "t3.medium",
			Name:                 "dev-box",
			CPUUtilizationAvg:    1.25,
			NetworkBytesPerDay:   1024 * 1024,
			DaysChecked:          14,
			EstimatedMonthlyCost: 30.37,
		}

		p := PresentIdleEC2Instance(inst)

		if !strings.Contains(p.Identifier, "dev-box") || !strings.Contains(p.Identifier, "i-1234") {
			t.Errorf("Identifier = %v, expected name+id", p.Identifier)
		}

		if p.EstimatedCost != "$30.37" {
			t.Errorf("EstimatedCost = %v", p.EstimatedCost)
		}

		if !strings.Contains(p.Metric, "1.25") || !strings.Contains(p.Metric, "MB/day") {
			t.Errorf("Metric = %v, expected CPU and network", p.Metric)
		}

		if !strings.Contains(p.Details, "t3.medium") || !strings.Contains(p.Details, "14") {
			t.Errorf("Details = %v, missing type or days", p.Details)
		}
	})

	t.Run("without name tag falls back to instance id", func(t *testing.T) {
		inst := model.EC2IdleInstanceInfo{
			InstanceID:   "i-5678",
			InstanceType: "m5.large",
		}

		p := PresentIdleEC2Instance(inst)

		if p.Identifier != "i-5678" {
			t.Errorf("Identifier = %v, want bare instance id", p.Identifier)
		}
	})
}

func TestPresentUnusedSecret(t *testing.T) {
	mockPricing := new(services.MockPricingService)
	mockPricing.On("CalculateSecretsManagerMonthlyCost", 1).Return(0.40)

	t.Run("with last accessed date", func(t *testing.T) {
		lastAccessed := time.Now().AddDate(0, 0, -10)
		secret := model.UnusedSecretInfo{
			Name:             "prod/db/password",
			LastAccessedDate: &lastAccessed,
		}

		p := PresentUnusedSecret(secret, mockPricing)

		if p.Identifier != "prod/db/password" {
			t.Errorf("Identifier = %v", p.Identifier)
		}

		if p.EstimatedCost != "$0.40" {
			t.Errorf("EstimatedCost = %v", p.EstimatedCost)
		}

		if p.Age != "10" {
			t.Errorf("Age = %v, want '10'", p.Age)
		}

		if !strings.Contains(p.Details, lastAccessed.Format(time.RFC3339)) {
			t.Errorf("Details = %v, missing formatted date", p.Details)
		}
	})

	t.Run("without last accessed date", func(t *testing.T) {
		secret := model.UnusedSecretInfo{
			Name: "test/key",
		}

		p := PresentUnusedSecret(secret, mockPricing)

		if p.Age != NAValue {
			t.Errorf("Age = %v, want N/A", p.Age)
		}

		if !strings.Contains(p.Details, NAValue) {
			t.Errorf("Details = %v, want N/A in details", p.Details)
		}
	})
}
