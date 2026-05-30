package wastetable

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	elbtypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/pricing"
	outputshared "github.com/elC0mpa/aws-doctor/utils/output_shared"
	wastesummary "github.com/elC0mpa/aws-doctor/utils/waste_summary"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// DrawWasteTable renders a table containing detected AWS waste.
func DrawWasteTable(out io.Writer, input model.RenderWasteInput, pricingSvc pricing.Service) {
	drawHeader(out, input.AccountID)

	if !hasAnyWaste(input) {
		if len(input.Errors) == 0 {
			_, _ = fmt.Fprintln(out, "\n"+text.FgHiGreen.Sprint(" ✅  Your account is healthy! No waste found."))
			return
		}

		_, _ = fmt.Fprintln(out, "\n"+text.FgYellow.Sprint(" ⚠️  No waste found, but some checks failed due to errors."))
	} else {
		drawWasteSections(out, input, pricingSvc)
		drawSummaryTable(out, input, pricingSvc)
	}

	if len(input.Errors) > 0 {
		drawErrorsSection(out, input.Errors)
	}
}

func drawHeader(out io.Writer, accountID string) {
	_, _ = fmt.Fprintf(out, "\n%s\n", text.FgHiWhite.Sprint(" 🏥 AWS DOCTOR CHECKUP"))
	_, _ = fmt.Fprintf(out, " Account ID: %s\n", text.FgBlue.Sprint(accountID))
	_, _ = fmt.Fprintln(out, text.FgHiBlue.Sprint(" ------------------------------------------------"))
}

func hasAnyWaste(input model.RenderWasteInput) bool {
	return hasEC2Waste(input) ||
		hasRDSWaste(input) ||
		hasS3Waste(input) ||
		hasNetworkWaste(input) ||
		hasServerlessWaste(input) ||
		hasECRWaste(input) ||
		len(input.CloudWatchLogGroups) > 0 ||
		len(input.UnusedSecrets) > 0 ||
		hasIAMWaste(input)
}

func hasEC2Waste(input model.RenderWasteInput) bool {
	return len(input.UnusedVolumes) > 0 ||
		len(input.StoppedVolumes) > 0 ||
		len(input.StoppedInstances) > 0 ||
		len(input.IdleEC2Instances) > 0 ||
		len(input.Ris) > 0 ||
		len(input.UnusedAMIs) > 0 ||
		len(input.OrphanedSnapshots) > 0 ||
		len(input.UnusedKeyPairs) > 0
}

func hasRDSWaste(input model.RenderWasteInput) bool {
	return len(input.RDSInstances) > 0 ||
		len(input.RDSSnapshots) > 0 ||
		len(input.RDSIdleInstances) > 0
}

func hasS3Waste(input model.RenderWasteInput) bool {
	return len(input.S3Buckets) > 0 ||
		len(input.S3MultipartUploads) > 0
}

func hasNetworkWaste(input model.RenderWasteInput) bool {
	return len(input.ElasticIPs) > 0 ||
		len(input.LoadBalancers) > 0 ||
		len(input.IdleNATGateways) > 0 ||
		len(input.IdleLoadBalancers) > 0
}

func hasServerlessWaste(input model.RenderWasteInput) bool {
	return len(input.OverProvisionedLambdas) > 0 ||
		len(input.IdleSageMakerEndpoints) > 0
}

func hasECRWaste(input model.RenderWasteInput) bool {
	return len(input.ECREmptyRepositories) > 0 ||
		len(input.ECRNoLifecyclePolicies) > 0 ||
		len(input.ECRUntaggedImages) > 0
}

func drawWasteSections(out io.Writer, input model.RenderWasteInput, pricingSvc pricing.Service) {
	drawStorageSections(out, input, pricingSvc)
	drawNetworkSections(out, input, pricingSvc)
	drawComputeSections(out, input)
	drawDatabaseSections(out, input)
	drawServerlessSections(out, input)
	drawContainerSections(out, input)
	drawSecretsManagerSections(out, input, pricingSvc)
	drawIAMSections(out, input)
}

