package output

import (
	"os"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/pricing"
	"github.com/elC0mpa/aws-doctor/utils/barchart"
	costtable "github.com/elC0mpa/aws-doctor/utils/cost_table"
	"github.com/elC0mpa/aws-doctor/utils/tui"
	wastetable "github.com/elC0mpa/aws-doctor/utils/waste_table"
	"golang.org/x/term"
)

type tableRenderer struct{}

func (r *tableRenderer) RenderCostComparison(input model.RenderCostComparisonInput) error {
	costtable.DrawCostTable(input)
	return nil
}

func (r *tableRenderer) RenderTrend(accountID string, costInfo []model.CostInfo, services []string) error {
	barchart.DrawTrendChart(accountID, costInfo)
	return nil
}

func (r *tableRenderer) RenderWaste(input model.RenderWasteInput, pricingSvc pricing.Service) error {
	wastetable.DrawWasteTable(os.Stdout, input, pricingSvc)
	return nil
}

func (r *tableRenderer) RenderWasteInteractive(accountID string, resultCh <-chan model.ScopeResult, scopes []string, pricingSvc pricing.Service) error {
	return tui.RenderWasteInteractive(accountID, resultCh, scopes, pricingSvc)
}

func (r *tableRenderer) IsInteractive() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
