package wastetable

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
	outputshared "github.com/elC0mpa/aws-doctor/utils/output_shared"
	wastesummary "github.com/elC0mpa/aws-doctor/utils/waste_summary"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// DrawWasteTable renders a table containing detected AWS waste.
func DrawWasteTable(input model.RenderWasteInput) {
	drawHeader(input.AccountID)

	if !hasAnyWaste(input) {
		fmt.Println("\n" + text.FgHiGreen.Sprint(" ✅  Your account is healthy! No waste found."))
		return
	}

	drawWasteSections(input)
	drawSummaryTable(input)
}

func drawHeader(accountID string) {
	fmt.Printf("\n%s\n", text.FgHiWhite.Sprint(" 🏥 AWS DOCTOR CHECKUP"))
	fmt.Printf(" Account ID: %s\n", text.FgBlue.Sprint(accountID))
	fmt.Println(text.FgHiBlue.Sprint(" ------------------------------------------------"))
}

func hasAnyWaste(input model.RenderWasteInput) bool {
	return len(input.ElasticIPs) > 0 ||
		len(input.UnusedVolumes) > 0 ||
		len(input.StoppedVolumes) > 0 ||
		len(input.StoppedInstances) > 0 ||
		len(input.Ris) > 0 ||
		len(input.LoadBalancers) > 0 ||
		len(input.UnusedAMIs) > 0 ||
		len(input.OrphanedSnapshots) > 0 ||
		len(input.UnusedKeyPairs) > 0 ||
		len(input.S3Buckets) > 0 ||
		len(input.S3MultipartUploads) > 0 ||
		len(input.CloudWatchLogGroups) > 0 ||
		len(input.RDSInstances) > 0 ||
		len(input.RDSSnapshots) > 0 ||
		len(input.RDSIdleInstances) > 0
}

func drawWasteSections(input model.RenderWasteInput) {
	if len(input.UnusedVolumes) > 0 || len(input.StoppedVolumes) > 0 {
		drawEBSTable(input.UnusedVolumes, input.StoppedVolumes)
	}

	if len(input.ElasticIPs) > 0 {
		drawElasticIPTable(input.ElasticIPs)
	}

	if len(input.StoppedInstances) > 0 || len(input.Ris) > 0 {
		drawEC2Table(input.StoppedInstances, input.Ris)
	}

	if len(input.LoadBalancers) > 0 {
		drawLoadBalancerTable(input.LoadBalancers)
	}

	if len(input.S3Buckets) > 0 || len(input.S3MultipartUploads) > 0 {
		drawS3Table(input.S3Buckets, input.S3MultipartUploads)
	}

	if len(input.CloudWatchLogGroups) > 0 {
		drawCloudWatchLogsTable(input.CloudWatchLogGroups)
	}

	if len(input.UnusedAMIs) > 0 {
		drawAMITable(input.UnusedAMIs)
	}

	if len(input.OrphanedSnapshots) > 0 {
		drawSnapshotTable(input.OrphanedSnapshots)
	}

	if len(input.UnusedKeyPairs) > 0 {
		drawKeyPairTable(input.UnusedKeyPairs)
	}

	if len(input.RDSInstances) > 0 || len(input.RDSSnapshots) > 0 || len(input.RDSIdleInstances) > 0 {
		drawRDSTable(input.RDSInstances, input.RDSSnapshots, input.RDSIdleInstances)
	}
}

