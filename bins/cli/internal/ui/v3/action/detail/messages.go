package detail

import (
	tea "charm.land/bubbletea/v2"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type installActionWorkflowFetchedMsg struct {
	installActionWorkflow *models.AppInstallActionWorkflow
	err                   error
}

func (m Model) fetchInstallActionWorkflowCmd() tea.Msg {
	// This runs in a goroutine automatically
	installActionWorkflow, _, err := m.api.GetInstallActionWorkflowRecentRuns(m.ctx, m.installID, m.actionWorkflowID, nil)
	return installActionWorkflowFetchedMsg{installActionWorkflow: installActionWorkflow, err: err}
}

type latestConfigFetchedMsg struct {
	config *models.AppActionWorkflowConfig
	err    error
}

func (m Model) fetchLatestConfigCmd() tea.Msg {
	// This runs in a goroutine automatically
	config, err := m.api.GetActionWorkflowLatestConfig(m.ctx, m.actionWorkflowID)
	return latestConfigFetchedMsg{config: config, err: err}
}
