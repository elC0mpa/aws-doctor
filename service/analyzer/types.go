package analyzer

import (
	"context"

	"github.com/elC0mpa/aws-doctor/model"
)

// WasteAnalyzer defines the standard contract for all AWS service waste detectors.
type WasteAnalyzer interface {
	Name() string
	Analyze(ctx context.Context, flags model.Flags) (model.ScopeResult, error)
}
