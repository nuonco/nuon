package orgs

import (
	"context"
	"strconv"

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
			"sandbox mode",
			"updated at",
		},
	}

	for _, org := range orgs {
		data = append(data, []string{
			*&org.ID,
			org.Name,
			org.StatusDescription,
			strconv.FormatBool(org.SandboxMode),
			org.UpdatedAt,
		})
	}
	view.Render(data)
}
