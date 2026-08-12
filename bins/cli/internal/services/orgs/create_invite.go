package orgs

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) CreateInvite(ctx context.Context, email, role string, asJSON bool) error {
	view := ui.NewGetView()
	if email == "" {
		return view.Error(fmt.Errorf("email is required"))
	}

	req := &models.ServiceCreateOrgInviteRequest{
		Email: &email,
	}
	if role != "" {
		req.RoleType = models.AppRoleType(role)
	}

	invite, err := s.api.CreateOrgInvite(ctx, req)
	if err != nil {
		return err
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
