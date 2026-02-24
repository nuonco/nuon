package common

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/nuonco/nuon/pkg/cli/styles"
)

var dialogBoxStyle = lipgloss.NewStyle().
	Padding(1, 0).
	Border(lipgloss.RoundedBorder()).
	BorderTop(true).
	BorderLeft(true).
	BorderRight(true).
	BorderForeground(styles.PrimaryColor).
	BorderBottom(true)

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
			Width(lipgloss.Width(req.Content)).
			Height(lipgloss.Height(req.Content)).
			Render(req.Content),
		lipgloss.WithWhitespaceChars("⦾∙"),
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Foreground(styles.Ghost)),
	)
	return dialog
}
