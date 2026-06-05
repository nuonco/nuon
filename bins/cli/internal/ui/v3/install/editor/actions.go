package editor

import (
	"fmt"

	"github.com/pkg/browser"
)

func (m *model) openInBrowser() {
	cfg, err := m.api.GetCLIConfig(m.ctx)
	if err != nil {
		m.setLogMessage("Could not get dashboard URL", "error")
		return
	}

	// Pattern: https://app.nuon.co/{org_id}/installs/{install_id}
	dashboardURL := fmt.Sprintf("%s/%s/installs/%s",
		cfg.DashboardURL,
		m.cfg.OrgID,
		m.installID,
	)

	browser.OpenURL(dashboardURL)
	m.setLogMessage("Opening in browser...", "info")
}
