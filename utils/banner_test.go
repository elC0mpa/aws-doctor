package utils //nolint:revive

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestDrawBanner(t *testing.T) {
	// Capture stdout to verify banner is drawn without panic
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Should not panic
	DrawBanner()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Verify some output was produced (ASCII art banner)
	if len(output) == 0 {
		t.Error("DrawBanner() produced no output")
	}

	// The banner should contain "AWS" somewhere (it's ASCII art of "AWS Doctor")
	// Note: ASCII art fonts may vary, so we just check for non-empty output
}

func TestDrawBanner_MultipleCallsNoPanic(_ *testing.T) {
	// Redirect stdout to discard
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	// Multiple calls should not panic
	for i := 0; i < 3; i++ {
		DrawBanner()
	}
}

func BenchmarkDrawBanner(b *testing.B) {
	// Redirect stdout to discard
	old := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = old }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DrawBanner()
	}
}
