package orgs

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	helpers "github.com/powertoolsdev/mono/bins/cli/internal"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) ListInvites(ctx context.Context, limit int, asJSON bool) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	view := ui.NewGetView()

	invites, err := s.listInvites(ctx, limit)
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
	view.Render(data)
	return nil
}

func (s *Service) listInvites(ctx context.Context, limit int) ([]*models.AppOrgInvite, error) {
	if !s.cfg.PaginationEnabled {
		invites, _, err := s.api.GetOrgInvites(ctx, &models.GetOrgInvitesQuery{
			Offset:            0,
			Limit:             limit,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, err
		}
		return invites, nil
	}

	fetchFn := func(ctx context.Context, offset, limit int) ([]*models.AppOrgInvite, bool, error) {
		invites, hasMore, err := s.api.GetOrgInvites(ctx, &models.GetOrgInvitesQuery{
			Offset:            offset,
			Limit:             limit,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, false, err
		}
		return invites, hasMore, nil
	}

	return helpers.BatchFetch(ctx, 10, limit, fetchFn)
}
