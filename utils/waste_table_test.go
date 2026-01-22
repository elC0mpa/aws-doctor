package utils

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

func TestPopulateEBSRows(t *testing.T) {
	tests := []struct {
		name    string
		volumes []types.Volume
		wantLen int
	}{
		{
			name:    "empty_volumes",
			volumes: []types.Volume{},
			wantLen: 0,
		},
		{
			name: "single_volume",
			volumes: []types.Volume{
				{
					VolumeId: aws.String("vol-12345"),
					Size:     aws.Int32(100),
				},
			},
			wantLen: 1,
		},
		{
			name: "multiple_volumes",
			volumes: []types.Volume{
				{VolumeId: aws.String("vol-111"), Size: aws.Int32(50)},
				{VolumeId: aws.String("vol-222"), Size: aws.Int32(100)},
				{VolumeId: aws.String("vol-333"), Size: aws.Int32(200)},
			},
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := populateEBSRows(tt.volumes)

			if len(rows) != tt.wantLen {
				t.Errorf("populateEBSRows() returned %d rows, want %d", len(rows), tt.wantLen)
				return
			}

			// Verify each row has 3 columns
			for i, row := range rows {
				if len(row) != 3 {
					t.Errorf("Row %d has %d columns, want 3", i, len(row))
				}
			}

			// Verify volume IDs are in the rows
			for i, vol := range tt.volumes {
				if rows[i][1] != *vol.VolumeId {
					t.Errorf("Row %d VolumeId = %v, want %v", i, rows[i][1], *vol.VolumeId)
				}
			}
		})
	}
}

func TestPopulateElasticIpRows(t *testing.T) {
	tests := []struct {
		name    string
		ips     []types.Address
		wantLen int
	}{
		{
			name:    "empty_ips",
			ips:     []types.Address{},
			wantLen: 0,
		},
		{
			name: "single_ip",
			ips: []types.Address{
				{
					PublicIp:     aws.String("1.2.3.4"),
					AllocationId: aws.String("eipalloc-12345"),
				},
			},
			wantLen: 1,
		},
		{
			name: "multiple_ips",
			ips: []types.Address{
				{PublicIp: aws.String("1.2.3.4"), AllocationId: aws.String("eipalloc-111")},
				{PublicIp: aws.String("5.6.7.8"), AllocationId: aws.String("eipalloc-222")},
			},
			wantLen: 2,
		},
		{
			name: "ip_with_nil_fields",
			ips: []types.Address{
				{PublicIp: nil, AllocationId: nil},
			},
			wantLen: 1,
		},
		{
			name: "ip_with_only_public_ip",
			ips: []types.Address{
				{PublicIp: aws.String("10.0.0.1"), AllocationId: nil},
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := populateElasticIpRows(tt.ips)

			if len(rows) != tt.wantLen {
				t.Errorf("populateElasticIpRows() returned %d rows, want %d", len(rows), tt.wantLen)
				return
			}

			// Verify each row has 3 columns
			for i, row := range rows {
				if len(row) != 3 {
					t.Errorf("Row %d has %d columns, want 3", i, len(row))
				}
			}
		})
	}
}

