package report

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/pricing"
	outputshared "github.com/elC0mpa/aws-doctor/utils/output_shared"
	wastesummary "github.com/elC0mpa/aws-doctor/utils/waste_summary"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

func (s *service) addWasteSections(m core.Maroto, input model.RenderWasteInput, pricingSvc pricing.Service) bool {
	hasWaste := s.addEBSWaste(m, input, pricingSvc)

	if s.addElasticIPWaste(m, input, pricingSvc) {
		hasWaste = true
	}

	if s.addEC2Waste(m, input) {
		hasWaste = true
	}

	if s.addLBWaste(m, input, pricingSvc) {
		hasWaste = true
	}

	if s.addS3Waste(m, input) {
		hasWaste = true
	}

	if s.addCloudWatchWaste(m, input) {
		hasWaste = true
	}

	if s.addAMIWaste(m, input) {
		hasWaste = true
	}

	if s.addSnapshotWaste(m, input) {
		hasWaste = true
	}

	if s.addKeyPairWaste(m, input) {
		hasWaste = true
	}

	if s.addRDSWaste(m, input) {
		hasWaste = true
	}

	if s.addNATGatewayWaste(m, input) {
		hasWaste = true
	}

	if s.addLambdaWaste(m, input) {
		hasWaste = true
	}

	if s.addSageMakerWaste(m, input) {
		hasWaste = true
	}

	if s.addECRWaste(m, input) {
		hasWaste = true
	}

	return hasWaste
}

func (s *service) addEBSWaste(m core.Maroto, input model.RenderWasteInput, pricingSvc pricing.Service) bool {
	if len(input.UnusedVolumes) == 0 && len(input.StoppedVolumes) == 0 {
		return false
	}

	s.addWasteSection(m, "EBS Volume Waste", []string{"Status", "Volume ID", "Size", "Est. Cost"})

	for _, v := range input.UnusedVolumes {
		p := outputshared.PresentEBSVolume(v, "unattached", pricingSvc)
		s.addWasteRow(m, []string{"Unattached", p.Identifier, p.Metric, p.EstimatedCost})
	}

	for _, v := range input.StoppedVolumes {
		p := outputshared.PresentEBSVolume(v, "stopped", pricingSvc)
		s.addWasteRow(m, []string{"Stopped Inst.", p.Identifier, p.Metric, p.EstimatedCost})
	}

	m.AddRow(5, col.New(12))

	return true
}

func (s *service) addElasticIPWaste(m core.Maroto, input model.RenderWasteInput, pricingSvc pricing.Service) bool {
	if len(input.ElasticIPs) == 0 {
		return false
	}

	s.addWasteSection(m, "Elastic IP Waste", []string{"Status", "IP Address", "Allocation ID", "Est. Cost"})

	for _, ip := range input.ElasticIPs {
		p := outputshared.PresentElasticIP(ip, pricingSvc)
		s.addWasteRow(m, []string{"Unassociated", p.Identifier, aws.ToString(ip.AllocationId), p.EstimatedCost})
	}

	m.AddRow(5, col.New(12))

	return true
}

func (s *service) addEC2Waste(m core.Maroto, input model.RenderWasteInput) bool {
	if len(input.StoppedInstances) == 0 && len(input.Ris) == 0 {
		return false
	}

	s.addWasteSection(m, "EC2 & Reserved Instance Waste", []string{"Status", "Instance ID", "Time Info"})

	for _, inst := range input.StoppedInstances {
		p := outputshared.PresentStoppedInstance(inst)
		timeInfo := outputshared.NAValue

		if p.Age != outputshared.NAValue {
			timeInfo = p.Age + " days ago"
		}

		s.addWasteRow(m, []string{"Stopped (>30d)", p.Identifier, timeInfo})
	}

	for _, ri := range input.Ris {
		p := outputshared.PresentReservedInstance(ri)
		timeInfo := ""

		if ri.DaysUntilExpiry >= 0 {
			timeInfo = fmt.Sprintf("In %d days", ri.DaysUntilExpiry)
		} else {
			timeInfo = fmt.Sprintf("%d days ago", -ri.DaysUntilExpiry)
		}

		s.addWasteRow(m, []string{fmt.Sprintf("RI (%s)", ri.Status), p.Identifier, timeInfo})
	}

	m.AddRow(5, col.New(12))

	return true
}

