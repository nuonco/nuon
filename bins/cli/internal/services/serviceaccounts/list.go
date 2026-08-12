package serviceaccounts

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) ListServiceAccounts(ctx context.Context, includeRunners bool, offset, limit int, asJSON bool) error {
	if s.cfg.OrgID == "" {
		return ui.PrintError(ui.ErrOrgNotSet())
	}

	view := ui.NewListView()

	accounts, _, err := s.api.ListServiceAccounts(ctx, includeRunners, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(accounts)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"NAME",
			"EMAIL",
			"ROLE",
			"CREATED AT",
		},
	}

	for _, a := range accounts {
		role := ""
		if len(a.Roles) > 0 {
			role = string(a.Roles[0].RoleType)
		}
		data = append(data, []string{
			a.ID,
			a.Name,
			a.Email,
			role,
			a.CreatedAt,
		})
	}

	view.Render(data)
	return nil
}
