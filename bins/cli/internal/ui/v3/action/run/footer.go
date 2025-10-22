package run

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/common"
	// "github.com/powertoolsdev/mono/pkg/cli/styles"
	"go.uber.org/zap"
)

func (m Model) logMessageView() string {
	if m.status.Message == "" {
		return ""
	}
	return common.StatusBar(m.status)
}

func (m *Model) setFooterContent() {
	/*
		renders two rows
		1. log message
		2. help footer

		If the log message is empty, that row is omitted.
	*/
	m.log.Info("setting footer content")

	// we have to handle this base case since the element widths are zero on init and we use the footer width
	// to determine some content widthS
	if m.footer.Width == 0 {
		content := "\n" + m.help.View(m.keys)
		m.footer.SetContent(content)
		return
	}

	// TODO(fd): refine
	// this catches the case where the footer is not wide enough to hold the content
	footerMaxContentWidth := m.footer.Width - 3
	if footerMaxContentWidth < 0 {
		content := "\n" + m.help.View(m.keys)
		m.footer.SetContent(content)
		return
	}

	// happy path
	// set help width
	m.help.Width = m.footer.Width

	sections := []string{}

	if m.status.Message != "" {
		sections = append(sections, m.logMessageView())
	}
	sections = append(sections, m.help.View(m.keys))

	// set contents

	content := lipgloss.JoinVertical(
		lipgloss.Top,
		sections...,
	)

	// set the content and adjust the footer height
	m.footer.SetContent(content)
	m.footer.Height = lipgloss.Height(content)

	m.log.Info(
		"setting footer content",
		zap.Int("content.length", len(content)),
		zap.Int("content.width", lipgloss.Width(content)),
		zap.Int("content.height", lipgloss.Height(content)),
	)
}
