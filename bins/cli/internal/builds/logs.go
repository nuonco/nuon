package builds

import (
	"context"
	"fmt"

	"github.com/pkg/browser"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Logs(ctx context.Context, appID, compID, buildID string, asJSON bool) error {
	if appID == "" {
		s.printAppNotSetMsg()
		return nil
	}

	compID, err := lookup.ComponentID(ctx, s.api, appID, compID)
	if err != nil {
		return ui.PrintError(err)
	}

	cfg, err := s.api.GetCLIConfig(ctx)
	if err != nil {
		return ui.PrintError(fmt.Errorf("couldn't get cli config: %w", err))
	}

	url := fmt.Sprintf("%s/%s/apps/%s/components/%s/builds/%s", cfg.DashboardURL, s.cfg.OrgID, appID, compID, buildID)
	browser.OpenURL(url)
	return nil
}
