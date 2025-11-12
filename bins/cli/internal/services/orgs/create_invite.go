package orgs

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) CreateInvite(ctx context.Context, email string, asJSON bool) error {
	view := ui.NewGetView()
	if email == "" {
		return view.Error(fmt.Errorf("email is required"))
	}

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