func TestPopulateInstanceRows(t *testing.T) {
	tests := []struct {
		name      string
		instances []types.Instance
		wantLen   int
	}{
		{
			name:      "empty_instances",
			instances: []types.Instance{},
			wantLen:   0,
		},
		{
			name: "single_instance_with_valid_date",
			instances: []types.Instance{
				{
					InstanceId:            aws.String("i-12345"),
					StateTransitionReason: aws.String("User initiated (2024-01-01 00:00:00 UTC)"),
				},
			},
			wantLen: 1,
		},
		{
			name: "instance_with_nil_reason",
			instances: []types.Instance{
				{
					InstanceId:            aws.String("i-67890"),
					StateTransitionReason: nil,
				},
			},
			wantLen: 1,
		},
		{
			name: "instance_with_invalid_date",
			instances: []types.Instance{
				{
					InstanceId:            aws.String("i-abcde"),
					StateTransitionReason: aws.String("Unknown reason"),
				},
			},
			wantLen: 1,
		},
		{
			name: "multiple_instances",
			instances: []types.Instance{
				{InstanceId: aws.String("i-111"), StateTransitionReason: aws.String("User initiated (2024-01-01 00:00:00 UTC)")},
				{InstanceId: aws.String("i-222"), StateTransitionReason: nil},
				{InstanceId: aws.String("i-333"), StateTransitionReason: aws.String("invalid")},
			},
			wantLen: 3,
		},
		{
			name: "instance_with_nil_instance_id",
			instances: []types.Instance{
				{
					InstanceId:            nil,
					StateTransitionReason: aws.String("User initiated (2024-01-01 00:00:00 UTC)"),
				},
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := populateInstanceRows(tt.instances)

			if len(rows) != tt.wantLen {
				t.Errorf("populateInstanceRows() returned %d rows, want %d", len(rows), tt.wantLen)
				return
			}

			// Verify each row has 3 columns
			for i, row := range rows {
				if len(row) != 3 {
					t.Errorf("Row %d has %d columns, want 3", i, len(row))
				}
			}
		})
	}
}

func TestPopulateInstanceRows_TimeInfo(t *testing.T) {
	// Test that the time info is calculated correctly
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30).Format("2006-01-02 15:04:05") + " UTC"

	instances := []types.Instance{
		{
			InstanceId:            aws.String("i-test"),
			StateTransitionReason: aws.String("User initiated (" + thirtyDaysAgo + ")"),
		},
	}

	rows := populateInstanceRows(instances)

	if len(rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(rows))
	}

	// The time info should contain "days ago"
	timeInfo := rows[0][2].(string)
	if timeInfo == "-" {
		t.Error("Expected time info to be calculated, got '-'")
	}
}

func TestPopulateRiRows(t *testing.T) {
	tests := []struct {
		name    string
		ris     []model.RiExpirationInfo
		wantLen int
	}{
		{
			name:    "empty_ris",
			ris:     []model.RiExpirationInfo{},
			wantLen: 0,
		},
		{
			name: "single_ri_expiring_soon",
			ris: []model.RiExpirationInfo{
				{
					ReservedInstanceId: "ri-12345",
					DaysUntilExpiry:    15,
					Status:             "EXPIRING SOON",
				},
			},
			wantLen: 1,
		},
		{
			name: "single_ri_expired",
			ris: []model.RiExpirationInfo{
				{
					ReservedInstanceId: "ri-67890",
					DaysUntilExpiry:    -10,
					Status:             "EXPIRED",
				},
			},
			wantLen: 1,
		},
		{
			name: "multiple_ris",
			ris: []model.RiExpirationInfo{
				{ReservedInstanceId: "ri-111", DaysUntilExpiry: 30},
				{ReservedInstanceId: "ri-222", DaysUntilExpiry: 0},
				{ReservedInstanceId: "ri-333", DaysUntilExpiry: -5},
			},
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := populateRiRows(tt.ris)

			if len(rows) != tt.wantLen {
				t.Errorf("populateRiRows() returned %d rows, want %d", len(rows), tt.wantLen)
				return
			}

			// Verify each row has 3 columns
			for i, row := range rows {
				if len(row) != 3 {
					t.Errorf("Row %d has %d columns, want 3", i, len(row))
				}
			}
		})
	}
}

func TestPopulateRiRows_TimeInfo(t *testing.T) {
	tests := []struct {
		name            string
		daysUntilExpiry int
		wantContains    string
	}{
		{
			name:            "expiring_in_future",
			daysUntilExpiry: 15,
			wantContains:    "In 15 days",
		},
		{
			name:            "expired_in_past",
			daysUntilExpiry: -10,
			wantContains:    "10 days ago",
		},
		{
			name:            "expires_today",
			daysUntilExpiry: 0,
			wantContains:    "In 0 days",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ris := []model.RiExpirationInfo{
				{ReservedInstanceId: "ri-test", DaysUntilExpiry: tt.daysUntilExpiry},
			}

			rows := populateRiRows(ris)
			timeInfo := rows[0][2].(string)

			if timeInfo != tt.wantContains {
				t.Errorf("Time info = %q, want %q", timeInfo, tt.wantContains)
			}
		})
	}
}

