package roles

import (
	"context"
	"strings"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) ListRoles(ctx context.Context, asJSON bool) error {
	if s.cfg.OrgID == "" {
		return ui.PrintError(ui.ErrOrgNotSet())
	}

	view := ui.NewListView()

	roles, err := s.api.ListRoles(ctx)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(roles)
		return nil
	}

	data := [][]string{
		{
			"ROLE TYPE",
			"TITLE",
			"APPLIES TO",
			"DESCRIPTION",
		},
	}

	for _, r := range roles {
		data = append(data, []string{
			string(r.RoleType),
			r.Title,
			strings.Join(r.AppliesTo, ", "),
			r.Description,
		})
	}

	view.Render(data)
	return nil
}
