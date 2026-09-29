package bubbles

import (
	"charm.land/lipgloss/v2"

	"github.com/nuonco/nuon/pkg/cli/styles"
)

var (
	BorderColor = styles.BorderInactiveColor

	BaseStyle = lipgloss.NewStyle().
			Padding(0, 1)

	InfoStyle = lipgloss.NewStyle().
			Foreground(styles.InfoColor).
			Bold(true).
			Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(styles.ErrorColor).
			Bold(true).
			Padding(0, 1)

	WarningStyle = lipgloss.NewStyle().
			Foreground(styles.WarningColor).
			Bold(true).
			Padding(0, 1)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(styles.SuccessColor).
			Bold(true).
			Padding(0, 1)

	FocusedStyle = lipgloss.NewStyle().
			Foreground(styles.PrimaryColor).
			Bold(true)

	BlurredStyle = lipgloss.NewStyle().
			Foreground(styles.SubtleColor)

	BorderStyle = lipgloss.NewStyle().
			Padding(1)

	FocusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(styles.PrimaryColor).
				Padding(1, 2)
)

var (
	EvaluationHeaderStyle = lipgloss.NewStyle().
				Foreground(styles.AccentColor).
				Bold(true).
				Underline(true).
				Margin(1, 0)

	EvaluationTipStyle = lipgloss.NewStyle().
				Foreground(styles.InfoColor).
				Italic(true).
				Padding(0, 1)
)
