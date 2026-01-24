package utils //nolint:revive

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func DrawCostTable(accountID string, lastTotalCost, currenttotalCost string, lastMonthGroups, currentMonthGroups *model.CostInfo, _ string) {
	fmt.Printf("\n%s\n", text.FgHiWhite.Sprint(" 💰 AWS COST DIAGNOSIS"))
	fmt.Printf(" Account ID: %s\n", text.FgBlue.Sprint(accountID))
	fmt.Println(text.FgHiBlue.Sprint(" ------------------------------------------------"))

	currentMonthHeader := fmt.Sprintf("Current Month\n(%s\n%s)", *currentMonthGroups.Start, *currentMonthGroups.End)
	lastMonthHeader := fmt.Sprintf("Last Month\n(%s\n%s)", *lastMonthGroups.Start, *lastMonthGroups.End)

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

	rows = append(rows, populateFirstRow(lastTotalCost, currenttotalCost))

	orderedServicesCosts := orderCostServices(&currentMonthGroups.CostGroup)

	for _, group := range orderedServicesCosts {
		rows = append(rows, populateRow(*lastMonthGroups, group))
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

func orderCostServices(costGroups *model.CostGroup) []model.ServiceCost {
	sortedServices := make([]model.ServiceCost, 0, len(*costGroups))
	for key, group := range *costGroups {
		sortedServices = append(sortedServices, model.ServiceCost{
			Name:   key,
			Amount: group.Amount,
			Unit:   group.Unit,
		})
	}

	sort.Slice(sortedServices, func(i, j int) bool {
		return sortedServices[i].Amount > sortedServices[j].Amount
	})

	return sortedServices
}

func populateFirstRow(lastTotalCost, currentTotalCost string) table.Row {
	currentTotalSplitted := strings.Split(currentTotalCost, " ")
	lastTotalSplitted := strings.Split(lastTotalCost, " ")

	currentTotalAmount, err := strconv.ParseFloat(currentTotalSplitted[0], 64)
	if err != nil {
		panic("Error parsing current month total cost")
	}

	lastTotalAmount, err := strconv.ParseFloat(lastTotalSplitted[0], 64)
	if err != nil {
		panic("Error parsing last month total cost")
	}

	difference := currentTotalAmount - lastTotalAmount

	row := make(table.Row, 4)
	row[0] = text.FgHiGreen.Sprint("Total Costs")
	row[1] = text.FgHiYellow.Sprintf("%s", lastTotalCost)
	row[2] = text.FgHiGreen.Sprintf("%s", currentTotalCost)
	row[3] = text.FgHiGreen.Sprintf("%.2f %s", difference, currentTotalSplitted[1])

	if difference > 0 {
		row[2] = text.FgHiRed.Sprintf("%s", currentTotalCost)
		row[0] = text.FgHiRed.Sprintf("Total Costs")
		row[3] = text.FgHiRed.Sprintf("%.2f %s", difference, currentTotalSplitted[1])
	}

	return row
}

func populateRow(lastMonthGroups model.CostInfo, currentMonthGroup model.ServiceCost) table.Row {
	row := make(table.Row, 4)

	serviceName := currentMonthGroup.Name
	lastMonthGroup := lastMonthGroups.CostGroup[serviceName]

	currentServiceCost := fmt.Sprintf("%.2f %s", currentMonthGroup.Amount, currentMonthGroup.Unit)
	lastServiceCost := fmt.Sprintf("%.2f %s", lastMonthGroup.Amount, lastMonthGroup.Unit)

	difference := currentMonthGroup.Amount - lastMonthGroup.Amount

	row[0] = text.FgGreen.Sprintf("%s", serviceName)
	row[1] = text.FgYellow.Sprintf("%s", lastServiceCost)
	row[2] = text.FgGreen.Sprintf("%s", currentServiceCost)
	row[3] = text.FgGreen.Sprintf("%.2f %s", difference, currentMonthGroup.Unit)

	if difference > 0 {
		row[0] = text.FgRed.Sprintf("%s", serviceName)
		row[2] = text.FgRed.Sprintf("%s", currentServiceCost)
		row[3] = text.FgRed.Sprintf("%.2f %s", difference, currentMonthGroup.Unit)
	}

	return row
}
