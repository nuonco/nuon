package orgs

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	helpers "github.com/powertoolsdev/mono/bins/cli/internal"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) VCSConnections(ctx context.Context, asJSON bool) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	view := ui.NewGetView()

	vcs, err := s.listVCSConnections(ctx)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(vcs)
		return nil
	}

	data := [][]string{
		{
			"GITHUB INSTALL ID",
		},
	}

	for _, v := range vcs {
		data = append(data, []string{
			*&v.GithubInstallID,
		})
	}

	view.Render(data)
	return nil
}

func (s *Service) listVCSConnections(ctx context.Context) ([]*models.AppVCSConnection, error) {
	if !s.cfg.PaginationEnabled {
		o, _, err := s.api.GetVCSConnections(ctx, &models.GetVCSConnectionsQuery{
			Offset:            0,
			Limit:             10,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, err
		}
		return o, nil
	}

	fetchFn := func(ctx context.Context, offset, limit int) ([]*models.AppVCSConnection, bool, error) {
		o, hasMore, err := s.api.GetVCSConnections(ctx, &models.GetVCSConnectionsQuery{
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
