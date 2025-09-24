package common

import (
	"github.com/charmbracelet/lipgloss"
)

var subtle = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
var dialogBoxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("#874BFD")).
	Padding(1, 0).
	BorderTop(true).
	BorderLeft(true).
	BorderRight(true).
	BorderBottom(true)

func FullPageDialog(w, h int, padding int, content string) string {
	// dialog that fits the w, h provided w/ content centered..
	dialog := lipgloss.Place(w, h,
		lipgloss.Center, lipgloss.Center,
		dialogBoxStyle.
			Width(lipgloss.Width(content)).
			Height(lipgloss.Height(content)).
			Render(content),
		lipgloss.WithWhitespaceChars("猫咪"),
		lipgloss.WithWhitespaceForeground(subtle),
	)
	return dialog
}
