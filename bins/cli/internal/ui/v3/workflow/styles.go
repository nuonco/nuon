package workflow

import "github.com/charmbracelet/lipgloss"

var appStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("#6133FF"))

var appStyleBlur = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.AdaptiveColor{Dark: "237", Light: "237"})

var appStyleFocus = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("#6133FF"))