func hasIAMWaste(input model.RenderWasteInput) bool {
	return len(input.UnusedIAMUsers) > 0 || len(input.RootUserWaste) > 0
}

func drawIAMSections(out io.Writer, input model.RenderWasteInput) {
	drawIAMTable(out, input.UnusedIAMUsers, input.RootUserWaste)
}

func drawStorageSections(out io.Writer, input model.RenderWasteInput, pricingSvc pricing.Service) {
	drawEBSTable(out, input.UnusedVolumes, input.StoppedVolumes, pricingSvc)

	drawS3Table(out, input.S3Buckets, input.S3MultipartUploads)

	drawSnapshotTable(out, input.OrphanedSnapshots)
}

func drawNetworkSections(out io.Writer, input model.RenderWasteInput, pricingSvc pricing.Service) {
	drawElasticIPTable(out, input.ElasticIPs, pricingSvc)

	drawLoadBalancerTable(out, input.LoadBalancers, input.IdleLoadBalancers, pricingSvc)

	drawNatGatewayTable(out, input.IdleNATGateways)
}

func drawComputeSections(out io.Writer, input model.RenderWasteInput) {
	drawEC2Table(out, input.StoppedInstances, input.Ris)

	drawIdleEC2Table(out, input.IdleEC2Instances)

	drawAMITable(out, input.UnusedAMIs)

	drawKeyPairTable(out, input.UnusedKeyPairs)

	drawCloudWatchLogsTable(out, input.CloudWatchLogGroups)
}

func drawIdleEC2Table(out io.Writer, instances []model.EC2IdleInstanceInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(out)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("Idle EC2 Instance Waste")

	t.AppendHeader(table.Row{"Status", "Identifier", "Type", "Utilization", "Est. Cost/Mo"})

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 5, Align: text.AlignRight},
	})

	rows := populateIdleEC2Rows(instances)

	if len(rows) > 0 {
		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiRed.Sprint("Idle (low CPU/network)")
	}

	t.AppendRows(rows)

	if t.Length() > 0 {
		t.Render()

		_, _ = fmt.Fprintln(out)
	}
}

func populateIdleEC2Rows(instances []model.EC2IdleInstanceInfo) []table.Row {
	rows := make([]table.Row, 0, len(instances))

	for _, inst := range instances {
		p := outputshared.PresentIdleEC2Instance(inst)
		rows = append(rows, table.Row{
			"",
			p.Identifier,
			inst.InstanceType,
			p.Metric,
			p.EstimatedCost,
		})
	}

	return rows
}

func drawDatabaseSections(out io.Writer, input model.RenderWasteInput) {
	drawRDSTable(out, input.RDSInstances, input.RDSSnapshots, input.RDSIdleInstances)
}

func drawServerlessSections(out io.Writer, input model.RenderWasteInput) {
	drawLambdaTable(out, input.OverProvisionedLambdas)

	drawSageMakerTable(out, input.IdleSageMakerEndpoints)
}

func drawContainerSections(out io.Writer, input model.RenderWasteInput) {
	drawECRTable(out, input.ECRNoLifecyclePolicies, input.ECREmptyRepositories, input.ECRUntaggedImages)
}

func drawSecretsManagerSections(out io.Writer, input model.RenderWasteInput, pricingSvc pricing.Service) {
	drawSecretsManagerTable(out, input.UnusedSecrets, pricingSvc)
}