func drawEBSTable(unusedEBSVolumeInfo []types.Volume, attachedToStoppedInstancesEBSVolumeInfo []types.Volume) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("EBS Volume Waste")

	t.AppendHeader(table.Row{"Status", "Volume ID", "Size (GiB)", "Est. Cost/Mo"})

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 3, Align: text.AlignRight},
		{Number: 4, Align: text.AlignRight},
	})

	if len(unusedEBSVolumeInfo) > 0 {
		statusAvailable := "Available (Unattached)"
		rows := populateEBSRows(unusedEBSVolumeInfo, "unattached")

		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiRed.Sprint(statusAvailable)

		t.AppendRows(rows)
	}

	if len(unusedEBSVolumeInfo) > 0 && len(attachedToStoppedInstancesEBSVolumeInfo) > 0 {
		t.AppendSeparator()
	}

	if len(attachedToStoppedInstancesEBSVolumeInfo) > 0 {
		statusStopped := "Attached to Stopped Instance"
		rows := populateEBSRows(attachedToStoppedInstancesEBSVolumeInfo, "stopped")

		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiRed.Sprint(statusStopped)

		t.AppendRows(rows)
	}

	if t.Length() > 0 {
		t.Render()
		fmt.Println()
	}
}

func drawEC2Table(instances []types.Instance, ris []model.RiExpirationInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("EC2 & Reserved Instance Waste")

	t.AppendHeader(table.Row{"Status", "Instance ID", "Time Info"})

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 3, Align: text.AlignRight},
	})

	var hasPreviousRows bool

	if len(instances) > 0 {
		statusLabel := "Stopped Instance(> 30 Days)"
		rows := populateInstanceRows(instances)

		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiRed.Sprint(statusLabel)

		t.AppendRows(rows)

		hasPreviousRows = true
	}

	if len(ris) > 0 {
		var expiring, expired []model.RiExpirationInfo

		for _, ri := range ris {
			if ri.Status == "EXPIRING SOON" {
				expiring = append(expiring, ri)
			} else {
				expired = append(expired, ri)
			}
		}

		if len(expiring) > 0 {
			if hasPreviousRows {
				t.AppendSeparator()
			}

			statusLabel := "Reserved Instance\n(Expiring Soon)"
			rows := populateRiRows(expiring)

			halfRow := len(rows) / 2
			rows[halfRow][0] = text.FgHiYellow.Sprint(statusLabel)

			t.AppendRows(rows)

			hasPreviousRows = true
		}

		if len(expired) > 0 {
			if hasPreviousRows {
				t.AppendSeparator()
			}

			statusLabel := "Reserved Instance\n(Recently Expired)"
			rows := populateRiRows(expired)

			halfRow := len(rows) / 2
			rows[halfRow][0] = text.FgHiRed.Sprint(statusLabel)

			t.AppendRows(rows)
		}
	}

	t.Render()
	fmt.Println()
}

func drawElasticIPTable(elasticIPInfo []types.Address) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("Elastic IP Waste")

	t.AppendHeader(table.Row{"Status", "IP Address", "Allocation ID", "Est. Cost/Mo"})

	statusUnused := "Unassociated"
	rows := populateElasticIPRows(elasticIPInfo)

	if len(rows) > 0 {
		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiRed.Sprint(statusUnused)
	}

	t.AppendRows(rows)
	t.Render()
	fmt.Println()
}

func populateEBSRows(volumes []types.Volume, status string) []table.Row {
	rows := make([]table.Row, 0, len(volumes))

	for _, vol := range volumes {
		p := outputshared.PresentEBSVolume(vol, status)
		rows = append(rows, table.Row{"", p.Identifier, p.Metric, p.EstimatedCost})
	}

	return rows
}

func populateElasticIPRows(ips []types.Address) []table.Row {
	rows := make([]table.Row, 0, len(ips))

	for _, ip := range ips {
		p := outputshared.PresentElasticIP(ip)
		rows = append(rows, table.Row{"", p.Identifier, aws.ToString(ip.AllocationId), p.EstimatedCost})
	}

	return rows
}

func populateInstanceRows(instances []types.Instance) []table.Row {
	rows := make([]table.Row, 0, len(instances))

	for _, instance := range instances {
		p := outputshared.PresentStoppedInstance(instance)
		timeInfo := outputshared.NAValue

		if p.Age != outputshared.NAValue {
			timeInfo = p.Age + " days ago"
		}

		rows = append(rows, table.Row{"", p.Identifier, timeInfo})
	}

	return rows
}

