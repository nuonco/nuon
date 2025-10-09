package action

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/powertoolsdev/mono/pkg/cli/styles"
)

func (m model) headerView() string {
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
	// 1. title and status from latest run
	title := ""
	status := ""
	latestStatus := ""
	if m.installActionWorkflow.ActionWorkflow != nil {
		title = m.installActionWorkflow.ActionWorkflow.Name
	}

	// Get status from the latest run (first in the list)
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
		status = fmt.Sprintf("%s %s ", styles.TextDim.Render("latest status:"), statusStyle.Render(latestStatus))
	}

	// top row has two sections:
	// [ title ] ... [ status ] with spacing between
	spacer := strings.Repeat(" ", int(math.Max(float64(m.width-2-lipgloss.Width(title)-lipgloss.Width(status)), float64(0))))
	topRow := lipgloss.NewStyle().Width(m.width).Render(
		lipgloss.JoinHorizontal(
			lipgloss.Center,
			title,
			spacer,
			status,
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
