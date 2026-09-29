package run

import (
	"charm.land/lipgloss/v2"

	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/common"

	"go.uber.org/zap"
)

func (m Model) logMessageView() string {
	if m.status.Message == "" {
		return ""
	}
	return common.StatusBar(m.status)
}

func (m *Model) setFooterContent() {
	m.log.Info("setting footer content")

	if m.footer.Width() == 0 {
		content := "\n" + m.help.View(m.keys)
		m.footer.SetContent(content)
		return
	}

	// TODO(fd): refine
	// this catches the case where the footer is not wide enough to hold the content
	footerMaxContentWidth := m.footer.Width() - 3
	if footerMaxContentWidth < 0 {
		content := "\n" + m.help.View(m.keys)
		m.footer.SetContent(content)
		return
	}

	m.help.SetWidth(m.footer.Width())

	sections := []string{}

	if m.status.Message != "" {
		sections = append(sections, m.logMessageView())
	}
	sections = append(sections, m.help.View(m.keys))

	content := lipgloss.JoinVertical(
		lipgloss.Top,
		sections...,
	)

	m.footer.SetContent(content)
	m.footer.SetHeight(lipgloss.Height(content))

	m.log.Info(
		"setting footer content",
		zap.Int("content.length", len(content)),
		zap.Int("content.width", lipgloss.Width(content)),
		zap.Int("content.height", lipgloss.Height(content)),
	)
}
