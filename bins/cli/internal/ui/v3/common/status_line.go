package common

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/nuonco/nuon/pkg/cli/styles"
)

var statusBarStyle = lipgloss.NewStyle().
	Padding(0)

type StatusBarRequest struct {
	Width   int
	Message string
	Level   string
}

func (r StatusBarRequest) getLevelColor() color.Color {
	style, ok := levelStyleMap[r.Level]
	if ok {
		return style
	}
	return styles.SubtleColor
}

func (r *StatusBarRequest) getLevelText() string {
	if r.Level == "default" {
		r.Level = "INFO"
	}
	return strings.ToUpper(r.Level)
}

func StatusBar(req StatusBarRequest) string {
	color := req.getLevelColor()
	status := lipgloss.NewStyle().Foreground(color).Padding(0, 1).Render(req.getLevelText())

	messageWidth := req.Width - lipgloss.Width(status)
	message := lipgloss.NewStyle().Foreground(styles.SubtleColor).Width(messageWidth).Render(req.Message)

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		status, message,
	)

}
