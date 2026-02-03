//go:build !windows

package utils //nolint:revive

import (
	"os"
	"strings"
)

func isBlueBackground() bool {
	raw := os.Getenv("COLORFGBG")

	if raw == "" {
		return false
	}

	parts := strings.Split(raw, ";")

	if len(parts) == 0 {
		return false
	}

	bg := strings.TrimSpace(parts[len(parts)-1])

	if bg == "" {
		return false
	}

	// ANSI 16-color backgrounds: 4 (blue) and 12 (bright blue).
	return bg == "4" || bg == "12"
}
