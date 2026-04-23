package report

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/pricing"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

//go:embed assets/logo-pdf.png
var logoBytes []byte

type service struct{}

// NewService creates a new report service.
func NewService() Service {
	return &service{}
}

func (s *service) GenerateCostComparisonReport(input model.RenderCostComparisonInput, reportPath string) (*string, error) {
	path := s.getReportPath(reportPath, "cost")

	m := maroto.New()

	s.addHeader(m, CostReport, input.AccountID)

	s.addCostComparisonTable(m, input)

	s.addFooter(m, CostReport)

	return s.generateAndSave(m, path)
}

func (s *service) GenerateTrendReport(accountID string, costInfo []model.CostInfo, services []string, reportPath string) (*string, error) {
	path := s.getReportPath(reportPath, "trend")

	m := maroto.New()

	s.addHeader(m, TrendReport, accountID)

	if err := s.addTrendContent(m, costInfo, services); err != nil {
		return nil, err
	}

	s.addFooter(m, TrendReport)

	return s.generateAndSave(m, path)
}

func (s *service) GenerateWasteReport(input model.RenderWasteInput, pricingSvc pricing.Service, reportPath string) (*string, error) {
	path := s.getReportPath(reportPath, "waste")

	m := maroto.New()

	s.addHeader(m, WasteReport, input.AccountID)

	hasWaste := s.addWasteSections(m, input, pricingSvc)

	if !hasWaste {
		m.AddRow(20,
			text.NewCol(12, "✅ Your account is healthy! No waste found.", props.Text{
				Size:  12,
				Style: fontstyle.Bold,
				Align: align.Center,
				Color: &props.Color{Red: 0, Green: 150, Blue: 0},
			}),
		)
	} else {
		s.addWasteSummary(m, input, pricingSvc)
	}

	s.addFooter(m, WasteReport)

	return s.generateAndSave(m, path)
}

func (s *service) getReportPath(reportPath, flow string) string {
	if reportPath != "" && reportPath != "DEFAULT" {
		return reportPath
	}

	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	documents := filepath.Join(home, "Documents")
	// If Documents folder doesn't exist (some Linux distros), use home
	if _, err := os.Stat(documents); os.IsNotExist(err) {
		documents = home
	}

	timestamp := time.Now().Format("20060102-150405")
	fileName := fmt.Sprintf("aws-doctor-%s-%s.pdf", flow, timestamp)

	return filepath.Join(documents, fileName)
}

func (s *service) generateAndSave(m core.Maroto, path string) (*string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	document, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	err = document.Save(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to save PDF: %w", err)
	}

	return &absPath, nil
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

func (s *service) addHeader(m core.Maroto, reportType Type, accountID string) {
	title := "AWS COST DIAGNOSIS"

	switch reportType {
	case TrendReport:
		title = "AWS COST TREND"
	case WasteReport:
		title = "AWS WASTE REPORT"
	}

	m.AddRow(15,
		image.NewFromBytesCol(3, logoBytes, extension.Png),
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

func (s *service) addFooter(m core.Maroto, reportType Type) {
	m.AddRow(10, line.NewCol(12))

	leftCol := col.New(6)

	if reportType == CostReport {
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