func populateRiRows(ris []model.RiExpirationInfo) []table.Row {
	rows := make([]table.Row, 0, len(ris))

	for _, ri := range ris {
		p := outputshared.PresentReservedInstance(ri)
		timeInfo := ""

		days := ri.DaysUntilExpiry
		if days >= 0 {
			timeInfo = fmt.Sprintf("In %d days", days)
		} else {
			timeInfo = fmt.Sprintf("%d days ago", -days)
		}

		rows = append(rows, table.Row{"", p.Identifier, timeInfo})
	}

	return rows
}

func drawLoadBalancerTable(loadBalancers []elbtypes.LoadBalancer) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("Load Balancer Waste")

	t.AppendHeader(table.Row{"Status", "Name", "Type", "Est. Cost/Mo"})

	statusUnused := "No Target Groups"
	rows := populateLoadBalancerRows(loadBalancers)

	if len(rows) > 0 {
		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiRed.Sprint(statusUnused)
	}

	t.AppendRows(rows)
	t.Render()
	fmt.Println()
}

func populateLoadBalancerRows(loadBalancers []elbtypes.LoadBalancer) []table.Row {
	rows := make([]table.Row, 0, len(loadBalancers))

	for _, lb := range loadBalancers {
		p := outputshared.PresentLoadBalancer(lb)
		// Details: "Created on 2026-03-21T08:00:00Z"
		// We want the name which is not in ResourceRow, oh wait, Identifier is LoadBalancerArn
		// In DrawLoadBalancerTable it used lb.LoadBalancerName
		rows = append(rows, table.Row{"", aws.ToString(lb.LoadBalancerName), p.Metric, p.EstimatedCost})
	}

	return rows
}

func drawAMITable(amis []model.AMIWasteInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("Unused AMI Waste (Verify before delete - may be used by ASGs/Launch Templates)")

	t.AppendHeader(table.Row{"Status", "AMI ID", "Name", "Age (Days)", "Max Savings/Mo"})

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 4, Align: text.AlignRight},
		{Number: 5, Align: text.AlignRight},
	})

	rows := populateAMIRows(amis)

	if len(rows) > 0 {
		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiYellow.Sprint("Unused*")
	}

	t.AppendRows(rows)
	t.Render()
	fmt.Println(text.FgHiYellow.Sprint(" * Warning: AMIs may be referenced by Auto Scaling Groups or Launch Templates"))
	fmt.Println()
}

func populateAMIRows(amis []model.AMIWasteInfo) []table.Row {
	rows := make([]table.Row, 0, len(amis))

	for _, ami := range amis {
		name := ami.Name
		if len(name) > 30 {
			name = name[:27] + "..."
		}

		p := outputshared.PresentAMI(ami)

		rows = append(rows, table.Row{
			"",
			p.Identifier,
			name,
			p.Age + " days",
			fmt.Sprintf("$%.2f", ami.MaxPotentialSaving),
		})
	}

	return rows
}

func drawSnapshotTable(snapshots []model.SnapshotWasteInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("EBS Snapshot Waste")

	t.AppendHeader(table.Row{"Status", "Snapshot ID", "Reason", "Size (GB)", "Max Savings/MO"})

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 4, Align: text.AlignRight},
		{Number: 5, Align: text.AlignRight},
	})

	// Separate orphaned and stale snapshots
	var orphaned, stale []model.SnapshotWasteInfo

	for _, snap := range snapshots {
		if snap.Category == model.SnapshotCategoryOrphaned {
			orphaned = append(orphaned, snap)
		} else {
			stale = append(stale, snap)
		}
	}

	var hasPreviousRows bool

	if len(orphaned) > 0 {
		statusLabel := "Orphaned(Volume Deleted)"
		rows := populateSnapshotRows(orphaned)

		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiRed.Sprint(statusLabel)

		t.AppendRows(rows)

		hasPreviousRows = true
	}

	if len(stale) > 0 {
		if hasPreviousRows {
			t.AppendSeparator()
		}

		statusLabel := "Stale(Old Backup > 90 days)"
		rows := populateSnapshotRows(stale)

		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiYellow.Sprint(statusLabel)

		t.AppendRows(rows)
	}

	t.Render()
	fmt.Println()
}

