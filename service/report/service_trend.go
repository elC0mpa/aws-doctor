package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/elC0mpa/aws-doctor/model"
	awscostexplorer "github.com/elC0mpa/aws-doctor/service/costexplorer"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/vicanso/go-charts/v2"
)

func (s *service) addTrendContent(m core.Maroto, costInfo []model.CostInfo, services []string) error {
	// Services info (Maroto layer for wrapping support)
	if len(services) > 0 {
		var fullNames []string

		for _, svc := range services {
			if fullName, ok := awscostexplorer.ServiceNameMap[strings.ToLower(svc)]; ok {
				fullNames = append(fullNames, fullName)
			} else {
				fullNames = append(fullNames, svc)
			}
		}

		servicesLabel := "Services Included: " + strings.Join(fullNames, ", ")

		// Calculate dynamic height (approx 100 chars per line at size 9)
		numLines := (len(servicesLabel) / 100) + 1
		m.AddRow(float64(numLines*7),
			text.NewCol(12, servicesLabel, props.Text{
				Size:  9,
				Style: fontstyle.Italic,
				Align: align.Center,
			}),
		)
		m.AddRow(5, col.New(12))
	}

	// Add Chart
	chartBytes, err := s.generateTrendChartImage(costInfo)
	if err != nil {
		return fmt.Errorf("failed to generate trend chart: %w", err)
	}

	m.AddRow(80,
		image.NewFromBytesCol(12, chartBytes, extension.Png),
	)

	m.AddRow(10, line.NewCol(12))

	// Add Breakdown Table
	s.addTrendBreakdownTable(m, costInfo)

	return nil
}

func (s *service) generateTrendChartImage(costInfo []model.CostInfo) ([]byte, error) {
	xAxisData := make([]string, 0, len(costInfo))
	seriesData := make([]charts.SeriesData, len(costInfo))

	// Find min and max amounts for scaling colors
	maxAmount := 0.0
	minAmount := -1.0

	for _, info := range costInfo {
		amt := info.CostGroup["Total"].Amount

		if amt > maxAmount {
			maxAmount = amt
		}

		if minAmount == -1.0 || amt < minAmount {
			minAmount = amt
		}
	}

	// Define our color poles
	// Dark Blue (Max)
	maxColor := struct{ r, g, b int }{31, 78, 121}
	// Light Blue (Min)
	minColor := struct{ r, g, b int }{189, 215, 238}

	for i, info := range costInfo {
		month := ""

		if info.Start != nil {
			if t, err := time.Parse("2006-01-02", *info.Start); err == nil {
				month = t.Format("Jan")
			}
		}

		amount := info.CostGroup["Total"].Amount

		xAxisData = append(xAxisData, fmt.Sprintf("%s: %.0f", month, amount))

		// Calculate color interpolation factor (0.0 to 1.0)
		factor := 1.0
		if maxAmount != minAmount {
			factor = (amount - minAmount) / (maxAmount - minAmount)
		}

		// Interpolate R, G, B
		r := uint8(float64(minColor.r) + factor*float64(maxColor.r-minColor.r))
		g := uint8(float64(minColor.g) + factor*float64(maxColor.g-minColor.g))
		b := uint8(float64(minColor.b) + factor*float64(maxColor.b-minColor.b))

		seriesData[i] = charts.SeriesData{
			Value: amount,
			Style: charts.Style{
				FillColor: charts.Color{R: r, G: g, B: b, A: 255},
			},
		}
	}

	p, err := charts.Render(
		charts.ChartOption{
			Title: charts.TitleOption{
				Text: "Monthly Total Costs",
				Left: charts.PositionCenter,
			},
			XAxis: charts.XAxisOption{
				Data: xAxisData,
			},
			SeriesList: charts.SeriesList{
				{
					Type: charts.ChartTypeBar,
					Data: seriesData,
				},
			},
			Height: 400,
			Width:  800,
		},
	)
	if err != nil {
		return nil, err
	}

	return p.Bytes()
}

func (s *service) addTrendBreakdownTable(m core.Maroto, costInfo []model.CostInfo) {
	m.AddRow(10,
		text.NewCol(12, "Monthly Total Costs Breakdown", props.Text{Style: fontstyle.Bold, Size: 11}),
	)

	// Determine unit
	unit := ""
	if len(costInfo) > 0 {
		unit = costInfo[0].CostGroup["Total"].Unit
	}

	// Table Header
	headerCols := []string{"Period"}
	for _, info := range costInfo {
		headerCols = append(headerCols, s.formatDateToMonthYear(info.Start))
	}

	serviceColSize := 3

	monthColSize := (12 - serviceColSize) / len(costInfo)

	if monthColSize == 0 {
		monthColSize = 1
	}

	headerRow := make([]core.Col, 0, len(headerCols))

	headerRow = append(headerRow, text.NewCol(serviceColSize, "Metric", props.Text{Style: fontstyle.Bold, Size: 9}))

	for i := 1; i < len(headerCols); i++ {
		headerRow = append(headerRow, text.NewCol(monthColSize, headerCols[i], props.Text{Style: fontstyle.Bold, Size: 8, Align: align.Right}))
	}

	m.AddRow(10, headerRow...)
	m.AddRow(2, line.NewCol(12))

	// Total Row
	totalRow := make([]core.Col, 0, len(costInfo)+1)

	totalRow = append(totalRow, text.NewCol(serviceColSize, "Total Cost", props.Text{Style: fontstyle.Bold, Size: 9}))

	for _, info := range costInfo {
		totalRow = append(totalRow, text.NewCol(monthColSize, fmt.Sprintf("%.2f", info.CostGroup["Total"].Amount), props.Text{Size: 8, Align: align.Right}))
	}

	m.AddRow(10, totalRow...)

	m.AddRow(5, text.NewCol(12, fmt.Sprintf("All amounts in %s", unit), props.Text{Size: 7, Align: align.Right, Top: 2}))
}
