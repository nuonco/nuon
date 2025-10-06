package bubbles

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/powertoolsdev/mono/pkg/cli/styles"
)

// Common styles and colors for consistent theming
// Using ANSI colors to respect terminal color schemes
var (
	// Neutral colors - use terminal's default colors
	BorderColor = styles.BorderInactiveColor

	// Base styles
	BaseStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Status message styles
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

	// Interactive styles
	FocusedStyle = lipgloss.NewStyle().
			Foreground(styles.PrimaryColor).
			Bold(true)

	BlurredStyle = lipgloss.NewStyle().
			Foreground(styles.SubtleColor)

	// Border styles: padded but border-less
	BorderStyle = lipgloss.NewStyle().
			Padding(1)

	FocusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(styles.PrimaryColor).
				Padding(1, 2)
)

// Evaluation journey specific styling
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

// Helper functions
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
