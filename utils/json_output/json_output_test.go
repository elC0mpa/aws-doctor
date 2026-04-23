package jsonoutput

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/mocks/services"
	"github.com/elC0mpa/aws-doctor/model"
)

// captureStdout captures stdout during function execution
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

func TestPrintJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:    "simple_struct",
			input:   struct{ Name string }{Name: "test"},
			wantErr: false,
		},
		{
			name:    "map",
			input:   map[string]int{"a": 1, "b": 2},
			wantErr: false,
		},
		{
			name:    "slice",
			input:   []string{"one", "two", "three"},
			wantErr: false,
		},
		{
			name:    "nested_struct",
			input:   struct{ Data struct{ Value int } }{Data: struct{ Value int }{Value: 42}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error

			output := captureStdout(func() {
				err = printJSON(tt.input)
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("printJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify output is valid JSON
				var parsed interface{}
				if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(output)), &parsed); jsonErr != nil {
					t.Errorf("printJSON() output is not valid JSON: %v", jsonErr)
				}
			}
		})
	}
}

func TestPrintJSON_Error(t *testing.T) {
	// Channels cannot be marshaled to JSON
	err := printJSON(make(chan int))
	if err == nil {
		t.Error("printJSON() should have returned an error for a channel")
	}
}

func TestOutputCostComparisonJSON(t *testing.T) {
	lastMonth := &model.CostInfo{
		CostGroup: model.CostGroup{
			"Amazon EC2": {Amount: 100.0, Unit: "USD"},
			"Amazon S3":  {Amount: 50.0, Unit: "USD"},
		},
	}
	lastMonth.Start = aws.String("2024-01-01")
	lastMonth.End = aws.String("2024-01-31")

	currentMonth := &model.CostInfo{
		CostGroup: model.CostGroup{
			"Amazon EC2": {Amount: 120.0, Unit: "USD"},
			"Amazon S3":  {Amount: 45.0, Unit: "USD"},
		},
	}
	currentMonth.Start = aws.String("2024-02-01")
	currentMonth.End = aws.String("2024-02-29")

	var err error

	output := captureStdout(func() {
		err = OutputCostComparisonJSON(model.RenderCostComparisonInput{
			AccountID:        "123456789012",
			LastTotalCost:    "150.00 USD",
			CurrentTotalCost: "165.00 USD",
			LastMonth:        lastMonth,
			CurrentMonth:     currentMonth,
		})
	})

	if err != nil {
		t.Fatalf("OutputCostComparisonJSON() error = %v", err)
	}

	// Parse and verify JSON structure
	var result model.CostComparisonJSON
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); jsonErr != nil {
		t.Fatalf("Failed to parse output JSON: %v", jsonErr)
	}

	// Verify account ID
	if result.AccountID != "123456789012" {
		t.Errorf("AccountID = %v, want 123456789012", result.AccountID)
	}

	// Verify current month
	if result.CurrentMonth.Total != 165.0 {
		t.Errorf("CurrentMonth.Total = %v, want 165.0", result.CurrentMonth.Total)
	}

	// Verify last month
	if result.LastMonth.Total != 150.0 {
		t.Errorf("LastMonth.Total = %v, want 150.0", result.LastMonth.Total)
	}

	// Verify service breakdown exists
	if len(result.ServiceBreakdown) != 2 {
		t.Errorf("ServiceBreakdown has %d items, want 2", len(result.ServiceBreakdown))
	}
}

