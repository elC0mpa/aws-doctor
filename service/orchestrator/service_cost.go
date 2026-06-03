package orchestrator

import (
	"context"
	"errors"

	"github.com/elC0mpa/aws-doctor/service/output"
	"github.com/elC0mpa/aws-doctor/utils/spinner"

	"github.com/elC0mpa/aws-doctor/model"
)

type costService struct {
	cfg CostConfig
}

// NewCostService creates a new cost orchestrator service.
func NewCostService(cfg CostConfig) CostService {
	return &costService{cfg: cfg}
}

func (s *costService) CompareCosts(generateReport bool, reportPath string) error {
	currentMonthData, err := s.cfg.CostService.GetCurrentMonthCostsByService(context.Background())
	if err != nil {
		return s.handleCostError(err)
	}

	lastMonthData, err := s.cfg.CostService.GetLastMonthCostsByService(context.Background())
	if err != nil {
		return s.handleCostError(err)
	}

	currentTotalCost, err := s.cfg.CostService.GetCurrentMonthTotalCosts(context.Background())
	if err != nil {
		return err
	}

	lastTotalCost, err := s.cfg.CostService.GetLastMonthTotalCosts(context.Background())
	if err != nil {
		return err
	}

	stsResult, err := s.cfg.STSService.GetCallerIdentity(context.Background())
	if err != nil {
		return err
	}

	spinner.StopSpinner()

	input := model.RenderCostComparisonInput{
		AccountID:        *stsResult.Account,
		LastTotalCost:    *lastTotalCost,
		CurrentTotalCost: *currentTotalCost,
		LastMonth:        lastMonthData,
		CurrentMonth:     currentMonthData,
	}

	if generateReport {
		path, err := s.cfg.ReportService.GenerateCostComparisonReport(input, reportPath)
		if err != nil {
			return err
		}

		output.PrintReportSuccess(*path)

		return nil
	}

	return s.cfg.Renderer.RenderCostComparison(input)
}

func (s *costService) handleCostError(err error) error {
	if errors.Is(err, model.ErrFirstDayOfMonth) {
		spinner.StopSpinner()
		output.PrintFirstDayOfMonthError()

		return nil
	}

	return err
}
