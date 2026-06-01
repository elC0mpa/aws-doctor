package cmd

import (
	"testing"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/orchestrator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	ec2        = "ec2"
	s3Const    = "s3"
	dev        = "dev"
	none       = "none"
	unknown    = "unknown"
	table      = "table"
	outputJSON = "json"
)

// MockWasteOrchestrator
type MockWasteOrchestrator struct{ mock.Mock }

func (m *MockWasteOrchestrator) AnalyzeWaste(flags model.Flags) error {
	return m.Called(flags).Error(0)
}

// MockCostOrchestrator
type MockCostOrchestrator struct{ mock.Mock }

func (m *MockCostOrchestrator) CompareCosts(generateReport bool, reportPath string) error {
	return m.Called(generateReport, reportPath).Error(0)
}

// MockTrendOrchestrator
type MockTrendOrchestrator struct{ mock.Mock }

func (m *MockTrendOrchestrator) AnalyzeTrends(trendChecks []string, generateReport bool, reportPath string) error {
	return m.Called(trendChecks, generateReport, reportPath).Error(0)
}

// MockSystemOrchestrator
type MockSystemOrchestrator struct{ mock.Mock }

func (m *MockSystemOrchestrator) Update() error  { return m.Called().Error(0) }
func (m *MockSystemOrchestrator) Version() error { return m.Called().Error(0) }
func (m *MockSystemOrchestrator) CheckForUpdateInBackground() <-chan model.VersionCheckResult {
	args := m.Called()

	ch := make(chan model.VersionCheckResult, 1)
	if val := args.Get(0); val != nil {
		ch <- val.(model.VersionCheckResult)
	}

	return ch
}

func setupTest() (*MockWasteOrchestrator, *MockCostOrchestrator, *MockTrendOrchestrator, *MockSystemOrchestrator, func()) {
	mockWaste := new(MockWasteOrchestrator)
	mockCost := new(MockCostOrchestrator)
	mockTrend := new(MockTrendOrchestrator)
	mockSystem := new(MockSystemOrchestrator)

	origWaste := buildWasteOrchestratorHook
	origCost := buildCostOrchestratorHook
	origTrend := buildTrendOrchestratorHook
	origSystem := buildSystemOrchestratorHook

	buildWasteOrchestratorHook = func() (orchestrator.WasteService, error) { return mockWaste, nil }
	buildCostOrchestratorHook = func() (orchestrator.CostService, error) { return mockCost, nil }
	buildTrendOrchestratorHook = func() (orchestrator.TrendService, error) { return mockTrend, nil }
	buildSystemOrchestratorHook = func() (orchestrator.SystemService, error) { return mockSystem, nil }

	mockSystem.On("CheckForUpdateInBackground").Return(model.VersionCheckResult{})

	return mockWaste, mockCost, mockTrend, mockSystem, func() {
		buildWasteOrchestratorHook = origWaste
		buildCostOrchestratorHook = origCost
		buildTrendOrchestratorHook = origTrend
		buildSystemOrchestratorHook = origSystem

		region = ""
		profile = ""
		outputFormat = table
		lambdaMemoryThreshold = 10
		rootCmd.PersistentFlags().Lookup("output").Changed = false
		rootCmd.PersistentFlags().Lookup("region").Changed = false
		rootCmd.PersistentFlags().Lookup("profile").Changed = false
	}
}

func TestExecuteVersion(t *testing.T) {
	_, _, _, mockSystem, teardown := setupTest()
	defer teardown()

	mockSystem.On("Version").Return(nil)

	rootCmd.SetArgs([]string{"version"})

	err := Execute(dev, none, unknown)
	assert.NoError(t, err)
	mockSystem.AssertExpectations(t)
}

func TestExecuteUpdate(t *testing.T) {
	_, _, _, mockSystem, teardown := setupTest()
	defer teardown()

	mockSystem.On("Update").Return(nil)

	rootCmd.SetArgs([]string{"update"})

	err := Execute(dev, none, unknown)
	assert.NoError(t, err)
	mockSystem.AssertExpectations(t)
}

func TestExecuteTrend(t *testing.T) {
	_, _, mockTrend, _, teardown := setupTest()
	defer teardown()

	mockTrend.On("AnalyzeTrends", mock.Anything, false, "").Return(nil)

	rootCmd.SetArgs([]string{"trend"})

	err := Execute(dev, none, unknown)
	assert.NoError(t, err)
	mockTrend.AssertExpectations(t)
}

func TestExecuteWaste(t *testing.T) {
	mockWaste, _, _, _, teardown := setupTest()
	defer teardown()

	mockWaste.On("AnalyzeWaste", mock.Anything).Return(nil)

	rootCmd.SetArgs([]string{"waste"})

	err := Execute(dev, none, unknown)
	assert.NoError(t, err)
	mockWaste.AssertExpectations(t)
}

func TestExecuteCost(t *testing.T) {
	_, mockCost, _, _, teardown := setupTest()
	defer teardown()

	mockCost.On("CompareCosts", false, "").Return(nil)

	rootCmd.SetArgs([]string{"cost"})

	err := Execute(dev, none, unknown)
	assert.NoError(t, err)
	mockCost.AssertExpectations(t)
}

func TestExecuteReportCost(t *testing.T) {
	_, mockCost, _, _, teardown := setupTest()
	defer teardown()

	mockCost.On("CompareCosts", true, mock.Anything).Return(nil)

	rootCmd.SetArgs([]string{"report", "cost", "--path", "test.pdf"})

	err := Execute(dev, none, unknown)
	assert.NoError(t, err)
	mockCost.AssertExpectations(t)
}

func TestExecuteReportWaste(t *testing.T) {
	mockWaste, _, _, _, teardown := setupTest()
	defer teardown()

	mockWaste.On("AnalyzeWaste", mock.MatchedBy(func(f model.Flags) bool {
		return f.Report == true
	})).Return(nil)

	rootCmd.SetArgs([]string{"report", "waste", "--path", "test.pdf"})

	err := Execute(dev, none, unknown)
	assert.NoError(t, err)
	mockWaste.AssertExpectations(t)
}

func TestExecuteReportTrend(t *testing.T) {
	_, _, mockTrend, _, teardown := setupTest()
	defer teardown()

	mockTrend.On("AnalyzeTrends", mock.Anything, true, mock.Anything).Return(nil)

	rootCmd.SetArgs([]string{"report", "trend", "--path", "test.pdf"})

	err := Execute(dev, none, unknown)
	assert.NoError(t, err)
	mockTrend.AssertExpectations(t)
}

func TestBuilders(t *testing.T) {
	t.Setenv("AWS_REGION", "us-east-1")
	t.Setenv("AWS_ACCESS_KEY_ID", "mock")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "mock")

	_, _ = buildSystemOrchestrator()
	_, _ = buildWasteOrchestrator()
	_, _ = buildCostOrchestrator()
	_, _ = buildTrendOrchestrator()
}
