package workflow

import (
	"charm.land/lipgloss/v2"

	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/common"
	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

var buttonStyle = lipgloss.NewStyle().
	Foreground(styles.SuccessColor).
	Padding(0, 3)

var dangerousButtonStyle = buttonStyle.
	Foreground(styles.WarningColor)

func (m model) getProgressPercentage() (int, int, float64) {
	if m.workflow == nil {
		return 0, 0, 0
	}
	return common.CalculateStepProgress(m.workflow.Steps)
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

// TODO(fd): write an is-retryable for a workflow step

// Actual View Code
func (m model) actionsOrMessage() string {
	// Show approving indicator when approval is in flight
	if m.approvingStep {
		return m.spinner.View() + " Approving..."
	}

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
	if m.autoRetryAll {
		content += dangerousButtonStyle.Render("[R] Auto-Retry ✓")
	} else {
		content += buttonStyle.Render("[R] Auto-Retry")
	}
	return content
}

func (m model) headerView() string {
	/*
		renders two rows
		1. title + action instructions
		2. details _ progress

		unless it's loading, in which case we render a single row
	*/
	content := ""
	if len(m.steps) == 0 {
		content += m.spinner.View() + " loading ..."
		m.header.SetContent(content)
		return appStyle.Render(m.header.View())
	}
	stepsFinished, stepsPending, progress := m.getProgressPercentage()
	m.progress.SetPercent(progress)

	content = common.RenderHeader(common.HeaderConfig{
		Width:         m.header.Width(),
		Title:         m.workflow.Name,
		Status:        m.workflow.Status.Status,
		StepsTotal:    len(m.workflow.Steps),
		StepsFinished: stepsFinished,
		StepsPending:  stepsPending,
		Progress:      progress,
		ProgressBar:   m.progress,
		Spinner:       m.spinner,
		Actions:       m.actionsOrMessage(),
		ShowPlanOnly:  m.workflow.PlanOnly,
		ShowSpinner:   true,
	})

	m.header.SetContent(content)
	return appStyle.Render(m.header.View())
}
