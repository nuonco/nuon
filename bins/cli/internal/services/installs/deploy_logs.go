package installs

import (
	"context"
	"fmt"

	"github.com/pkg/browser"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	logs "github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/logs"
)

func (s *Service) DeployLogs(ctx context.Context, installID, deployID, installComponentID string, asJSON bool) error {

	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	// fetch deploy
	deploy, err := s.api.GetInstallDeploy(ctx, installID, deployID)
	if err != nil {
		return ui.PrintError(err)
	}

	if !s.cfg.Preview {
		// open in browser
		cfg, err := s.api.GetCLIConfig(ctx)
		if err != nil {
			ui.PrintError(err)
		}
		url := fmt.Sprintf("%s/%s/installs/%s/components/%s/deploys/%s", cfg.DashboardURL, s.cfg.OrgID, installID, installComponentID, deployID)
		browser.OpenURL(url)
	} else {
		// open in tui
		logs.LogStreamApp(ctx, s.cfg, s.api, installID, deployID, deploy.LogStream.ID)
	}

	return nil
}
