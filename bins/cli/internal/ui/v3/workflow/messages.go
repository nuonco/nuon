package workflow

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nuonco/nuon-go/models"
)

// Messages are events that we respond to in our Update function. This
// particular one indicates that the timer has ticked.
type tickMsg time.Time

func tick() tea.Msg {
	time.Sleep(time.Second * 5)
	return tickMsg{}
}

type workflowFetchedMsg struct {
	workflow *models.AppWorkflow
	stack    *models.AppInstallStack
	err      error
}

func (m model) fetchWorkflowCmd() tea.Msg {
	// This runs in a goroutine automatically
	workflow, err := m.api.GetWorkflow(m.ctx, m.workflowID)
	return workflowFetchedMsg{workflow: workflow, err: err}
}

type stackFetchedMsg struct {
	stack *models.AppInstallStack
	err   error
}

func (m model) fetchStackCmd() tea.Msg {
	// This runs in a goroutine automatically
	stack, err := m.api.GetInstallStack(m.ctx, m.installID)
	return stackFetchedMsg{stack: stack, err: err}
}

type createWorkflowStepApprovalResponseMsg struct {
	selectedStepApprovalResponse *models.AppWorkflowStepApprovalResponse
	err                          error
}

func (m model) createWorkflowStepApprovalResponseCmd() tea.Msg {
	// This runs in a goroutine automatically
	req := &models.ServiceCreateWorkflowStepApprovalResponseRequest{
		ResponseType: models.AppWorkflowStepResponseTypeApprove,
		Note:         "",
	}
	approvalResponse, err := m.api.CreateWorkflowStepApprovalResponse(m.ctx, m.workflowID, m.selectedStep.ID, m.selectedStep.Approval.ID, req)
	m.setLogMessage(fmt.Sprintf("[%s] step approved %s", approvalResponse.Type, approvalResponse.InstallWorkflowStepApprovalID), "success")
	return createWorkflowStepApprovalResponseMsg{selectedStepApprovalResponse: approvalResponse, err: err}
}

type cancelWorkflowMsg struct {
	err error
}

func (m model) cancelWorkflowCmd() tea.Msg {
	// This runs in a goroutine automatically
	_, err := m.api.CancelWorkflow(m.ctx, m.workflowID)

	m.setLogMessage(fmt.Sprintf("[%s] workflow cancelled", m.workflowID), "error")

	return cancelWorkflowMsg{err: err}
}

type approveAllMsg struct {
	approved int
	err      error
}

func (m model) approveAllCmd() tea.Msg {
	// This runs in a goroutine automatically
	approved := 0
	for _, step := range m.workflow.Steps {
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
			return approveAllMsg{err: err}
		}
		m.selectedStepApprovalResponse = resp
		approved += 1
		m.fetchWorkflowCmd()
	}

	return approveAllMsg{approved: approved, err: nil}
}

type getWorkflowStepApprovalContentsMsg struct {
	raw     interface{}
	loading bool
	err     error
}

func (m model) getWorkflowStepApprovalContentsCmd() tea.Msg {
	resp, err := m.api.GetWorkflowStepApprovalContents(
		m.ctx,
		m.workflowID,
		m.selectedStep.ID,
		m.selectedStep.Approval.ID,
	)
	return getWorkflowStepApprovalContentsMsg{raw: resp, err: err}
}
