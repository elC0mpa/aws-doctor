package output

import (
	"fmt"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/pricing"
	csvoutput "github.com/elC0mpa/aws-doctor/utils/csv_output"
)

type csvRenderer struct{}

func (r *csvRenderer) RenderCostComparison(input model.RenderCostComparisonInput) error {
	return csvoutput.OutputCostComparisonCSV(input)
}

func (r *csvRenderer) RenderTrend(accountID string, costInfo []model.CostInfo, services []string) error {
	return csvoutput.OutputTrendCSV(costInfo, services)
}

func (r *csvRenderer) RenderWaste(input model.RenderWasteInput, pricingSvc pricing.Service) error {
	return csvoutput.OutputWasteCSV(input, pricingSvc)
}

func (r *csvRenderer) RenderWasteInteractive(accountID string, resultCh <-chan model.ScopeResult, scopes []string, pricingSvc pricing.Service) error {
	return fmt.Errorf("interactive mode not supported for CSV format")
}

func (r *csvRenderer) IsInteractive() bool {
	return false
}
