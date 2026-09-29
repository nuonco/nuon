package creator

import (
	"fmt"

	"github.com/pkg/browser"
)

func (m *model) openInBrowser() {
	if !m.success || m.installID == "" {
		m.setLogMessage("No install created yet", "warning")
		return
	}

	cfg, err := m.api.GetCLIConfig(m.ctx)
	if err != nil {
		m.setLogMessage("Could not get dashboard URL", "error")
		return
	}

	dashboardURL := fmt.Sprintf("%s/%s/installs/%s",
		cfg.DashboardURL,
		m.cfg.OrgID,
		m.installID,
	)

	browser.OpenURL(dashboardURL)
	m.setLogMessage("Opening in browser...", "info")
}