func (s *service) addLBWaste(m core.Maroto, input model.RenderWasteInput, pricingSvc pricing.Service) bool {
	if len(input.LoadBalancers) == 0 && len(input.IdleLoadBalancers) == 0 {
		return false
	}

	s.addWasteSection(m, "Load Balancer Waste", []string{"Status", "Name", "Type", "Est. Cost"})

	for _, lb := range input.LoadBalancers {
		p := outputshared.PresentLoadBalancer(lb, pricingSvc)
		s.addWasteRow(m, []string{"Unused", aws.ToString(lb.LoadBalancerName), p.Metric, p.EstimatedCost})
	}

	for _, lb := range input.IdleLoadBalancers {
		p := outputshared.PresentIdleLoadBalancer(lb)
		s.addWasteRow(m, []string{"Idle", lb.Name, p.Metric, p.EstimatedCost})
	}

	m.AddRow(5, col.New(12))

	return true
}

func (s *service) addS3Waste(m core.Maroto, input model.RenderWasteInput) bool {
	if len(input.S3Buckets) == 0 && len(input.S3MultipartUploads) == 0 {
		return false
	}

	s.addWasteSection(m, "S3 Bucket Waste", []string{"Status", "Bucket Name", "Info"})

	for _, b := range input.S3Buckets {
		p := outputshared.PresentS3Bucket(b)
		s.addWasteRow(m, []string{"No Lifecycle", p.Identifier, fmt.Sprintf("Created: %s", b.CreationDate.Format("2006-01-02"))})
	}

	for _, b := range input.S3MultipartUploads {
		p := outputshared.PresentS3MultipartUpload(b)
		s.addWasteRow(m, []string{"Incomplete MP", p.Identifier, p.Metric})
	}

	m.AddRow(5, col.New(12))

	return true
}

func (s *service) addCloudWatchWaste(m core.Maroto, input model.RenderWasteInput) bool {
	if len(input.CloudWatchLogGroups) == 0 {
		return false
	}

	s.addWasteSection(m, "CloudWatch Log Group Waste", []string{"Status", "Log Group Name", "Size", "Est. Cost"})

	for _, lg := range input.CloudWatchLogGroups {
		p := outputshared.PresentCloudWatchLogGroup(lg)
		s.addWasteRow(m, []string{"No Retention", p.Identifier, p.Metric, p.EstimatedCost})
	}

	m.AddRow(5, col.New(12))

	return true
}

func (s *service) addAMIWaste(m core.Maroto, input model.RenderWasteInput) bool {
	if len(input.UnusedAMIs) == 0 {
		return false
	}

	s.addWasteSection(m, "Unused AMI Waste", []string{"Status", "AMI ID", "Age", "Max Savings"})

	for _, ami := range input.UnusedAMIs {
		p := outputshared.PresentAMI(ami)
		s.addWasteRow(m, []string{"Unused", p.Identifier, p.Age + " days", fmt.Sprintf("$%.2f", ami.MaxPotentialSaving)})
	}

	m.AddRow(5, col.New(12))

	return true
}

func (s *service) addSnapshotWaste(m core.Maroto, input model.RenderWasteInput) bool {
	if len(input.OrphanedSnapshots) == 0 {
		return false
	}

	s.addWasteSection(m, "EBS Snapshot Waste", []string{"Status", "Snapshot ID", "Size", "Max Savings"})

	for _, snap := range input.OrphanedSnapshots {
		p := outputshared.PresentSnapshot(snap)
		s.addWasteRow(m, []string{string(snap.Category), p.Identifier, p.Metric, p.EstimatedCost})
	}

	m.AddRow(5, col.New(12))

	return true
}

func (s *service) addKeyPairWaste(m core.Maroto, input model.RenderWasteInput) bool {
	if len(input.UnusedKeyPairs) == 0 {
		return false
	}

	s.addWasteSection(m, "Unused Key Pair Waste", []string{"Status", "Key Name", "Age"})

	for _, kp := range input.UnusedKeyPairs {
		p := outputshared.PresentKeyPair(kp)
		s.addWasteRow(m, []string{"Unused", p.Identifier, p.Age + " days"})
	}

	m.AddRow(5, col.New(12))

	return true
}

