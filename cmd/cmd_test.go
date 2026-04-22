package cmd

import (
	"errors"
	"testing"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/elC0mpa/aws-doctor/service/orchestrator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	ec2     = "ec2"
	s3Const = "s3"
	dev     = "dev"
	none    = "none"
	unknown = "unknown"
)

// MockOrchestrator is a mock implementation of the orchestrator service.
type MockOrchestrator struct {
	mock.Mock
}

func (m *MockOrchestrator) Orchestrate(flags model.Flags) error {
	args := m.Called(flags)

	return args.Error(0)
}

func setupTest() (*MockOrchestrator, func()) {
	mockOrch := new(MockOrchestrator)
	originalBuilder := orchestratorBuilder
	orchestratorBuilder = func(needsAWS bool) (orchestrator.Service, error) {
		return mockOrch, nil
	}

	return mockOrch, func() {
		orchestratorBuilder = originalBuilder
		// Reset persistent flags to default
		region = ""
		profile = ""
		outputFormat = "table"
		lambdaMemoryThreshold = 10 // reset to default
	}
}

func TestExecuteVersion(t *testing.T) {
	mockOrch, teardown := setupTest()
	defer teardown()

	mockOrch.On("Orchestrate", mock.MatchedBy(func(f model.Flags) bool {
		return f.Version == true
	})).Return(nil)

	rootCmd.SetArgs([]string{"version"})

	err := Execute(dev, none, unknown)
	assert.NoError(t, err)
	mockOrch.AssertExpectations(t)
}

func TestExecuteUpdate(t *testing.T) {
	mockOrch, teardown := setupTest()
	defer teardown()

	mockOrch.On("Orchestrate", mock.MatchedBy(func(f model.Flags) bool {
		return f.Update == true
	})).Return(nil)

	rootCmd.SetArgs([]string{"update"})

	err := Execute(dev, none, unknown)
	assert.NoError(t, err)
	mockOrch.AssertExpectations(t)
}

func TestExecuteTrend(t *testing.T) {
	mockOrch, teardown := setupTest()
	defer teardown()

	mockOrch.On("Orchestrate", mock.MatchedBy(func(f model.Flags) bool {
		return f.Trend == true
	})).Return(nil)

	rootCmd.SetArgs([]string{"trend"})

	err := Execute(dev, none, unknown)
	assert.NoError(t, err)
	mockOrch.AssertExpectations(t)
}

func TestExecuteWaste(t *testing.T) {
	mockOrch, teardown := setupTest()
	defer teardown()

	mockOrch.On("Orchestrate", mock.MatchedBy(func(f model.Flags) bool {
		return f.Waste == true && len(f.WasteChecks) == 2 && f.WasteChecks[0] == ec2 && f.WasteChecks[1] == s3Const
	})).Return(nil)

	rootCmd.SetArgs([]string{"waste", ec2, s3Const})

	err := Execute(dev, none, unknown)
	assert.NoError(t, err)
	mockOrch.AssertExpectations(t)
}

func TestExecuteWasteComma(t *testing.T) {
	mockOrch, teardown := setupTest()
	defer teardown()

	mockOrch.On("Orchestrate", mock.MatchedBy(func(f model.Flags) bool {
		return f.Waste == true && len(f.WasteChecks) == 2 && f.WasteChecks[0] == ec2 && f.WasteChecks[1] == s3Const
	})).Return(nil)

	rootCmd.SetArgs([]string{"waste", ec2 + "," + s3Const})

	err := Execute(dev, none, unknown)
	assert.NoError(t, err)
	mockOrch.AssertExpectations(t)
}

func TestExecuteCost(t *testing.T) {
	mockOrch, teardown := setupTest()
	defer teardown()

	mockOrch.On("Orchestrate", mock.MatchedBy(func(f model.Flags) bool {
		return f.Waste == false && f.Trend == false
	})).Return(nil)

	rootCmd.SetArgs([]string{"cost"})

	err := Execute(dev, none, unknown)
	assert.NoError(t, err)
	mockOrch.AssertExpectations(t)
}

func TestPersistentFlags(t *testing.T) {
	mockOrch, teardown := setupTest()
	defer teardown()

	mockOrch.On("Orchestrate", mock.MatchedBy(func(f model.Flags) bool {
		return f.Region == "us-west-2" && f.Profile == "test-profile" && f.Output == "json"
	})).Return(nil)

	rootCmd.SetArgs([]string{"cost", "--region", "us-west-2", "--profile", "test-profile", "--output", "json"})

	err := Execute(dev, none, unknown)
	assert.NoError(t, err)
	mockOrch.AssertExpectations(t)
}

func TestExecuteTrendArgs(t *testing.T) {
	mockOrch, teardown := setupTest()
	defer teardown()

	mockOrch.On("Orchestrate", mock.MatchedBy(func(f model.Flags) bool {
		return f.Trend == true && len(f.TrendChecks) == 2 && f.TrendChecks[0] == ec2 && f.TrendChecks[1] == s3Const
	})).Return(nil)

	rootCmd.SetArgs([]string{"trend", ec2 + "," + s3Const})

	err := Execute(dev, none, unknown)
	assert.NoError(t, err)
	mockOrch.AssertExpectations(t)
}