func drawSecretsManagerTable(out io.Writer, secrets []model.UnusedSecretInfo, pricingSvc pricing.Service) {
	t := table.NewWriter()
	t.SetOutputMirror(out)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("Secrets Manager Waste")

	t.AppendHeader(table.Row{"Secret Name", "Last Accessed", "Est. Cost/Mo"})

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 3, Align: text.AlignRight},
	})

	for _, s := range secrets {
		row := outputshared.PresentUnusedSecret(s, pricingSvc)
		t.AppendRow(table.Row{row.Identifier, row.Age + " days ago", row.EstimatedCost})
	}

	if t.Length() > 0 {
		t.Render()

		_, _ = fmt.Fprintln(out)
	}
}

func drawEBSTable(out io.Writer, unusedEBSVolumeInfo []ec2types.Volume, attachedToStoppedInstancesEBSVolumeInfo []ec2types.Volume, pricingSvc pricing.Service) {
	t := table.NewWriter()
	t.SetOutputMirror(out)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("EBS Volume Waste")

	t.AppendHeader(table.Row{"Status", "Volume ID", "Size (GiB)", "Est. Cost/Mo"})

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 3, Align: text.AlignRight},
		{Number: 4, Align: text.AlignRight},
	})

	if len(unusedEBSVolumeInfo) > 0 {
		statusAvailable := "Available (Unattached)"
		rows := populateEBSRows(unusedEBSVolumeInfo, "unattached", pricingSvc)

		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiRed.Sprint(statusAvailable)

		t.AppendRows(rows)
	}

	if len(unusedEBSVolumeInfo) > 0 && len(attachedToStoppedInstancesEBSVolumeInfo) > 0 {
		t.AppendSeparator()
	}

	if len(attachedToStoppedInstancesEBSVolumeInfo) > 0 {
		statusStopped := "Attached to Stopped Instance"
		rows := populateEBSRows(attachedToStoppedInstancesEBSVolumeInfo, "stopped", pricingSvc)

		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiRed.Sprint(statusStopped)

		t.AppendRows(rows)
	}

	if t.Length() > 0 {
		t.Render()

		_, _ = fmt.Fprintln(out)
	}
}

func drawEC2Table(out io.Writer, instances []ec2types.Instance, ris []model.RiExpirationInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(out)
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

	if t.Length() > 0 {
		t.Render()

		_, _ = fmt.Fprintln(out)
	}
}

func drawElasticIPTable(out io.Writer, elasticIPInfo []ec2types.Address, pricingSvc pricing.Service) {
	t := table.NewWriter()
	t.SetOutputMirror(out)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("Elastic IP Waste")

	t.AppendHeader(table.Row{"Status", "IP Address", "Allocation ID", "Est. Cost/Mo"})

	statusUnused := "Unassociated"
	rows := populateElasticIPRows(elasticIPInfo, pricingSvc)

	if len(rows) > 0 {
		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiRed.Sprint(statusUnused)
	}

	t.AppendRows(rows)

	if t.Length() > 0 {
		t.Render()

		_, _ = fmt.Fprintln(out)
	}
}

func populateEBSRows(volumes []ec2types.Volume, status string, pricingSvc pricing.Service) []table.Row {
	rows := make([]table.Row, 0, len(volumes))

	for _, vol := range volumes {
		p := outputshared.PresentEBSVolume(vol, status, pricingSvc)
		rows = append(rows, table.Row{"", p.Identifier, p.Metric, p.EstimatedCost})
	}

	return rows
}

func populateElasticIPRows(ips []ec2types.Address, pricingSvc pricing.Service) []table.Row {
	rows := make([]table.Row, 0, len(ips))

	for _, ip := range ips {
		p := outputshared.PresentElasticIP(ip, pricingSvc)
		rows = append(rows, table.Row{"", p.Identifier, aws.ToString(ip.AllocationId), p.EstimatedCost})
	}

	return rows
}

