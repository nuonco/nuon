package orgs

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) CreateInvite(ctx context.Context, email string, asJSON bool) {
	view := ui.NewGetView()

	invite, err := s.api.CreateOrgInvite(ctx, &models.ServiceCreateOrgInviteRequest{
		Email: email,
	})
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(invite)
		return
	}

	data := [][]string{
		{
			"ID",
			"EMAIL",
			"STATUS",
		},
		{
			invite.ID,
			invite.Email,
			string(invite.Status),
		},
	}
	view.Render(data)
}
