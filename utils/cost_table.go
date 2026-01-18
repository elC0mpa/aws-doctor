package utils

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func DrawCostTable(accountId string, lastTotalCost, currenttotalCost string, lastMonthGroups, currentMonthGroups *model.CostInfo, costAggregation string) {
	currentMonthHeader := fmt.Sprintf("Current Month\n(%s\n%s)", *currentMonthGroups.Start, *currentMonthGroups.End)
	lastMonthHeader := fmt.Sprintf("Last Month\n(%s\n%s)", *lastMonthGroups.Start, *lastMonthGroups.End)

	rowHeader := table.Row{
		"Account ID",
		"Service",
		lastMonthHeader,
		currentMonthHeader,
		"Difference",
	}

	tw := table.Table{}

	tw.AppendHeader(rowHeader)
	var rows []table.Row

	rows = append(rows, populateFirstRow(lastTotalCost, currenttotalCost))

	orderedServicesCosts := orderCostServices(&currentMonthGroups.CostGroup)

	for _, group := range orderedServicesCosts {
		rows = append(rows, populateRow(*lastMonthGroups, group))
	}

	halfRow := len(rows) / 2
	rows[halfRow][0] = text.FgBlue.Sprintf("%s", accountId)
	tw.AppendRows(rows)
	tw.SetStyle(table.StyleRounded)
	tw.SetColumnConfigs([]table.ColumnConfig{
		{
			Number:       1,
			VAlignHeader: text.VAlignMiddle,
		},
		{
			Number:       2,
			VAlignHeader: text.VAlignMiddle,
		},
		{
			Number: 3,
			Align:  text.AlignRight,
		},
		{
			Number: 4,
			Align:  text.AlignRight,
		},
		{
			Number:       5,
			Align:        text.AlignRight,
			VAlignHeader: text.VAlignMiddle,
		},
	})
	fmt.Println(tw.Render())
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

	row := make(table.Row, 5)
	row[0] = ""
	row[1] = text.FgHiGreen.Sprint("Total Costs")
	row[2] = text.FgHiYellow.Sprintf("%s", lastTotalCost)
	row[3] = text.FgHiGreen.Sprintf("%s", currentTotalCost)
	row[4] = text.FgHiGreen.Sprintf("%.2f %s", difference, currentTotalSplitted[1])

	if difference > 0 {
		row[3] = text.FgHiRed.Sprintf("%s", currentTotalCost)
		row[1] = text.FgHiRed.Sprintf("Total Costs")
		row[4] = text.FgHiRed.Sprintf("%.2f %s", difference, currentTotalSplitted[1])
	}

	return row
}

func populateRow(lastMonthGroups model.CostInfo, currentMonthGroup model.ServiceCost) table.Row {
	row := make(table.Row, 5)

	serviceName := currentMonthGroup.Name
	lastMonthGroup := lastMonthGroups.CostGroup[serviceName]

	currentServiceCost := fmt.Sprintf("%.2f %s", currentMonthGroup.Amount, currentMonthGroup.Unit)
	lastServiceCost := fmt.Sprintf("%.2f %s", lastMonthGroup.Amount, lastMonthGroup.Unit)

	difference := currentMonthGroup.Amount - lastMonthGroup.Amount

	row[0] = ""
	row[1] = text.FgGreen.Sprintf("%s", serviceName)
	row[2] = text.FgYellow.Sprintf("%s", lastServiceCost)
	row[3] = text.FgGreen.Sprintf("%s", currentServiceCost)
	row[4] = text.FgGreen.Sprintf("%.2f %s", difference, currentMonthGroup.Unit)

	if difference > 0 {
		row[1] = text.FgRed.Sprintf("%s", serviceName)
		row[3] = text.FgRed.Sprintf("%s", currentServiceCost)
		row[4] = text.FgRed.Sprintf("%.2f %s", difference, currentMonthGroup.Unit)
	}

	return row
}
