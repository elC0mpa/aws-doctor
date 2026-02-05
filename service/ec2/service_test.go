package awscostexplorer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elC0mpa/aws-doctor/mocks/awsinterfaces"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetElasticIPAddressesInfo(t *testing.T) {
	mockClient := new(awsinterfaces.MockEC2Client)
	s := &service{client: mockClient}

	// Setup mock data
	mockClient.On("DescribeAddresses", mock.Anything, mock.Anything, mock.Anything).Return(&ec2.DescribeAddressesOutput{
		Addresses: []types.Address{
			// Unused IP
			{
				AllocationId:  aws.String("eipalloc-1"),
				PublicIp:      aws.String("1.2.3.4"),
				AssociationId: nil,
			},
			// Attached to Instance
			{
				AllocationId:  aws.String("eipalloc-2"),
				PublicIp:      aws.String("5.6.7.8"),
				AssociationId: aws.String("eipassoc-2"),
				InstanceId:    aws.String("i-12345"),
			},
			// Attached to Network Interface (ELB)
			{
				AllocationId:       aws.String("eipalloc-3"),
				PublicIp:           aws.String("9.10.11.12"),
				AssociationId:      aws.String("eipassoc-3"),
				InstanceId:         nil,
				NetworkInterfaceId: aws.String("eni-3"),
			},
		},
	}, nil)

	mockClient.On("DescribeNetworkInterfaces", mock.Anything, &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: []string{"eni-3"},
	}, mock.Anything).Return(&ec2.DescribeNetworkInterfacesOutput{
		NetworkInterfaces: []types.NetworkInterface{
			{
				InterfaceType: types.NetworkInterfaceTypeInterface,
				Description:   aws.String("ELB app/my-lb/123"),
			},
		},
	}, nil)

	// Execute
	result, err := s.GetElasticIPAddressesInfo(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result.UnusedElasticIPAddresses, 1)
	assert.Contains(t, result.UnusedElasticIPAddresses, "eipalloc-1")
	assert.Len(t, result.UsedElasticIPAddresses, 2)

	// Check items in used list
	foundInstanceIP := false
	foundELBIP := false

	for _, ip := range result.UsedElasticIPAddresses {
		if ip.AllocationID == "eipalloc-2" {
			assert.Equal(t, "5.6.7.8", ip.IPAddress)
			assert.Equal(t, "ec2", ip.ResourceType)

			foundInstanceIP = true
		}

		if ip.AllocationID == "eipalloc-3" {
			assert.Equal(t, "9.10.11.12", ip.IPAddress)
			assert.Equal(t, string(types.NetworkInterfaceTypeLoadBalancer), ip.ResourceType)

			foundELBIP = true
		}
	}

	assert.True(t, foundInstanceIP, "Should find instance attached IP")
	assert.True(t, foundELBIP, "Should find ELB attached IP")

	mockClient.AssertExpectations(t)
}

func TestGetElasticIPAddressesInfo_Error(t *testing.T) {
	mockClient := new(awsinterfaces.MockEC2Client)
	s := &service{client: mockClient}

	mockClient.On("DescribeAddresses", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("API error"))

	_, err := s.GetElasticIPAddressesInfo(context.Background())
	assert.Error(t, err)
	mockClient.AssertExpectations(t)
}

func TestGetElasticIPAddressesInfo_ENIError(t *testing.T) {
	mockClient := new(awsinterfaces.MockEC2Client)
	s := &service{client: mockClient}

	mockClient.On("DescribeAddresses", mock.Anything, mock.Anything, mock.Anything).Return(&ec2.DescribeAddressesOutput{
		Addresses: []types.Address{
			{
				AllocationId:       aws.String("eipalloc-3"),
				AssociationId:      aws.String("eipassoc-3"),
				InstanceId:         nil,
				NetworkInterfaceId: aws.String("eni-3"),
			},
		},
	}, nil)

	mockClient.On("DescribeNetworkInterfaces", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("ENI error"))

	_, err := s.GetElasticIPAddressesInfo(context.Background())
	assert.Error(t, err)
	mockClient.AssertExpectations(t)
}

