package watch

import (
	"charm.land/lipgloss/v2"

	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/common"
	"github.com/nuonco/nuon/pkg/cli/styles"
)

func (m model) logMessageView() string {
	if m.status.Message == "" {
		return ""
	}
	return common.StatusBar(m.status)
}

func (m model) footerView() string {
	if m.footer.Width() == 0 {
		content := "\n" + m.help.View(m.keys)
		m.footer.SetContent(content)
		return m.footer.View()
	}
	footerMaxContentWidth := m.footer.Width() - 3
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

	m.footer.SetContent(lipgloss.JoinVertical(
		lipgloss.Top,
		sections...,
	))
	return m.footer.View()
}