func populateSnapshotRows(snapshots []model.SnapshotWasteInfo) []table.Row {
	rows := make([]table.Row, 0, len(snapshots))

	for _, snap := range snapshots {
		p := outputshared.PresentSnapshot(snap)
		rows = append(rows, table.Row{
			"",
			p.Identifier,
			snap.Reason,
			p.Metric,
			p.EstimatedCost + "/mo",
		})
	}

	return rows
}

func drawKeyPairTable(keyPairs []model.KeyPairWasteInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("Unused EC2 Key Pair Waste")

	t.AppendHeader(table.Row{"Status", "Key Name", "Key Pair ID", "Age (Days)"})

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 4, Align: text.AlignRight},
	})

	statusUnused := "Unused"
	rows := populateKeyPairRows(keyPairs)

	if len(rows) > 0 {
		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiRed.Sprint(statusUnused)
	}

	t.AppendRows(rows)
	t.Render()
	fmt.Println()
}

func populateKeyPairRows(keyPairs []model.KeyPairWasteInfo) []table.Row {
	rows := make([]table.Row, 0, len(keyPairs))

	for _, kp := range keyPairs {
		p := outputshared.PresentKeyPair(kp)
		rows = append(rows, table.Row{
			"",
			p.Identifier,
			kp.KeyPairID,
			p.Age + " days",
		})
	}

	return rows
}

func drawS3Table(buckets []model.S3BucketWasteInfo, multipartBuckets []model.S3MultipartUploadWasteInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("S3 Bucket Waste")

	t.AppendHeader(table.Row{"Status", "Bucket Name", "Info"})

	var hasPreviousRows bool

	if len(buckets) > 0 {
		statusLabel := "No Lifecycle Policy"
		rows := populateS3Rows(buckets)

		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiRed.Sprint(statusLabel)

		t.AppendRows(rows)

		hasPreviousRows = true
	}

	if len(multipartBuckets) > 0 {
		if hasPreviousRows {
			t.AppendSeparator()
		}

		statusLabel := "Incomplete Multipart"
		rows := populateS3MultipartRows(multipartBuckets)

		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiRed.Sprint(statusLabel)

		t.AppendRows(rows)
	}

	t.Render()
	fmt.Println()
}

func populateS3Rows(buckets []model.S3BucketWasteInfo) []table.Row {
	rows := make([]table.Row, 0, len(buckets))

	for _, bucket := range buckets {
		p := outputshared.PresentS3Bucket(bucket)
		rows = append(rows, table.Row{
			"",
			p.Identifier,
			fmt.Sprintf("Created on %s", bucket.CreationDate.Format("2006-01-02")),
		})
	}

	return rows
}

func populateS3MultipartRows(buckets []model.S3MultipartUploadWasteInfo) []table.Row {
	rows := make([]table.Row, 0, len(buckets))

	for _, bucket := range buckets {
		p := outputshared.PresentS3MultipartUpload(bucket)
		rows = append(rows, table.Row{
			"",
			p.Identifier,
			p.Metric,
		})
	}

	return rows
}

func drawCloudWatchLogsTable(logGroups []model.CloudWatchLogsWasteInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("CloudWatch Log Group Waste")

	t.AppendHeader(table.Row{"Status", "Log Group Name", "Size", "Created On", "Est. Cost/Mo"})

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 3, Align: text.AlignRight},
		{Number: 5, Align: text.AlignRight},
	})

	statusLabel := "No Retention Policy"
	rows := populateCloudWatchLogsRows(logGroups)

	if len(rows) > 0 {
		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiRed.Sprint(statusLabel)
	}

	t.AppendRows(rows)
	t.Render()
	fmt.Println()
}