func TestPopulateLoadBalancerRows(t *testing.T) {
	tests := []struct {
		name          string
		loadBalancers []elbtypes.LoadBalancer
		wantLen       int
	}{
		{
			name:          "empty_load_balancers",
			loadBalancers: []elbtypes.LoadBalancer{},
			wantLen:       0,
		},
		{
			name: "single_alb",
			loadBalancers: []elbtypes.LoadBalancer{
				{
					LoadBalancerName: aws.String("my-alb"),
					Type:             elbtypes.LoadBalancerTypeEnumApplication,
				},
			},
			wantLen: 1,
		},
		{
			name: "single_nlb",
			loadBalancers: []elbtypes.LoadBalancer{
				{
					LoadBalancerName: aws.String("my-nlb"),
					Type:             elbtypes.LoadBalancerTypeEnumNetwork,
				},
			},
			wantLen: 1,
		},
		{
			name: "multiple_load_balancers",
			loadBalancers: []elbtypes.LoadBalancer{
				{LoadBalancerName: aws.String("alb-1"), Type: elbtypes.LoadBalancerTypeEnumApplication},
				{LoadBalancerName: aws.String("nlb-1"), Type: elbtypes.LoadBalancerTypeEnumNetwork},
				{LoadBalancerName: aws.String("gwlb-1"), Type: elbtypes.LoadBalancerTypeEnumGateway},
			},
			wantLen: 3,
		},
		{
			name: "load_balancer_with_nil_name",
			loadBalancers: []elbtypes.LoadBalancer{
				{
					LoadBalancerName: nil,
					Type:             elbtypes.LoadBalancerTypeEnumApplication,
				},
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := populateLoadBalancerRows(tt.loadBalancers)

			if len(rows) != tt.wantLen {
				t.Errorf("populateLoadBalancerRows() returned %d rows, want %d", len(rows), tt.wantLen)
				return
			}

			// Verify each row has 3 columns
			for i, row := range rows {
				if len(row) != 3 {
					t.Errorf("Row %d has %d columns, want 3", i, len(row))
				}
			}
		})
	}
}

func TestPopulateLoadBalancerRows_Values(t *testing.T) {
	loadBalancers := []elbtypes.LoadBalancer{
		{
			LoadBalancerName: aws.String("test-alb"),
			Type:             elbtypes.LoadBalancerTypeEnumApplication,
		},
	}

	rows := populateLoadBalancerRows(loadBalancers)

	if len(rows) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(rows))
	}

	// Column 0 is status placeholder (empty)
	if rows[0][0] != "" {
		t.Errorf("Column 0 should be empty, got %v", rows[0][0])
	}

	// Column 1 is name
	if rows[0][1] != "test-alb" {
		t.Errorf("Column 1 = %v, want 'test-alb'", rows[0][1])
	}

	// Column 2 is type
	if rows[0][2] != "application" {
		t.Errorf("Column 2 = %v, want 'application'", rows[0][2])
	}
}

