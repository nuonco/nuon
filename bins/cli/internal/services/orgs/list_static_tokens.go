package orgs

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) ListStaticTokens(ctx context.Context, asJSON bool) error {
	if s.cfg.OrgID == "" {
		return ui.PrintError(ui.ErrOrgNotSet())
	}

	view := ui.NewListView()

	tokens, err := s.api.ListStaticTokens(ctx)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(tokens)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"NAME",
			"ROLE",
			"EXPIRES AT",
			"CREATED AT",
			"CREATED BY",
		},
	}

	for _, t := range tokens {
		data = append(data, []string{
			t.ID,
			t.Name,
			t.Role,
			t.ExpiresAt,
			t.CreatedAt,
			t.CreatedByID,
		})
	}

	view.Render(data)
	return nil
}
