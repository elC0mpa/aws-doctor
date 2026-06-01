// Package orchestrator coordinates the execution of various AWS service checks.
package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/elC0mpa/aws-doctor/model"
	awscostexplorer "github.com/elC0mpa/aws-doctor/service/costexplorer"
	"github.com/elC0mpa/aws-doctor/utils/slice"
	"github.com/google/go-github/v62/github"
	"golang.org/x/sync/errgroup"
)

// EC2-related thresholds for the various waste checks.
const (
	ec2StoppedDays            = 30
	ec2RiExpiringDays         = 30
	ec2AmiStaleDays           = 90
	ec2SnapshotStaleDays      = 90
	ec2IdleDays               = 14
	ec2IdleCPUPercent         = 5.0
	ec2IdleNetworkBytesPerDay = 5 * 1024 * 1024 // 5 MB/day, combined in+out
)

// Lookback windows and thresholds for the remaining waste checks.
const (
	sagemakerIdleDays  = 14
	vpcNatIdleDays     = 7
	elbIdleDays        = 7
	rdsIdleDays        = 7
	rdsSnapshotDays    = 30
	lambdaLookbackDays = 14
)

// --- System Service ---

type systemService struct {
	cfg SystemConfig
}

// NewSystemService creates a new system orchestrator service.
func NewSystemService(cfg SystemConfig) SystemService {
	return &systemService{cfg: cfg}
}

func (s *systemService) Version() error {
	s.cfg.OutputService.StopSpinner()
	s.cfg.OutputService.RenderVersion(s.cfg.VersionInfo)

	return nil
}

func (s *systemService) Update() error {
	s.cfg.OutputService.StopSpinner()

	err := s.cfg.UpdateService.Update()
	if err == nil {
		return nil
	}

	if errors.Is(err, model.ErrHomebrewInstall) {
		s.cfg.OutputService.PrintHomebrewUpdate()
		return nil
	}

	if errors.Is(err, model.ErrGoInstall) {
		s.cfg.OutputService.PrintGoInstallUpdate()
		return nil
	}

	if errors.Is(err, model.ErrAlreadyLatest) {
		s.cfg.OutputService.PrintAlreadyLatest(s.cfg.VersionInfo.Version)
		return nil
	}

	if errors.Is(err, model.ErrRateLimit) {
		s.cfg.OutputService.PrintRateLimitError()
		return err
	}

	var rateLimitErr *github.RateLimitError
	if errors.As(err, &rateLimitErr) {
		s.cfg.OutputService.PrintRateLimitError()
		return err
	}

	s.cfg.OutputService.PrintUpdateError(err)

	return err
}

func (s *systemService) CheckForUpdateInBackground() <-chan model.VersionCheckResult {
	versionCh := make(chan model.VersionCheckResult, 1)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		latest, err := s.cfg.UpdateService.CheckForUpdate(ctx)
		versionCh <- model.VersionCheckResult{LatestVersion: latest, Err: err}
	}()

	return versionCh
}

// --- Cost Service ---

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

// --- Trend Service ---

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

// --- Waste Service ---

type wasteService struct {
	cfg WasteConfig
}

// NewWasteService creates a new waste orchestrator service.
func NewWasteService(cfg WasteConfig) WasteService {
	return &wasteService{cfg: cfg}
}

func (s *wasteService) AnalyzeWaste(flags model.Flags) error {
	ctx := context.Background()

	s.loadPricing(ctx)

	stsResult, err := s.cfg.STSService.GetCallerIdentity(ctx)
	if err != nil {
		return err
	}

	resultCh := make(chan model.ScopeResult, 20)
	g, ctx := errgroup.WithContext(ctx)

	// Ensure hardcoded thresholds are still populated
	flags.EC2StoppedDays = ec2StoppedDays
	flags.EC2RiExpiringDays = ec2RiExpiringDays
	flags.EC2AmiStaleDays = ec2AmiStaleDays
	flags.EC2SnapshotStaleDays = ec2SnapshotStaleDays
	flags.EC2IdleDays = ec2IdleDays
	flags.EC2IdleCPUPercent = ec2IdleCPUPercent
	flags.EC2IdleNetworkBytesPerDay = ec2IdleNetworkBytesPerDay
	flags.SageMakerIdleDays = sagemakerIdleDays
	flags.VPCNatIdleDays = vpcNatIdleDays
	flags.ELBIdleDays = elbIdleDays
	flags.RDSIdleDays = rdsIdleDays
	flags.RDSSnapshotDays = rdsSnapshotDays
	flags.LambdaLookbackDays = lambdaLookbackDays

	analyzers := s.cfg.Registry.GetAnalyzers()
	for _, a := range analyzers {
		if !shouldRunCheck(flags.WasteChecks, a.Name()) {
			continue
		}

		analyzer := a

		g.Go(func() error {
			res, err := analyzer.Analyze(ctx, flags)

			res.Scope = analyzer.TabName()

			if err != nil {
				res.Err = err
			}

			if res.Err != nil {
				if res.Input.Errors == nil {
					res.Input.Errors = make(map[string]string)
				}

				res.Input.Errors[res.Scope] = res.Err.Error()
			}

			resultCh <- res

			return nil
		})
	}

	// Wait and close channel in background
	go func() {
		_ = g.Wait()

		close(resultCh)
	}()

	isInteractive := s.cfg.OutputService.IsInteractive() && !flags.Report
	if isInteractive {
		s.cfg.OutputService.StopSpinner()

		var scopes []string

		for _, a := range analyzers {
			if shouldRunCheck(flags.WasteChecks, a.Name()) {
				scopes = append(scopes, a.TabName())
			}
		}

		err := s.cfg.OutputService.RenderWasteInteractive(*stsResult.Account, resultCh, scopes, s.cfg.PricingService)
		workflowErr := g.Wait()

		if err != nil {
			s.cfg.OutputService.PrintWasteError(err)
			return err
		}

		return workflowErr
	}

	finalInput := model.RenderWasteInput{AccountID: *stsResult.Account}
	for res := range resultCh {
		finalInput.Merge(res.Input)
	}

	if err := g.Wait(); err != nil {
		s.cfg.OutputService.StopSpinner()
		return err
	}

	s.cfg.OutputService.StopSpinner()

	if flags.Report {
		return s.handleWasteReport(finalInput, flags.ReportPath)
	}

	return s.cfg.OutputService.RenderWaste(finalInput, s.cfg.PricingService)
}

func shouldRunCheck(wasteChecks []string, name string) bool {
	return len(wasteChecks) == 0 || slice.ContainsIgnoreCase(wasteChecks, name)
}

func (s *wasteService) handleWasteReport(input model.RenderWasteInput, reportPath string) error {
	path, err := s.cfg.ReportService.GenerateWasteReport(input, s.cfg.PricingService, reportPath)
	if err != nil {
		return err
	}

	s.cfg.OutputService.PrintReportSuccess(*path)

	return nil
}

func (s *wasteService) loadPricing(ctx context.Context) {
	s.cfg.OutputService.SetSpinnerMessage("Gathering pricing data...")

	pricingCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	if err := s.cfg.PricingService.LoadRegionRates(pricingCtx); err != nil {
		fmt.Fprintf(os.Stderr, "warning: pricing API partial failure, falling back to defaults: %v\n", err)
	}

	s.cfg.OutputService.SetSpinnerMessage("Please wait while data is being fetched...")
}
