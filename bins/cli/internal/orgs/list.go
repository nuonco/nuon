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

	curID := s.cfg.GetString("org_id")

	data := [][]string{
		{
			" NAME",
			"ID",
			"STATUS",
			"SANDBOX MODE",
			"UPDATED AT",
		},
	}

	for _, org := range orgs {
		if curID != "" {
			if org.ID == curID {
				org.Name = "*" + org.Name
			} else {
				org.Name = " " + org.Name
			}
		}
		data = append(data, []string{
			org.Name,
			org.ID,
			org.StatusDescription,
			strconv.FormatBool(org.SandboxMode),
			org.UpdatedAt,
		})
	}
	view.Render(data)
	return nil
}
