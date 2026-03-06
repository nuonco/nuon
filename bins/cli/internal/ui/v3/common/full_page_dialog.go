package common

import (
	"image/color"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/nuonco/nuon/pkg/cli/styles"
)

var hasDarkBG = lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
var lightDark = lipgloss.LightDark(hasDarkBG)

var dialogBoxStyle = lipgloss.NewStyle().
	Padding(1, 3).
	Align(lipgloss.Center).
	Border(lipgloss.RoundedBorder()).
	BorderTop(true).
	BorderLeft(true).
	BorderRight(true).
	BorderForeground(styles.PrimaryColor).
	BorderBottom(true)

var bgCharColor = lightDark(lipgloss.Color("#E8E8E8"), lipgloss.Color("#1A1A1A"))

var levelStyleMap = map[string]color.Color{
	"default": styles.SubtleColor,
	"warning": styles.WarningColor,
	"error":   styles.ErrorColor,
	"info":    styles.InfoColor,
	"success": styles.SuccessColor,
}

type FullPageDialogRequest struct {
	Width   int
	Height  int
	Padding int
	Content string
	Level   string
}

func (r FullPageDialogRequest) getLevelStyle() color.Color {
	style, ok := levelStyleMap[r.Level]
	if ok {
		return style
	}
	return styles.SubtleColor
}

func FullPageDialog(req FullPageDialogRequest) string {
	// dialog that fits the w, h provided w/ content centered..
	levelStyle := req.getLevelStyle()
	dialog := lipgloss.Place(req.Width, req.Height,
		lipgloss.Center, lipgloss.Center,
		dialogBoxStyle.
			BorderForeground(levelStyle).
			Render(req.Content),
		lipgloss.WithWhitespaceChars("⦾∙"),
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Foreground(bgCharColor)),
	)
	return dialog
}
