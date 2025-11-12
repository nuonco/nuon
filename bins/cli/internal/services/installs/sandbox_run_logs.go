package installs

import (
	"context"
	"fmt"

	"github.com/pkg/browser"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) SandboxRunLogs(ctx context.Context, installID, runID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}
	cfg, err := s.api.GetCLIConfig(ctx)
	if err != nil {
		return ui.PrintError(fmt.Errorf("couldn't get cli config: %w", err))
	}

	url := fmt.Sprintf("%s/%s/installs/%s/runs/%s", cfg.DashboardURL, s.cfg.OrgID, installID, runID)
	browser.OpenURL(url)
	return nil
}
