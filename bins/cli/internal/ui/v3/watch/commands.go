package watch

import (
	tea "charm.land/bubbletea/v2"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type workflowsFetchedMsg struct {
	workflows []*models.AppWorkflow
	err       error
}

func (m model) fetchWorkflowsCmd() tea.Msg {
	workflows, _, err := m.api.GetWorkflows(m.ctx, m.installID, nil)
	return workflowsFetchedMsg{workflows: workflows, err: err}
}
