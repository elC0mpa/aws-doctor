package output

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	cetypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/elC0mpa/aws-doctor/model"
	"github.com/stretchr/testify/assert"
)

func TestRealRenderer_PrintMethods(t *testing.T) {
	r := &realRenderer{}

	t.Run("PrintAlreadyLatest", func(t *testing.T) {
		output := captureStdout(func() {
			r.PrintAlreadyLatest("v1.0.0")
		})
		assert.Contains(t, output, "v1.0.0 is already the latest version")
	})

	t.Run("PrintHomebrewUpdate", func(t *testing.T) {
		output := captureStdout(func() {
			r.PrintHomebrewUpdate()
		})
		assert.Contains(t, output, "brew upgrade aws-doctor")
	})

	t.Run("PrintGoInstallUpdate", func(t *testing.T) {
		output := captureStdout(func() {
			r.PrintGoInstallUpdate()
		})
		assert.Contains(t, output, "reinstall with the script")
	})

	t.Run("PrintRateLimitError", func(t *testing.T) {
		output := captureStdout(func() {
			r.PrintRateLimitError()
		})
		assert.Contains(t, output, "rate limits")
	})

	t.Run("PrintUpdateError", func(t *testing.T) {
		output := captureStdout(func() {
			r.PrintUpdateError(assert.AnError)
		})
		assert.Contains(t, output, "failed to check for updates")
	})

	t.Run("PrintWasteError", func(t *testing.T) {
		output := captureStdout(func() {
			r.PrintWasteError(assert.AnError)
		})
		assert.Contains(t, output, "failed to run interactive waste rendering")
	})

	t.Run("RenderVersion", func(t *testing.T) {
		v := model.VersionInfo{Version: "v1.2.3", Commit: "abc", Date: "today"}
		output := captureStdout(func() {
			r.RenderVersion(v)
		})
		assert.Contains(t, output, "aws-doctor version v1.2.3")
	})

	t.Run("PrintReportSuccess", func(t *testing.T) {
		output := captureStdout(func() {
			r.PrintReportSuccess("/path/to/report")
		})
		assert.Contains(t, output, "Report generated successfully")
		assert.Contains(t, output, "/path/to/report")
	})

	t.Run("PrintFirstDayOfMonthError", func(t *testing.T) {
		output := captureStdout(func() {
			r.PrintFirstDayOfMonthError()
		})
		assert.Contains(t, output, "first day of the month")
	})

	t.Run("PrintNewVersionAvailable", func(t *testing.T) {
		output := captureStderr(func() {
			r.PrintNewVersionAvailable("v1.0.0", "v1.1.0")
		})
		assert.Contains(t, output, "v1.0.0 → v1.1.0")
	})

	t.Run("SpinnerMethods", func(t *testing.T) {
		r.StopSpinner()
		r.SetSpinnerMessage("test")
	})

	t.Run("DrawMethods_SmokeTest", func(t *testing.T) {
		// Just ensure they don't panic with valid inputs
		costInfo := &model.CostInfo{
			DateInterval: cetypes.DateInterval{
				Start: aws.String("2024-01-01"),
				End:   aws.String("2024-01-31"),
			},
		}
		input := model.RenderCostComparisonInput{
			LastMonth:    costInfo,
			CurrentMonth: costInfo,
		}

		captureStdout(func() {
			r.DrawCostTable(input)
			r.DrawTrendChart("123", []model.CostInfo{*costInfo})
			r.DrawWasteTable(model.RenderWasteInput{}, nil)
		})
	})

	t.Run("OutputMethods_SmokeTest", func(t *testing.T) {
		costInfo := &model.CostInfo{
			DateInterval: cetypes.DateInterval{
				Start: aws.String("2024-01-01"),
				End:   aws.String("2024-01-31"),
			},
		}
		input := model.RenderCostComparisonInput{
			LastMonth:    costInfo,
			CurrentMonth: costInfo,
		}

		captureStdout(func() {
			_ = r.OutputCostComparisonJSON(input)
			_ = r.OutputCostComparisonCSV(input)
			_ = r.OutputTrendJSON("123", []model.CostInfo{*costInfo}, []string{})
			_ = r.OutputTrendCSV([]model.CostInfo{*costInfo}, []string{})
			_ = r.OutputWasteJSON(model.RenderWasteInput{}, nil)
			_ = r.OutputWasteCSV(model.RenderWasteInput{}, nil)
		})
	})
}

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	_ = w.Close()

	os.Stdout = old

	var buf bytes.Buffer

	_, _ = io.Copy(&buf, r)

	return buf.String()
}

func captureStderr(f func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	f()

	_ = w.Close()

	os.Stderr = old

	var buf bytes.Buffer

	_, _ = io.Copy(&buf, r)

	return buf.String()
}
