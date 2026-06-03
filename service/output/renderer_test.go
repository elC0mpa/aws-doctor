package output

import (
	"testing"

	"github.com/elC0mpa/aws-doctor/model"
)

func newString(s string) *string { return &s }

func TestJSONRenderer(t *testing.T) {
	r := &jsonRenderer{}

	info := &model.CostInfo{}
	info.Start = newString("start")
	info.End = newString("end")

	err := r.RenderCostComparison(model.RenderCostComparisonInput{
		CurrentMonth: info,
		LastMonth:    info,
	})
	if err != nil {
		t.Errorf("RenderCostComparison failed: %v", err)
	}

	err = r.RenderTrend("account", []model.CostInfo{}, []string{})
	if err != nil {
		t.Errorf("RenderTrend failed: %v", err)
	}

	err = r.RenderWaste(model.RenderWasteInput{}, nil)
	if err != nil {
		t.Errorf("RenderWaste failed: %v", err)
	}

	err = r.RenderWasteInteractive("account", nil, nil, nil)
	if err == nil {
		t.Error("RenderWasteInteractive should return an error for JSON")
	}

	if r.IsInteractive() {
		t.Error("JSON renderer should not be interactive")
	}
}

func TestCSVRenderer(t *testing.T) {
	r := &csvRenderer{}

	info := &model.CostInfo{}
	info.Start = newString("start")
	info.End = newString("end")

	err := r.RenderCostComparison(model.RenderCostComparisonInput{
		CurrentMonth: info,
		LastMonth:    info,
	})
	if err != nil {
		t.Errorf("RenderCostComparison failed: %v", err)
	}

	err = r.RenderTrend("account", []model.CostInfo{}, []string{})
	if err != nil {
		t.Errorf("RenderTrend failed: %v", err)
	}

	err = r.RenderWaste(model.RenderWasteInput{}, nil)
	if err != nil {
		t.Errorf("RenderWaste failed: %v", err)
	}

	err = r.RenderWasteInteractive("account", nil, nil, nil)
	if err == nil {
		t.Error("RenderWasteInteractive should return an error for CSV")
	}

	if r.IsInteractive() {
		t.Error("CSV renderer should not be interactive")
	}
}

func TestTableRenderer(t *testing.T) {
	// For table renderer, Draw functions write directly to stdout. We can capture it.
	r := &tableRenderer{}

	info := &model.CostInfo{}
	info.Start = newString("start")
	info.End = newString("end")

	captureOutput(func() {
		err := r.RenderCostComparison(model.RenderCostComparisonInput{
			CurrentMonth: info,
			LastMonth:    info,
		})
		if err != nil {
			t.Errorf("RenderCostComparison failed: %v", err)
		}
	})

	captureOutput(func() {
		err := r.RenderTrend("account", []model.CostInfo{}, []string{})
		if err != nil {
			t.Errorf("RenderTrend failed: %v", err)
		}
	})

	captureOutput(func() {
		err := r.RenderWaste(model.RenderWasteInput{}, nil)
		if err != nil {
			t.Errorf("RenderWaste failed: %v", err)
		}
	})

	// We intentionally do not test RenderWasteInteractive here to avoid
	// initializing the bubble tea UI during unit tests, which can block.

	// IsInteractive depends on os.Stdout being a terminal. We can't guarantee
	// true or false reliably across all CI environments, but we can verify it doesn't panic.
	_ = r.IsInteractive()
}
