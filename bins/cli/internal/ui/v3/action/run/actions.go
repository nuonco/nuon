package run

import (
	"fmt"

	"github.com/pkg/browser"
)

// openInBrowser opens the run in the browser
func (m *Model) openInBrowser() {
	if m.run == nil {
		m.setLogMessage("No run loaded yet", "error")
		return
	}

	// Build the URL for the run
	// Pattern: https://app.nuon.co/{org_id}/installs/{install_id}/actions/{action_workflow_id}/runs/{run_id}
	dashboardURL := fmt.Sprintf("%s/%s/installs/%s/actions/%s/runs/%s",
		"https://app.nuon.co", // TODO: make this configurable
		m.cfg.OrgID,
		m.installID,
		m.actionWorkflowID,
		m.runID,
	)

	browser.OpenURL(dashboardURL)
	m.setLogMessage("Opened in browser", "success")
}
