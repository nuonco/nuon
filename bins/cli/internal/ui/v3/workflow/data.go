package workflow

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
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
			stepItem := listStep{step: step, spinnerView: m.spinner.View()}
			stepsList = append(stepsList, stepItem)
		}
	}
	return stepsList
}

// updateListSpinnerViews updates the spinner view string on all list items
// so in-progress step spinners animate at the same rate as the header spinner.
func (m *model) updateListSpinnerViews() {
	items := m.stepsList.Items()
	if len(items) == 0 {
		return
	}
	sv := m.spinner.View()
	for i, item := range items {
		if step, ok := item.(listStep); ok {
			step.spinnerView = sv
			items[i] = step
		}
	}
	m.stepsList.SetItems(items)
}

// hasRetryableCurrentStep returns true when any step is in error and retryable.
// Mirrors the target selection logic in retryAllCmd.
func (m *model) hasRetryableCurrentStep() bool {
	if m.workflow == nil {
		return false
	}
	for _, step := range m.workflow.Steps {
		if step.Status != nil && step.Status.Status == models.AppStatusError && step.Retryable {
			return true
		}
	}
	return false
}

// TODO: put this in a goroutine
func (m *model) handleWorkflowFetched(msg workflowFetchedMsg) tea.Cmd {
	workflow := msg.workflow
	err := msg.err
	if err != nil {
		m.setLogMessage(fmt.Sprintf("[error] failed to fetch data: %s", err), "error")
		return nil
	} else if workflow == nil {
		m.setLogMessage("something unexpected has taken place", "error")
		return nil
	}
	m.workflow = workflow
	if msg.policies != nil {
		policyNames := make(map[string]string)
		for _, policy := range msg.policies.Policies {
			if policy == nil || policy.ID == "" {
				continue
			}
			name := policy.Name
			if name == "" {
				name = policy.ID
			}
			policyNames[policy.ID] = name
		}
		m.policyNames = policyNames
	}
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
	m.syncHelmDiffExplorer()

	// TODO(fd): hoist into an isDone method
	if generics.SliceContains(m.workflow.Status.Status, []models.AppStatus{models.AppStatusCancelled, models.AppStatusError, models.AppStatusSuccess}) {
		m.keys.CancelWorkflow.SetEnabled(false)
	}

	m.populateStepDetailView(false)
	m.loading = false

	if m.autoRetryAll && !m.retryInFlight && m.hasRetryableCurrentStep() {
		m.retryInFlight = true
		return m.retryAllCmd
	}
	// clear the in-flight flag once the step is no longer in error
	if m.retryInFlight && !m.hasRetryableCurrentStep() {
		m.retryInFlight = false
	}
	return nil
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
	return m.fetchWorkflowCmd
}

func (m *model) handleApproveAll(msg approveAllMsg) []tea.Cmd {
	cmds := []tea.Cmd{}
	if msg.err != nil {
		m.setLogMessage(fmt.Sprintf("%s", msg.err), "error")
	}
	m.workflowApprovalConf = false
	m.approvingStep = false
	m.populateStepDetailView(true)
	if msg.approved > 0 {
		m.setLogMessage(fmt.Sprintf("approved %02d workflows", msg.approved), "success")
		cmds = append(cmds, m.fetchWorkflowCmd)
	} else {
		m.setLogMessage("nothing to approve", "warning")
	}
	return cmds
}

func (m *model) handleRetryAll(msg retryAllMsg) []tea.Cmd {
	cmds := []tea.Cmd{}
	if msg.err != nil {
		m.retryInFlight = false // allow a retry attempt on the next poll
		m.setLogMessage(fmt.Sprintf("retry error: %s", msg.err), "error")
	}
	if msg.retried > 0 {
		m.setLogMessage("retrying from last failed step", "success")
		cmds = append(cmds, m.fetchWorkflowCmd)
	} else if msg.err == nil {
		m.setLogMessage("no retryable failed steps found", "warning")
	}
	return cmds
}

func (m *model) handleGetWorkflowStepApprovalContents(msg getWorkflowStepApprovalContentsMsg) []tea.Cmd {

	if msg.err != nil {
		m.approvalContents = approvalContents{error: msg.err, raw: msg.raw, loading: false}
		m.syncHelmDiffExplorer()
		return []tea.Cmd{}
	}
	contents, err := interfaceToMap(msg.raw)
	if err != nil {
		m.approvalContents = approvalContents{error: err, raw: msg.raw, loading: false}
		m.syncHelmDiffExplorer()
		return []tea.Cmd{}
	}
	m.approvalContents = approvalContents{error: err, raw: msg.raw, loading: false, contents: contents}
	m.syncHelmDiffExplorer()
	m.populateStepDetailView(false)
	m.setLogMessage(
		fmt.Sprintf("workflow content fetched %02d keys", len(contents)),
		"info",
	)

	return []tea.Cmd{}
}
