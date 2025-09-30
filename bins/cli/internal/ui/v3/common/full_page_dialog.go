package common

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/styles"
)

var dialogBoxStyle = lipgloss.NewStyle().
	Padding(1, 0).
	Border(lipgloss.RoundedBorder()).
	BorderTop(true).
	BorderLeft(true).
	BorderRight(true).
	BorderForeground(styles.PrimaryColor).
	BorderBottom(true)

var levelStyleMap = map[string]lipgloss.CompleteAdaptiveColor{
	"default": styles.SubtleColor,
	"warning": styles.WarningColor,
	"error":   styles.ErrorColor,
	"info":    styles.InfoColor,
}

type FullPageDialogRequest struct {
	Width   int
	Height  int
	Padding int
	Content string
	Level   string
}

func (r FullPageDialogRequest) getLevelStyle() lipgloss.CompleteAdaptiveColor {
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
			Width(lipgloss.Width(req.Content)).
			Height(lipgloss.Height(req.Content)).
			Render(req.Content),
		lipgloss.WithWhitespaceChars("猫咪"),
		lipgloss.WithWhitespaceForeground(levelStyle),
	)
	return dialog
}
