package run

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (m *Model) setHeaderContent() {
	if m.run == nil {
		content := fmt.Sprintf("%s loading", m.spinner.View())
		m.header.SetContent(content)
		return
	}

	content := ""

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
		icon := styles.GetStepStatusIcon(string(runStatus))
		title = statusStyle.Render(fmt.Sprintf("%s ", icon)) + title
	}
	status = statusStyle.Render(fmt.Sprintf(" [%s]", runStatus))

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
	if m.run.ID != "" {
		details = styles.TextSubtle.Width(m.width).Render(m.run.ID)
	}
	bottomRow := details

	content = lipgloss.JoinVertical(lipgloss.Right, topRow, bottomRow)
	m.header.SetContent(content)
	m.header.SetHeight(lipgloss.Height(content))
}
