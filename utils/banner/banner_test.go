package banner

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureOutput(f func()) string {
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

func TestBannerTitleColorName(t *testing.T) {
	tests := []struct {
		color bannerColor
		want  string
	}{
		{bannerCocaColaRed, "CocaColaRed"},
		{bannerBMWBlue, "BMWBlue"},
		{bannerColor(-1), ""},
		{bannerColor(100), ""},
	}

	for _, tt := range tests {
		if got := bannerTitleColorName(tt.color); got != tt.want {
			t.Errorf("bannerTitleColorName(%d) = %q, want %q", tt.color, got, tt.want)
		}
	}
}

func TestBannerTitleColorFromEnv(t *testing.T) {
	os.Setenv(bannerTitleColorEnv, "AmazonOrange")
	defer os.Unsetenv(bannerTitleColorEnv)

	color, ok := bannerTitleColorFromEnv()
	if !ok || color != bannerAmazonOrange {
		t.Errorf("bannerTitleColorFromEnv() = %v, %v, want %v, true", color, ok, bannerAmazonOrange)
	}

	os.Setenv(bannerTitleColorEnv, "InvalidColor")
	_, ok = bannerTitleColorFromEnv()
	if ok {
		t.Error("bannerTitleColorFromEnv() should return false for invalid color")
	}
}

func TestBannerTitleColor(t *testing.T) {
	// Test default
	os.Unsetenv(bannerTitleColorEnv)
	os.Unsetenv("COLORFGBG")
	if color := bannerTitleColor(); color != bannerTitleColorDefault {
		t.Errorf("bannerTitleColor() = %v, want %v", color, bannerTitleColorDefault)
	}

	// Test blue background
	os.Setenv("COLORFGBG", "15;4")
	if color := bannerTitleColor(); color != bannerTitleColorBlueBackground {
		t.Errorf("bannerTitleColor() with blue bg = %v, want %v", color, bannerTitleColorBlueBackground)
	}
	os.Unsetenv("COLORFGBG")

	// Test from env
	os.Setenv(bannerTitleColorEnv, "CocaColaRed")
	if color := bannerTitleColor(); color != bannerCocaColaRed {
		t.Errorf("bannerTitleColor() from env = %v, want %v", color, bannerCocaColaRed)
	}
	os.Unsetenv(bannerTitleColorEnv)
}

func TestPrintCenteredLines(t *testing.T) {
	lines := []string{"ABC", "DEFG"}
	output := captureOutput(func() {
		printCenteredLines(lines, 10)
	})

	if !strings.Contains(output, "ABC") || !strings.Contains(output, "DEFG") {
		t.Error("printCenteredLines() output missing lines")
	}
}

func TestDrawBannerTitle(t *testing.T) {
	output := captureOutput(func() {
		DrawBannerTitle()
	})

	if len(output) == 0 {
		t.Error("DrawBannerTitle() produced no output")
	}
}

func TestDrawBannerTitle_NonTerminal(t *testing.T) {
	// When stdout is a pipe, term.GetSize should fail
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	// We don't use captureOutput here because we want to specifically 
	// have os.Stdout be a pipe during the term.GetSize call
	DrawBannerTitle()

	os.Stdout = oldStdout
	w.Close()
	r.Close()
}