func TestGetUnusedElasticIPAddressesInfo(t *testing.T) {
	mockClient := new(awsinterfaces.MockEC2Client)
	s := &service{client: mockClient}

	mockClient.On("DescribeAddresses", mock.Anything, mock.Anything, mock.Anything).Return(&ec2.DescribeAddressesOutput{
		Addresses: []types.Address{
			{AllocationId: aws.String("eipalloc-1"), AssociationId: nil},
			{AllocationId: aws.String("eipalloc-2"), AssociationId: aws.String("assoc-1")},
		},
	}, nil)

	result, err := s.GetUnusedElasticIPAddressesInfo(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "eipalloc-1", *result[0].AllocationId)
	mockClient.AssertExpectations(t)
}

func TestGetUnusedEBSVolumes(t *testing.T) {
	mockClient := new(awsinterfaces.MockEC2Client)
	s := &service{client: mockClient}

	mockClient.On("DescribeVolumes", mock.Anything, mock.Anything, mock.Anything).Return(&ec2.DescribeVolumesOutput{
		Volumes: []types.Volume{
			{VolumeId: aws.String("vol-1"), State: types.VolumeStateAvailable},
		},
	}, nil)

	result, err := s.GetUnusedEBSVolumes(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "vol-1", *result[0].VolumeId)
	mockClient.AssertExpectations(t)
}

func TestGetReservedInstanceExpiringOrExpired30DaysWaste(t *testing.T) {
	mockClient := new(awsinterfaces.MockEC2Client)
	s := &service{client: mockClient}

	now := time.Now()
	expiringSoon := now.Add(5 * 24 * time.Hour)
	recentlyExpired := now.Add(-5 * 24 * time.Hour)
	activeFuture := now.Add(60 * 24 * time.Hour)

	mockClient.On("DescribeReservedInstances", mock.Anything, mock.Anything, mock.Anything).Return(&ec2.DescribeReservedInstancesOutput{
		ReservedInstances: []types.ReservedInstances{
			{
				ReservedInstancesId: aws.String("ri-1"),
				State:               types.ReservedInstanceStateActive,
				End:                 &expiringSoon,
			},
			{
				ReservedInstancesId: aws.String("ri-2"),
				State:               types.ReservedInstanceStateRetired,
				End:                 &recentlyExpired,
			},
			{
				ReservedInstancesId: aws.String("ri-3"),
				State:               types.ReservedInstanceStateActive,
				End:                 &activeFuture,
			},
		},
	}, nil)

	result, err := s.GetReservedInstanceExpiringOrExpired30DaysWaste(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Verify contents
	foundExpiring := false
	foundExpired := false

	for _, r := range result {
		if r.ReservedInstanceID == "ri-1" {
			assert.Equal(t, "EXPIRING SOON", r.Status)

			foundExpiring = true
		}

		if r.ReservedInstanceID == "ri-2" {
			assert.Equal(t, "RECENTLY EXPIRED", r.Status)

			foundExpired = true
		}
	}

	assert.True(t, foundExpiring)
	assert.True(t, foundExpired)
	mockClient.AssertExpectations(t)
}

func TestGetStoppedInstancesInfo(t *testing.T) {
	mockClient := new(awsinterfaces.MockEC2Client)
	s := &service{client: mockClient}

	now := time.Now()
	oldDate := now.Add(-35*24*time.Hour).Format("2006-01-02 15:04:05") + " GMT"
	recentDate := now.Add(-5*24*time.Hour).Format("2006-01-02 15:04:05") + " GMT"

	// Mock DescribeInstances
	mockClient.On("DescribeInstances", mock.Anything, mock.Anything, mock.Anything).Return(&ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{
						InstanceId:            aws.String("i-old"),
						StateTransitionReason: aws.String("User initiated (" + oldDate + ")"),
						BlockDeviceMappings: []types.InstanceBlockDeviceMapping{
							{
								Ebs: &types.EbsInstanceBlockDevice{
									VolumeId: aws.String("vol-old"),
								},
							},
						},
					},
					{
						InstanceId:            aws.String("i-recent"),
						StateTransitionReason: aws.String("User initiated (" + recentDate + ")"),
					},
				},
			},
		},
	}, nil)

	// Mock DescribeVolumes (for the stopped instance's volume)
	mockClient.On("DescribeVolumes", mock.Anything, &ec2.DescribeVolumesInput{
		VolumeIds: []string{"vol-old"},
	}, mock.Anything).Return(&ec2.DescribeVolumesOutput{
		Volumes: []types.Volume{
			{
				VolumeId: aws.String("vol-old"),
				Size:     aws.Int32(50),
			},
		},
	}, nil)

	instances, volumes, err := s.GetStoppedInstancesInfo(context.Background())

	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assert.Equal(t, "i-old", *instances[0].InstanceId)
	assert.Len(t, volumes, 1)
	assert.Equal(t, "vol-old", *volumes[0].VolumeId)
	mockClient.AssertExpectations(t)
}

