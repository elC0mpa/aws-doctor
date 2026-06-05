package orchestrator

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/elC0mpa/aws-doctor/service/output"
	"github.com/elC0mpa/aws-doctor/utils/spinner"

	"github.com/elC0mpa/aws-doctor/model"

	"github.com/elC0mpa/aws-doctor/utils/slice"

	"golang.org/x/sync/errgroup"
)

type wasteService struct {
	cfg WasteConfig
}

// NewWasteService creates a new waste orchestrator service.
func NewWasteService(cfg WasteConfig) WasteService {
	return &wasteService{cfg: cfg}
}

func (s *wasteService) AnalyzeWaste(flags model.Flags) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.loadPricing(ctx)

	stsResult, err := s.cfg.STSService.GetCallerIdentity(ctx)
	if err != nil {
		return err
	}

	resultCh := make(chan model.ScopeResult, 20)
	g, ctx := errgroup.WithContext(ctx)

	analyzers := s.cfg.Registry.GetAnalyzers()
	for _, a := range analyzers {
		if !shouldRunCheck(flags.WasteChecks, a.Name()) {
			continue
		}

		analyzer := a

		g.Go(func() error {
			res, err := analyzer.Analyze(ctx, flags)

			res.Scope = analyzer.TabName()
			res.Input.Flags = flags

			if err != nil {
				res.Err = err
			}

			if res.Err != nil {
				if res.Input.Errors == nil {
					res.Input.Errors = make(map[string]string)
				}

				res.Input.Errors[res.Scope] = res.Err.Error()
			}

			select {
			case resultCh <- res:
			case <-ctx.Done():
				return ctx.Err()
			}

			return nil
		})
	}

	// Wait and close channel in background
	go func() {
		_ = g.Wait()

		close(resultCh)
	}()

	isInteractive := s.cfg.Renderer.IsInteractive() && !flags.Report
	if isInteractive {
		spinner.StopSpinner()

		var scopes []string

		for _, a := range analyzers {
			if shouldRunCheck(flags.WasteChecks, a.Name()) {
				scopes = append(scopes, a.TabName())
			}
		}

		err := s.cfg.Renderer.RenderWasteInteractive(*stsResult.Account, resultCh, scopes, s.cfg.PricingService)

		cancel() // ensure context is cancelled to release blocked analyzers

		workflowErr := g.Wait()

		if err != nil {
			output.PrintWasteError(err)
			return err
		}

		return workflowErr
	}

	finalInput := model.RenderWasteInput{AccountID: *stsResult.Account}
	for res := range resultCh {
		finalInput.Merge(res.Input)
	}

	cancel() // ensure context is cancelled

	if err := g.Wait(); err != nil {
		spinner.StopSpinner()
		return err
	}

	spinner.StopSpinner()

	if flags.Report {
		return s.handleWasteReport(finalInput, flags.ReportPath)
	}

	return s.cfg.Renderer.RenderWaste(finalInput, s.cfg.PricingService)
}

func shouldRunCheck(wasteChecks []string, name string) bool {
	return len(wasteChecks) == 0 || slice.ContainsIgnoreCase(wasteChecks, name)
}

func (s *wasteService) handleWasteReport(input model.RenderWasteInput, reportPath string) error {
	path, err := s.cfg.ReportService.GenerateWasteReport(input, s.cfg.PricingService, reportPath)
	if err != nil {
		return err
	}

	output.PrintReportSuccess(*path)

	return nil
}

func (s *wasteService) loadPricing(ctx context.Context) {
	spinner.SetMessage("Gathering pricing data...")

	pricingCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	if err := s.cfg.PricingService.LoadRegionRates(pricingCtx); err != nil {
		fmt.Fprintf(os.Stderr, "warning: pricing API partial failure, falling back to defaults: %v\n", err)
	}

	spinner.SetMessage("Please wait while data is being fetched...")
}
