package orgs

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) CreateInvite(ctx context.Context, email, firstName, lastName string, asJSON bool) error {
	view := ui.NewGetView()
	if email == "" {
		return view.Error(fmt.Errorf("email is required"))
	}

	invite, err := s.api.CreateOrgInvite(ctx, &models.ServiceCreateOrgInviteRequest{
		Email:     &email,
		FirstName: firstName,
		LastName:  lastName,
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
			"FIRST NAME",
			"LAST NAME",
			"STATUS",
		},
		{
			invite.ID,
			invite.Email,
			invite.FirstName,
			invite.LastName,
			string(invite.Status),
		},
	}
	view.Render(data)
	return nil
}
