package serviceaccounts

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) DeleteServiceAccount(ctx context.Context, accountID string, asJSON bool) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	view := ui.NewGetView()
	if accountID == "" {
		return view.Error(fmt.Errorf("id is required"))
	}

	if err := s.api.DeleteServiceAccount(ctx, accountID); err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(map[string]string{"id": accountID, "status": "deleted"})
		return nil
	}

	view.Render([][]string{
		{"id", accountID},
		{"status", "deleted"},
	})
	return nil
}
