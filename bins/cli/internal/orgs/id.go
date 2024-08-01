package orgs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) ID(ctx context.Context, asJSON bool) error {
	view := ui.NewGetView()

	if asJSON {
		ui.PrintJSON(s.cfg.OrgID)
		return nil
	}

	view.Render([][]string{
		{"id", s.cfg.OrgID},
	})
	return nil
}
