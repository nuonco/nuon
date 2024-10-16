package installs

import (
	"context"
	"fmt"

	"github.com/pkg/browser"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) DeployLogs(ctx context.Context, installID, deployID, installComponentID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	cfg, err := s.api.GetCLIConfig(ctx)
	if err != nil {
		return ui.PrintError(fmt.Errorf("couldn't get cli config: %w", err))
	}

	url := fmt.Sprintf("%s/%s/installs/%s/components/%s/deploys/%s", cfg.DashboardURL, s.cfg.OrgID, installID, installComponentID, deployID)
	browser.OpenURL(url)

	return nil
}
