package workflow

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
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
func (m *model) fetchWorkflow() {
	m.loading = true
	workflow, err := m.api.GetWorkflow(m.ctx, m.workflowID)
	if err != nil {
		m.logMessage = fmt.Sprintf("[error] failed to fetch data: %s", err)
		return
	}
	if m.stepApprovalConf {
		// in this case, do not override the message
		m.logMessage = fmt.Sprintf("[%s] fetched workflow id:%s", time.Now().String(), workflow.ID)
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
	m.getInstallStack()
}

func (m *model) approveWorkflowStep() {
	req := &models.ServiceCreateWorkflowStepApprovalResponseRequest{
		ResponseType: models.AppWorkflowStepResponseTypeApprove,
		Note:         "",
	}
	resp, err := m.api.CreateWorkflowStepApprovalResponse(m.ctx, m.workflowID, m.selectedStep.ID, m.selectedStep.Approval.ID, req)
	if err != nil {
		m.error = err
		return
	}
	m.selectedStepApprovalResponse = resp
	m.logMessage = fmt.Sprintf("[%s] step approved %s", resp.Type, resp.InstallWorkflowStepApprovalID)
	m.loading = false
	m.stepApprovalConf = false
}

func (m *model) cancelWorkflow() {
	m.loading = true
	_, err := m.api.CancelWorkflow(m.ctx, m.workflowID)
	if err != nil {
		m.error = err
		return
	}
	m.loading = false
	m.logMessage = "workflow has been cancelled"
	m.resetSelected()
	m.fetchWorkflow()
}

func (m *model) getInstallStack() {
	m.stackLoading = true
	stack, err := m.api.GetInstallStack(m.ctx, m.installID)
	if err != nil {
		m.error = err
	}
	m.stack = stack
	m.stackLoading = false
	m.populateStepDetailView(false)
}

func (m *model) approveAll() {

	m.loading = true
	m.logMessage = "approving all workflows"
	approved := 0
	for i, step := range m.workflow.Steps {
		if step.Approval == nil || step.Approval.Response != nil {
			continue
		}
		req := &models.ServiceCreateWorkflowStepApprovalResponseRequest{
			ResponseType: models.AppWorkflowStepResponseTypeApprove,
			Note:         "",
		}
		resp, err := m.api.CreateWorkflowStepApprovalResponse(m.ctx, m.workflowID, step.ID, step.Approval.ID, req)
		if err != nil {
			m.error = err
			m.logMessage = fmt.Sprintf("%s", err)
			return
		}
		m.selectedStepApprovalResponse = resp
		m.logMessage = fmt.Sprintf("[%02d] step \"%s\" approved", i, step.Name)
		approved += 1
		m.fetchWorkflow()
	}
	m.loading = false
	m.workflowApprovalConf = false
	m.stepApprovalConf = false
	m.populateStepDetailView(true)
}
