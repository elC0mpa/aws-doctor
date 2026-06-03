package output

import (
	"fmt"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/pricing"
	jsonoutput "github.com/elC0mpa/aws-doctor/utils/json_output"
)

type jsonRenderer struct{}

func (r *jsonRenderer) RenderCostComparison(input model.RenderCostComparisonInput) error {
	return jsonoutput.OutputCostComparisonJSON(input)
}

func (r *jsonRenderer) RenderTrend(accountID string, costInfo []model.CostInfo, services []string) error {
	return jsonoutput.OutputTrendJSON(accountID, costInfo, services)
}

func (r *jsonRenderer) RenderWaste(input model.RenderWasteInput, pricingSvc pricing.Service) error {
	return jsonoutput.OutputWasteJSON(input, pricingSvc)
}

func (r *jsonRenderer) RenderWasteInteractive(accountID string, resultCh <-chan model.ScopeResult, scopes []string, pricingSvc pricing.Service) error {
	return fmt.Errorf("interactive mode not supported for JSON format")
}

func (r *jsonRenderer) IsInteractive() bool {
	return false
}
