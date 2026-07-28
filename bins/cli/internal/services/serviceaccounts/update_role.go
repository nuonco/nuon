package serviceaccounts

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) UpdateServiceAccountRole(ctx context.Context, accountID, role string, asJSON bool) error {
	if s.cfg.OrgID == "" {
		return ui.PrintError(ui.ErrOrgNotSet())
	}

	view := ui.NewGetView()
	if accountID == "" {
		return view.Error(fmt.Errorf("id is required"))
	}
	if role == "" {
		return view.Error(fmt.Errorf("role is required"))
	}

	account, err := s.api.UpdateServiceAccountRole(ctx, accountID, &models.ServiceUpdateServiceAccountRoleRequest{
		Role: &role,
	})
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(account)
		return nil
	}

	role_ := ""
	if len(account.Roles) > 0 {
		role_ = string(account.Roles[0].RoleType)
	}

	view.Render([][]string{
		{"id", account.ID},
		{"email", account.Email},
		{"role", role_},
	})
	return nil
}
