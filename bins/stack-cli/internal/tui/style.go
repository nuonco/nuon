package tui

import "charm.land/lipgloss/v2"

const (
	totalSteps   = 6
	maxBodyWidth = 100
	minWidth     = 70
	minHeight    = 18
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Padding(0, 1)
	headerMetaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Padding(0, 1)
	ruleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	stepActive  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	stepDone    = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	stepPending = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	stepSep     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	footerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Padding(0, 1)
	footerKey   = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

	bodyPadStyle = lipgloss.NewStyle().Padding(1, 2)
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	kvKeyStyle   = lipgloss.NewStyle().Bold(true)

	focusedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	checkboxOn    = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	checkboxOff   = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	requiredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	buttonOn      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("16")).Background(lipgloss.Color("39")).Padding(0, 2)
	buttonOff     = lipgloss.NewStyle().Foreground(lipgloss.Color("250")).Padding(0, 2)
	errStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))

	outerStyle = lipgloss.NewStyle().
			Margin(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39")).
			Padding(0, 1)

	jsonKey     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	jsonString  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	jsonNumber  = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	jsonLiteral = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	jsonPunct   = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)

func renderButton(label string, focused, disabled bool) string {
	switch {
	case disabled:
		return dimStyle.Padding(0, 2).Render(label)
	case focused:
		return buttonOn.Render(label)
	default:
		return buttonOff.Render(label)
	}
}
