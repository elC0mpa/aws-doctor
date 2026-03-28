package report

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elC0mpa/aws-doctor/model"
	outputshared "github.com/elC0mpa/aws-doctor/utils/output_shared"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

type service struct{}

// NewService creates a new report service.
func NewService() Service {
	return &service{}
}

func (s *service) GenerateCostComparisonReport(input model.RenderCostComparisonInput, reportPath string) error {
	path := s.getReportPath(reportPath, "cost")

	m := maroto.New()

	s.addHeader(m, CostReport, input.AccountID)

	// Get unit from data
	_, unit := outputshared.ParseCostString(input.CurrentTotalCost)

	// Format months
	lastMonthLabel := s.formatDateToMonthYear(input.LastMonth.Start)
	currentMonthLabel := s.formatDateToMonthYear(input.CurrentMonth.Start)

	// Table Header
	m.AddRow(12,
		text.NewCol(4, "Service", props.Text{Style: fontstyle.Bold, Size: 10}),
		col.New(3).Add(
			text.New(fmt.Sprintf("%s (%s)", lastMonthLabel, unit), props.Text{Style: fontstyle.Bold, Size: 10, Align: align.Right}),
			text.New(s.formatDayRange(input.LastMonth.Start, input.LastMonth.End), props.Text{Size: 8, Align: align.Right, Top: 5}),
		),
		col.New(3).Add(
			text.New(fmt.Sprintf("%s (%s)", currentMonthLabel, unit), props.Text{Style: fontstyle.Bold, Size: 10, Align: align.Right}),
			text.New(s.formatDayRange(input.CurrentMonth.Start, input.CurrentMonth.End), props.Text{Size: 8, Align: align.Right, Top: 5}),
		),
		text.NewCol(2, fmt.Sprintf("Diff (%s)", unit), props.Text{Style: fontstyle.Bold, Size: 10, Align: align.Right}),
	)
	m.AddRow(2, line.NewCol(12))

	// Total Row
	lastTotalAmount, _ := outputshared.ParseCostString(input.LastTotalCost)
	currentTotalAmount, _ := outputshared.ParseCostString(input.CurrentTotalCost)
	difference := currentTotalAmount - lastTotalAmount

	totalColor := &props.Color{Red: 0, Green: 150, Blue: 0} // Green
	if difference > 0 {
		totalColor = &props.Color{Red: 200, Green: 0, Blue: 0} // Red
	}

	m.AddRow(8,
		text.NewCol(4, "Total Costs", props.Text{Style: fontstyle.Bold, Size: 10, Color: totalColor}),
		text.NewCol(3, fmt.Sprintf("%.2f", lastTotalAmount), props.Text{Size: 10, Align: align.Right}),
		text.NewCol(3, fmt.Sprintf("%.2f", currentTotalAmount), props.Text{Size: 10, Align: align.Right, Color: totalColor}),
		text.NewCol(2, fmt.Sprintf("%.2f", difference), props.Text{Size: 10, Align: align.Right, Color: totalColor}),
	)
	m.AddRow(2, line.NewCol(12))

	// Service Breakdown
	orderedServices := outputshared.OrderCostServices(&input.CurrentMonth.CostGroup)
	for _, currentSvc := range orderedServices {
		lastMonthGroup := input.LastMonth.CostGroup[currentSvc.Name]
		diff := currentSvc.Amount - lastMonthGroup.Amount

		rowColor := &props.Color{Red: 0, Green: 150, Blue: 0} // Green
		if diff > 0 {
			rowColor = &props.Color{Red: 200, Green: 0, Blue: 0} // Red
		}

		m.AddRow(8,
			text.NewCol(4, currentSvc.Name, props.Text{Size: 9, Color: rowColor}),
			text.NewCol(3, fmt.Sprintf("%.2f", lastMonthGroup.Amount), props.Text{Size: 9, Align: align.Right}),
			text.NewCol(3, fmt.Sprintf("%.2f", currentSvc.Amount), props.Text{Size: 9, Align: align.Right, Color: rowColor}),
			text.NewCol(2, fmt.Sprintf("%.2f", diff), props.Text{Size: 9, Align: align.Right, Color: rowColor}),
		)
	}

	s.addFooter(m, CostReport)

	return s.generateAndSave(m, path)
}

