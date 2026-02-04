//go:build windows

package utils //nolint:revive

import (
	"os"

	"golang.org/x/sys/windows"
)

func isBlueBackground() bool {
	handle := windows.Handle(os.Stdout.Fd())

	var info windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(handle, &info); err != nil {
		return false
	}

	const backgroundBlue = 0x0010

	return info.Attributes&backgroundBlue != 0
}
