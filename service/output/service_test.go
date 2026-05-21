package output

import (
	"errors"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cetypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/elC0mpa/aws-doctor/mocks/services"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/pricing"
)

func runSilent(f func()) {
	oldOut := os.Stdout
	oldErr := os.Stderr
	_, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	defer func() {
		_ = w.Close()

		os.Stdout = oldOut
		os.Stderr = oldErr
	}()

	f()
}

func TestIsInteractive(t *testing.T) {
	s1 := NewService("table")
	if !s1.IsInteractive() {
		t.Error("Expected IsInteractive to be true for table format")
	}

	s2 := NewService("json")
	if s2.IsInteractive() {
		t.Error("Expected IsInteractive to be false for json format")
	}
}

func TestRenderWasteInteractive_Smoke(t *testing.T) {
	oldFn := renderWasteInteractiveFn

	defer func() { renderWasteInteractiveFn = oldFn }()

	called := false
	renderWasteInteractiveFn = func(accountID string, resultCh <-chan model.ScopeResult, scopes []string, pricingSvc pricing.Service) error {
		called = true
		return nil
	}

	s := NewService("table")
	resultCh := make(chan model.ScopeResult)

	go func() {
		close(resultCh)
	}()

	err := s.RenderWasteInteractive("123456789012", resultCh, []string{"EC2"}, nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !called {
		t.Error("Expected renderWasteInteractiveFn to be called")
	}
}

func TestService_WrapperMethods(t *testing.T) {
	sTable := NewService("table")
	sJSON := NewService("json")
	sCSV := NewService("csv")

	runSilent(func() {
		sTable.PrintAlreadyLatest("v1.0.0")
		sTable.PrintHomebrewUpdate()
		sTable.PrintGoInstallUpdate()
		sTable.PrintRateLimitError()
		sTable.PrintUpdateError(errors.New("test"))
		sTable.PrintWasteError(errors.New("test"))
		sTable.PrintFirstDayOfMonthError()
		sTable.PrintNewVersionAvailable("v1.0.0", "v1.1.0")
		sTable.SetSpinnerMessage("loading")
		sTable.StopSpinner()
		sTable.PrintReportSuccess("/path/to/report")
		sTable.RenderVersion(model.VersionInfo{Version: "v1.0.0"})

		costInfo := &model.CostInfo{
			DateInterval: cetypes.DateInterval{
				Start: aws.String("2024-01-01"),
				End:   aws.String("2024-01-31"),
			},
		}
		inputCC := model.RenderCostComparisonInput{
			LastMonth:    costInfo,
			CurrentMonth: costInfo,
		}
		_ = sTable.RenderCostComparison(inputCC)
		_ = sJSON.RenderCostComparison(inputCC)
		_ = sCSV.RenderCostComparison(inputCC)

		_ = sTable.RenderTrend("123456789012", []model.CostInfo{*costInfo}, []string{"EC2"})
		_ = sJSON.RenderTrend("123456789012", []model.CostInfo{*costInfo}, []string{"EC2"})
		_ = sCSV.RenderTrend("123456789012", []model.CostInfo{*costInfo}, []string{"EC2"})

		inputWaste := model.RenderWasteInput{}
		mockPricingSvc := services.NewMockPricingService()
		_ = sTable.RenderWaste(inputWaste, mockPricingSvc)
		_ = sJSON.RenderWaste(inputWaste, mockPricingSvc)
		_ = sCSV.RenderWaste(inputWaste, mockPricingSvc)
	})
}