func populateInstanceRows(instances []ec2types.Instance) []table.Row {
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

func drawLoadBalancerTable(out io.Writer, loadBalancers []elbtypes.LoadBalancer, idleLoadBalancers []model.ELBIdleInfo, pricingSvc pricing.Service) {
	t := table.NewWriter()
	t.SetOutputMirror(out)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("Load Balancer Waste")

	t.AppendHeader(table.Row{"Status", "Name", "Type", "Est. Cost/Mo"})

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 4, Align: text.AlignRight},
	})

	var hasPreviousRows bool

	if len(loadBalancers) > 0 {
		statusUnused := "No Target Groups"
		rows := populateLoadBalancerRows(loadBalancers, pricingSvc)

		if len(rows) > 0 {
			halfRow := len(rows) / 2
			rows[halfRow][0] = text.FgHiRed.Sprint(statusUnused)
		}

		t.AppendRows(rows)

		hasPreviousRows = true
	}

	if len(idleLoadBalancers) > 0 {
		if hasPreviousRows {
			t.AppendSeparator()
		}

		statusIdle := "Idle (0 connections)"
		rows := populateIdleLoadBalancerRows(idleLoadBalancers)

		if len(rows) > 0 {
			halfRow := len(rows) / 2
			rows[halfRow][0] = text.FgHiYellow.Sprint(statusIdle)
		}

		t.AppendRows(rows)
	}

	if t.Length() > 0 {
		t.Render()

		_, _ = fmt.Fprintln(out)
	}
}

func populateIdleLoadBalancerRows(idleLBs []model.ELBIdleInfo) []table.Row {
	rows := make([]table.Row, 0, len(idleLBs))

	for _, lb := range idleLBs {
		p := outputshared.PresentIdleLoadBalancer(lb)
		rows = append(rows, table.Row{"", lb.Name, p.Metric, p.EstimatedCost})
	}

	return rows
}

func populateLoadBalancerRows(loadBalancers []elbtypes.LoadBalancer, pricingSvc pricing.Service) []table.Row {
	rows := make([]table.Row, 0, len(loadBalancers))

	for _, lb := range loadBalancers {
		p := outputshared.PresentLoadBalancer(lb, pricingSvc)
		// Details: "Created on 2026-03-21T08:00:00Z"
		// We want the name which is not in ResourceRow, oh wait, Identifier is LoadBalancerArn
		// In DrawLoadBalancerTable it used lb.LoadBalancerName
		rows = append(rows, table.Row{"", aws.ToString(lb.LoadBalancerName), p.Metric, p.EstimatedCost})
	}

	return rows
}