func (s *service) addRDSWaste(m core.Maroto, input model.RenderWasteInput) bool {
	if len(input.RDSInstances) == 0 && len(input.RDSSnapshots) == 0 && len(input.RDSIdleInstances) == 0 {
		return false
	}

	s.addWasteSection(m, "RDS Waste", []string{"Status", "Identifier", "Engine", "Est. Cost"})

	for _, inst := range input.RDSInstances {
		p := outputshared.PresentRDSInstance(inst)
		s.addWasteRow(m, []string{"Stopped", p.Identifier, inst.Engine, p.EstimatedCost})
	}

	for _, snap := range input.RDSSnapshots {
		p := outputshared.PresentRDSSnapshot(snap)
		s.addWasteRow(m, []string{"Old Snapshot", p.Identifier, snap.Engine, p.EstimatedCost})
	}

	for _, inst := range input.RDSIdleInstances {
		p := outputshared.PresentRDSIdleInstance(inst)
		s.addWasteRow(m, []string{"Idle", p.Identifier, inst.Engine, p.EstimatedCost})
	}

	m.AddRow(5, col.New(12))

	return true
}

func (s *service) addNATGatewayWaste(m core.Maroto, input model.RenderWasteInput) bool {
	if len(input.IdleNATGateways) == 0 {
		return false
	}

	s.addWasteSection(m, "NAT Gateway Waste", []string{"Status", "NAT Gateway ID", "VPC ID", "Est. Cost"})

	for _, ng := range input.IdleNATGateways {
		p := outputshared.PresentIdleNATGateway(ng)
		s.addWasteRow(m, []string{"Idle", p.Identifier, ng.VPCID, p.EstimatedCost})
	}

	m.AddRow(5, col.New(12))

	return true
}

func (s *service) addLambdaWaste(m core.Maroto, input model.RenderWasteInput) bool {
	if len(input.OverProvisionedLambdas) == 0 {
		return false
	}

	m.AddRow(10,
		text.NewCol(12, "Lambda Over-Provisioned Memory", props.Text{Style: fontstyle.Bold, Size: 11}),
	)

	m.AddRow(10,
		text.NewCol(5, "Function", props.Text{Style: fontstyle.Bold, Size: 9}),
		text.NewCol(3, "Used/Configured", props.Text{Style: fontstyle.Bold, Size: 9}),
		text.NewCol(2, "Utilization", props.Text{Style: fontstyle.Bold, Size: 9, Align: align.Right}),
		text.NewCol(2, "Recommended", props.Text{Style: fontstyle.Bold, Size: 9, Align: align.Right}),
	)

	m.AddRow(2, line.NewCol(12))

	for _, fn := range input.OverProvisionedLambdas {
		m.AddRow(8,
			text.NewCol(5, fn.FunctionName, props.Text{Size: 7}),
			text.NewCol(3, fmt.Sprintf("%d / %d MB", fn.MaxMemoryUsedMB, fn.ConfiguredMemoryMB), props.Text{Size: 8}),
			text.NewCol(2, fmt.Sprintf("%.1f%%", fn.MemoryUtilization), props.Text{Size: 8, Align: align.Right}),
			text.NewCol(2, fmt.Sprintf("%d MB", fn.RecommendedMemoryMB), props.Text{Size: 8, Align: align.Right}),
		)
	}

	m.AddRow(5, col.New(12))

	return true
}

func (s *service) addSageMakerWaste(m core.Maroto, input model.RenderWasteInput) bool {
	if len(input.IdleSageMakerEndpoints) == 0 {
		return false
	}

	s.addWasteSection(m, "SageMaker Endpoints (Idle)", []string{"Status", "Endpoint", "Variants", "Est. Cost"})

	for _, ep := range input.IdleSageMakerEndpoints {
		p := outputshared.PresentIdleSageMakerEndpoint(ep)
		s.addWasteRow(m, []string{"Idle", p.Identifier, p.Details, p.EstimatedCost})
	}

	m.AddRow(5, col.New(12))

	return true
}

