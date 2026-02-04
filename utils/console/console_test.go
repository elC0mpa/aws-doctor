package console

import (
	"os"
	"testing"
)

func TestIsBlueBackground(t *testing.T) {
	tests := []struct {
		envValue string
		want     bool
	}{
		{"15;4", true},
		{"15;12", true},
		{"15;0", false},
		{"", false},
		{"4", true},
		{";", false},
		{"  ", false},
	}

	for _, tt := range tests {
		if tt.envValue != "" {
			_ = os.Setenv("COLORFGBG", tt.envValue)
		} else {
			_ = os.Unsetenv("COLORFGBG")
		}

		if got := IsBlueBackground(); got != tt.want {
			t.Errorf("IsBlueBackground() for COLORFGBG=%q = %v, want %v", tt.envValue, got, tt.want)
		}
	}

	_ = os.Unsetenv("COLORFGBG")
}
