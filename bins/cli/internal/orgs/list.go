package orgs

import (
	"context"
	"strconv"

	"github.com/nuonco/nuon-go/models"
	helpers "github.com/powertoolsdev/mono/bins/cli/internal"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, asJSON bool) error {
	view := ui.NewGetView()

	orgs, err := s.list(ctx)
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

func (s *Service) list(ctx context.Context) ([]*models.AppOrg, error) {
	if !s.cfg.PaginationEnabled {
		o, _, err := s.api.GetOrgs(ctx, &models.GetOrgsQuery{
			Offset:            0,
			Limit:             10,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, err
		}
		return o, nil
	}

	fetchFn := func(ctx context.Context, offset, limit int) ([]*models.AppOrg, bool, error) {
		o, hasMore, err := s.api.GetOrgs(ctx, &models.GetOrgsQuery{
			Offset:            offset,
			Limit:             limit,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, false, err
		}
		return o, hasMore, nil
	}

	return helpers.BatchFetch(ctx, 10, 50, fetchFn)
}
