package orgs

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/pkg/browser"
)

func (s *Service) ConnectGithub(ctx context.Context, asJSON bool) error {
	if s.cfg.OrgID == "" {
		return ui.PrintError(ui.ErrOrgNotSet())
	}

	cfg, err := s.api.GetCLIConfig(ctx)
	if err != nil {
		return ui.PrintError(fmt.Errorf("couldn't get cli config: %w", err))
	}

	url := fmt.Sprintf("%s/api/connect-github?org_id=%s", cfg.DashboardURL, s.cfg.OrgID)

	if asJSON {
		ui.PrintJSON(map[string]string{"url": url})
		return nil
	}

	ui.PrintLn("opening GitHub connection flow")
	if err := browser.OpenURL(url); err != nil {
		return err
	}
	return nil
}