func TestGetUnusedAMIs(t *testing.T) {
	mockClient := new(awsinterfaces.MockEC2Client)
	s := &service{client: mockClient}

	staleDays := 90
	oldDate := time.Now().AddDate(0, 0, -100).Format(time.RFC3339)
	recentDate := time.Now().AddDate(0, 0, -10).Format(time.RFC3339)

	// Mock DescribeInstances (to find used AMIs)
	mockClient.On("DescribeInstances", mock.Anything, mock.Anything, mock.Anything).Return(&ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{ImageId: aws.String("ami-used")},
				},
			},
		},
	}, nil)

	// Mock DescribeImages
	mockClient.On("DescribeImages", mock.Anything, &ec2.DescribeImagesInput{
		Owners: []string{"self"},
	}, mock.Anything).Return(&ec2.DescribeImagesOutput{
		Images: []types.Image{
			// Used AMI (should be filtered out even if old)
			{
				ImageId:      aws.String("ami-used"),
				CreationDate: aws.String(oldDate),
			},
			// Unused but Recent AMI (should be filtered out)
			{
				ImageId:      aws.String("ami-recent"),
				CreationDate: aws.String(recentDate),
			},
			// Unused and Old AMI (Waste!)
			{
				ImageId:      aws.String("ami-waste"),
				CreationDate: aws.String(oldDate),
				Name:         aws.String("Old Backup"),
				BlockDeviceMappings: []types.BlockDeviceMapping{
					{
						Ebs: &types.EbsBlockDevice{
							SnapshotId: aws.String("snap-1"),
							VolumeSize: aws.Int32(100),
						},
					},
				},
			},
		},
	}, nil)

	result, err := s.GetUnusedAMIs(context.Background(), staleDays)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "ami-waste", result[0].ImageID)
	assert.Equal(t, int64(100), result[0].SnapshotSizeGB)
	assert.Equal(t, 5.0, result[0].MaxPotentialSaving) // 100 * 0.05
	mockClient.AssertExpectations(t)
}

