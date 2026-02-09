package workflow

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
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
	policies *models.AppAppPoliciesConfig
	err      error
}

func (m model) fetchWorkflowCmd() tea.Msg {
	// This runs in a goroutine automatically
	workflow, err := m.api.GetWorkflow(m.ctx, m.workflowID)
	if err != nil {
		return workflowFetchedMsg{workflow: workflow, err: err}
	}

	var policiesConfig *models.AppAppPoliciesConfig
	if workflow != nil && workflow.OwnerType == "installs" {
		install, installErr := m.api.GetInstall(m.ctx, workflow.OwnerID)
		if installErr == nil && install.AppID != "" {
			policiesConfig, _ = m.api.GetLatestAppPoliciesConfig(m.ctx, install.AppID)
		}
	}

	return workflowFetchedMsg{workflow: workflow, policies: policiesConfig, err: nil}
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
	selectedStepApprovalResponse *models.ServiceCreateWorkflowStepApprovalResponseResponse
	err                          error
}

// makeApproveStepCmd creates a command that captures the necessary values upfront
// so the UI can update immediately while the API call runs async
func (m model) makeApproveStepCmd() tea.Cmd {
	// Guard against nil selectedStep or Approval
	if m.selectedStep == nil || m.selectedStep.Approval == nil {
		return func() tea.Msg {
			return createWorkflowStepApprovalResponseMsg{err: fmt.Errorf("no step selected for approval")}
		}
	}

	// Capture values needed for the API call
	api := m.api
	ctx := m.ctx
	workflowID := m.workflowID
	stepID := m.selectedStep.ID
	approvalID := m.selectedStep.Approval.ID

	return func() tea.Msg {
		req := &models.ServiceCreateWorkflowStepApprovalResponseRequest{
			ResponseType: models.AppWorkflowStepResponseTypeApprove,
			Note:         "",
		}
		resp, err := api.CreateWorkflowStepApprovalResponse(ctx, workflowID, stepID, approvalID, req)
		return createWorkflowStepApprovalResponseMsg{selectedStepApprovalResponse: resp, err: err}
	}
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
