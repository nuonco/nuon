package workflow

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/common"
	"github.com/powertoolsdev/mono/pkg/cli/styles"
)

func (m model) logMessageView() string {
	if m.status.Message == "" {
		return ""
	}
	return common.StatusBar(m.status)
}

// TODO: Log Message should be its own component tbh
func (m model) footerView() string {
	/*

		renders two rows
		1. log message
		2. help footer

		If the log message is empty, that row is omitted.

	*/
	// we have to handle this base case since the element widths are zero on init
	if m.footer.Width == 0 {
		content := "\n" + m.help.View(m.keys)
		m.footer.SetContent(content)
		return m.footer.View()
	}
	footerMaxContentWidth := m.footer.Width - 3
	if footerMaxContentWidth < 0 {
		content := "\n" + m.help.View(m.keys)
		m.footer.SetContent(content)
		return m.footer.View()
	}

	sections := []string{}

	if m.status.Message != "" {
		sections = append(sections, m.logMessageView())
	}
	sections = append(sections, styles.HelpStyle.Render(m.help.View(m.keys)))

	// set contents
	m.footer.SetContent(lipgloss.JoinVertical(
		lipgloss.Top,
		sections...,
	))
	return m.footer.View()
}
