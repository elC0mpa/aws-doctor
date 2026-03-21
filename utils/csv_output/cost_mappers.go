package csvoutput

import (
	"github.com/elC0mpa/aws-doctor/model"
	outputshared "github.com/elC0mpa/aws-doctor/utils/output_shared"
)

func mapTotalRow(lastTotal, currentTotal string) []string {
	lastAmount, lastUnit := outputshared.ParseCostString(lastTotal)
	currentAmount, currentUnit := outputshared.ParseCostString(currentTotal)
	diff := currentAmount - lastAmount

	return []string{
		"Total Costs",
		outputshared.FormatCost(lastAmount, lastUnit),
		outputshared.FormatCost(currentAmount, currentUnit),
		outputshared.FormatCost(diff, currentUnit),
	}
}

func mapServiceRow(lastMonth model.CostInfo, currentService model.ServiceCost) []string {
	lastService := lastMonth.CostGroup[currentService.Name]
	diff := currentService.Amount - lastService.Amount

	return []string{
		currentService.Name,
		outputshared.FormatCost(lastService.Amount, lastService.Unit),
		outputshared.FormatCost(currentService.Amount, currentService.Unit),
		outputshared.FormatCost(diff, currentService.Unit),
	}
}

func orderCostServices(costGroups *model.CostGroup) []model.ServiceCost {
	return outputshared.OrderCostServices(costGroups)
}