func populateCloudWatchLogsRows(logGroups []model.CloudWatchLogsWasteInfo) []table.Row {
	rows := make([]table.Row, 0, len(logGroups))

	for _, lg := range logGroups {
		p := outputshared.PresentCloudWatchLogGroup(lg)
		rows = append(rows, table.Row{
			"",
			p.Identifier,
			p.Metric,
			lg.CreationTime.Format("2006-01-02"),
			p.EstimatedCost,
		})
	}

	return rows
}

func drawSummaryTable(input model.RenderWasteInput) {
	categories, totalCost := wastesummary.Compute(input)

	if len(categories) == 0 {
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("Waste Summary")

	t.AppendHeader(table.Row{"Category", "Count", "Est. Monthly Cost"})
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 2, Align: text.AlignRight},
		{Number: 3, Align: text.AlignRight},
	})

	for _, cat := range categories {
		costStr := outputshared.NAValue
		if cat.Cost > 0 {
			costStr = fmt.Sprintf("$%.2f", cat.Cost)
		}

		t.AppendRow(table.Row{cat.Name, cat.Count, costStr})
	}

	t.AppendSeparator()
	t.AppendRow(table.Row{
		text.FgHiWhite.Sprint("Total Estimated Monthly Waste"),
		"",
		text.FgHiRed.Sprintf("$%.2f", totalCost),
	})

	t.Render()
	fmt.Println(text.FgHiYellow.Sprint(" * Estimates based on us-east-1 pricing. Actual costs may vary by region."))
	fmt.Println()
}

func drawRDSTable(instances []model.RDSInstanceWasteInfo, snapshots []model.RDSSnapshotWasteInfo, idleInstances []model.RDSIdleInstanceInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("RDS Waste")

	t.AppendHeader(table.Row{"Status", "Identifier", "Engine", "Info", "Est. Cost/Mo"})

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 5, Align: text.AlignRight},
	})

	var hasPreviousRows bool

	if len(instances) > 0 {
		statusLabel := "Stopped Instance"
		rows := populateRDSInstanceRows(instances)

		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiRed.Sprint(statusLabel)

		t.AppendRows(rows)

		hasPreviousRows = true
	}

	if len(snapshots) > 0 {
		if hasPreviousRows {
			t.AppendSeparator()
		}

		statusLabel := "Old Snapshot (> 30 days)"
		rows := populateRDSSnapshotRows(snapshots)

		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiYellow.Sprint(statusLabel)

		t.AppendRows(rows)

		hasPreviousRows = true
	}

	if len(idleInstances) > 0 {
		if hasPreviousRows {
			t.AppendSeparator()
		}

		statusLabel := "Idle (0 connections)"
		rows := populateRDSIdleRows(idleInstances)

		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiYellow.Sprint(statusLabel)

		t.AppendRows(rows)
	}

	t.Render()
	fmt.Println()
}

func populateRDSIdleRows(instances []model.RDSIdleInstanceInfo) []table.Row {
	rows := make([]table.Row, 0, len(instances))

	for _, inst := range instances {
		p := outputshared.PresentRDSIdleInstance(inst)
		rows = append(rows, table.Row{
			"",
			p.Identifier,
			inst.Engine,
			p.Metric,
			p.EstimatedCost,
		})
	}

	return rows
}

func populateRDSInstanceRows(instances []model.RDSInstanceWasteInfo) []table.Row {
	rows := make([]table.Row, 0, len(instances))

	for _, inst := range instances {
		p := outputshared.PresentRDSInstance(inst)
		rows = append(rows, table.Row{
			"",
			p.Identifier,
			inst.Engine,
			p.Metric,
			p.EstimatedCost,
		})
	}

	return rows
}

func populateRDSSnapshotRows(snapshots []model.RDSSnapshotWasteInfo) []table.Row {
	rows := make([]table.Row, 0, len(snapshots))

	for _, snap := range snapshots {
		p := outputshared.PresentRDSSnapshot(snap)
		rows = append(rows, table.Row{
			"",
			p.Identifier,
			snap.Engine,
			fmt.Sprintf("%s days old, %d GB", p.Age, snap.AllocatedStorage),
			p.EstimatedCost,
		})
	}

	return rows
}
