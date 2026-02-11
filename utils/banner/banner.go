package banner

import (
	"fmt"
	"os"
	"strings"

	"github.com/elC0mpa/aws-doctor/utils/ansi"
	"github.com/elC0mpa/aws-doctor/utils/console"
	"golang.org/x/term"
)

type bannerColor int

const (
	bannerCocaColaRed bannerColor = iota
	bannerFacebookBlue
	bannerTwitterBlue
	bannerLinkedInBlue
	bannerIBMBlue
	bannerYouTubeRed
	bannerSpotifyGreen
	bannerNetflixRed
	bannerTwitchPurple
	bannerYahooPurple
	bannerAmazonOrange
	bannerIntelBlue
	bannerWhatsAppGreen
	bannerAndroidGreen
	bannerSkypeBlue
	bannerStarbucksGreen
	bannerPinterestRed
	bannerAirbnbPink
	bannerFantaOrange
	bannerBMWBlue
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

var bannerTitleColorNames = []string{
	"CocaColaRed",
	"FacebookBlue",
	"TwitterBlue",
	"LinkedInBlue",
	"IBMBlue",
	"YouTubeRed",
	"SpotifyGreen",
	"NetflixRed",
	"TwitchPurple",
	"YahooPurple",
	"AmazonOrange",
	"IntelBlue",
	"WhatsAppGreen",
	"AndroidGreen",
	"SkypeBlue",
	"StarbucksGreen",
	"PinterestRed",
	"AirbnbPink",
	"FantaOrange",
	"BMWBlue",
}

const (
	bannerTitleColorDefault        = bannerSkypeBlue
	bannerTitleColorBlueBackground = bannerAmazonOrange
	bannerTitleColorEnv            = "AWS_DOCTOR_BANNER_COLOR"
)

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
			fmt.Fprint(os.Stderr, strings.Repeat(" ", pad))
		}

		fmt.Fprintln(os.Stderr, line)
	}
}

func bannerTitleColor() bannerColor {
	if color, ok := bannerTitleColorFromEnv(); ok {
		return color
	}

	if console.IsBlueBackground() {
		return bannerTitleColorBlueBackground
	}

	return bannerTitleColorDefault
}

func bannerTitleColorFromEnv() (bannerColor, bool) {
	raw := strings.TrimSpace(os.Getenv(bannerTitleColorEnv))

	if raw == "" {
		return 0, false
	}

	for idx, color := range bannerTitleColors {
		name := bannerTitleColorName(bannerColor(idx))
		if strings.EqualFold(raw, name) || raw == color {
			return bannerColor(idx), true
		}
	}

	return 0, false
}

func bannerTitleColorName(color bannerColor) string {
	if color < 0 || int(color) >= len(bannerTitleColorNames) {
		return ""
	}

	return bannerTitleColorNames[int(color)]
}

// DrawBannerTitle prints the application title banner to stdout.
func DrawBannerTitle() {
	ansi.EnableANSI()

	width := 80

	if w, _, err := term.GetSize(int(os.Stderr.Fd())); err == nil {
		width = w
	}

	fmt.Fprint(os.Stderr, bannerTitleColors[bannerTitleColor()])
	printCenteredLines(titleLines, width)
	fmt.Fprint(os.Stderr, "\x1b[0m")
}
