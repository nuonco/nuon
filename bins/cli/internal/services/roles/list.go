package roles

import (
	"context"
	"strings"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// AssignmentIdentifier returns what --role expects for the role: its type for
// managed roles, its id for custom ones, which all share the "custom" type.
// Mirrors app.Role.AssignmentIdentifier.
func AssignmentIdentifier(role *models.AppRole) string {
	if role.RoleType == models.AppRoleTypeCustom {
		return role.ID
	}
	return string(role.RoleType)
}

func (s *Service) ListRoles(ctx context.Context, asJSON bool) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	view := ui.NewListView()

	allRoles, err := s.api.ListRoles(ctx)
	if err != nil {
		return view.Error(err)
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
			"ROLE",
			"TITLE",
			"APPLIES TO",
			"DESCRIPTION",
		},
	}

	for _, r := range roles {
		data = append(data, []string{
			AssignmentIdentifier(r),
			r.Title,
			strings.Join(r.AppliesTo, ", "),
			r.Description,
		})
	}

	view.Render(data)
	return nil
}
