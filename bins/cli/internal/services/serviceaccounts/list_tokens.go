package serviceaccounts

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) ListTokens(ctx context.Context, accountID, search string, offset, limit int, asJSON bool) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	view := ui.NewListView()

	tokens, _, err := s.api.ListTokens(ctx, accountID, search, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
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
			"ACCOUNT",
			"ROLE",
			"EXPIRES AT",
			"CREATED AT",
		},
	}

	for _, t := range tokens {
		account := t.AccountEmail
		if t.AccountName != "" {
			account = t.AccountName
		}

		data = append(data, []string{
			t.ID,
			t.Name,
			account,
			t.AccountRole,
			t.ExpiresAt,
			t.CreatedAt,
		})
	}

	view.Render(data)
	return nil
}
