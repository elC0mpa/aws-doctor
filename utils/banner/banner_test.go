package banner

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureOutput(f func()) string {
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
	_ = os.Setenv(bannerTitleColorEnv, "AmazonOrange")

	defer func() { _ = os.Unsetenv(bannerTitleColorEnv) }()

	color, ok := bannerTitleColorFromEnv()
	if !ok || color != bannerAmazonOrange {
		t.Errorf("bannerTitleColorFromEnv() = %v, %v, want %v, true", color, ok, bannerAmazonOrange)
	}

	_ = os.Setenv(bannerTitleColorEnv, "InvalidColor")

	_, ok = bannerTitleColorFromEnv()
	if ok {
		t.Error("bannerTitleColorFromEnv() should return false for invalid color")
	}
}

func TestBannerTitleColor(t *testing.T) {
	// Test default
	_ = os.Unsetenv(bannerTitleColorEnv)
	_ = os.Unsetenv("COLORFGBG")

	if color := bannerTitleColor(); color != bannerTitleColorDefault {
		t.Errorf("bannerTitleColor() = %v, want %v", color, bannerTitleColorDefault)
	}

	// Test blue background
	_ = os.Setenv("COLORFGBG", "15;4")

	if color := bannerTitleColor(); color != bannerTitleColorBlueBackground {
		t.Errorf("bannerTitleColor() with blue bg = %v, want %v", color, bannerTitleColorBlueBackground)
	}

	_ = os.Unsetenv("COLORFGBG")

	// Test from env
	_ = os.Setenv(bannerTitleColorEnv, "CocaColaRed")

	if color := bannerTitleColor(); color != bannerCocaColaRed {
		t.Errorf("bannerTitleColor() from env = %v, want %v", color, bannerCocaColaRed)
	}

	_ = os.Unsetenv(bannerTitleColorEnv)
}

func TestPrintCenteredLines(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		width int
	}{
		{
			name:  "ASCII lines",
			lines: []string{"ABC", "DEFG"},
			width: 10,
		},
		{
			name:  "Unicode lines",
			lines: []string{"┌─┐", "│A│", "└─┘"},
			width: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				printCenteredLines(tt.lines, tt.width)
			})

			for _, line := range tt.lines {
				if !strings.Contains(output, line) {
					t.Errorf("printCenteredLines() output missing line: %q", line)
				}
			}
		})
	}
}

func TestDrawBannerTitle_NonTerminal(t *testing.T) {
	// Mock non-terminal
	orig := isTerminal
	isTerminal = func(fd int) bool { return false }

	defer func() { isTerminal = orig }()

	// When stderr is a pipe (non-TTY), nothing should be printed.
	output := captureOutput(func() {
		DrawBannerTitle()
	})

	const want = ""
	if output != want {
		t.Errorf("DrawBannerTitle() non-TTY output = %q, want %q", output, want)
	}
}

func TestDrawBannerTitle(t *testing.T) {
	// Mock terminal
	orig := isTerminal
	isTerminal = func(fd int) bool { return true }

	defer func() { isTerminal = orig }()

	output := captureOutput(func() {
		DrawBannerTitle()
	})

	if len(output) == 0 {
		t.Error("DrawBannerTitle() produced no output in mock TTY mode")
	}
}
