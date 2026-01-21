package utils

import (
	"bytes"
	"encoding/json"
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

// captureStdout captures stdout during function execution
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
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
		err = OutputCostComparisonJSON("123456789012", 150.0, 165.0, lastMonth, currentMonth)
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
		err = OutputTrendJSON("123456789012", costInfo)
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
		err = OutputTrendJSON("123456789012", costInfo)
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
	elasticIPs := []types.Address{
		{PublicIp: aws.String("1.2.3.4"), AllocationId: aws.String("eipalloc-123")},
	}

	unusedVolumes := []types.Volume{
		{VolumeId: aws.String("vol-123"), Size: aws.Int32(100)},
	}

	stoppedVolumes := []types.Volume{
		{VolumeId: aws.String("vol-456"), Size: aws.Int32(200)},
	}

	stoppedInstances := []types.Instance{
		{
			InstanceId:            aws.String("i-123"),
			StateTransitionReason: aws.String("User initiated (2024-01-01 00:00:00 UTC)"),
		},
	}

	ris := []model.RiExpirationInfo{
		{
			ReservedInstanceId: "ri-123",
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

	var err error
	output := captureStdout(func() {
		err = OutputWasteJSON("123456789012", elasticIPs, unusedVolumes, stoppedVolumes, ris, stoppedInstances, loadBalancers)
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
	}

	if len(result.UnusedEBSVolumes) != 1 {
		t.Errorf("UnusedEBSVolumes has %d items, want 1", len(result.UnusedEBSVolumes))
	}

	if len(result.StoppedVolumes) != 1 {
		t.Errorf("StoppedVolumes has %d items, want 1", len(result.StoppedVolumes))
	}

	if len(result.StoppedInstances) != 1 {
		t.Errorf("StoppedInstances has %d items, want 1", len(result.StoppedInstances))
	}

	if len(result.ReservedInstances) != 1 {
		t.Errorf("ReservedInstances has %d items, want 1", len(result.ReservedInstances))
	}

	if len(result.UnusedLoadBalancers) != 1 {
		t.Errorf("UnusedLoadBalancers has %d items, want 1", len(result.UnusedLoadBalancers))
	}
}

func TestOutputWasteJSON_NoWaste(t *testing.T) {
	var err error
	output := captureStdout(func() {
		err = OutputWasteJSON("123456789012", nil, nil, nil, nil, nil, nil)
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
	stoppedInstances := []types.Instance{
		{
			InstanceId:            aws.String("i-123"),
			StateTransitionReason: nil, // No reason
		},
	}

	var err error
	output := captureStdout(func() {
		err = OutputWasteJSON("123456789012", nil, nil, nil, nil, stoppedInstances, nil)
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
	stoppedInstances := []types.Instance{
		{
			InstanceId:            aws.String("i-123"),
			StateTransitionReason: aws.String("invalid reason without date"),
		},
	}

	var err error
	output := captureStdout(func() {
		err = OutputWasteJSON("123456789012", nil, nil, nil, nil, stoppedInstances, nil)
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

func BenchmarkOutputWasteJSON(b *testing.B) {
	elasticIPs := make([]types.Address, 10)
	for i := 0; i < 10; i++ {
		elasticIPs[i] = types.Address{
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
		OutputWasteJSON("123456789012", elasticIPs, nil, nil, nil, nil, nil)
	}
}
