package workflow

import (
	"fmt"

	"github.com/pkg/browser"
)

func (m model) openInBrowser() {
	cliCfg, err := m.api.GetCLIConfig(m.ctx)
	if err != nil || cliCfg.DashboardURL == "" {
		return
	}
	dashboardURL := cliCfg.DashboardURL

	var url string
	if m.workflow != nil && m.workflow.OwnerType == "app_branches" {
		appID := m.workflow.Metadata["app_id"]
		branchID := m.workflow.OwnerID
		runID := m.workflow.Metadata["run_id"]
		if appID != "" && branchID != "" && runID != "" {
			url = fmt.Sprintf("%s/%s/apps/%s/branches/%s/runs/%s", dashboardURL, m.cfg.OrgID, appID, branchID, runID)
		}
	}

	if url == "" {
		if m.installID != "" {
			url = fmt.Sprintf("%s/%s/installs/%s/workflows/%s", dashboardURL, m.cfg.OrgID, m.installID, m.workflowID)
		} else {
			url = fmt.Sprintf("%s/%s/workflows/%s", dashboardURL, m.cfg.OrgID, m.workflowID)
		}
	}

	if m.selectedStep != nil {
		url += fmt.Sprintf("?target=%s", m.selectedStep.ID)
	}
	browser.OpenURL(url)
}

func (m *model) openQuickLink() {
	if m.stack != nil && len(m.stack.Versions) > 0 {
		browser.OpenURL(m.stack.Versions[0].QuickLinkURL)
	}
}

func (m *model) openTemplateLink() {
	if m.stack != nil && len(m.stack.Versions) > 0 {
		browser.OpenURL(m.stack.Versions[0].TemplateURL)
	}
}
