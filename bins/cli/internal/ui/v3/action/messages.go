package action

import (
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

type installActionWorkflowFetchedMsg struct {
	installActionWorkflow *models.AppInstallActionWorkflow
	err                   error
}

func (m model) fetchInstallActionWorkflowCmd() tea.Msg {
	// This runs in a goroutine automatically
	installActionWorkflow, _, err := m.api.GetInstallActionWorkflowRecentRuns(m.ctx, m.installID, m.actionWorkflowID, nil)
	return installActionWorkflowFetchedMsg{installActionWorkflow: installActionWorkflow, err: err}
}

type latestConfigFetchedMsg struct {
	config *models.AppActionWorkflowConfig
	err    error
}

func (m model) fetchLatestConfigCmd() tea.Msg {
	// This runs in a goroutine automatically
	config, err := m.api.GetActionWorkflowLatestConfig(m.ctx, m.actionWorkflowID)
	return latestConfigFetchedMsg{config: config, err: err}
}
