package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func DrawWasteTable(accountId string, elasticIpInfo []types.Address, unusedEBSVolumeInfo []types.Volume, attachedToStoppedInstancesEBSVolumeInfo []types.Volume, expireReservedInstancesInfo []model.RiExpirationInfo, instancesStoppedMoreThan30Days []types.Instance) {
	if len(unusedEBSVolumeInfo) > 0 || len(attachedToStoppedInstancesEBSVolumeInfo) > 0 {
		drawEBSTable(unusedEBSVolumeInfo, attachedToStoppedInstancesEBSVolumeInfo)
	}

	if len(elasticIpInfo) > 0 {
		drawElasticIpTable(elasticIpInfo)
	}

	if len(instancesStoppedMoreThan30Days) > 0 || len(expireReservedInstancesInfo) > 0 {
		drawEC2Table(instancesStoppedMoreThan30Days, expireReservedInstancesInfo)
	}
}

func drawEBSTable(unusedEBSVolumeInfo []types.Volume, attachedToStoppedInstancesEBSVolumeInfo []types.Volume) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("🗑EBS Volume Waste")

	t.AppendHeader(table.Row{"Status", "Volume ID", "Size (GiB)"})

	t.SetColumnConfigs([]table.ColumnConfig{
		{
			Number: 3,
			Align:  text.AlignRight,
		},
	})

	statusAvailable := "Available (Unattached)"
	rows := populateEBSRows(unusedEBSVolumeInfo)

	halfRow := len(rows) / 2
	rows[halfRow][0] = text.FgHiRed.Sprint(statusAvailable)

	t.AppendRows(rows)

	rows = []table.Row{}

	if len(unusedEBSVolumeInfo) > 0 && len(attachedToStoppedInstancesEBSVolumeInfo) > 0 {
		t.AppendSeparator()
	}

	statusStopped := "Attached to Stopped Instance"
	rows = populateEBSRows(attachedToStoppedInstancesEBSVolumeInfo)

	halfRow = len(rows) / 2
	rows[halfRow][0] = text.FgHiRed.Sprint(statusStopped)

	t.AppendRows(rows)

	if t.Length() > 0 {
		t.Render()
		fmt.Println()
	}
}

func drawEC2Table(instances []types.Instance, ris []model.RiExpirationInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("🗑EC2 & Reserved Instance Waste")

	t.AppendHeader(table.Row{"Status", "Instance ID", "Time Info"})

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 3, Align: text.AlignRight},
	})

	var hasPreviousRows bool

	if len(instances) > 0 {
		statusLabel := "Stopped Instance(> 30 Days)"
		rows := populateInstanceRows(instances)

		// Apply grouped status label
		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiRed.Sprint(statusLabel)

		t.AppendRows(rows)
		hasPreviousRows = true
	}

	// --- SECTION 2 & 3: Reserved Instances ---
	if len(ris) > 0 {
		// We split RIs into two groups for better status labeling
		var expiring, expired []model.RiExpirationInfo
		for _, ri := range ris {
			if ri.Status == "EXPIRING SOON" {
				expiring = append(expiring, ri)
			} else {
				expired = append(expired, ri)
			}
		}

		// Group: Expiring Soon
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

		// Group: Recently Expired
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

func drawElasticIpTable(elasticIpInfo []types.Address) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("🗑Elastic IP Waste")

	t.AppendHeader(table.Row{"Status", "IP Address", "Allocation ID"})

	statusUnused := "Unassociated"
	rows := populateElasticIpRows(elasticIpInfo)

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

func populateElasticIpRows(ips []types.Address) []table.Row {
	var rows []table.Row

	for _, ip := range ips {
		publicIp := ""
		if ip.PublicIp != nil {
			publicIp = *ip.PublicIp
		}

		allocationId := ""
		if ip.AllocationId != nil {
			allocationId = *ip.AllocationId
		}

		rows = append(rows, table.Row{
			"",
			publicIp,
			allocationId,
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
		stoppedAt, err := ParseTransitionDate(reason)
		if err == nil {
			days := int(now.Sub(stoppedAt).Hours() / 24)
			timeInfo = fmt.Sprintf("%d days ago", days)
		}

		instanceId := ""
		if instance.InstanceId != nil {
			instanceId = *instance.InstanceId
		}

		rows = append(rows, table.Row{
			"", // Placeholder for Status
			instanceId,
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
			ri.ReservedInstanceId,
			timeInfo,
		})
	}
	return rows
}