func TestGetOrphanedSnapshots(t *testing.T) {
	mockClient := new(awsinterfaces.MockEC2Client)
	s := &service{client: mockClient}

	staleDays := 90
	now := time.Now()
	oldDate := now.AddDate(0, 0, -100)
	recentDate := now.AddDate(0, 0, -10)

	// Mock DescribeSnapshots
	mockClient.On("DescribeSnapshots", mock.Anything, &ec2.DescribeSnapshotsInput{
		OwnerIds: []string{"self"},
	}, mock.Anything).Return(&ec2.DescribeSnapshotsOutput{
		Snapshots: []types.Snapshot{
			// Case 1: Used by AMI (Skip)
			{
				SnapshotId: aws.String("snap-ami"),
				VolumeId:   aws.String("vol-1"),
				StartTime:  &oldDate,
			},
			// Case 2: Orphaned (Volume deleted)
			{
				SnapshotId: aws.String("snap-orphaned"),
				VolumeId:   aws.String("vol-deleted"),
				StartTime:  &recentDate, // Date doesn't matter for orphaned
				VolumeSize: aws.Int32(50),
			},
			// Case 3: Stale (Volume exists, but old)
			{
				SnapshotId: aws.String("snap-stale"),
				VolumeId:   aws.String("vol-exists"),
				StartTime:  &oldDate,
				VolumeSize: aws.Int32(100),
			},
			// Case 4: Active (Volume exists, recent) - Not returned
			{
				SnapshotId: aws.String("snap-active"),
				VolumeId:   aws.String("vol-exists"),
				StartTime:  &recentDate,
			},
		},
	}, nil)

	// Mock DescribeVolumes (to find existing volumes)
	mockClient.On("DescribeVolumes", mock.Anything, mock.Anything, mock.Anything).Return(&ec2.DescribeVolumesOutput{
		Volumes: []types.Volume{
			{VolumeId: aws.String("vol-1")},
			{VolumeId: aws.String("vol-exists")},
		},
	}, nil)

	// Mock DescribeImages (to find snapshots used by AMIs)
	mockClient.On("DescribeImages", mock.Anything, &ec2.DescribeImagesInput{
		Owners: []string{"self"},
	}, mock.Anything).Return(&ec2.DescribeImagesOutput{
		Images: []types.Image{
			{
				ImageId: aws.String("ami-1"),
				BlockDeviceMappings: []types.BlockDeviceMapping{
					{
						Ebs: &types.EbsBlockDevice{
							SnapshotId: aws.String("snap-ami"),
						},
					},
				},
			},
		},
	}, nil)

	result, err := s.GetOrphanedSnapshots(context.Background(), staleDays)

	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Verify Orphaned
	var (
		orphaned model.SnapshotWasteInfo
		stale    model.SnapshotWasteInfo
	)

	for _, r := range result {
		if r.SnapshotID == "snap-orphaned" {
			orphaned = r
		}

		if r.SnapshotID == "snap-stale" {
			stale = r
		}
	}

	assert.Equal(t, "snap-orphaned", orphaned.SnapshotID)
	assert.Equal(t, model.SnapshotCategoryOrphaned, orphaned.Category)
	assert.False(t, orphaned.VolumeExists)

	assert.Equal(t, "snap-stale", stale.SnapshotID)
	assert.Equal(t, model.SnapshotCategoryStale, stale.Category)
	assert.True(t, stale.VolumeExists)

	mockClient.AssertExpectations(t)
}

func TestGetUnusedKeyPairs(t *testing.T) {
	mockClient := new(awsinterfaces.MockEC2Client)
	s := &service{client: mockClient}

	now := time.Now()
	createTime := now.AddDate(0, 0, -10)

	// Mock DescribeKeyPairs
	mockClient.On("DescribeKeyPairs", mock.Anything, mock.Anything, mock.Anything).Return(&ec2.DescribeKeyPairsOutput{
		KeyPairs: []types.KeyPairInfo{
			{
				KeyName:    aws.String("key-used"),
				KeyPairId:  aws.String("key-1"),
				CreateTime: &createTime,
			},
			{
				KeyName:    aws.String("key-unused"),
				KeyPairId:  aws.String("key-2"),
				CreateTime: &createTime,
			},
		},
	}, nil)

	// Mock DescribeInstances
	mockClient.On("DescribeInstances", mock.Anything, mock.Anything, mock.Anything).Return(&ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{KeyName: aws.String("key-used")},
				},
			},
		},
	}, nil)

	result, err := s.GetUnusedKeyPairs(context.Background())

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "key-unused", result[0].KeyName)
	assert.Equal(t, "key-2", result[0].KeyPairID)
	assert.Equal(t, 10, result[0].DaysSinceCreate)
	mockClient.AssertExpectations(t)
}

