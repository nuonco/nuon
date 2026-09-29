package detail

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/nuonco/nuon/pkg/cli/styles"
)

func (m Model) headerView() string {
	content := ""
	if m.installActionWorkflow == nil {
		content += m.spinner.View() + " loading ..."
		m.header.SetContent(content)
		return appStyle.Render(m.header.View())
	}

	title := m.installActionWorkflow.ActionWorkflow.Name
	status := ""
	prompt := styles.TextSuccess.Padding(0, 1).Render("[E] Execute this Action")

	latestStatus := ""
	if len(m.installActionWorkflow.Runs) > 0 {
		latestRun := m.installActionWorkflow.Runs[0]
		latestStatus = latestRun.Status
		statusStyle := styles.GetRunStatusStyle(latestStatus)

		if latestStatus == "in_progress" {
			title = m.spinner.View() + " " + title
		} else {
			icon := styles.GetRunStatusIcon(latestStatus)
			title = statusStyle.Render(fmt.Sprintf("%s ", icon)) + title
		}
		status = statusStyle.Render(fmt.Sprintf(" [%s]", latestStatus))
	}

	left := lipgloss.JoinHorizontal(lipgloss.Left, title, status)
	right := lipgloss.JoinHorizontal(lipgloss.Left, prompt)
	spacer := strings.Repeat(" ", max(m.width-2-lipgloss.Width(left)-lipgloss.Width(right), 0))

	topRow := lipgloss.NewStyle().Width(m.width).Render(
		lipgloss.JoinHorizontal(
			lipgloss.Center,
			left,
			spacer,
			right,
		),
	)

	details := ""
	if m.installActionWorkflow.ActionWorkflowID != "" {
		details = styles.TextSubtle.Width(m.width).Render(m.installActionWorkflow.ActionWorkflowID)
	}
	bottomRow := details

	content = lipgloss.JoinVertical(lipgloss.Right, topRow, bottomRow)
	m.header.SetContent(content)
	return appStyle.Render(m.header.View())
}