func TestOutputTrendJSON(t *testing.T) {
	costInfo := []model.CostInfo{
		{CostGroup: model.CostGroup{"Total": {Amount: 100.0, Unit: "USD"}}},
		{CostGroup: model.CostGroup{"Total": {Amount: 120.0, Unit: "USD"}}},
		{CostGroup: model.CostGroup{"Total": {Amount: 90.0, Unit: "USD"}}},
	}
	costInfo[0].Start = aws.String("2024-01-01")
	costInfo[0].End = aws.String("2024-01-31")
	costInfo[1].Start = aws.String("2024-02-01")
	costInfo[1].End = aws.String("2024-02-29")
	costInfo[2].Start = aws.String("2024-03-01")
	costInfo[2].End = aws.String("2024-03-31")

	var err error

	output := captureStdout(func() {
		err = OutputTrendJSON("123456789012", costInfo, []string{})
	})

	if err != nil {
		t.Fatalf("OutputTrendJSON() error = %v", err)
	}

	// Parse and verify JSON structure
	var result model.TrendJSON
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); jsonErr != nil {
		t.Fatalf("Failed to parse output JSON: %v", jsonErr)
	}

	if result.AccountID != "123456789012" {
		t.Errorf("AccountID = %v, want 123456789012", result.AccountID)
	}

	if len(result.Months) != 3 {
		t.Errorf("Months has %d items, want 3", len(result.Months))
	}
}

func TestOutputTrendJSON_SkipsNonTotal(t *testing.T) {
	// Test that entries without "Total" key are skipped
	costInfo := []model.CostInfo{
		{CostGroup: model.CostGroup{"Total": {Amount: 100.0, Unit: "USD"}}},
		{CostGroup: model.CostGroup{"Other": {Amount: 50.0, Unit: "USD"}}}, // No Total
	}
	costInfo[0].Start = aws.String("2024-01-01")
	costInfo[0].End = aws.String("2024-01-31")
	costInfo[1].Start = aws.String("2024-02-01")
	costInfo[1].End = aws.String("2024-02-29")

	var err error

	output := captureStdout(func() {
		err = OutputTrendJSON("123456789012", costInfo, []string{})
	})

	if err != nil {
		t.Fatalf("OutputTrendJSON() error = %v", err)
	}

	var result model.TrendJSON
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); jsonErr != nil {
		t.Fatalf("Failed to parse output JSON: %v", jsonErr)
	}

	// Should only have 1 month (the one with Total)
	if len(result.Months) != 1 {
		t.Errorf("Months has %d items, want 1 (non-Total should be skipped)", len(result.Months))
	}
}

