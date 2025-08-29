package orgs

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) ListInvites(ctx context.Context, offset, limit int, asJSON bool) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	view := ui.NewListView()

	invites, hasMore, err := s.listInvites(ctx, offset, limit)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(invites)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"EMAIL",
			"STATUS",
		},
	}

	for _, invite := range invites {
		data = append(data, []string{
			invite.ID,
			invite.Email,
			string(invite.Status),
		})
	}
	view.RenderPaging(data, offset, limit, hasMore)
	return nil
}

func (s *Service) listInvites(ctx context.Context, offset, limit int) ([]*models.AppOrgInvite, bool, error) {
	invites, hasMore, err := s.api.GetOrgInvites(ctx, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, hasMore, err
	}
	return invites, hasMore, nil
}