func drawAMITable(out io.Writer, amis []model.AMIWasteInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(out)
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

	_, _ = fmt.Fprintln(out, text.FgHiYellow.Sprint(" * Warning: AMIs may be referenced by Auto Scaling Groups or Launch Templates"))
	_, _ = fmt.Fprintln(out)
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

func drawSnapshotTable(out io.Writer, snapshots []model.SnapshotWasteInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(out)
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

	if t.Length() > 0 {
		t.Render()

		_, _ = fmt.Fprintln(out)
	}
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

func drawKeyPairTable(out io.Writer, keyPairs []model.KeyPairWasteInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(out)
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

	if t.Length() > 0 {
		t.Render()

		_, _ = fmt.Fprintln(out)
	}
}

func drawNatGatewayTable(out io.Writer, natGateways []model.NATGatewayWasteInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(out)
	t.SetStyle(table.StyleRounded)
	t.SetTitle(" %s ", text.FgHiYellow.Sprint("NAT Gateway Waste"))

	t.AppendHeader(table.Row{"Status", "NAT Gateway ID", "Metric", "Est. Cost/Mo"})

	rows := populateNatGatewayRows(natGateways)
	if len(rows) > 0 {
		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiRed.Sprint("Idle (> 7 Days)")
	}

	t.AppendRows(rows)

	if t.Length() > 0 {
		t.Render()

		_, _ = fmt.Fprintln(out)
	}
}

func populateNatGatewayRows(natGateways []model.NATGatewayWasteInfo) []table.Row {
	rows := make([]table.Row, 0, len(natGateways))

	for _, ng := range natGateways {
		p := outputshared.PresentIdleNATGateway(ng)
		rows = append(rows, table.Row{
			"",
			p.Identifier,
			p.Metric,
			p.EstimatedCost,
		})
	}

	return rows
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

func drawS3Table(out io.Writer, buckets []model.S3BucketWasteInfo, multipartBuckets []model.S3MultipartUploadWasteInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(out)
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

	if t.Length() > 0 {
		t.Render()

		_, _ = fmt.Fprintln(out)
	}
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

func drawCloudWatchLogsTable(out io.Writer, logGroups []model.CloudWatchLogsWasteInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(out)
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

	if t.Length() > 0 {
		t.Render()

		_, _ = fmt.Fprintln(out)
	}
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

func drawSummaryTable(out io.Writer, input model.RenderWasteInput, pricingSvc pricing.Service) {
	categories, totalCost := wastesummary.Compute(input, pricingSvc)

	if len(categories) == 0 {
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(out)
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

	_, _ = fmt.Fprintln(out, text.FgHiYellow.Sprint(" * Estimates use AWS Pricing API rates for the configured region, falling back to us-east-1 defaults when unavailable."))
	_, _ = fmt.Fprintln(out)
}

func drawRDSTable(out io.Writer, instances []model.RDSInstanceWasteInfo, snapshots []model.RDSSnapshotWasteInfo, idleInstances []model.RDSIdleInstanceInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(out)
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

	if t.Length() > 0 {
		t.Render()

		_, _ = fmt.Fprintln(out)
	}
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

func drawLambdaTable(out io.Writer, lambdas []model.LambdaOverProvisionedInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(out)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("Lambda Over-Provisioned Memory")

	t.AppendHeader(table.Row{"Status", "Function Name", "Runtime", "Memory (Configured)", "Memory (Max Used)", "Utilization", "Recommended"})

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 4, Align: text.AlignRight},
		{Number: 5, Align: text.AlignRight},
		{Number: 6, Align: text.AlignRight},
		{Number: 7, Align: text.AlignRight},
	})

	statusLabel := "Over-Provisioned"
	rows := populateLambdaRows(lambdas)

	if len(rows) > 0 {
		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiYellow.Sprint(statusLabel)
	}

	t.AppendRows(rows)

	if t.Length() > 0 {
		t.Render()

		_, _ = fmt.Fprintln(out)
	}
}

func populateLambdaRows(lambdas []model.LambdaOverProvisionedInfo) []table.Row {
	rows := make([]table.Row, 0, len(lambdas))

	for _, fn := range lambdas {
		p := outputshared.PresentLambdaOverProvisioned(fn)
		rows = append(rows, table.Row{
			"",
			p.Identifier,
			fn.Runtime,
			fmt.Sprintf("%d MB", fn.ConfiguredMemoryMB),
			fmt.Sprintf("%d MB", fn.MaxMemoryUsedMB),
			p.Metric,
			fmt.Sprintf("%d MB", fn.RecommendedMemoryMB),
		})
	}

	return rows
}

func drawSageMakerTable(out io.Writer, endpoints []model.IdleSageMakerEndpointInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(out)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("SageMaker Endpoints (Idle)")

	t.AppendHeader(table.Row{"Status", "Endpoint", "Variants", "Days Checked", "Est. Cost/Mo"})

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 4, Align: text.AlignRight},
		{Number: 5, Align: text.AlignRight},
	})

	statusLabel := "Idle"
	rows := populateSageMakerRows(endpoints)

	if len(rows) > 0 {
		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiRed.Sprint(statusLabel)
	}

	t.AppendRows(rows)

	if t.Length() > 0 {
		t.Render()

		_, _ = fmt.Fprintln(out)
	}
}

func populateSageMakerRows(endpoints []model.IdleSageMakerEndpointInfo) []table.Row {
	rows := make([]table.Row, 0, len(endpoints))

	for _, ep := range endpoints {
		p := outputshared.PresentIdleSageMakerEndpoint(ep)
		rows = append(rows, table.Row{
			"",
			p.Identifier,
			p.Details,
			fmt.Sprintf("%d days", ep.DaysChecked),
			p.EstimatedCost,
		})
	}

	return rows
}

func drawECRTable(out io.Writer, noPolicy []model.ECRNoLifecyclePolicyInfo, empty []model.ECREmptyRepositoryInfo, untagged []model.ECRUntaggedImageInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(out)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("ECR Repository Waste")

	t.AppendHeader(table.Row{"Status", "Repository Name", "Metric", "Est. Cost/Mo", "Details"})

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 4, Align: text.AlignRight},
	})

	var hasPreviousRows bool

	if len(noPolicy) > 0 {
		statusLabel := "No Lifecycle Policy"
		rows := populateECRNoPolicyRows(noPolicy)
		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiYellow.Sprint(statusLabel)
		t.AppendRows(rows)

		hasPreviousRows = true
	}

	if len(empty) > 0 {
		if hasPreviousRows {
			t.AppendSeparator()
		}

		statusLabel := "Empty Repository"
		rows := populateECREmptyRows(empty)
		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiYellow.Sprint(statusLabel)
		t.AppendRows(rows)

		hasPreviousRows = true
	}

	if len(untagged) > 0 {
		if hasPreviousRows {
			t.AppendSeparator()
		}

		statusLabel := "Untagged Images"
		rows := populateECRUntaggedRows(untagged)
		halfRow := len(rows) / 2
		rows[halfRow][0] = text.FgHiRed.Sprint(statusLabel)
		t.AppendRows(rows)
	}

	if t.Length() > 0 {
		t.Render()

		_, _ = fmt.Fprintln(out)
	}
}