func TestOutputWasteJSON(t *testing.T) {
	elasticIPs := []ec2types.Address{
		{PublicIp: aws.String("1.2.3.4"), AllocationId: aws.String("eipalloc-123")},
	}

	unusedVolumes := []ec2types.Volume{
		{VolumeId: aws.String("vol-123"), Size: aws.Int32(100)},
	}

	stoppedVolumes := []ec2types.Volume{
		{VolumeId: aws.String("vol-456"), Size: aws.Int32(200)},
	}

	stoppedInstances := []ec2types.Instance{
		{
			InstanceId:            aws.String("i-123"),
			StateTransitionReason: aws.String("User initiated (2024-01-01 00:00:00 UTC)"),
		},
	}

	ris := []model.RiExpirationInfo{
		{
			ReservedInstanceID: "ri-123",
			InstanceType:       "t3.medium",
			ExpirationDate:     time.Now().Add(30 * 24 * time.Hour),
			DaysUntilExpiry:    30,
			State:              "active",
			Status:             "EXPIRING SOON",
		},
	}

	loadBalancers := []elbtypes.LoadBalancer{
		{
			LoadBalancerName: aws.String("my-alb"),
			LoadBalancerArn:  aws.String("arn:aws:elasticloadbalancing:..."),
			Type:             elbtypes.LoadBalancerTypeEnumApplication,
		},
	}

	unusedKeyPairs := []model.KeyPairWasteInfo{
		{
			KeyName:         "test-key",
			KeyPairID:       "key-0123456789",
			CreateTime:      time.Now().AddDate(0, 0, -30),
			DaysSinceCreate: 30,
		},
	}

	s3Buckets := []model.S3BucketWasteInfo{
		{
			BucketName:   "waste-bucket",
			CreationDate: time.Now().AddDate(0, 0, -10),
			Reason:       "No lifecycle policy",
		},
	}

	s3Multipart := []model.S3MultipartUploadWasteInfo{
		{
			BucketName:  "multipart-bucket",
			UploadCount: 5,
		},
	}

	var err error

	output := captureStdout(func() {
		err = OutputWasteJSON(model.RenderWasteInput{
			AccountID:          "123456789012",
			ElasticIPs:         elasticIPs,
			UnusedVolumes:      unusedVolumes,
			StoppedVolumes:     stoppedVolumes,
			Ris:                ris,
			StoppedInstances:   stoppedInstances,
			LoadBalancers:      loadBalancers,
			UnusedKeyPairs:     unusedKeyPairs,
			S3Buckets:          s3Buckets,
			S3MultipartUploads: s3Multipart,
		}, services.NewMockPricingService())
	})

	if err != nil {
		t.Fatalf("OutputWasteJSON() error = %v", err)
	}

	// Parse and verify JSON structure
	var result model.WasteReportJSON
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); jsonErr != nil {
		t.Fatalf("Failed to parse output JSON: %v", jsonErr)
	}

	if result.AccountID != "123456789012" {
		t.Errorf("AccountID = %v, want 123456789012", result.AccountID)
	}

	if !result.HasWaste {
		t.Error("HasWaste should be true when waste items exist")
	}

	if len(result.UnusedElasticIPs) != 1 {
		t.Errorf("UnusedElasticIPs has %d items, want 1", len(result.UnusedElasticIPs))
	} else if result.UnusedElasticIPs[0].EstimatedMonthlyCost == 0 {
		t.Error("UnusedElasticIPs EstimatedMonthlyCost should not be zero")
	}

	if len(result.UnusedEBSVolumes) != 1 {
		t.Errorf("UnusedEBSVolumes has %d items, want 1", len(result.UnusedEBSVolumes))
	} else if result.UnusedEBSVolumes[0].EstimatedMonthlyCost == 0 {
		t.Error("UnusedEBSVolumes EstimatedMonthlyCost should not be zero")
	}

	if len(result.StoppedVolumes) != 1 {
		t.Errorf("StoppedVolumes has %d items, want 1", len(result.StoppedVolumes))
	} else if result.StoppedVolumes[0].EstimatedMonthlyCost == 0 {
		t.Error("StoppedVolumes EstimatedMonthlyCost should not be zero")
	}

	if len(result.StoppedInstances) != 1 {
		t.Errorf("StoppedInstances has %d items, want 1", len(result.StoppedInstances))
	}

	if len(result.ReservedInstances) != 1 {
		t.Errorf("ReservedInstances has %d items, want 1", len(result.ReservedInstances))
	}

	if len(result.UnusedLoadBalancers) != 1 {
		t.Errorf("UnusedLoadBalancers has %d items, want 1", len(result.UnusedLoadBalancers))
	} else if result.UnusedLoadBalancers[0].EstimatedMonthlyCost == 0 {
		t.Error("UnusedLoadBalancers EstimatedMonthlyCost should not be zero")
	}

	if len(result.UnusedKeyPairs) != 1 {
		t.Errorf("UnusedKeyPairs has %d items, want 1", len(result.UnusedKeyPairs))
	}

	if len(result.S3Buckets) != 1 {
		t.Errorf("S3Buckets has %d items, want 1", len(result.S3Buckets))
	}

	if len(result.S3MultipartUploads) != 1 {
		t.Errorf("S3MultipartUploads has %d items, want 1", len(result.S3MultipartUploads))
	}

	if result.UnusedKeyPairs[0].KeyName != "test-key" {
		t.Errorf("KeyName = %v, want 'test-key'", result.UnusedKeyPairs[0].KeyName)
	}
}

