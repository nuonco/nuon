package serviceaccounts

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) CreateServiceAccount(ctx context.Context, name, role string, asJSON bool) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	view := ui.NewGetView()
	if name == "" {
		return view.Error(fmt.Errorf("name is required"))
	}
	if role == "" {
		return view.Error(fmt.Errorf("role is required"))
	}

	account, err := s.api.CreateServiceAccount(ctx, &models.ServiceCreateServiceAccountRequest{
		Name: &name,
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
		{"name", account.Name},
		{"email", account.Email},
		{"role", role_},
	})
	return nil
}
