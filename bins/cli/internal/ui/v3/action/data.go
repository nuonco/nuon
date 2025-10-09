package action

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
)

func (m *model) handleInstallActionWorkflowFetched(msg installActionWorkflowFetchedMsg) {
	installActionWorkflow := msg.installActionWorkflow
	err := msg.err
	if err != nil {
		m.setLogMessage(fmt.Sprintf("[error] failed to fetch action workflow: %s", err), "error")
		m.error = err
		return
	} else if installActionWorkflow == nil {
		m.setLogMessage("no action workflow data returned", "error")
		return
	}

	m.installActionWorkflow = installActionWorkflow
	m.workflowLoading = false

	// Populate the runs list
	runsList := []list.Item{}
	if installActionWorkflow.Runs != nil {
		for _, run := range installActionWorkflow.Runs {
			runItem := listRun{run: run, name: installActionWorkflow.ActionWorkflow.Name}
			runsList = append(runsList, runItem)
		}
	}
	m.runsList.SetItems(runsList)
	m.loading = false
}

func (m *model) handleLatestConfigFetched(msg latestConfigFetchedMsg) {
	config := msg.config
	err := msg.err
	if err != nil {
		m.setLogMessage(fmt.Sprintf("[error] failed to fetch latest config: %s", err), "error")
		return
	}

	m.latestConfig = config
	m.configLoading = false
	m.populateActionConfigView(true)
}
