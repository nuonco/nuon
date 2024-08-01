package orgs

import (
	"context"
	"strconv"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, asJSON bool) error {
	view := ui.NewGetView()

	orgs, err := s.api.GetOrgs(ctx)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(orgs)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"NAME",
			"STATUS",
			"SANDBOX MODE",
			"UPDATED AT",
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
	return nil
}
