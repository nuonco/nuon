package orgs

import (
	"context"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) CreateInvite(ctx context.Context, email string, asJSON bool) error {
	view := ui.NewGetView()

	invite, err := s.api.CreateOrgInvite(ctx, &models.ServiceCreateOrgInviteRequest{
		Email: &email,
	})
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(invite)
		return nil
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
	return nil
}
