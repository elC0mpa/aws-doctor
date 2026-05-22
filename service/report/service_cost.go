package report

import (
	"fmt"

	"github.com/elC0mpa/aws-doctor/model"
	outputshared "github.com/elC0mpa/aws-doctor/utils/output_shared"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

func (s *service) addCostComparisonTable(m core.Maroto, input model.RenderCostComparisonInput) {
	_, unit := outputshared.ParseCostString(input.CurrentTotalCost)

	s.addCostTableHeader(m, input, unit)
	s.addCostTotalRow(m, input)
	s.addCostServiceBreakdown(m, input)
}

func (s *service) addCostTableHeader(m core.Maroto, input model.RenderCostComparisonInput, unit string) {
	lastMonthLabel := s.formatDateToMonthYear(input.LastMonth.Start)
	currentMonthLabel := s.formatDateToMonthYear(input.CurrentMonth.Start)

	m.AddRow(
		12,
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
}

func (s *service) addCostTotalRow(m core.Maroto, input model.RenderCostComparisonInput) {
	lastTotalAmount, _ := outputshared.ParseCostString(input.LastTotalCost)
	currentTotalAmount, _ := outputshared.ParseCostString(input.CurrentTotalCost)
	difference := currentTotalAmount - lastTotalAmount

	totalColor := &props.Color{Red: 0, Green: 150, Blue: 0} // Green
	if difference > 0 {
		totalColor = &props.Color{Red: 200, Green: 0, Blue: 0} // Red
	}

	m.AddRow(
		8,
		text.NewCol(4, "Total Costs", props.Text{Style: fontstyle.Bold, Size: 10, Color: totalColor}),
		text.NewCol(3, fmt.Sprintf("%.2f", lastTotalAmount), props.Text{Size: 10, Align: align.Right}),
		text.NewCol(3, fmt.Sprintf("%.2f", currentTotalAmount), props.Text{Size: 10, Align: align.Right, Color: totalColor}),
		text.NewCol(2, fmt.Sprintf("%.2f", difference), props.Text{Size: 10, Align: align.Right, Color: totalColor}),
	)
	m.AddRow(2, line.NewCol(12))
}

func (s *service) addCostServiceBreakdown(m core.Maroto, input model.RenderCostComparisonInput) {
	orderedServices := outputshared.OrderCostServices(&input.CurrentMonth.CostGroup)
	for _, currentSvc := range orderedServices {
		lastMonthGroup := input.LastMonth.CostGroup[currentSvc.Name]
		diff := currentSvc.Amount - lastMonthGroup.Amount

		rowColor := &props.Color{Red: 0, Green: 150, Blue: 0} // Green
		if diff > 0 {
			rowColor = &props.Color{Red: 200, Green: 0, Blue: 0} // Red
		}

		m.AddRow(
			8,
			text.NewCol(4, currentSvc.Name, props.Text{Size: 9, Color: rowColor}),
			text.NewCol(3, fmt.Sprintf("%.2f", lastMonthGroup.Amount), props.Text{Size: 9, Align: align.Right}),
			text.NewCol(3, fmt.Sprintf("%.2f", currentSvc.Amount), props.Text{Size: 9, Align: align.Right, Color: rowColor}),
			text.NewCol(2, fmt.Sprintf("%.2f", diff), props.Text{Size: 9, Align: align.Right, Color: rowColor}),
		)
	}
}