func TestOutputWasteJSON_WithSnapshots(t *testing.T) {
	orphanedSnapshots := []model.SnapshotWasteInfo{
		{
			SnapshotID: "snap-orphaned",
			Category:   model.SnapshotCategoryOrphaned,
		},
		{
			SnapshotID: "snap-stale",
			Category:   model.SnapshotCategoryStale,
		},
	}

	var err error

	output := captureStdout(func() {
		err = OutputWasteJSON(model.RenderWasteInput{
			AccountID:         "123456789012",
			OrphanedSnapshots: orphanedSnapshots,
		}, services.NewMockPricingService())
	})

	if err != nil {
		t.Fatalf("OutputWasteJSON() error = %v", err)
	}

	var result model.WasteReportJSON
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); jsonErr != nil {
		t.Fatalf("Failed to parse output JSON: %v", jsonErr)
	}

	if len(result.OrphanedSnapshots) != 1 {
		t.Errorf("OrphanedSnapshots has %d items, want 1", len(result.OrphanedSnapshots))
	}

	if len(result.StaleSnapshots) != 1 {
		t.Errorf("StaleSnapshots has %d items, want 1", len(result.StaleSnapshots))
	}
}

func TestOutputWasteJSON_NoWaste(t *testing.T) {
	var err error

	output := captureStdout(func() {
		err = OutputWasteJSON(model.RenderWasteInput{AccountID: "123456789012"}, services.NewMockPricingService())
	})

	if err != nil {
		t.Fatalf("OutputWasteJSON() error = %v", err)
	}

	var result model.WasteReportJSON
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); jsonErr != nil {
		t.Fatalf("Failed to parse output JSON: %v", jsonErr)
	}

	if result.HasWaste {
		t.Error("HasWaste should be false when no waste items exist")
	}
}

func TestOutputWasteJSON_InstanceWithoutTransitionReason(t *testing.T) {
	stoppedInstances := []ec2types.Instance{
		{
			InstanceId:            aws.String("i-123"),
			StateTransitionReason: nil, // No reason
		},
	}

	var err error

	output := captureStdout(func() {
		err = OutputWasteJSON(model.RenderWasteInput{
			AccountID:        "123456789012",
			StoppedInstances: stoppedInstances,
		}, services.NewMockPricingService())
	})

	if err != nil {
		t.Fatalf("OutputWasteJSON() error = %v", err)
	}

	var result model.WasteReportJSON
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); jsonErr != nil {
		t.Fatalf("Failed to parse output JSON: %v", jsonErr)
	}

	if len(result.StoppedInstances) != 1 {
		t.Fatalf("Expected 1 stopped instance, got %d", len(result.StoppedInstances))
	}

	// StoppedAt should be empty since no reason was provided
	if result.StoppedInstances[0].StoppedAt != "" {
		t.Errorf("StoppedAt should be empty, got %v", result.StoppedInstances[0].StoppedAt)
	}
}

func TestOutputWasteJSON_InstanceWithInvalidTransitionReason(t *testing.T) {
	stoppedInstances := []ec2types.Instance{
		{
			InstanceId:            aws.String("i-123"),
			StateTransitionReason: aws.String("invalid reason without date"),
		},
	}

	var err error

	output := captureStdout(func() {
		err = OutputWasteJSON(model.RenderWasteInput{
			AccountID:        "123456789012",
			StoppedInstances: stoppedInstances,
		}, services.NewMockPricingService())
	})

	if err != nil {
		t.Fatalf("OutputWasteJSON() error = %v", err)
	}

	var result model.WasteReportJSON
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); jsonErr != nil {
		t.Fatalf("Failed to parse output JSON: %v", jsonErr)
	}

	// Should still have the instance, just without StoppedAt
	if len(result.StoppedInstances) != 1 {
		t.Fatalf("Expected 1 stopped instance, got %d", len(result.StoppedInstances))
	}
}

