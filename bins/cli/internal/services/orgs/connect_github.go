package orgs

import (
	"context"
	"fmt"

	"github.com/pkg/browser"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) ConnectGithub(ctx context.Context) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	cfg, err := s.api.GetCLIConfig(ctx)
	if err != nil {
		return ui.PrintError(fmt.Errorf("couldn't get cli config: %w", err))
	}

	url := fmt.Sprintf("%s/api/connect-github?org_id=%s", cfg.DashboardURL, s.cfg.OrgID)
	return browser.OpenURL(url)
}
