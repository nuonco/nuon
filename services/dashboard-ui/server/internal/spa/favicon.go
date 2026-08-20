package spa

import (
	"fmt"
	"html"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

const (
	byocDefaultBackground = "#662F9D"
	byocMarkDimOpacity    = "0.4"

	nuonMark = `<g transform="translate(100.585 100.945) scale(1.42) translate(-100.585 -100.945)"><path d="M121.15 40.3118L97.9645 53.715V75.4151L79.1959 64.5597H79.1852L56.8232 77.492V148.651L79.1852 161.584H79.1959L103.205 147.699V126.951L121.161 137.325L144.346 123.922V53.715L121.161 40.3118H121.15ZM62.0528 80.5216L79.1745 70.6297H79.1852L97.9538 81.4744V117.862L62.0528 97.1151V80.5216ZM97.9538 144.669L79.1745 155.514L62.0528 145.622V103.174L97.9538 123.922V144.669ZM139.095 120.881L121.15 131.255L103.205 120.892V84.504L139.106 105.251V120.881H139.095ZM139.095 99.192L103.194 78.4447V56.7447L121.15 46.3711L139.095 56.7447V99.192Z" fill="%s"/></g>`
)

var hexColorPattern = regexp.MustCompile(`^#?(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

func generateBYOCFavicon(cfg *internal.Config, l *zap.Logger) []byte {
	if !cfg.IsBYOC {
		return nil
	}

	iconText := strings.TrimSpace(cfg.BYOCIconText)
	color := strings.TrimSpace(cfg.BYOCColor)
	if iconText == "" && color == "" {
		return nil
	}

	background, hasColor := resolveBYOCBackground(color)
	if color != "" && !hasColor {
		l.Warn("invalid NUON_BYOC_COLOR, falling back to default purple",
			zap.String("color", cfg.BYOCColor))
	}

	glyphs := ""
	if iconText != "" {
		runes := []rune(iconText)
		if len(runes) > 2 {
			runes = runes[:2]
		}
		if isValidIconText(string(runes)) {
			glyphs = string(runes)
		} else {
			l.Warn("invalid NUON_BYOC_ICON_TEXT, must be printable characters",
				zap.String("icon_text", cfg.BYOCIconText))
		}
	}

	if glyphs == "" && !hasColor {
		return nil
	}

	foreground := contrastingForeground(background)
	mark := fmt.Sprintf(nuonMark, foreground)

	content := mark
	if glyphs != "" {
		fit := ""
		if len([]rune(glyphs)) == 2 {
			fit = ` textLength="170" lengthAdjust="spacingAndGlyphs"`
		}
		content = fmt.Sprintf(
			`<g opacity="%s">%s</g><text x="100.5" y="100.5" fill="%s" font-size="%d" font-family="system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif" font-weight="650" text-anchor="middle" dominant-baseline="central"%s>%s</text>`,
			byocMarkDimOpacity, mark, foreground, glyphFontSize(glyphs), fit, html.EscapeString(glyphs))
	}

	svg := fmt.Sprintf(
		`<svg width="201" height="201" viewBox="0 0 201 201" xmlns="http://www.w3.org/2000/svg"><rect x="0.5" y="0.5" width="200" height="200" rx="44" fill="%s"/>%s</svg>`,
		background, content)

	return []byte(svg)
}

func resolveBYOCBackground(color string) (string, bool) {
	c := strings.TrimSpace(color)
	if c == "" || !hexColorPattern.MatchString(c) {
		return byocDefaultBackground, false
	}
	return "#" + strings.TrimPrefix(c, "#"), true
}

func isValidIconText(s string) bool {
	runes := []rune(s)
	if len(runes) < 1 || len(runes) > 2 {
		return false
	}
	for _, r := range runes {
		if !unicode.IsPrint(r) || unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func glyphFontSize(glyphs string) int {
	if len([]rune(glyphs)) == 1 {
		return 150
	}
	return 130
}

func contrastingForeground(hexColor string) string {
	lum := relativeLuminance(hexColor)
	contrastWithWhite := 1.05 / (lum + 0.05)
	contrastWithBlack := (lum + 0.05) / 0.05
	if contrastWithBlack > contrastWithWhite {
		return "#000000"
	}
	return "#FFFFFF"
}

func relativeLuminance(hexColor string) float64 {
	hexDigits := strings.TrimPrefix(hexColor, "#")
	if len(hexDigits) == 3 {
		hexDigits = strings.Repeat(string(hexDigits[0]), 2) +
			strings.Repeat(string(hexDigits[1]), 2) +
			strings.Repeat(string(hexDigits[2]), 2)
	}

	channel := func(offset int) float64 {
		v, _ := strconv.ParseUint(hexDigits[offset:offset+2], 16, 8)
		c := float64(v) / 255.0
		if c <= 0.03928 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}

	return 0.2126*channel(0) + 0.7152*channel(2) + 0.0722*channel(4)
}
