package orgs

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) UpdateUserRole(ctx context.Context, userID, role string, asJSON bool) error {
	if s.cfg.OrgID == "" {
		return ui.PrintError(ui.ErrOrgNotSet())
	}

	view := ui.NewGetView()
	if userID == "" {
		return view.Error(fmt.Errorf("user id is required"))
	}
	if role == "" {
		return view.Error(fmt.Errorf("role is required"))
	}

	roleType := models.AppRoleType(role)
	account, err := s.api.UpdateOrgAccountRole(ctx, userID, &models.ServiceUpdateOrgAccountRoleRequest{
		RoleType: &roleType,
	})
	if err != nil {
		return err
	}

	if asJSON {
		ui.PrintJSON(account)
		return nil
	}

	data := [][]string{
		{"id", account.ID},
		{"email", account.Email},
	}
	view.Render(data)
	return nil
}