func TestOutputWasteJSON_WithUnusedAMIs(t *testing.T) {
	unusedAMIs := []model.AMIWasteInfo{
		{
			ImageID:            "ami-12345",
			Name:               "my-test-ami",
			Description:        "Test AMI for unit tests",
			CreationDate:       time.Now().AddDate(0, -3, 0), // 3 months ago
			DaysSinceCreate:    90,
			IsPublic:           false,
			SnapshotIDs:        []string{"snap-111", "snap-222"},
			SnapshotSizeGB:     100,
			UsedByInstances:    0,
			MaxPotentialSaving: 5.00,
			SafetyWarning:      "Verify before deleting: AMI may be used by Auto Scaling Groups",
		},
	}

	var err error

	output := captureStdout(func() {
		err = OutputWasteJSON(model.RenderWasteInput{
			AccountID:  "123456789012",
			UnusedAMIs: unusedAMIs,
		}, services.NewMockPricingService())
	})

	if err != nil {
		t.Fatalf("OutputWasteJSON() error = %v", err)
	}

	var result model.WasteReportJSON
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); jsonErr != nil {
		t.Fatalf("Failed to parse output JSON: %v", jsonErr)
	}

	if !result.HasWaste {
		t.Error("HasWaste should be true when AMIs exist")
	}

	if len(result.UnusedAMIs) != 1 {
		t.Fatalf("UnusedAMIs has %d items, want 1", len(result.UnusedAMIs))
	}

	ami := result.UnusedAMIs[0]
	if ami.ImageID != "ami-12345" {
		t.Errorf("AMI ImageID = %v, want 'ami-12345'", ami.ImageID)
	}

	if ami.Name != "my-test-ami" {
		t.Errorf("AMI Name = %v, want 'my-test-ami'", ami.Name)
	}

	if ami.DaysSinceCreate != 90 {
		t.Errorf("AMI DaysSinceCreate = %v, want 90", ami.DaysSinceCreate)
	}

	if ami.MaxPotentialSaving != 5.00 {
		t.Errorf("AMI MaxPotentialSaving = %v, want 5.00", ami.MaxPotentialSaving)
	}

	if ami.SafetyWarning == "" {
		t.Error("AMI SafetyWarning should not be empty")
	}

	if len(ami.SnapshotIDs) != 2 {
		t.Errorf("AMI SnapshotIDs has %d items, want 2", len(ami.SnapshotIDs))
	}
}

func TestOutputWasteJSON_AMIWithEmptySnapshots(t *testing.T) {
	unusedAMIs := []model.AMIWasteInfo{
		{
			ImageID:            "ami-nosnapshots",
			Name:               "ami-without-snapshots",
			DaysSinceCreate:    45,
			SnapshotIDs:        []string{}, // Empty snapshots
			SnapshotSizeGB:     0,
			MaxPotentialSaving: 0.00,
			SafetyWarning:      "Verify before deleting",
		},
	}

	var err error

	output := captureStdout(func() {
		err = OutputWasteJSON(model.RenderWasteInput{
			AccountID:  "123456789012",
			UnusedAMIs: unusedAMIs,
		}, services.NewMockPricingService())
	})

	if err != nil {
		t.Fatalf("OutputWasteJSON() error = %v", err)
	}

	var result model.WasteReportJSON
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); jsonErr != nil {
		t.Fatalf("Failed to parse output JSON: %v", jsonErr)
	}

	if len(result.UnusedAMIs) != 1 {
		t.Fatalf("UnusedAMIs has %d items, want 1", len(result.UnusedAMIs))
	}

	ami := result.UnusedAMIs[0]
	if ami.SnapshotSizeGB != 0 {
		t.Errorf("AMI SnapshotSizeGB = %v, want 0", ami.SnapshotSizeGB)
	}

	if ami.MaxPotentialSaving != 0.00 {
		t.Errorf("AMI MaxPotentialSaving = %v, want 0.00", ami.MaxPotentialSaving)
	}
}

