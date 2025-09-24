package workflow

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/styles"
)

func (m model) logMessageView() string {
	// log message for footer with padding or truncation

	if m.logMessage == "" {
		return ""
	}
	footerMaxContentWidth := m.footer.Width - 2
	content := ""
	padding := ""
	caret := "> "
	if m.loading {
		caret = m.spinner.View()
	}
	if len(m.logMessage) < footerMaxContentWidth {
		repeatCount := footerMaxContentWidth - len(m.logMessage)
		if repeatCount > 0 {
			padding = strings.Repeat(" ", repeatCount)
		}
		content += styles.LogMessageStyle.Render(
			lipgloss.JoinHorizontal(lipgloss.Center,
				caret,
				fmt.Sprintf("%s%s", m.logMessage, padding),
			),
		)
	} else if len(m.logMessage) > footerMaxContentWidth {
		contentWidth := len(m.logMessage) - footerMaxContentWidth
		truncatedLogMessage := m.logMessage[:contentWidth]
		content += styles.LogMessageStyle.Render(
			lipgloss.JoinHorizontal(lipgloss.Center,
				caret,
				fmt.Sprintf("%s", truncatedLogMessage),
			),
		)
	}
	return content
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

	logMessage := m.logMessageView()
	helpView := styles.HelpStyle.Render(m.help.View(m.keys))

	// set contents
	m.footer.SetContent(lipgloss.JoinVertical(
		lipgloss.Top,
		logMessage,
		helpView,
	))
	return m.footer.View()
}