func populateECRNoPolicyRows(repos []model.ECRNoLifecyclePolicyInfo) []table.Row {
	rows := make([]table.Row, 0, len(repos))

	for _, repo := range repos {
		p := outputshared.PresentECRNoLifecyclePolicy(repo)
		rows = append(rows, table.Row{
			"",
			p.Identifier,
			p.Metric,
			p.EstimatedCost,
			p.Details,
		})
	}

	return rows
}

func populateECREmptyRows(repos []model.ECREmptyRepositoryInfo) []table.Row {
	rows := make([]table.Row, 0, len(repos))

	for _, repo := range repos {
		p := outputshared.PresentECREmptyRepository(repo)
		rows = append(rows, table.Row{
			"",
			p.Identifier,
			p.Metric,
			p.EstimatedCost,
			p.Details,
		})
	}

	return rows
}

func populateECRUntaggedRows(repos []model.ECRUntaggedImageInfo) []table.Row {
	rows := make([]table.Row, 0, len(repos))

	for _, repo := range repos {
		p := outputshared.PresentECRUntaggedImages(repo)
		rows = append(rows, table.Row{
			"",
			p.Identifier,
			p.Metric,
			p.EstimatedCost,
			p.Details,
		})
	}

	return rows
}

// RenderScopeTable renders the waste table for a specific scope.
func RenderScopeTable(scope string, input model.RenderWasteInput, pricingSvc pricing.Service) string {
	var buf strings.Builder

	out := &buf

	switch scope {
	case "EC2":
		drawEC2Table(out, input.StoppedInstances, input.Ris)
		drawIdleEC2Table(out, input.IdleEC2Instances)
		drawAMITable(out, input.UnusedAMIs)
		drawKeyPairTable(out, input.UnusedKeyPairs)
		drawEBSTable(out, input.UnusedVolumes, input.StoppedVolumes, pricingSvc)
		drawSnapshotTable(out, input.OrphanedSnapshots)
		drawElasticIPTable(out, input.ElasticIPs, pricingSvc)
	case "VPC":
		drawNatGatewayTable(out, input.IdleNATGateways)
	case "ELB":
		drawLoadBalancerTable(out, input.LoadBalancers, input.IdleLoadBalancers, pricingSvc)
	case "S3":
		drawS3Table(out, input.S3Buckets, input.S3MultipartUploads)
	case "CloudWatch":
		drawCloudWatchLogsTable(out, input.CloudWatchLogGroups)
	case "RDS":
		drawRDSTable(out, input.RDSInstances, input.RDSSnapshots, input.RDSIdleInstances)
	case "Lambda":
		drawLambdaTable(out, input.OverProvisionedLambdas)
	case "SageMaker":
		drawSageMakerTable(out, input.IdleSageMakerEndpoints)
	case "ECR":
		drawECRTable(out, input.ECRNoLifecyclePolicies, input.ECREmptyRepositories, input.ECRUntaggedImages)
	case "SecretsManager":
		drawSecretsManagerTable(out, input.UnusedSecrets, pricingSvc)
	case "IAM":
		drawIAMTable(out, input.UnusedIAMUsers, input.RootUserWaste)
	case "Summary":
		drawSummaryTable(out, input, pricingSvc)
	}

	return buf.String()
}

