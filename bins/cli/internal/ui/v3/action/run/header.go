package run

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/common"
	"github.com/powertoolsdev/mono/pkg/cli/styles"
)

func (m *Model) setHeaderContent() {
	/*
		renders two rows
		1. title + status indicator + browser link
		2. run ID

		unless it's loading, in which case we render a single row

		NOTE: this method just sets the content and should be called whenever the data changes.
		      to USE this view use m.header.View
	*/
	if m.run == nil {
		content := fmt.Sprintf("%s loading", m.spinner.View())
		m.header.SetContent(content)
		return
	}

	content := ""

	// top header row
	// [[⚪️title [status]] ... [[B] Browser]]
	// 1. title and status from run
	// 2. Browser link prompt

	title := "Action Workflow Run"
	if m.run.InstallActionWorkflow != nil && m.run.InstallActionWorkflow.ActionWorkflow != nil {
		title = m.run.InstallActionWorkflow.ActionWorkflow.Name
	}

	status := ""
	prompt := styles.TextSuccess.Padding(0, 1).Render("[B] Open in Browser")

	runStatus := models.AppStatus(m.run.Status)
	if m.run.StatusV2 != nil && m.run.StatusV2.Status != "" {
		runStatus = m.run.StatusV2.Status
	}

	statusStyle := styles.GetStatusStyle(runStatus)

	if runStatus == "in_progress" || runStatus == "running" {
		title = m.spinner.View() + " " + title
	} else {
		icon := getStepStatusIcon(string(runStatus))
		title = statusStyle.Render(fmt.Sprintf("%s ", icon)) + title
	}
	status = statusStyle.Render(fmt.Sprintf(" [%s]", runStatus))

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

	// bottom row - run ID
	details := ""
	if m.run.ID != "" {
		details = styles.TextSubtle.Width(m.width).Render(m.run.ID)
	}
	bottomRow := details

	content = lipgloss.JoinVertical(lipgloss.Right, topRow, bottomRow)
	m.header.SetContent(content)
	m.header.Height = lipgloss.Height(content)
}