func (s *service) GenerateTrendReport(accountID string, costInfo []model.CostInfo, services []string, reportPath string) error {
	path := s.getReportPath(reportPath, "trend")
	fmt.Printf("Generating trend report at: %s\n", path)
	// PDF generation logic to be implemented by the user
	return nil
}

func (s *service) GenerateWasteReport(input model.RenderWasteInput, reportPath string) error {
	path := s.getReportPath(reportPath, "waste")
	fmt.Printf("Generating waste report at: %s\n", path)
	// PDF generation logic to be implemented by the user
	return nil
}

func (s *service) getReportPath(reportPath, flow string) string {
	if reportPath != "" && reportPath != "DEFAULT" {
		return reportPath
	}

	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("aws-doctor-%s-%s.pdf", flow, timestamp)
}

func (s *service) generateAndSave(m core.Maroto, path string) error {
	document, err := m.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate PDF: %w", err)
	}

	err = document.Save(path)
	if err != nil {
		return fmt.Errorf("failed to save PDF: %w", err)
	}

	return nil
}

func (s *service) formatDateToMonthYear(dateStr *string) string {
	if dateStr == nil {
		return ""
	}
	t, err := time.Parse("2006-01-02", *dateStr)
	if err != nil {
		return *dateStr
	}
	return t.Format("January 2006")
}

func (s *service) formatDayRange(start, end *string) string {
	if start == nil || end == nil {
		return ""
	}

	t1, err1 := time.Parse("2006-01-02", *start)
	t2, err2 := time.Parse("2006-01-02", *end)

	if err1 != nil || err2 != nil {
		return fmt.Sprintf("(%s to %s)", *start, *end)
	}

	return fmt.Sprintf("(%s to %s)", s.getDayWithSuffix(t1.Day()), s.getDayWithSuffix(t2.Day()))
}

func (s *service) getDayWithSuffix(day int) string {
	suffix := "th"
	switch day % 10 {
	case 1:
		if day%100 != 11 {
			suffix = "st"
		}
	case 2:
		if day%100 != 12 {
			suffix = "nd"
		}
	case 3:
		if day%100 != 13 {
			suffix = "rd"
		}
	}
	return fmt.Sprintf("%d%s", day, suffix)
}

func (s *service) addHeader(m core.Maroto, reportType ReportType, accountID string) {
	title := "AWS COST DIAGNOSIS"
	switch reportType {
	case TrendReport:
		title = "AWS COST TREND"
	case WasteReport:
		title = "AWS WASTE REPORT"
	}

	m.AddRow(15,
		image.NewFromFileCol(3, "assets/logo-pdf.png"),
		text.NewCol(6, title, props.Text{
			Size:  16,
			Style: fontstyle.Bold,
			Align: align.Center,
			Top:   5,
		}),
		text.NewCol(3, time.Now().Format("January 02, 2006"), props.Text{
			Size:  12,
			Align: align.Right,
			Top:   7,
		}),
	)
	m.AddRow(10,
		text.NewCol(12, fmt.Sprintf("Account ID: %s", accountID), props.Text{
			Size:  12,
			Align: align.Center,
		}),
	)
	m.AddRow(5, line.NewCol(12))
}

func (s *service) addFooter(m core.Maroto, reportType ReportType) {
	m.AddRow(10, line.NewCol(12))

	var leftCol core.Col = col.New(6)
	if reportType == CostReport || reportType == TrendReport {
		leftCol = text.NewCol(6, "All costs shown in this report are Unblended.", props.Text{
			Size:  8,
			Align: align.Left,
			Style: fontstyle.Underline,
		})
	}

	m.AddRow(15,
		leftCol,
		text.NewCol(6, "Report generated by aws-doctor", props.Text{
			Size:      8,
			Align:     align.Right,
			Hyperlink: aws.String("https://github.com/elC0mpa/aws-doctor"),
		}),
	)
}
