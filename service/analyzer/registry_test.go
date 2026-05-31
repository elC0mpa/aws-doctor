package analyzer

import (
	"context"
	"testing"

	"github.com/elC0mpa/aws-doctor/model"
)

type mockAnalyzer struct {
	name string
}

func (m *mockAnalyzer) Analyze(ctx context.Context, flags model.Flags) (model.ScopeResult, error) {
	return model.ScopeResult{}, nil
}

func (m *mockAnalyzer) Name() string {
	return m.name
}

func (m *mockAnalyzer) TabName() string {
	return m.name
}

func TestRegistry(t *testing.T) {
	r := NewRegistry()

	analyzers := r.GetAnalyzers()
	if len(analyzers) != 0 {
		t.Errorf("Expected 0 analyzers, got %d", len(analyzers))
	}

	mock1 := &mockAnalyzer{name: "analyzer1"}
	mock2 := &mockAnalyzer{name: "analyzer2"}

	r.Register(mock1)
	r.Register(mock2)

	analyzers = r.GetAnalyzers()
	if len(analyzers) != 2 {
		t.Fatalf("Expected 2 analyzers, got %d", len(analyzers))
	}

	if analyzers[0].Name() != "analyzer1" {
		t.Errorf("Expected first analyzer to be analyzer1, got %s", analyzers[0].Name())
	}

	if analyzers[1].Name() != "analyzer2" {
		t.Errorf("Expected second analyzer to be analyzer2, got %s", analyzers[1].Name())
	}
}
