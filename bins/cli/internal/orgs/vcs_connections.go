package orgs

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) VCSConnections(ctx context.Context, offset, limit int, asJSON bool) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	view := ui.NewListView()

	vcs, hasMore, err := s.listVCSConnections(ctx, offset, limit)
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

	view.RenderPaging(data, offset, limit, hasMore)
	return nil
}

func (s *Service) listVCSConnections(ctx context.Context, offset, limit int) ([]*models.AppVCSConnection, bool, error) {
	o, hasMore, err := s.api.GetVCSConnections(ctx, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, hasMore, err
	}
	return o, hasMore, nil
}
