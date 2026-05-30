package services

import (
	"context"

	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/mock"
)

// MockAnalyzer is a mock implementation of analyzer.WasteAnalyzer.
type MockAnalyzer struct {
	mock.Mock
}

// Name returns the name of the mock analyzer.
func (m *MockAnalyzer) Name() string {
	args := m.Called()
	return args.String(0)
}

// TabName returns the tab name of the mock analyzer.
func (m *MockAnalyzer) TabName() string {
	args := m.Called()
	return args.String(0)
}

// Analyze performs the mock analysis.
func (m *MockAnalyzer) Analyze(ctx context.Context, flags model.Flags) (model.ScopeResult, error) {
	args := m.Called(ctx, flags)
	return args.Get(0).(model.ScopeResult), args.Error(1)
}
