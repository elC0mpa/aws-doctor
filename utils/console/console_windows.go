//go:build windows

package console

import (
	"os"

	"golang.org/x/sys/windows"
)

// IsBlueBackground returns true if the terminal background color is blue.
func IsBlueBackground() bool {
	handle := windows.Handle(os.Stdout.Fd())

	var info windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(handle, &info); err != nil {
		return false
	}

	const backgroundBlue = 0x0010

	return info.Attributes&backgroundBlue != 0
}
