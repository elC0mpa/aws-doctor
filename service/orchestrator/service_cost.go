package orchestrator

import (
	"context"
	"errors"

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

	s.cfg.OutputService.StopSpinner()

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

		s.cfg.OutputService.PrintReportSuccess(*path)

		return nil
	}

	return s.cfg.OutputService.RenderCostComparison(input)
}

func (s *costService) handleCostError(err error) error {
	if errors.Is(err, model.ErrFirstDayOfMonth) {
		s.cfg.OutputService.StopSpinner()
		s.cfg.OutputService.PrintFirstDayOfMonthError()

		return nil
	}

	return err
}
