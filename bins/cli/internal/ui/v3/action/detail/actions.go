package detail

import (
	"fmt"

	"github.com/pkg/browser"
	"golang.design/x/clipboard"
)

func (m Model) openInBrowser() {
	if m.installActionWorkflow == nil {
		m.setLogMessage("no action workflow loaded", "warning")
		return
	}

	// Construct dashboard URL
	// Pattern: https://app.nuon.co/installs/{install_id}/actions/{action_workflow_id}
	dashboardURL := fmt.Sprintf("%s/%s/installs/%s/actions/%s",
		"https://app.nuon.co", // TODO: make this configurable
		m.cfg.OrgID,
		m.installID,
		m.actionWorkflowID,
	)

	browser.OpenURL(dashboardURL)
}

func (m *Model) copyActionWorkflowID() {
	clipboard.Write(clipboard.FmtText, []byte(m.actionWorkflowID))
	m.setLogMessage("action workflow id copied to clipboard", "info")
}
