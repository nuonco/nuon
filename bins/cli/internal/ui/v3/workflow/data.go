package workflow

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/common"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func indexOf(array []int64, value int64) int {
	for k, v := range array {
		if v == value {
			return k
		}
	}
	return -1
}

// TODO(fd): deprecate
func (m *model) getSteps() [][]*models.AppWorkflowStep {
	groups := []int64{}
	for _, step := range m.workflow.Steps {
		if !generics.SliceContains(step.GroupIdx, groups) {
			groups = append(groups, step.GroupIdx)
		}
	}
	stepsList := make([][]*models.AppWorkflowStep, len(groups))
	for _, step := range m.workflow.Steps {
		idx := indexOf(groups, step.GroupIdx)
		innerList := stepsList[idx]
		innerList = append(innerList, step)
		stepsList[idx] = innerList
	}
	return stepsList
}

func (m *model) getFlatSteps() []list.Item {
	stepsList := []list.Item{}
	for _, innerStepList := range m.steps {
		for _, step := range innerStepList {
			stepItem := listStep{step: step}
			stepsList = append(stepsList, stepItem)
		}
	}
	return stepsList
}

// TODO: put this in a goroutine
func (m *model) handleWorkflowFetched(msg workflowFetchedMsg) {
	workflow := msg.workflow
	err := msg.err
	if err != nil {
		m.setLogMessage(fmt.Sprintf("[error] failed to fetch data: %s", err), "error")
		return
	} else if workflow == nil {
		m.setLogMessage("something unexpected has taken place", "error")
		return
	}
	m.workflow = workflow
	// set progress from workflow steps
	_, _, progress := m.getProgressPercentage()
	m.progress.SetPercent(progress)
	// populate the nested step list
	stepsList := m.getSteps()
	m.steps = stepsList
	flatSteps := m.getFlatSteps() // flat steps is a flat list of sorted steps
	m.stepsList.SetItems(flatSteps)
	m.loading = false

	//
	if m.selectedStep != nil {
		// TODO(fd): factor this out and use it in setSelected
		item := m.stepsList.Items()[m.selectedIndex]
		// coerce to our type so we can use the niecities to grab the step details
		m.selectedStep = item.(listStep).Step()

	}

	// TODO(fd): hoist into an isDone method
	if generics.SliceContains(m.workflow.Status.Status, []models.AppStatus{models.AppStatusCancelled, models.AppStatusError, models.AppStatusSuccess}) {
		m.keys.CancelWorkflow.SetEnabled(false)
	}

	m.populateStepDetailView(false)
	m.loading = false

	// In watch mode, check for action-required conditions and set exit
	if m.watchMode {
		m.checkWatchModeExit()
	}
}

// checkWatchModeExit checks if we should exit in watch mode due to action-required conditions
func (m *model) checkWatchModeExit() {
	if m.workflow == nil {
		return
	}

	// Skip exit checks while an action is being processed
	if m.actionInFlight || m.approvingStep {
		return
	}

	// Check for approval-pending steps
	for _, step := range m.workflow.Steps {
		if step.Approval != nil && step.Status != nil {
			if step.Status.Status == models.AppStatusApprovalDashAwaiting && !step.Finished {
				m.exitCode = ExitCodeApprovalRequired
				m.exitReason = exitReason{
					stepName: step.Name,
					stepID:   step.ID,
				}
				m.triggerExit()
				return
			}
		}
	}

	// Check for failed steps (error, cancelled, discarded)
	// We check for error/cancelled/discarded first to report the actual failure
	failedStatuses := []models.AppStatus{
		models.AppStatusError,
		models.AppStatusCancelled,
		models.AppStatusDiscarded,
	}
	for _, step := range m.workflow.Steps {
		if step.Status != nil && generics.SliceContains(step.Status.Status, failedStatuses) {
			m.exitCode = ExitCodeStepFailed
			m.exitReason = exitReason{
				stepName: step.Name,
				stepID:   step.ID,
			}
			if step.Status.StatusHumanDescription != "" {
				m.exitReason.errorMessage = step.Status.StatusHumanDescription
			}
			m.triggerExit()
			return
		}
	}

	// Check for not-attempted steps - find the last processed step and report its status
	hasNotAttempted := false
	var firstNotAttemptedIdx int64 = -1
	for _, step := range m.workflow.Steps {
		if step.Status != nil && step.Status.Status == models.AppStatusNotDashAttempted && step.Finished {
			hasNotAttempted = true
			if firstNotAttemptedIdx == -1 || step.Idx < firstNotAttemptedIdx {
				firstNotAttemptedIdx = step.Idx
			}
		}
	}
	if hasNotAttempted {
		// Find the last processed step (step before first not-attempted, or highest processed index)
		var lastProcessed *models.AppWorkflowStep
		for _, step := range m.workflow.Steps {
			if step.Status == nil {
				continue
			}
			// If we have a not-attempted step, look for the step just before it
			if firstNotAttemptedIdx > 0 && step.Idx == firstNotAttemptedIdx-1 {
				lastProcessed = step
				break
			}
			// Fallback: find highest index step that was processed
			status := step.Status.Status
			if status != models.AppStatusNotDashAttempted && status != models.AppStatusPending {
				if lastProcessed == nil || step.Idx > lastProcessed.Idx {
					lastProcessed = step
				}
			}
		}

		if lastProcessed != nil && lastProcessed.Status != nil {
			m.exitCode = ExitCodeStepFailed
			m.exitReason = exitReason{
				stepName: lastProcessed.Name,
				stepID:   lastProcessed.ID,
			}
			if lastProcessed.Status.StatusHumanDescription != "" {
				m.exitReason.errorMessage = lastProcessed.Status.StatusHumanDescription
			} else {
				m.exitReason.errorMessage = fmt.Sprintf("Step ended with status: %s", lastProcessed.Status.Status)
			}
			m.triggerExit()
			return
		}
	}

	// Check if workflow was cancelled
	if m.workflow.Status != nil && m.workflow.Status.Status == models.AppStatusCancelled {
		m.exitCode = ExitCodeCancelled
		m.exitReason = exitReason{workflowStatus: "cancelled"}
		m.triggerExit()
		return
	}

	// Check if all steps are finished (success case)
	// We rely on step statuses rather than workflow status to avoid premature exit
	_, pending, _ := common.CalculateStepProgress(m.workflow.Steps)
	if pending == 0 && len(m.workflow.Steps) > 0 {
		m.exitCode = ExitCodeSuccess
		m.exitReason = exitReason{workflowStatus: "success"}
		m.triggerExit()
	}
}

