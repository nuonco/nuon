package orgs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) ListInvites(ctx context.Context, limit int64, asJSON bool) {
	view := ui.NewGetView()

	invites, err := s.api.GetOrgInvites(ctx, &limit)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(invites)
		return
	}

	data := [][]string{
		{
			"id",
			"email",
			"status",
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
}
