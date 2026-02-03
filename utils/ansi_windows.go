//go:build windows

package utils //nolint:revive

import (
	"os"

	"golang.org/x/sys/windows"
)

// EnableANSI enables ANSI escape sequence processing on Windows consoles.
func EnableANSI() {
	handle := windows.Handle(os.Stdout.Fd())

	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return
	}

	const enableVirtualTerminalProcessing = 0x0004

	_ = windows.SetConsoleMode(handle, mode|enableVirtualTerminalProcessing)
}