func TestGetResourceTypeFromDescription(t *testing.T) {
	// Create a service instance for testing the method
	// We don't need a real client since getResourceTypeFromDescription doesn't use it
	s := &service{}

	tests := []struct {
		name        string
		description string
		want        types.NetworkInterfaceType
	}{
		// Application Load Balancer cases
		{
			name:        "alb_standard_format",
			description: "ELB app/my-load-balancer/abc123",
			want:        types.NetworkInterfaceTypeLoadBalancer,
		},
		{
			name:        "alb_lowercase",
			description: "elb app/test-alb/def456",
			want:        types.NetworkInterfaceTypeLoadBalancer,
		},
		{
			name:        "alb_mixed_case",
			description: "ELB APP/MyALB/xyz789",
			want:        types.NetworkInterfaceTypeLoadBalancer,
		},

		// Network Load Balancer cases
		{
			name:        "nlb_standard_format",
			description: "ELB net/my-nlb/abc123",
			want:        types.NetworkInterfaceTypeNetworkLoadBalancer,
		},
		{
			name:        "nlb_lowercase",
			description: "elb net/test-nlb/def456",
			want:        types.NetworkInterfaceTypeNetworkLoadBalancer,
		},

		// NAT Gateway cases
		{
			name:        "nat_gateway_standard",
			description: "Interface for NAT Gateway nat-0abc123def456",
			want:        types.NetworkInterfaceTypeNatGateway,
		},
		{
			name:        "nat_gateway_hyphenated",
			description: "nat-gateway interface",
			want:        types.NetworkInterfaceTypeNatGateway,
		},
		{
			name:        "nat_gateway_with_id",
			description: "NAT Gateway nat-12345",
			want:        types.NetworkInterfaceTypeNatGateway,
		},

		// Global Accelerator cases
		{
			name:        "global_accelerator",
			description: "AWS GlobalAccelerator managed interface",
			want:        types.NetworkInterfaceTypeGlobalAcceleratorManaged,
		},
		{
			name:        "global_accelerator_lowercase",
			description: "globalaccelerator endpoint",
			want:        types.NetworkInterfaceTypeGlobalAcceleratorManaged,
		},

		// VPC Endpoint cases
		{
			name:        "vpc_endpoint_standard",
			description: "VPC Endpoint Interface vpce-0abc123",
			want:        types.NetworkInterfaceTypeVpcEndpoint,
		},
		{
			name:        "vpc_endpoint_with_id",
			description: "Interface for vpce-12345678",
			want:        types.NetworkInterfaceTypeVpcEndpoint,
		},

		// Transit Gateway cases
		{
			name:        "transit_gateway_standard",
			description: "Transit Gateway Attachment tgw-attach-123",
			want:        types.NetworkInterfaceTypeTransitGateway,
		},
		{
			name:        "transit_gateway_with_id",
			description: "Network interface for tgw-12345",
			want:        types.NetworkInterfaceTypeTransitGateway,
		},

		// Lambda cases
		{
			name:        "lambda_standard",
			description: "AWS Lambda VPC ENI-my-function-abc123",
			want:        types.NetworkInterfaceTypeLambda,
		},
		{
			name:        "lambda_lowercase",
			description: "aws lambda function interface",
			want:        types.NetworkInterfaceTypeLambda,
		},

		// API Gateway cases
		{
			name:        "api_gateway_standard",
			description: "API Gateway managed interface",
			want:        types.NetworkInterfaceTypeApiGatewayManaged,
		},
		{
			name:        "api_gateway_lowercase",
			description: "api gateway endpoint",
			want:        types.NetworkInterfaceTypeApiGatewayManaged,
		},

		// IoT Rules cases
		{
			name:        "iot_rules",
			description: "IoT Rules managed interface",
			want:        types.NetworkInterfaceTypeIotRulesManaged,
		},

		// Gateway Load Balancer cases
		{
			name:        "gwlb_standard",
			description: "Gateway Load Balancer Endpoint",
			want:        types.NetworkInterfaceTypeGatewayLoadBalancer,
		},

		// Custom resource types (returned as NetworkInterfaceType strings)
		{
			name:        "redshift_cluster",
			description: "Redshift cluster my-cluster",
			want:        types.NetworkInterfaceType("redshift_cluster"),
		},
		{
			name:        "rds_database",
			description: "RDS database instance",
			want:        types.NetworkInterfaceType("rds_database"),
		},
		{
			name:        "directory_service",
			description: "Directory Service interface",
			want:        types.NetworkInterfaceType("directory_service"),
		},
		{
			name:        "fsx_filesystem",
			description: "FSx file system interface",
			want:        types.NetworkInterfaceType("fsx"),
		},

		// Default/fallback cases
		{
			name:        "empty_description",
			description: "",
			want:        types.NetworkInterfaceType("interface"),
		},
		{
			name:        "unknown_description",
			description: "Some random network interface",
			want:        types.NetworkInterfaceType("interface"),
		},
		{
			name:        "ec2_instance_description",
			description: "Primary network interface",
			want:        types.NetworkInterfaceType("interface"),
		},
		{
			name:        "ecs_task_description",
			description: "ecs-task/12345",
			want:        types.NetworkInterfaceType("interface"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.getResourceTypeFromDescription(tt.description)
			if got != tt.want {
				t.Errorf("getResourceTypeFromDescription(%q) = %v, want %v", tt.description, got, tt.want)
			}
		})
	}
}

