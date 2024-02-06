package orgs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s *Service) ListHealthChecks(ctx context.Context, limit int64, asJSON bool) {
	view := ui.NewGetView()

	healthChecks, err := s.api.GetOrgHealthChecks(ctx, generics.ToPtr(limit))
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(healthChecks)
		return
	}

	data := [][]string{
		{
			"created at",
			"status",
			"description",
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
}
