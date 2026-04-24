package output

import (
	"bytes"
	"io"
	"os"
	"testing"

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