func BenchmarkPopulateEBSRows(b *testing.B) {
	volumes := make([]types.Volume, 50)
	for i := 0; i < 50; i++ {
		volumes[i] = types.Volume{
			VolumeId: aws.String("vol-" + string(rune('a'+i%26))),
			Size:     aws.Int32(int32(100 + i*10)),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		populateEBSRows(volumes)
	}
}

func BenchmarkPopulateInstanceRows(b *testing.B) {
	instances := make([]types.Instance, 20)
	for i := 0; i < 20; i++ {
		instances[i] = types.Instance{
			InstanceId:            aws.String("i-" + string(rune('a'+i%26))),
			StateTransitionReason: aws.String("User initiated (2024-01-15 10:00:00 UTC)"),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		populateInstanceRows(instances)
	}
}

// captureWasteOutput captures stdout during function execution
func captureWasteOutput(f func()) string {
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

func TestDrawWasteTable_NoWaste(t *testing.T) {
	output := captureWasteOutput(func() {
		DrawWasteTable("123456789012", nil, nil, nil, nil, nil, nil, nil, nil)
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
		DrawWasteTable("123456789012", elasticIPs, nil, nil, nil, nil, nil, nil, nil)
	})

	if !strings.Contains(output, "Elastic IP") {
		t.Error("DrawWasteTable() with elastic IPs missing Elastic IP section")
	}
}

func TestDrawWasteTable_WithEBSVolumes(t *testing.T) {
	unusedVolumes := []types.Volume{
		{VolumeId: aws.String("vol-123"), Size: aws.Int32(100)},
	}

	output := captureWasteOutput(func() {
		DrawWasteTable("123456789012", nil, unusedVolumes, nil, nil, nil, nil, nil, nil)
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
		DrawWasteTable("123456789012", nil, nil, nil, nil, stoppedInstances, nil, nil, nil)
	})

	if !strings.Contains(output, "EC2") || !strings.Contains(output, "Reserved Instance") {
		t.Error("DrawWasteTable() with stopped instances missing EC2 section")
	}
}

func TestDrawWasteTable_WithReservedInstances(t *testing.T) {
	ris := []model.RiExpirationInfo{
		{
			ReservedInstanceId: "ri-123",
			DaysUntilExpiry:    15,
			Status:             "EXPIRING SOON",
		},
	}

	output := captureWasteOutput(func() {
		DrawWasteTable("123456789012", nil, nil, nil, ris, nil, nil, nil, nil)
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
		},
	}

	output := captureWasteOutput(func() {
		DrawWasteTable("123456789012", nil, nil, nil, nil, nil, loadBalancers, nil, nil)
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
		{VolumeId: aws.String("vol-123"), Size: aws.Int32(100)},
	}
	stoppedVolumes := []types.Volume{
		{VolumeId: aws.String("vol-456"), Size: aws.Int32(200)},
	}
	ris := []model.RiExpirationInfo{
		{ReservedInstanceId: "ri-123", DaysUntilExpiry: 15, Status: "EXPIRING SOON"},
	}
	stoppedInstances := []types.Instance{
		{InstanceId: aws.String("i-123"), StateTransitionReason: aws.String("User initiated (2024-01-01 00:00:00 UTC)")},
	}
	loadBalancers := []elbtypes.LoadBalancer{
		{LoadBalancerName: aws.String("my-alb"), Type: elbtypes.LoadBalancerTypeEnumApplication},
	}

	output := captureWasteOutput(func() {
		DrawWasteTable("123456789012", elasticIPs, unusedVolumes, stoppedVolumes, ris, stoppedInstances, loadBalancers, nil, nil)
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
		{VolumeId: aws.String("vol-111"), Size: aws.Int32(100)},
		{VolumeId: aws.String("vol-222"), Size: aws.Int32(200)},
	}
	stoppedVolumes := []types.Volume{
		{VolumeId: aws.String("vol-333"), Size: aws.Int32(300)},
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
		{VolumeId: aws.String("vol-111"), Size: aws.Int32(100)},
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
		{VolumeId: aws.String("vol-333"), Size: aws.Int32(300)},
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
		{ReservedInstanceId: "ri-123", DaysUntilExpiry: 15, Status: "EXPIRING SOON"},
		{ReservedInstanceId: "ri-456", DaysUntilExpiry: -5, Status: "EXPIRED"},
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
		{ReservedInstanceId: "ri-123", DaysUntilExpiry: 15, Status: "EXPIRING SOON"},
	}

	output := captureWasteOutput(func() {
		drawEC2Table(nil, ris)
	})

	if !strings.Contains(output, "Expiring Soon") {
		t.Error("drawEC2Table() with only expiring RIs missing Expiring Soon status")
	}
}

func TestDrawElasticIpTable(t *testing.T) {
	elasticIPs := []types.Address{
		{PublicIp: aws.String("1.2.3.4"), AllocationId: aws.String("eipalloc-123")},
		{PublicIp: aws.String("5.6.7.8"), AllocationId: aws.String("eipalloc-456")},
	}

	output := captureWasteOutput(func() {
		drawElasticIpTable(elasticIPs)
	})

	if !strings.Contains(output, "Elastic IP Waste") {
		t.Error("drawElasticIpTable() missing title")
	}

	if !strings.Contains(output, "1.2.3.4") {
		t.Error("drawElasticIpTable() missing IP address")
	}

	if !strings.Contains(output, "eipalloc-123") {
		t.Error("drawElasticIpTable() missing allocation ID")
	}
}

func TestDrawLoadBalancerTable(t *testing.T) {
	loadBalancers := []elbtypes.LoadBalancer{
		{LoadBalancerName: aws.String("my-alb"), Type: elbtypes.LoadBalancerTypeEnumApplication},
		{LoadBalancerName: aws.String("my-nlb"), Type: elbtypes.LoadBalancerTypeEnumNetwork},
	}

	output := captureWasteOutput(func() {
		drawLoadBalancerTable(loadBalancers)
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

func TestPopulateSnapshotRows(t *testing.T) {
	tests := []struct {
		name      string
		snapshots []model.SnapshotWasteInfo
		wantLen   int
	}{
		{
			name:      "empty_snapshots",
			snapshots: []model.SnapshotWasteInfo{},
			wantLen:   0,
		},
		{
			name: "single_orphaned_snapshot",
			snapshots: []model.SnapshotWasteInfo{
				{
					SnapshotId:          "snap-12345",
					VolumeId:            "vol-deleted",
					SizeGB:              100,
					Category:            model.SnapshotCategoryOrphaned,
					Reason:              "Volume Deleted",
					MaxPotentialSavings: 5.0,
				},
			},
			wantLen: 1,
		},
		{
			name: "single_stale_snapshot",
			snapshots: []model.SnapshotWasteInfo{
				{
					SnapshotId:          "snap-67890",
					VolumeId:            "vol-exists",
					SizeGB:              200,
					Category:            model.SnapshotCategoryStale,
					Reason:              "Old Backup",
					MaxPotentialSavings: 10.0,
				},
			},
			wantLen: 1,
		},
		{
			name: "multiple_snapshots",
			snapshots: []model.SnapshotWasteInfo{
				{SnapshotId: "snap-111", SizeGB: 50, Reason: "Volume Deleted", MaxPotentialSavings: 2.5},
				{SnapshotId: "snap-222", SizeGB: 100, Reason: "Old Backup", MaxPotentialSavings: 5.0},
				{SnapshotId: "snap-333", SizeGB: 200, Reason: "Volume Deleted", MaxPotentialSavings: 10.0},
			},
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := populateSnapshotRows(tt.snapshots)

			if len(rows) != tt.wantLen {
				t.Errorf("populateSnapshotRows() returned %d rows, want %d", len(rows), tt.wantLen)
				return
			}

			// Verify each row has 5 columns (Status, Snapshot ID, Reason, Size, Max Savings)
			for i, row := range rows {
				if len(row) != 5 {
					t.Errorf("Row %d has %d columns, want 5", i, len(row))
				}
			}

			// Verify snapshot IDs are in the rows
			for i, snap := range tt.snapshots {
				if rows[i][1] != snap.SnapshotId {
					t.Errorf("Row %d SnapshotID = %v, want %v", i, rows[i][1], snap.SnapshotId)
				}
				if rows[i][2] != snap.Reason {
					t.Errorf("Row %d Reason = %v, want %v", i, rows[i][2], snap.Reason)
				}
			}
		})
	}
}

func TestDrawSnapshotTable(t *testing.T) {
	orphanedSnapshot := model.SnapshotWasteInfo{
		SnapshotId:          "snap-orphan1",
		VolumeId:            "vol-deleted",
		SizeGB:              100,
		Category:            model.SnapshotCategoryOrphaned,
		Reason:              "Volume Deleted",
		MaxPotentialSavings: 5.0,
	}
	staleSnapshot := model.SnapshotWasteInfo{
		SnapshotId:          "snap-stale1",
		VolumeId:            "vol-exists",
		SizeGB:              200,
		Category:            model.SnapshotCategoryStale,
		Reason:              "Old Backup",
		MaxPotentialSavings: 10.0,
	}

	allSnapshots := []model.SnapshotWasteInfo{orphanedSnapshot, staleSnapshot}

	output := captureWasteOutput(func() {
		drawSnapshotTable(allSnapshots)
	})

	if !strings.Contains(output, "EBS Snapshot Waste") {
		t.Error("drawSnapshotTable() missing title")
	}

	if !strings.Contains(output, "snap-orphan1") {
		t.Error("drawSnapshotTable() missing orphaned snapshot ID")
	}

	if !strings.Contains(output, "snap-stale1") {
		t.Error("drawSnapshotTable() missing stale snapshot ID")
	}

	if !strings.Contains(output, "Volume Deleted") {
		t.Error("drawSnapshotTable() missing 'Volume Deleted' reason")
	}

	if !strings.Contains(output, "Old Backup") {
		t.Error("drawSnapshotTable() missing 'Old Backup' reason")
	}

	if !strings.Contains(output, "Max Potential Savings") {
		t.Error("drawSnapshotTable() missing savings disclaimer")
	}
}

func TestDrawSnapshotTable_OnlyOrphaned(t *testing.T) {
	snapshots := []model.SnapshotWasteInfo{
		{
			SnapshotId:          "snap-orphan1",
			SizeGB:              100,
			Category:            model.SnapshotCategoryOrphaned,
			Reason:              "Volume Deleted",
			MaxPotentialSavings: 5.0,
		},
	}

	output := captureWasteOutput(func() {
		drawSnapshotTable(snapshots)
	})

	if !strings.Contains(output, "Orphaned") {
		t.Error("drawSnapshotTable() with only orphaned snapshots missing Orphaned status")
	}
}

func TestDrawSnapshotTable_OnlyStale(t *testing.T) {
	snapshots := []model.SnapshotWasteInfo{
		{
			SnapshotId:          "snap-stale1",
			SizeGB:              200,
			Category:            model.SnapshotCategoryStale,
			Reason:              "Old Backup",
			MaxPotentialSavings: 10.0,
		},
	}

	output := captureWasteOutput(func() {
		drawSnapshotTable(snapshots)
	})

	if !strings.Contains(output, "Stale") {
		t.Error("drawSnapshotTable() with only stale snapshots missing Stale status")
	}
}

func TestDrawWasteTable_WithSnapshots(t *testing.T) {
	snapshots := []model.SnapshotWasteInfo{
		{
			SnapshotId:          "snap-12345",
			SizeGB:              100,
			Category:            model.SnapshotCategoryOrphaned,
			Reason:              "Volume Deleted",
			MaxPotentialSavings: 5.0,
		},
	}

	output := captureWasteOutput(func() {
		DrawWasteTable("123456789012", nil, nil, nil, nil, nil, nil, nil, snapshots)
	})

	if !strings.Contains(output, "EBS Snapshot") {
		t.Error("DrawWasteTable() with snapshots missing Snapshot section")
	}
}

func BenchmarkDrawWasteTable(b *testing.B) {
	elasticIPs := []types.Address{
		{PublicIp: aws.String("1.2.3.4"), AllocationId: aws.String("eipalloc-123")},
	}
	unusedVolumes := []types.Volume{
		{VolumeId: aws.String("vol-123"), Size: aws.Int32(100)},
	}

	// Redirect stdout to discard
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DrawWasteTable("123456789012", elasticIPs, unusedVolumes, nil, nil, nil, nil, nil, nil)
	}
}