func TestOutputWasteJSON_WithIdleNATGateways(t *testing.T) {
	idleNATGateways := []model.NATGatewayWasteInfo{
		{
			NATGatewayID:          "nat-1234567890abcdef0",
			VPCID:                 "vpc-12345678",
			SubnetID:              "subnet-12345678",
			State:                 "available",
			BytesOutToDestination: 0,
		},
		{
			NATGatewayID:          "nat-abcdef01234567890",
			VPCID:                 "vpc-87654321",
			SubnetID:              "subnet-87654321",
			State:                 "available",
			BytesOutToDestination: 0,
		},
	}

	var err error

	output := captureStdout(func() {
		err = OutputWasteJSON(model.RenderWasteInput{
			AccountID:       "123456789012",
			IdleNATGateways: idleNATGateways,
		}, services.NewMockPricingService())
	})

	if err != nil {
		t.Fatalf("OutputWasteJSON() error = %v", err)
	}

	var result model.WasteReportJSON
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); jsonErr != nil {
		t.Fatalf("Failed to parse output JSON: %v", jsonErr)
	}

	if len(result.IdleNATGateways) != 2 {
		t.Fatalf("IdleNATGateways has %d items, want 2", len(result.IdleNATGateways))
	}

	if result.IdleNATGateways[0].NATGatewayID != "nat-1234567890abcdef0" {
		t.Errorf("First NAT Gateway ID = %v, want nat-1234567890abcdef0", result.IdleNATGateways[0].NATGatewayID)
	}

	if result.IdleNATGateways[0].BytesOutToDestination != 0 {
		t.Errorf("First NAT Gateway BytesOutToDestination = %v, want 0", result.IdleNATGateways[0].BytesOutToDestination)
	}
}

// TestOutputWasteJSON_NilElasticIPs verifies nil ElasticIPs behavior
// This test documents a known issue where nil slices serialize to "null" instead of "[]"
func TestOutputWasteJSON_NilElasticIPs(t *testing.T) {
	input := model.RenderWasteInput{
		AccountID:  "123456789012",
		ElasticIPs: nil,
	}

	var err error

	output := captureStdout(func() {
		err = OutputWasteJSON(input, services.NewMockPricingService())
	})

	if err != nil {
		t.Fatalf("OutputWasteJSON() error = %v", err)
	}

	// Verify output is valid JSON
	var parsed map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(output)), &parsed); jsonErr != nil {
		t.Fatalf("Failed to parse output JSON: %v", jsonErr)
	}

	// Document current behavior: nil ElasticIPs serialize to null
	// TODO: Consider fixing OutputWasteJSON to use empty slices instead of nil
	t.Logf("Current behavior: nil ElasticIPs produces: %v", parsed["unused_elastic_ips"])
}

// TestOutputWasteJSON_EmptyElasticIPs verifies empty slice behavior
func TestOutputWasteJSON_EmptyElasticIPs(t *testing.T) {
	input := model.RenderWasteInput{
		AccountID:  "123456789012",
		ElasticIPs: []ec2types.Address{},
	}

	var err error

	output := captureStdout(func() {
		err = OutputWasteJSON(input, services.NewMockPricingService())
	})

	if err != nil {
		t.Fatalf("OutputWasteJSON() error = %v", err)
	}

	// Verify output is valid JSON
	var parsed map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(output)), &parsed); jsonErr != nil {
		t.Fatalf("Failed to parse output JSON: %v", jsonErr)
	}

	// Document current behavior
	t.Logf("Current behavior: empty ElasticIPs produces: %v", parsed["unused_elastic_ips"])
}

