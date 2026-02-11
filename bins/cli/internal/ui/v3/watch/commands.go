package watch

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type tickMsg time.Time

func tick() tea.Msg {
	time.Sleep(time.Second * 5)
	return tickMsg{}
}

type workflowsFetchedMsg struct {
	workflows []*models.AppWorkflow
	err       error
}

func (m model) fetchWorkflowsCmd() tea.Msg {
	workflows, _, err := m.api.GetWorkflows(m.ctx, m.installID, nil)
	return workflowsFetchedMsg{workflows: workflows, err: err}
}