func TestCommandFailures(t *testing.T) {
	originalBuilder := orchestratorBuilder
	orchestratorBuilder = func(needsAWS bool) (orchestrator.Service, error) {
		return nil, errors.New("builder error")
	}

	defer func() { orchestratorBuilder = originalBuilder }()

	commands := [][]string{
		{"cost"},
		{"trend"},
		{"waste"},
		{"update"},
		{"version"},
	}

	for _, cmdArgs := range commands {
		t.Run(cmdArgs[0], func(t *testing.T) {
			rootCmd.SetArgs(cmdArgs)

			err := Execute(dev, none, unknown)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "builder error")
		})
	}
}

func TestBuildOrchestratorNoAWS(t *testing.T) {
	orch, err := buildOrchestrator(false)
	assert.NoError(t, err)
	assert.NotNil(t, orch)
}

func TestBuildOrchestratorAWS(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	t.Setenv("AWS_REGION", "us-east-1")

	orch, err := buildOrchestrator(true)
	assert.NoError(t, err)
	assert.NotNil(t, orch)
}

func TestExecuteReportCost(t *testing.T) {
	mockOrch, teardown := setupTest()
	defer teardown()

	mockOrch.On("Orchestrate", mock.MatchedBy(func(f model.Flags) bool {
		return f.Report == true && f.Waste == false && f.Trend == false
	})).Return(nil)

	rootCmd.SetArgs([]string{"report", "cost"})

	err := Execute(dev, none, unknown)
	assert.NoError(t, err)
	mockOrch.AssertExpectations(t)
}

func TestExecuteReportWaste(t *testing.T) {
	mockOrch, teardown := setupTest()
	defer teardown()

	mockOrch.On("Orchestrate", mock.MatchedBy(func(f model.Flags) bool {
		return f.Report == true && f.Waste == true && f.Trend == false && len(f.WasteChecks) == 0
	})).Return(nil)

	rootCmd.SetArgs([]string{"report", "waste"})

	err := Execute(dev, none, unknown)
	assert.NoError(t, err)
	mockOrch.AssertExpectations(t)
}

func TestExecuteReportWasteSelective(t *testing.T) {
	mockOrch, teardown := setupTest()
	defer teardown()

	mockOrch.On("Orchestrate", mock.MatchedBy(func(f model.Flags) bool {
		return f.Report == true && f.Waste == true && f.Trend == false &&
			len(f.WasteChecks) == 2 && f.WasteChecks[0] == ec2 && f.WasteChecks[1] == s3Const
	})).Return(nil)

	rootCmd.SetArgs([]string{"report", "waste", ec2, s3Const})

	err := Execute(dev, none, unknown)
	assert.NoError(t, err)
	mockOrch.AssertExpectations(t)
}

func TestExecuteReportTrend(t *testing.T) {
	mockOrch, teardown := setupTest()
	defer teardown()

	mockOrch.On("Orchestrate", mock.MatchedBy(func(f model.Flags) bool {
		return f.Report == true && f.Trend == true && f.Waste == false
	})).Return(nil)

	rootCmd.SetArgs([]string{"report", "trend"})

	err := Execute(dev, none, unknown)
	assert.NoError(t, err)
	mockOrch.AssertExpectations(t)
}

func TestExecuteReportTrendSelective(t *testing.T) {
	mockOrch, teardown := setupTest()
	defer teardown()

	mockOrch.On("Orchestrate", mock.MatchedBy(func(f model.Flags) bool {
		return f.Report == true && f.Trend == true && f.Waste == false &&
			len(f.TrendChecks) == 1 && f.TrendChecks[0] == ec2
	})).Return(nil)

	rootCmd.SetArgs([]string{"report", "trend", ec2})

	err := Execute(dev, none, unknown)
	assert.NoError(t, err)
	mockOrch.AssertExpectations(t)
}

func TestExecuteReportCostCustomPath(t *testing.T) {
	mockOrch, teardown := setupTest()
	defer teardown()

	// When --path is used without a value, it defaults to "DEFAULT" per NoOptDefVal
	mockOrch.On("Orchestrate", mock.MatchedBy(func(f model.Flags) bool {
		return f.Report == true && f.Waste == false && f.Trend == false && f.ReportPath == "DEFAULT"
	})).Return(nil)

	rootCmd.SetArgs([]string{"report", "cost", "--path"})

	err := Execute(dev, none, unknown)
	assert.NoError(t, err)
	mockOrch.AssertExpectations(t)
}

func TestExecuteWasteLambdaMemoryThreshold(t *testing.T) {
	mockOrch, teardown := setupTest()
	defer teardown()

	mockOrch.On("Orchestrate", mock.MatchedBy(func(f model.Flags) bool {
		return f.Waste == true && f.LambdaMemoryThreshold == 20
	})).Return(nil)

	rootCmd.SetArgs([]string{"waste", "--lambda-memory-threshold", "20"})

	err := Execute(dev, none, unknown)
	assert.NoError(t, err)
	mockOrch.AssertExpectations(t)
}

func TestExecuteWasteLambdaMemoryThresholdDefault(t *testing.T) {
	mockOrch, teardown := setupTest()
	defer teardown()

	// Default value is 10
	mockOrch.On("Orchestrate", mock.MatchedBy(func(f model.Flags) bool {
		return f.Waste == true && f.LambdaMemoryThreshold == 10
	})).Return(nil)

	rootCmd.SetArgs([]string{"waste"})

	err := Execute(dev, none, unknown)
	assert.NoError(t, err)
	mockOrch.AssertExpectations(t)
}