func TestGetResourceTypeFromDescription_CaseInsensitivity(t *testing.T) {
	s := &service{}

	// Test that matching is case-insensitive
	casePairs := []struct {
		lower string
		upper string
		mixed string
	}{
		{"elb app/test/123", "ELB APP/TEST/123", "Elb App/Test/123"},
		{"nat gateway", "NAT GATEWAY", "Nat Gateway"},
		{"aws lambda", "AWS LAMBDA", "Aws Lambda"},
		{"vpc endpoint", "VPC ENDPOINT", "Vpc Endpoint"},
	}

	for _, pair := range casePairs {
		lowerResult := s.getResourceTypeFromDescription(pair.lower)
		upperResult := s.getResourceTypeFromDescription(pair.upper)
		mixedResult := s.getResourceTypeFromDescription(pair.mixed)

		if lowerResult != upperResult || upperResult != mixedResult {
			t.Errorf("Case sensitivity issue: lower=%v, upper=%v, mixed=%v for inputs %q/%q/%q",
				lowerResult, upperResult, mixedResult, pair.lower, pair.upper, pair.mixed)
		}
	}
}

func TestGetResourceTypeFromDescription_Priority(t *testing.T) {
	s := &service{}

	// Test that when multiple keywords could match, the first matching condition wins
	// Based on the order in the implementation
	tests := []struct {
		name        string
		description string
		want        types.NetworkInterfaceType
		reason      string
	}{
		{
			name:        "alb_before_generic_elb",
			description: "ELB app/lb-name/123 some elb",
			want:        types.NetworkInterfaceTypeLoadBalancer,
			reason:      "app/ should match before any other ELB pattern",
		},
		{
			name:        "nlb_before_generic_elb",
			description: "ELB net/lb-name/123",
			want:        types.NetworkInterfaceTypeNetworkLoadBalancer,
			reason:      "net/ should identify NLB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.getResourceTypeFromDescription(tt.description)
			if got != tt.want {
				t.Errorf("getResourceTypeFromDescription(%q) = %v, want %v (%s)",
					tt.description, got, tt.want, tt.reason)
			}
		})
	}
}

func BenchmarkGetResourceTypeFromDescription(b *testing.B) {
	s := &service{}
	descriptions := []string{
		"ELB app/my-load-balancer/abc123",
		"Interface for NAT Gateway nat-0abc123def456",
		"AWS Lambda VPC ENI-my-function-abc123",
		"Primary network interface",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, desc := range descriptions {
			s.getResourceTypeFromDescription(desc)
		}
	}
}
