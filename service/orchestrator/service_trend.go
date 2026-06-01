package orchestrator

import (
	"context"
	"strings"

	awscostexplorer "github.com/elC0mpa/aws-doctor/service/costexplorer"
)

type trendService struct {
	cfg TrendConfig
}

// NewTrendService creates a new trend orchestrator service.
func NewTrendService(cfg TrendConfig) TrendService {
	return &trendService{cfg: cfg}
}

func (s *trendService) AnalyzeTrends(trendChecks []string, generateReport bool, reportPath string) error {
	var mappedServices []string

	for _, svc := range trendChecks {
		if mapped, ok := awscostexplorer.ServiceNameMap[strings.ToLower(svc)]; ok {
			mappedServices = append(mappedServices, mapped)
		}
	}

	costInfo, err := s.cfg.CostService.GetLastSixMonthsCosts(context.Background(), mappedServices)
	if err != nil {
		return err
	}

	stsResult, err := s.cfg.STSService.GetCallerIdentity(context.Background())
	if err != nil {
		return err
	}

	s.cfg.OutputService.StopSpinner()

	if generateReport {
		path, err := s.cfg.ReportService.GenerateTrendReport(*stsResult.Account, costInfo, trendChecks, reportPath)
		if err != nil {
			return err
		}

		s.cfg.OutputService.PrintReportSuccess(*path)

		return nil
	}

	return s.cfg.OutputService.RenderTrend(*stsResult.Account, costInfo, trendChecks)
}
