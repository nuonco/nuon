package orgs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, asJSON bool) {
	view := ui.NewGetView()

	orgs, err := s.api.GetOrgs(ctx)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(orgs)
		return
	}

	data := [][]string{
		{
			"id",
			"name",
			"status",
			"updated at",
		},
	}

	for _, org := range orgs {
		data = append(data, []string{
			*&org.ID,
			org.Name,
			org.StatusDescription,
			org.UpdatedAt,
		})
	}
	view.Render(data)
}