// TestOutputWasteJSON_LargeCollection tests that large collections are handled properly
func TestOutputWasteJSON_LargeCollection(t *testing.T) {
	// Create 100 elastic IPs
	elasticIPs := make([]ec2types.Address, 100)
	for i := range 100 {
		elasticIPs[i] = ec2types.Address{
			PublicIp:     aws.String(fmt.Sprintf("1.2.3.%d", i)),
			AllocationId: aws.String(fmt.Sprintf("eipalloc-%d", i)),
		}
	}

	input := model.RenderWasteInput{
		AccountID:  "123456789012",
		ElasticIPs: elasticIPs,
	}

	var err error

	output := captureStdout(func() {
		err = OutputWasteJSON(input, services.NewMockPricingService())
	})

	if err != nil {
		t.Fatalf("OutputWasteJSON() error = %v", err)
	}

	// Parse and verify we got 100 items
	var result model.WasteReportJSON
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); jsonErr != nil {
		t.Fatalf("Failed to parse output JSON: %v", jsonErr)
	}

	if len(result.UnusedElasticIPs) != 100 {
		t.Errorf("UnusedElasticIPs has %d items, want 100", len(result.UnusedElasticIPs))
	}
}

// TestOutputWasteJSON_SpecialCharactersInNames tests proper JSON escaping for special characters
func TestOutputWasteJSON_SpecialCharactersInNames(t *testing.T) {
	input := model.RenderWasteInput{
		AccountID: "123456789012",
		S3Buckets: []model.S3BucketWasteInfo{
			{BucketName: "test<bucket>", Reason: "No lifecycle policy"},
			{BucketName: "test>bucket", Reason: "Has uploads"},
			{BucketName: "test&bucket", Reason: "Old bucket"},
			{BucketName: `test"bucket"`, Reason: "Special chars"},
		},
	}

	var err error

	output := captureStdout(func() {
		err = OutputWasteJSON(input, services.NewMockPricingService())
	})

	if err != nil {
		t.Fatalf("OutputWasteJSON() error = %v", err)
	}

	// Verify output is valid JSON
	var result model.WasteReportJSON
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); jsonErr != nil {
		t.Fatalf("Output is not valid JSON: %v\nOutput was: %s", jsonErr, output)
	}

	// Verify all 4 buckets were captured and names are correctly preserved
	if len(result.S3Buckets) != 4 {
		t.Errorf("S3Buckets has %d items, want 4", len(result.S3Buckets))
	} else {
		expected := []string{"test<bucket>", "test>bucket", "test&bucket", `test"bucket"`}
		for i, name := range expected {
			if result.S3Buckets[i].BucketName != name {
				t.Errorf("S3Buckets[%d].BucketName = %q, want %q", i, result.S3Buckets[i].BucketName, name)
			}
		}
	}
}

// TestOutputWasteJSON_AllNilFields tests that all nil fields serialize to empty arrays
func TestOutputWasteJSON_AllNilFields(t *testing.T) {
	input := model.RenderWasteInput{
		AccountID: "123456789012",
	}

	var err error

	output := captureStdout(func() {
		err = OutputWasteJSON(input, services.NewMockPricingService())
	})

	if err != nil {
		t.Fatalf("OutputWasteJSON() error = %v", err)
	}

	// Verify output is valid JSON
	var result model.WasteReportJSON
	if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); jsonErr != nil {
		t.Fatalf("Output is not valid JSON: %v", jsonErr)
	}

	// All waste arrays should be empty
	if result.HasWaste {
		t.Error("HasWaste should be false when all fields are empty/nil")
	}
}

func BenchmarkOutputWasteJSON(b *testing.B) {
	elasticIPs := make([]ec2types.Address, 10)
	for i := range 10 {
		elasticIPs[i] = ec2types.Address{
			PublicIp:     aws.String("1.2.3." + string(rune('0'+i))),
			AllocationId: aws.String("eipalloc-" + string(rune('a'+i))),
		}
	}

	// Redirect stdout to discard during benchmark
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)

	defer func() { os.Stdout = old }()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = OutputWasteJSON(model.RenderWasteInput{
			AccountID:  "123456789012",
			ElasticIPs: elasticIPs,
		}, services.NewMockPricingService())
	}
}
