package orgs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) APIToken(ctx context.Context, asJSON bool) {
	view := ui.NewGetView()

	if asJSON {
		ui.PrintJSON(s.cfg.APIToken)
		return
	}

	view.Render([][]string{
		{"api-token", s.cfg.APIToken},
	})
}
