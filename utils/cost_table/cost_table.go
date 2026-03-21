package costtable

import (
	"fmt"
	"os"

	"github.com/elC0mpa/aws-doctor/model"
	outputshared "github.com/elC0mpa/aws-doctor/utils/output_shared"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// DrawCostTable renders a table comparing costs between months.
func DrawCostTable(input model.RenderCostComparisonInput) {
	fmt.Printf("\n%s\n", text.FgHiWhite.Sprint(" 💰 AWS COST DIAGNOSIS"))
	fmt.Printf(" Account ID: %s\n", text.FgBlue.Sprint(input.AccountID))
	fmt.Println(text.FgHiBlue.Sprint(" ------------------------------------------------"))

	currentMonthHeader := fmt.Sprintf("Current Month\n(%s\n%s)", *input.CurrentMonth.Start, *input.CurrentMonth.End)
	lastMonthHeader := fmt.Sprintf("Last Month\n(%s\n%s)", *input.LastMonth.Start, *input.LastMonth.End)

	rowHeader := table.Row{
		"Service",
		lastMonthHeader,
		currentMonthHeader,
		"Difference",
	}

	tw := table.NewWriter()
	tw.SetOutputMirror(os.Stdout)
	tw.AppendHeader(rowHeader)

	var rows []table.Row

	rows = append(rows, populateFirstRow(input.LastTotalCost, input.CurrentTotalCost))

	orderedServicesCosts := outputshared.OrderCostServices(&input.CurrentMonth.CostGroup)

	for _, group := range orderedServicesCosts {
		rows = append(rows, populateRow(*input.LastMonth, group))
	}

	tw.AppendRows(rows)
	tw.SetStyle(table.StyleRounded)

	tw.SetColumnConfigs([]table.ColumnConfig{
		{
			Number:       1,
			VAlignHeader: text.VAlignMiddle,
		},
		{
			Number: 2,
			Align:  text.AlignRight,
		},
		{
			Number: 3,
			Align:  text.AlignRight,
		},
		{
			Number:       4,
			Align:        text.AlignRight,
			VAlignHeader: text.VAlignMiddle,
		},
	})
	tw.Render()
}

func populateFirstRow(lastTotalCost, currentTotalCost string) table.Row {
	lastTotalAmount, _ := outputshared.ParseCostString(lastTotalCost)
	currentTotalAmount, unit := outputshared.ParseCostString(currentTotalCost)

	difference := currentTotalAmount - lastTotalAmount

	row := make(table.Row, 4)
	row[0] = text.FgHiGreen.Sprint("Total Costs")
	row[1] = text.FgHiYellow.Sprintf("%s", lastTotalCost)
	row[2] = text.FgHiGreen.Sprintf("%s", currentTotalCost)
	row[3] = text.FgHiGreen.Sprintf("%s", outputshared.FormatCost(difference, unit))

	if difference > 0 {
		row[2] = text.FgHiRed.Sprintf("%s", currentTotalCost)
		row[0] = text.FgHiRed.Sprintf("Total Costs")
		row[3] = text.FgHiRed.Sprintf("%s", outputshared.FormatCost(difference, unit))
	}

	return row
}

func populateRow(lastMonthGroups model.CostInfo, currentMonthGroup model.ServiceCost) table.Row {
	row := make(table.Row, 4)

	serviceName := currentMonthGroup.Name
	lastMonthGroup := lastMonthGroups.CostGroup[serviceName]

	currentServiceCost := outputshared.FormatCost(currentMonthGroup.Amount, currentMonthGroup.Unit)
	lastServiceCost := outputshared.FormatCost(lastMonthGroup.Amount, lastMonthGroup.Unit)

	difference := currentMonthGroup.Amount - lastMonthGroup.Amount

	row[0] = text.FgGreen.Sprintf("%s", serviceName)
	row[1] = text.FgYellow.Sprintf("%s", lastServiceCost)
	row[2] = text.FgGreen.Sprintf("%s", currentServiceCost)
	row[3] = text.FgGreen.Sprintf("%s", outputshared.FormatCost(difference, currentMonthGroup.Unit))

	if difference > 0 {
		row[0] = text.FgRed.Sprintf("%s", serviceName)
		row[2] = text.FgRed.Sprintf("%s", currentServiceCost)
		row[3] = text.FgRed.Sprintf("%s", outputshared.FormatCost(difference, currentMonthGroup.Unit))
	}

	return row
}
