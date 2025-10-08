package workflow

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/pkg/generics"
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
