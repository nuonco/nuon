package orgs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s *Service) ListHealthChecks(ctx context.Context, limit int64, asJSON bool) error {
	view := ui.NewGetView()

	healthChecks, err := s.api.GetOrgHealthChecks(ctx, generics.ToPtr(limit))
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(healthChecks)
		return nil
	}

	data := [][]string{
		{
			"CREATED AT",
			"STATUS",
			"DESCRIPTION",
		},
	}

	for _, hc := range healthChecks {
		data = append(data, []string{
			hc.CreatedAt,
			string(hc.Status),
			hc.StatusDescription,
		})
	}
	view.Render(data)
	return nil
}