func drawIAMTable(out io.Writer, users []model.IAMUserWasteInfo, root []model.IAMRootUserWasteInfo) {
	if len(users) == 0 && len(root) == 0 {
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(out)
	t.SetStyle(table.StyleRounded)
	t.SetTitle("IAM Waste & Security")

	t.AppendHeader(table.Row{"Resource", "Issue", "Action Required"})

	if len(root) > 0 {
		for _, r := range root {
			p := outputshared.PresentIAMRootWaste(r)
			t.AppendRow(table.Row{
				text.FgHiYellow.Sprint("Root Account"),
				p.Details,
				"Enable Virtual MFA immediately",
			})
		}
	}

	for _, u := range users {
		p := outputshared.PresentIAMUser(u)
		issue := fmt.Sprintf("Pwd: %s | Keys: %s", p.Age, p.Metric)
		t.AppendRow(table.Row{
			fmt.Sprintf("User: %s", p.Identifier),
			issue,
			"Delete or disable user",
		})
	}

	t.Render()
}

func drawErrorsSection(out io.Writer, errors map[string]string) {
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, text.FgHiRed.Sprint(" ⚠️  ERRORS ENCOUNTERED DURING SCAN"))
	_, _ = fmt.Fprintln(out, text.FgRed.Sprint(" ------------------------------------------------"))

	t := table.NewWriter()
	t.SetOutputMirror(out)
	t.AppendHeader(table.Row{"Scope", "Error"})

	var scopes []string
	for scope := range errors {
		scopes = append(scopes, scope)
	}

	sort.Strings(scopes)

	for _, scope := range scopes {
		t.AppendRow(table.Row{scope, errors[scope]})
	}

	t.SetStyle(table.StyleRounded)
	t.Render()

	_, _ = fmt.Fprintln(out)
}

// RenderErrorsTable renders the errors as a string for the TUI.
func RenderErrorsTable(errors map[string]string) string {
	if len(errors) == 0 {
		return ""
	}

	b := &strings.Builder{}
	_, _ = fmt.Fprintln(b, text.FgHiRed.Sprint(" ⚠️  ERRORS ENCOUNTERED DURING SCAN"))
	_, _ = fmt.Fprintln(b, text.FgRed.Sprint(" ------------------------------------------------"))

	t := table.NewWriter()
	t.SetOutputMirror(b)
	t.AppendHeader(table.Row{"Scope", "Error"})

	var scopes []string
	for scope := range errors {
		scopes = append(scopes, scope)
	}

	sort.Strings(scopes)

	for _, scope := range scopes {
		t.AppendRow(table.Row{scope, errors[scope]})
	}

	t.SetStyle(table.StyleRounded)
	t.Render()

	return b.String()
}
