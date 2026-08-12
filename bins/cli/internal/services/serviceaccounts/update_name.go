package serviceaccounts

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) UpdateServiceAccountName(ctx context.Context, accountID, name string, asJSON bool) error {
	if s.cfg.OrgID == "" {
		return ui.PrintError(ui.ErrOrgNotSet())
	}

	view := ui.NewGetView()
	if accountID == "" {
		return view.Error(fmt.Errorf("id is required"))
	}
	if name == "" {
		return view.Error(fmt.Errorf("name is required"))
	}

	account, err := s.api.UpdateServiceAccount(ctx, accountID, &models.ServiceUpdateServiceAccountRequest{
		Name: &name,
	})
	if err != nil {
		return err
	}

	if asJSON {
		ui.PrintJSON(account)
		return nil
	}

	view.Render([][]string{
		{"id", account.ID},
		{"name", account.Name},
		{"email", account.Email},
	})
	return nil
}
