package roles

import (
	"context"
	"strings"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) ListRoles(ctx context.Context, asJSON bool) error {
	if s.cfg.OrgID == "" {
		return ui.PrintError(ui.ErrOrgNotSet())
	}

	view := ui.NewListView()

	allRoles, err := s.api.ListRoles(ctx)
	if err != nil {
		return err
	}

	// Only list roles that can be assigned somewhere; held-only roles
	// (deprecated or machine-only) carry no assignment contexts.
	roles := make([]*models.AppRole, 0, len(allRoles))
	for _, r := range allRoles {
		if len(r.AppliesTo) > 0 {
			roles = append(roles, r)
		}
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
