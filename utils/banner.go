package utils //nolint:revive

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

const bannerTitleColor = "\x1b[38;2;80;160;255m"

var titleLines = []string{
"  █████╗  ██╗    ██╗ ███████╗        ██████╗   ██████╗   ██████╗ ████████╗  ██████╗  ██████╗ ",
" ██╔══██╗ ██║    ██║ ██╔════╝        ██╔══██╗ ██╔═══██╗ ██╔════╝ ╚══██╔══╝ ██╔═══██╗ ██╔══██╗",
" ███████║ ██║ █╗ ██║ ███████╗ █████╗ ██║  ██║ ██║   ██║ ██║         ██║    ██║   ██║ ██████╔╝",
" ██╔══██║ ██║███╗██║ ╚════██║ ╚════╝ ██║  ██║ ██║   ██║ ██║         ██║    ██║   ██║ ██╔══██╗",
" ██║  ██║ ╚███╔███╔╝ ███████║        ██████╔╝ ╚██████╔╝ ╚██████╗    ██║    ╚██████╔╝ ██║  ██║",
" ╚═╝  ╚═╝  ╚══╝╚══╝  ╚══════╝        ╚═════╝   ╚═════╝   ╚═════╝    ╚═╝     ╚═════╝  ╚═╝  ╚═╝",
}


func printCenteredLines(lines []string, width int) {
	for _, line := range lines {
		pad := 0
		if width > len(line) {
			pad = (width - len(line)) / 2
		}
		if pad > 0 {
			fmt.Print(strings.Repeat(" ", pad))
		}
		fmt.Println(line)
	}
}

func DrawBannerTitle() {
	EnableANSI()
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
	}

	fmt.Print(bannerTitleColor)
	printCenteredLines(titleLines, width)
	fmt.Print("\x1b[0m")
}
