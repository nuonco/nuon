package detail

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/common"
	"github.com/powertoolsdev/mono/pkg/cli/styles"
)

func (m Model) headerView() string {
	/*
		renders two rows
		1. title + status indicator
		2. action workflow ID

		unless it's loading, in which case we render a single row
	*/
	content := ""
	if m.installActionWorkflow == nil {
		content += m.spinner.View() + " loading ..."
		m.header.SetContent(content)
		return appStyle.Render(m.header.View())
	}

	// top header row
	// [[⚪️title [status]] ... [[E] Execute]]
	// 1. title and status from latest run
	// 2. Action Indicator/Prompt

	title := m.installActionWorkflow.ActionWorkflow.Name
	status := ""
	prompt := styles.TextSuccess.Padding(0, 1).Render("[E] Execute this Action")

	latestStatus := ""
	// get status from the latest run (first in the list)
	if len(m.installActionWorkflow.Runs) > 0 {
		latestRun := m.installActionWorkflow.Runs[0]
		latestStatus = latestRun.Status
		statusStyle := getRunStatusStyle(latestStatus)

		if latestStatus == "in_progress" {
			title = m.spinner.View() + " " + title
		} else {
			icon := getRunStatusIcon(latestStatus)
			title = statusStyle.Render(fmt.Sprintf("%s ", icon)) + title
		}
		status = statusStyle.Render(fmt.Sprintf(" [%s]", latestStatus))
	}

	left := lipgloss.JoinHorizontal(lipgloss.Left, title, status)
	right := lipgloss.JoinHorizontal(lipgloss.Left, prompt)
	spacer := strings.Repeat(" ", common.Max(m.width-2-lipgloss.Width(left)-lipgloss.Width(right), 0))

	// top row has two sections:
	// [ title ] ... [ status ] with spacing between
	topRow := lipgloss.NewStyle().Width(m.width).Render(
		lipgloss.JoinHorizontal(
			lipgloss.Center,
			left,
			spacer,
			right,
		),
	)

	// bottom row - action workflow ID
	details := ""
	if m.installActionWorkflow.ActionWorkflowID != "" {
		details = styles.TextSubtle.Width(m.width).Render(m.installActionWorkflow.ActionWorkflowID)
	}
	bottomRow := details

	content = lipgloss.JoinVertical(lipgloss.Right, topRow, bottomRow)
	m.header.SetContent(content)
	return appStyle.Render(m.header.View())
}
