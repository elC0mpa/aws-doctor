package utils //nolint:revive

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

type BannerColor int

const (
	BannerCocaColaRed BannerColor = iota
	BannerFacebookBlue
	BannerTwitterBlue
	BannerLinkedInBlue
	BannerIBMBlue
	BannerYouTubeRed
	BannerSpotifyGreen
	BannerNetflixRed
	BannerTwitchPurple
	BannerYahooPurple
	BannerAmazonOrange
	BannerIntelBlue
	BannerWhatsAppGreen
	BannerAndroidGreen
	BannerSkypeBlue
	BannerStarbucksGreen
	BannerPinterestRed
	BannerAirbnbPink
	BannerFantaOrange
	BannerBMWBlue
)

var bannerTitleColors = []string{
	"\x1b[38;2;228;0;43m",   // Coca-Cola Red
	"\x1b[38;2;24;119;242m", // Facebook Blue
	"\x1b[38;2;29;161;242m", // Twitter/X Blue
	"\x1b[38;2;10;102;194m", // LinkedIn Blue
	"\x1b[38;2;15;98;254m",  // IBM Blue
	"\x1b[38;2;255;0;0m",    // YouTube Red
	"\x1b[38;2;30;215;96m",  // Spotify Green
	"\x1b[38;2;229;9;20m",   // Netflix Red
	"\x1b[38;2;145;70;255m", // Twitch Purple
	"\x1b[38;2;95;39;205m",  // Yahoo Purple
	"\x1b[38;2;255;153;0m",  // Amazon Orange
	"\x1b[38;2;0;113;197m",  // Intel Blue
	"\x1b[38;2;37;211;102m", // WhatsApp Green
	"\x1b[38;2;61;220;132m", // Android Green
	"\x1b[38;2;0;175;240m",  // Skype Blue
	"\x1b[38;2;0;112;74m",   // Starbucks Green
	"\x1b[38;2;189;8;28m",   // Pinterest Red
	"\x1b[38;2;255;90;95m",  // Airbnb Pink
	"\x1b[38;2;255;114;0m",  // Fanta Orange
	"\x1b[38;2;0;152;218m",  // BMW Blue
}

const bannerTitleColor = BannerSkypeBlue

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

	fmt.Print(bannerTitleColors[bannerTitleColor])
	printCenteredLines(titleLines, width)
	fmt.Print("\x1b[0m")
}
