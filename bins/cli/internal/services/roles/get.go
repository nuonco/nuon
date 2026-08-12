package roles

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) GetRole(ctx context.Context, roleID string, asJSON bool) error {
	if s.cfg.OrgID == "" {
		s.printOrgNotSetMsg()
		return nil
	}

	view := ui.NewGetView()

	role, err := s.api.GetRole(ctx, roleID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(role)
		return nil
	}

	view.Render(roleDetail(role))
	return nil
}

func roleDetail(role *models.AppRole) [][]string {
	rows := [][]string{
		{"id", role.ID},
		{"title", role.Title},
		{"description", role.Description},
		{"managed", strconv.FormatBool(role.Managed)},
		{"assign with", AssignmentIdentifier(role)},
		{"applies to", strings.Join(role.AppliesTo, ", ")},
	}

	for _, entry := range permissionEntries(role) {
		rows = append(rows, []string{"permission", renderPermissionEntry(entry)})
	}

	return rows
}

func permissionEntries(role *models.AppRole) []*models.AppPermissionEntry {
	var entries []*models.AppPermissionEntry
	for _, policy := range role.Policies {
		entries = append(entries, policy.ScopedPermissions...)
	}
	return entries
}

// renderPermissionEntry prints an entry in the same grammar --permission
// accepts, so a role can be read back and edited without translation.
func renderPermissionEntry(entry *models.AppPermissionEntry) string {
	verbs := slices.Clone(entry.Permissions)
	sortVerbs(verbs)

	out := fmt.Sprintf("%s:%s:%s", strings.Join(verbs, ","), entry.ResourceType, entry.ResourceID)
	if entry.ScopeID != "" {
		out += fmt.Sprintf(":scope=%s", entry.ScopeID)
	}
	return out
}

// sortVerbs orders verbs as create,read,update,delete rather than
// alphabetically, matching how the API and dashboard present them.
func sortVerbs(verbs []string) {
	order := map[string]int{"all": 0, "create": 1, "read": 2, "update": 3, "delete": 4}
	for i := 1; i < len(verbs); i++ {
		for j := i; j > 0 && order[verbs[j]] < order[verbs[j-1]]; j-- {
			verbs[j], verbs[j-1] = verbs[j-1], verbs[j]
		}
	}
}
