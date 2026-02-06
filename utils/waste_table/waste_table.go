package wastetable

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/utils/ec2"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// DrawWasteTable renders a table containing detected AWS waste.
func DrawWasteTable(input model.RenderWasteInput) {
	fmt.Printf("\n%s\n", text.FgHiWhite.Sprint(" 🏥 AWS DOCTOR CHECKUP"))
	fmt.Printf(" Account ID: %s\n", text.FgBlue.Sprint(input.AccountID))
	fmt.Println(text.FgHiBlue.Sprint(" ------------------------------------------------"))

	hasWaste := len(input.ElasticIPs) > 0 ||
		len(input.UnusedVolumes) > 0 ||
		len(input.StoppedVolumes) > 0 ||
		len(input.StoppedInstances) > 0 ||
		len(input.Ris) > 0 ||
		len(input.LoadBalancers) > 0 ||
		len(input.UnusedAMIs) > 0 ||
		len(input.OrphanedSnapshots) > 0 ||
		len(input.UnusedKeyPairs) > 0

	if !hasWaste {
		fmt.Println("\n" + text.FgHiGreen.Sprint(" ✅  Your account is healthy! No waste found."))
		return
	}

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

	if len(input.UnusedAMIs) > 0 {
		drawAMITable(input.UnusedAMIs)
	}

	if len(input.OrphanedSnapshots) > 0 {
		drawSnapshotTable(input.OrphanedSnapshots)
	}

	if len(input.UnusedKeyPairs) > 0 {
		drawKeyPairTable(input.UnusedKeyPairs)
	}
}

func drawEBSTable(unusedEBSVolumeInfo []types.Volume, attachedToStoppedInstancesEBSVolumeInfo []types.Volume) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("EBS Volume Waste")

	t.AppendHeader(table.Row{"Status", "Volume ID", "Size (GiB)"})

	t.SetColumnConfigs([]table.ColumnConfig{
		{
			Number: 3,
			Align:  text.AlignRight,
		},
	})

	if len(unusedEBSVolumeInfo) > 0 {
		statusAvailable := "Available (Unattached)"
		rows := populateEBSRows(unusedEBSVolumeInfo)

		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiRed.Sprint(statusAvailable)

		t.AppendRows(rows)
	}

	if len(unusedEBSVolumeInfo) > 0 && len(attachedToStoppedInstancesEBSVolumeInfo) > 0 {
		t.AppendSeparator()
	}

	if len(attachedToStoppedInstancesEBSVolumeInfo) > 0 {
		statusStopped := "Attached to Stopped Instance"
		rows := populateEBSRows(attachedToStoppedInstancesEBSVolumeInfo)

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

	t.AppendHeader(table.Row{"Status", "IP Address", "Allocation ID"})

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

func populateEBSRows(volumes []types.Volume) []table.Row {
	var rows []table.Row

	for _, vol := range volumes {
		rows = append(rows, table.Row{
			"",
			*vol.VolumeId,
			fmt.Sprintf("%d GiB", *vol.Size),
		})
	}

	return rows
}

func populateElasticIPRows(ips []types.Address) []table.Row {
	var rows []table.Row

	for _, ip := range ips {
		publicIP := ""
		if ip.PublicIp != nil {
			publicIP = *ip.PublicIp
		}

		allocationID := ""
		if ip.AllocationId != nil {
			allocationID = *ip.AllocationId
		}

		rows = append(rows, table.Row{
			"",
			publicIP,
			allocationID,
		})
	}

	return rows
}

func populateInstanceRows(instances []types.Instance) []table.Row {
	var rows []table.Row

	now := time.Now()

	for _, instance := range instances {
		// Parse date for display
		reason := ""
		if instance.StateTransitionReason != nil {
			reason = *instance.StateTransitionReason
		}

		timeInfo := "-"

		stoppedAt, err := ec2.ParseTransitionDate(reason)
		if err == nil {
			days := int(now.Sub(stoppedAt).Hours() / 24)
			timeInfo = fmt.Sprintf("%d days ago", days)
		}

		instanceID := ""
		if instance.InstanceId != nil {
			instanceID = *instance.InstanceId
		}

		rows = append(rows, table.Row{
			"", // Placeholder for Status
			instanceID,
			timeInfo,
		})
	}

	return rows
}

func populateRiRows(ris []model.RiExpirationInfo) []table.Row {
	var rows []table.Row

	for _, ri := range ris {
		timeInfo := ""
		if ri.DaysUntilExpiry >= 0 {
			timeInfo = fmt.Sprintf("In %d days", ri.DaysUntilExpiry)
		} else {
			timeInfo = fmt.Sprintf("%d days ago", -ri.DaysUntilExpiry)
		}

		rows = append(rows, table.Row{
			"",
			ri.ReservedInstanceID,
			timeInfo,
		})
	}

	return rows
}

func drawLoadBalancerTable(loadBalancers []elbtypes.LoadBalancer) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("Load Balancer Waste")

	t.AppendHeader(table.Row{"Status", "Name", "Type"})

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
	var rows []table.Row

	for _, lb := range loadBalancers {
		name := aws.ToString(lb.LoadBalancerName)
		lbType := string(lb.Type)

		rows = append(rows, table.Row{
			"",
			name,
			lbType,
		})
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
	var rows []table.Row

	for _, ami := range amis {
		name := ami.Name
		if len(name) > 30 {
			name = name[:27] + "..."
		}

		rows = append(rows, table.Row{
			"",
			ami.ImageID,
			name,
			fmt.Sprintf("%d days", ami.DaysSinceCreate),
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
	var rows []table.Row

	for _, snap := range snapshots {
		rows = append(rows, table.Row{
			"",
			snap.SnapshotID,
			snap.Reason,
			fmt.Sprintf("%d GB", snap.SizeGB),
			fmt.Sprintf("$%.2f/mo", snap.MaxPotentialSavings),
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
	var rows []table.Row

	for _, kp := range keyPairs {
		rows = append(rows, table.Row{
			"",
			kp.KeyName,
			kp.KeyPairID,
			fmt.Sprintf("%d days", kp.DaysSinceCreate),
		})
	}

	return rows
}
