package utils //nolint:revive

import (
	_ "embed"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/term"
)

//go:embed banner_raw.txt
var bannerRawFile string

var (
	bannerByWidth map[int]string
	bannerWidths []int
)

const bannerSize = 200 // 32, 64, 96, 128, 160, 200

func init() {
	bannerByWidth = parseBannerFile(bannerRawFile)
	for width := range bannerByWidth {
		bannerWidths = append(bannerWidths, width)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(bannerWidths)))
}

// bannerRaw* contain ANSI escape sequences as literal \x1b strings.
// They are expanded to real escape codes at init time.
func expandBanner(raw string) string {
	return strings.ReplaceAll(raw, "\\x1b", "\x1b")
}

func parseBannerFile(data string) map[int]string {
	byWidth := make(map[int]string)
	lines := strings.Split(strings.ReplaceAll(data, "\r\n", "\n"), "\n")
	currentWidth := 0
	var buf []string

	flush := func() {
		if currentWidth == 0 {
			return
		}
		joined := strings.Join(buf, "\n")
		if joined != "" {
			joined += "\n"
		}
		byWidth[currentWidth] = expandBanner(joined)
		buf = nil
		currentWidth = 0
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "WIDTH ") {
			flush()
			widthStr := strings.TrimSpace(strings.TrimPrefix(line, "WIDTH "))
			if widthStr == "" {
				continue
			}
			width, err := strconv.Atoi(widthStr)
			if err != nil {
				continue
			}
			currentWidth = width
			continue
		}
		if strings.TrimSpace(line) == "END" {
			flush()
			continue
		}
		if currentWidth != 0 {
			buf = append(buf, line)
		}
	}
	flush()

	return byWidth
}

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

func drawBannerImage() {
	fmt.Print(bannerByWidth[bannerSize])
	fmt.Println()
}

func drawBannerTitle(width int) {
	fmt.Print("\x1b[1;37m")
	printCenteredLines(titleLines, width)
	fmt.Print("\x1b[0m")
}

// DrawBanner prints the application banner to stdout.
func DrawBanner() {
	EnableANSI()
	width := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
	}

	// drawBannerImage()
	drawBannerTitle(width)
}
