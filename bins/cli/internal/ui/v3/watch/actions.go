package watch

import (
	"fmt"
	"strings"

	"github.com/pkg/browser"
)

func (m *model) openWorkflowInBrowser() {
	if m.selectedWorkflow == nil {
		return
	}

	dashboardURL := strings.Replace(m.cfg.APIURL, "api", "app", 1)
	url := fmt.Sprintf("%s/%s/installs/%s/workflows/%s", dashboardURL, m.cfg.OrgID, m.installID, m.selectedWorkflow.ID)
	if err := browser.OpenURL(url); err != nil {
		m.setLogMessage("failed to open browser: "+err.Error(), "error")
		return
	}
	m.setLogMessage("opened in browser", "success")
}
