package workflow

import (
	"fmt"
	"strings"

	"github.com/pkg/browser"
)

func (m model) openInBrowser() {
	// TODO)(fd): fix this a bit
	dashboardURL := strings.Replace(m.cfg.APIURL, "api", "app", 1)
	var url string
	if m.installID != "" {
		url = fmt.Sprintf("%s/%s/installs/%s/workflows/%s", dashboardURL, m.cfg.OrgID, m.installID, m.workflowID)
	} else {
		url = fmt.Sprintf("%s/%s/workflows/%s", dashboardURL, m.cfg.OrgID, m.workflowID)
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