// triggerExit either starts the countdown or quits immediately based on exitWait
func (m *model) triggerExit() {
	// Don't trigger exit while approving
	if m.approvingStep {
		return
	}
	if m.exitWait > 0 && !m.exitWaiting {
		m.exitWaiting = true
		m.exitCountdown = m.exitWait
	} else if m.exitWait == 0 {
		m.quitting = true
	}
}

// resetExitCountdown resets the exit countdown when user interacts with the TUI
func (m *model) resetExitCountdown() {
	if m.exitWaiting {
		m.exitCountdown = m.exitWait
	}
}

// cancelExitCountdown cancels the exit countdown (e.g., when a new step begins)
func (m *model) cancelExitCountdown() {
	m.exitWaiting = false
	m.exitCountdown = 0
	m.exitCode = 0
	m.exitReason = exitReason{}
	m.actionInFlight = true // skip exit checks until action response is received
}

func (m *model) handleStackFetched(msg stackFetchedMsg) {
	stack := msg.stack
	err := msg.err
	m.stack = stack
	if err != nil {
		m.error = err
	}
}

func (m *model) handleWorkflowStepApprovalResponseCreated(msg createWorkflowStepApprovalResponseMsg) tea.Cmd {
	resp := msg.selectedStepApprovalResponse
	err := msg.err
	if err != nil {
		m.error = err
	}
	m.selectedStepApprovalResponse = resp
	m.loading = false
	m.stepApprovalConf = false
	m.approvingStep = false
	m.actionInFlight = false // action completed, re-enable exit checks
	// after a step is approved, we want to immediately fetch the workflow to get the upated version
	return m.fetchWorkflowCmd
}

func (m *model) handleCancelWorkflow(msg cancelWorkflowMsg) tea.Cmd {
	m.loading = false
	_, err := m.api.CancelWorkflow(m.ctx, m.workflowID)
	if msg.err != nil {
		m.error = err
	}
	m.setLogMessage("workflow has been cancelled", "error")
	m.resetSelected()
	m.resetWorkflowCancelationConf()
	m.actionInFlight = false // action completed, re-enable exit checks
	return m.fetchWorkflowCmd
}

func (m *model) handleApproveAll(msg approveAllMsg) []tea.Cmd {
	cmds := []tea.Cmd{}
	if msg.err != nil {
		m.setLogMessage(fmt.Sprintf("%s", msg.err), "error")
	}
	m.workflowApprovalConf = false
	m.approvingStep = false
	m.actionInFlight = false // action completed, re-enable exit checks
	m.populateStepDetailView(true)
	if msg.approved > 0 {
		m.setLogMessage(fmt.Sprintf("approved %02d workflows", msg.approved), "success")
		cmds = append(cmds, m.fetchWorkflowCmd)
	} else {
		m.setLogMessage("nothing to approve", "warning")
	}
	return cmds
}

func (m *model) handleGetWorkflowStepApprovalContents(msg getWorkflowStepApprovalContentsMsg) []tea.Cmd {

	if msg.err != nil {
		m.approvalContents = approvalContents{error: msg.err, raw: msg.raw, loading: false}
		return []tea.Cmd{}
	}
	contents, err := interfaceToMap(msg.raw)
	if err != nil {
		m.approvalContents = approvalContents{error: err, raw: msg.raw, loading: false}
		return []tea.Cmd{}
	}
	m.approvalContents = approvalContents{error: err, raw: msg.raw, loading: false, contents: contents}
	m.populateStepDetailView(false)
	m.setLogMessage(
		fmt.Sprintf("workflow content fetched %02d keys", len(contents)),
		"info",
	)

	return []tea.Cmd{}
}
