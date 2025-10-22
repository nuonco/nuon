package run

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/action/run/steps"
)

// fetchInstallActionWorkflowRunCmd fetches the workflow run data
func (m Model) fetchInstallActionWorkflowRunCmd() tea.Msg {
	// This runs in a goroutine automatically
	run, err := m.api.GetInstallActionWorkflowRun(m.ctx, m.installID, m.runID)
	return installActionWorkflowRunFetchedMsg{run: run, err: err}
}

// handleInstallActionWorkflowRunFetched handles the fetched run data
func (m *Model) handleInstallActionWorkflowRunFetched(msg installActionWorkflowRunFetchedMsg) {
	run := msg.run
	err := msg.err
	if err != nil {
		m.setLogMessage(fmt.Sprintf("[error] failed to fetch workflow run: %s", err), "error")
		m.error = err
		return
	} else if run == nil {
		m.setLogMessage("no workflow run data returned", "error")
		return
	}

	// we only want to create the steps component once, otherwise, we'll keep making a fresh one
	// and we'll lose logs
	if m.run == nil { // NOTE(fd): so we only create the component the first time we fetch the run
		m.run = run
		m.stepsView = steps.New(m.ctx, m.api, m.stepsWidth, m.stepsHeight, m.run)
	} else {
		m.run = run
		m.stepsView.SetRun(m.run)
	}

	m.loading = false
	m.setLogMessage("install action workflow run refreshed", "info")
	m.setContent() // sets sidebar content
}
