//go:build windows

package console

import (
	"os"
	"strings"

	"golang.org/x/sys/windows"
)

// IsBlueBackground returns true if the terminal background color is blue.
func IsBlueBackground() bool {
	if raw := os.Getenv("COLORFGBG"); raw != "" {
		parts := strings.Split(raw, ";")
		if len(parts) > 0 {
			bg := strings.TrimSpace(parts[len(parts)-1])
			if bg == "4" || bg == "12" {
				return true
			}
		}
	}

	handle := windows.Handle(os.Stdout.Fd())

	var info windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(handle, &info); err != nil {
		return false
	}

	const backgroundBlue = 0x0010

	return info.Attributes&backgroundBlue != 0
}
