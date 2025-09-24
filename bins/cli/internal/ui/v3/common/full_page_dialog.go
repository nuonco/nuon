package common

import (
	"github.com/charmbracelet/lipgloss"
)

var subtle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
var warning = lipgloss.AdaptiveColor{Light: "209", Dark: "130"}
var error = lipgloss.AdaptiveColor{Light: "160", Dark: "52"}
var info = lipgloss.AdaptiveColor{Light: "27", Dark: "17"}
var dialogBoxStyle = lipgloss.NewStyle().
	Padding(1, 0).
	Border(lipgloss.RoundedBorder()).
	BorderTop(true).
	BorderLeft(true).
	BorderRight(true).
	BorderForeground(lipgloss.Color("#874BFD")).
	BorderBottom(true)

var levelStyleMap = map[string]lipgloss.AdaptiveColor{
	"default": subtle,
	"warning": warning,
	"error":   error,
	"info":    info,
}

type FullPageDialogRequest struct {
	Width   int
	Height  int
	Padding int
	Content string
	Level   string
}

func (r FullPageDialogRequest) getLevelStyle() lipgloss.AdaptiveColor {
	style, ok := levelStyleMap[r.Level]
	if ok {
		return style
	}
	return subtle
}

func FullPageDialog(req FullPageDialogRequest) string {
	// dialog that fits the w, h provided w/ content centered..
	levelStyle := req.getLevelStyle()
	dialog := lipgloss.Place(req.Width, req.Height,
		lipgloss.Center, lipgloss.Center,
		dialogBoxStyle.
			Width(lipgloss.Width(req.Content)).
			Height(lipgloss.Height(req.Content)).
			Render(req.Content),
		lipgloss.WithWhitespaceChars("猫咪"),
		lipgloss.WithWhitespaceForeground(levelStyle),
	)
	return dialog
}
