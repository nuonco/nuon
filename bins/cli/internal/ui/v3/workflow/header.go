package workflow

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/styles"
	"github.com/powertoolsdev/mono/pkg/generics"
)

var buttonStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FFF7DB")).
	Background(lipgloss.Color("#888B7E")).
	Padding(0, 3)

var dangerousButtonStyle = buttonStyle.
	Foreground(lipgloss.Color("#FFF7DB")).
	Background(lipgloss.Color("#F25D94"))

func (m model) getProgressPercentage() (int, int, float64) {
	stepsTotal := len(m.workflow.Steps)
	stepsFinished := 0
	for _, step := range m.workflow.Steps {
		if step.Finished {
			stepsFinished++
		}
	}
	stepsPending := stepsTotal - stepsFinished
	progress := float64(stepsFinished) / float64(stepsTotal)
	return stepsFinished, stepsPending, progress
}

// action checks for workflows
func (m model) workflowIsCancellable() bool {
	// is this workflow cancellable
	cancellableStatuses := []models.AppStatus{
		models.AppStatusPending,
		models.AppStatusInDashProgress,
		models.AppStatusRetrying,
		models.AppStatusApproved,
		// models.AppStatus
	}
	return generics.SliceContains(m.workflow.Status.Status, cancellableStatuses)
}

func (m model) workflowIsApprovable() bool {
	// is this workflow approvable
	approvableStatuses := []models.AppStatus{
		models.AppStatusInDashProgress,
		models.AppStatusRetrying,
		// models.AppStatus
	}
	return generics.SliceContains(m.workflow.Status.Status, approvableStatuses)
}

// TOOD(fd): write an is-retryable for a workflow step

// Actual View Code
func (m model) actionsOrMessage() string {
	cancellable := m.workflowIsCancellable()
	approvable := m.workflowIsApprovable()
	_, _, progress := m.getProgressPercentage()
	content := ""
	if !cancellable && !approvable && progress == 1.0 {
		return styles.TextSuccess.Render("All changes have been approved.")
	}
	if approvable {
		content += buttonStyle.Render("[A] Approve All")
	}
	if cancellable {
		content += dangerousButtonStyle.Render("[C] Cancel Workflow") + " "
	}
	return content
}

func (m model) headerView() string {
	/*
		renders two rows
		1. title + action instructions
		2. details _ progress
	*/
	content := ""
	if len(m.steps) == 0 {
		content += m.spinner.View() + " loading ..."
		m.header.SetContent(content)
		return appStyle.Render(m.header.View())
	}
	stepsFinished, stepsPending, progress := m.getProgressPercentage()
	m.progress.SetPercent(progress)

	// top header row
	// 1. title
	style := styles.GetStatusStyle(m.workflow.Status.Status)
	title := ""
	if m.workflow.Status.Status == models.AppStatusInDashProgress {
		title += m.spinner.View()
	} else {
		icon := getStatusIcon(m.workflow.Status.Status)
		title += style.Render(fmt.Sprintf("%s ", icon))
	}

	title += fmt.Sprintf("%s ", m.workflow.Name)
	if m.workflow.PlanOnly {
		title += styles.TextDim.Render("[Plan Only] ")
	}
	title += fmt.Sprintf("(%02d steps)", len(m.workflow.Steps))

	if m.workflow.Status.Status != models.AppStatusInDashProgress {
		title += style.Render(fmt.Sprintf("[%s]", m.workflow.Status.Status))
	}

	// 2. actions
	actions := m.actionsOrMessage()
	spacer := m.spinner.View()
	spacerCount := m.header.Width - (lipgloss.Width(title) + lipgloss.Width(actions))
	if spacerCount > 0 {
		spacer = strings.Repeat(" ", spacerCount)
	}
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, title, spacer, actions)

	// bottom row
	details := styles.TextDim.Render("  pending: ") + fmt.Sprintf("%02d ", stepsPending)
	details += styles.TextDim.Render("total:  ") + fmt.Sprintf("%02d ", len(m.workflow.Steps))
	details += styles.TextDim.Render("completed: ") + fmt.Sprintf("%02d", stepsFinished)
	spacerBottom := ""
	spacerBottomCount := m.header.Width - (lipgloss.Width(details) + m.progress.Width)
	if spacerBottomCount > 0 {
		spacerBottom = strings.Repeat(" ", spacerBottomCount)
	}
	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, details, spacerBottom, m.progress.ViewAs(progress))

	content += lipgloss.JoinVertical(lipgloss.Center, topRow, bottomRow)
	m.header.SetContent(content)
	return appStyle.Render(m.header.View())
}
