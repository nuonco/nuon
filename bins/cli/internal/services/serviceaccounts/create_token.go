package serviceaccounts

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) CreateServiceAccountToken(ctx context.Context, accountID, duration string, invalidate, asJSON bool) error {
	if s.cfg.OrgID == "" {
		return ui.PrintError(ui.ErrOrgNotSet())
	}

	view := ui.NewGetView()
	if accountID == "" {
		return view.Error(fmt.Errorf("id is required"))
	}

	resp, err := s.api.CreateServiceAccountToken(ctx, accountID, &models.ServiceCreateServiceAccountTokenRequest{
		Duration:   &duration,
		Invalidate: invalidate,
	})
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(resp)
		return nil
	}

	view.Render([][]string{
		{"token", resp.Token},
	})
	ui.Println("\nCopy this token now. For security, it won't be shown again.")
	return nil
}