func (s *service) addWasteSummary(m core.Maroto, input model.RenderWasteInput, pricingSvc pricing.Service) {
	categories, totalCost := wastesummary.Compute(input, pricingSvc)

	if len(categories) == 0 {
		return
	}

	s.addWasteSection(m, "Waste Summary", []string{"Category", "Count", "Est. Monthly Cost"})

	for _, cat := range categories {
		costStr := outputshared.NAValue

		if cat.Cost > 0 {
			costStr = fmt.Sprintf("$%.2f", cat.Cost)
		}

		s.addWasteRow(m, []string{cat.Name, fmt.Sprintf("%d", cat.Count), costStr})
	}

	m.AddRow(2, line.NewCol(12))

	m.AddRow(10,
		text.NewCol(4, "Total Estimated Monthly Waste", props.Text{Style: fontstyle.Bold, Size: 10}),
		text.NewCol(4, "", props.Text{}),
		text.NewCol(4, fmt.Sprintf("$%.2f", totalCost), props.Text{Style: fontstyle.Bold, Size: 10, Align: align.Right, Color: &props.Color{Red: 200, Green: 0, Blue: 0}}),
	)
}

func (s *service) addWasteSection(m core.Maroto, title string, headers []string) {
	m.AddRow(10,
		text.NewCol(12, title, props.Text{Style: fontstyle.Bold, Size: 11}),
	)

	if len(headers) == 3 {
		m.AddRow(10,
			text.NewCol(4, headers[0], props.Text{Style: fontstyle.Bold, Size: 9}),
			text.NewCol(4, headers[1], props.Text{Style: fontstyle.Bold, Size: 9}),
			text.NewCol(4, headers[2], props.Text{Style: fontstyle.Bold, Size: 9, Align: align.Right}),
		)
	} else {
		m.AddRow(10,
			text.NewCol(3, headers[0], props.Text{Style: fontstyle.Bold, Size: 9}),
			text.NewCol(4, headers[1], props.Text{Style: fontstyle.Bold, Size: 9}),
			text.NewCol(3, headers[2], props.Text{Style: fontstyle.Bold, Size: 9, Align: align.Right}),
			text.NewCol(2, headers[3], props.Text{Style: fontstyle.Bold, Size: 9, Align: align.Right}),
		)
	}

	m.AddRow(2, line.NewCol(12))
}

func (s *service) addWasteRow(m core.Maroto, values []string) {
	if len(values) == 3 {
		m.AddRow(8,
			text.NewCol(4, values[0], props.Text{Size: 8}),
			text.NewCol(4, values[1], props.Text{Size: 8}),
			text.NewCol(4, values[2], props.Text{Size: 8, Align: align.Right}),
		)
	} else {
		m.AddRow(8,
			text.NewCol(3, values[0], props.Text{Size: 8}),
			text.NewCol(4, values[1], props.Text{Size: 8}),
			text.NewCol(3, values[2], props.Text{Size: 8, Align: align.Right}),
			text.NewCol(2, values[3], props.Text{Size: 8, Align: align.Right}),
		)
	}
}

func (s *service) addECRWaste(m core.Maroto, input model.RenderWasteInput) bool {
	if len(input.ECRNoLifecyclePolicies) == 0 && len(input.ECREmptyRepositories) == 0 && len(input.ECRUntaggedImages) == 0 {
		return false
	}

	s.addWasteSection(m, "ECR Repository Waste", []string{"Status", "Repository", "Info", "Est. Cost"})

	for _, repo := range input.ECRNoLifecyclePolicies {
		p := outputshared.PresentECRNoLifecyclePolicy(repo)
		s.addWasteRow(m, []string{"No Policy", p.Identifier, p.Details, p.EstimatedCost})
	}

	for _, repo := range input.ECREmptyRepositories {
		p := outputshared.PresentECREmptyRepository(repo)
		s.addWasteRow(m, []string{"Empty", p.Identifier, p.Details, p.EstimatedCost})
	}

	for _, repo := range input.ECRUntaggedImages {
		p := outputshared.PresentECRUntaggedImages(repo)
		s.addWasteRow(m, []string{"Untagged", p.Identifier, p.Details, p.EstimatedCost})
	}

	m.AddRow(5, col.New(12))

	return true
}
