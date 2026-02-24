package styles

import (
	"charm.land/lipgloss/v2"
)

// NOTE: use the colors from colors.go
var Link = lipgloss.NewStyle().Foreground(lipgloss.Color("20")).Underline(true)

var TextGhost = lipgloss.NewStyle().Italic(true).Foreground(Ghost)
var TextBold = lipgloss.NewStyle().Bold(true)
var TextDim = lipgloss.NewStyle().Foreground(Dim)
