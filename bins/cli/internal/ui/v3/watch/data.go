package watch

import (
	"fmt"

	"charm.land/bubbles/v2/list"
)

func (m *model) getWorkflowItems() []list.Item {
	items := []list.Item{}
	for _, workflow := range m.workflows {
		items = append(items, listWorkflow{workflow: workflow})
	}
	return items
}

func (m *model) handleWorkflowsFetched(msg workflowsFetchedMsg) {
	workflows := msg.workflows
	err := msg.err
	if err != nil {
		m.setLogMessage(fmt.Sprintf("[error] failed to fetch workflows: %s", err), "error")
		m.error = err
		m.loading = false
		return
	}
	if workflows == nil {
		m.setLogMessage("no workflows found", "info")
		m.loading = false
		return
	}

	m.workflows = workflows
	items := m.getWorkflowItems()
	m.workflowsList.SetItems(items)
	m.loading = false

	if m.selectedWorkflow != nil && m.selectedIndex >= 0 && m.selectedIndex < len(items) {
		item := m.workflowsList.Items()[m.selectedIndex]
		m.selectedWorkflow = item.(listWorkflow).Workflow()
		m.populateWorkflowDetailView()
	}
}
