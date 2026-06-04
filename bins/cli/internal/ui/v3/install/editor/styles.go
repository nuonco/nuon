package editor

import (
	"charm.land/lipgloss/v2"
	"github.com/nuonco/nuon/pkg/cli/styles"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(styles.PrimaryColor).
			Bold(true).
			Padding(1, 0)

	labelStyle = lipgloss.NewStyle().
			Foreground(styles.TextColor).
			Bold(true)

	descStyle = styles.TextDim.Italic(true)

	focusedInputStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(styles.BorderActiveColor).
				Padding(0, 1)

	blurredInputStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(styles.BorderInactiveColor).
				Padding(0, 1)

	groupTitleStyle = lipgloss.NewStyle().
			Foreground(styles.SecondaryColor).
			Bold(true)
)

func groupHeaderStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderLeft(false).
		BorderRight(false).
		BorderTop(false).
		BorderForeground(styles.SubtleColor).
		Width(width).
		Padding(0, 1)
}

func groupInputsStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width-4).
		Padding(0, 1)
}
