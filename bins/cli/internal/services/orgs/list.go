package orgs

import (
	"context"
	"strconv"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, offset, limit int, asJSON bool) error {
	view := ui.NewListView()

	orgs, hasMore, err := s.list(ctx, offset, limit)
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
	view.RenderPaging(data, offset, limit, hasMore)
	return nil
}

func (s *Service) list(ctx context.Context, offset, limit int) ([]*models.AppOrg, bool, error) {
	o, hasMore, err := s.api.GetOrgs(ctx, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, hasMore, err
	}
	return o, hasMore, nil
}
