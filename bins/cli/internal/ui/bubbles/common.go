package bubbles

import (
	"github.com/charmbracelet/lipgloss"
)

// Common styles and colors for consistent theming
// Using ANSI colors to respect terminal color schemes
var (
	// Primary brand colors - use terminal's bright colors
	PrimaryColor   = lipgloss.Color("12")  // Bright Blue
	SecondaryColor = lipgloss.Color("14")  // Bright Cyan
	AccentColor    = lipgloss.Color("11")  // Bright Yellow
	
	// Status colors - use standard ANSI colors
	SuccessColor = lipgloss.Color("2")   // Green
	ErrorColor   = lipgloss.Color("1")   // Red
	WarningColor = lipgloss.Color("3")   // Yellow
	InfoColor    = lipgloss.Color("4")   // Blue
	
	// Neutral colors - use terminal's default colors
	TextColor     = lipgloss.Color("")    // Default foreground
	SubtleColor   = lipgloss.Color("8")   // Bright Black (typically gray)
	BorderColor   = lipgloss.Color("8")   // Bright Black (typically gray)
	
	// Base styles
	BaseStyle = lipgloss.NewStyle().
			Padding(0, 1)
	
	// Status message styles
	InfoStyle = lipgloss.NewStyle().
			Foreground(InfoColor).
			Bold(true).
			Padding(0, 1)
	
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true).
			Padding(0, 1)
	
	WarningStyle = lipgloss.NewStyle().
			Foreground(WarningColor).
			Bold(true).
			Padding(0, 1)
	
	SuccessStyle = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Bold(true).
			Padding(0, 1)
	
	// Interactive styles
	FocusedStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)
	
	BlurredStyle = lipgloss.NewStyle().
			Foreground(SubtleColor)
	
	// Border styles
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BorderColor).
			Padding(1, 2)
	
	FocusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(PrimaryColor).
				Padding(1, 2)
)

// Evaluation journey specific styling
var (
	EvaluationHeaderStyle = lipgloss.NewStyle().
				Foreground(AccentColor).
				Bold(true).
				Underline(true).
				Margin(1, 0)
	
	EvaluationTipStyle = lipgloss.NewStyle().
				Foreground(InfoColor).
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