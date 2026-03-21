package csvoutput

import (
	"strings"

	"github.com/elC0mpa/aws-doctor/model"
	outputshared "github.com/elC0mpa/aws-doctor/utils/output_shared"
)

func mapTrendRows(monthlyCosts []model.CostInfo, services []string) [][]string {
	presented := outputshared.PresentTrend(monthlyCosts)
	result := make([][]string, 0, len(presented))

	servicesStr := strings.Join(services, ", ")

	for _, row := range presented {
		csvRow := row.ToSlice()
		if len(services) > 0 {
			csvRow = append(csvRow, servicesStr)
		}
		result = append(result, csvRow)
	}

	return result
}
