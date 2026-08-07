package authz

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func TestStandardOrgRoles(t *testing.T) {
	const orgID = "org_one"

	roles := standardOrgRoles(orgID)

	seen := map[app.RoleType]app.Role{}
	for _, role := range roles {
		seen[role.RoleType] = role
	}

	require.Len(t, seen, len(roles), "duplicate role types in standardOrgRoles")
	require.NotContains(t, seen, app.RoleTypeOrgBuilder, "org_builder is deprecated and must not be created for new orgs")

	for _, roleType := range []app.RoleType{
		app.RoleTypeOrgAdmin,
		app.RoleTypeOrgSupport,
		app.RoleTypeOrgReadOnly,
		app.RoleTypeInstaller,
		app.RoleTypeRunner,
	} {
		require.Contains(t, seen, roleType)
	}

	for _, role := range roles {
		require.Len(t, role.Policies, 1, "role %s", role.RoleType)
		require.Len(t, role.Policies[0].Permissions, 1, "role %s policies must only carry the bare org key", role.RoleType)
		require.NotNil(t, role.Policies[0].Permissions[orgID], "role %s must key its permission on the org ID", role.RoleType)
		require.NotEmpty(t, role.Title, "role %s must carry display metadata", role.RoleType)
		require.NotEmpty(t, role.Description, "role %s must carry display metadata", role.RoleType)
		require.True(t, role.Managed, "standard roles are managed")
	}

	fullContexts := []string{
		app.RoleContextTeam,
		app.RoleContextServiceAccount,
		app.RoleContextAPIToken,
		app.RoleContextTrustPolicy,
	}
	require.ElementsMatch(t, fullContexts, seen[app.RoleTypeOrgAdmin].Contexts)
	require.ElementsMatch(t, fullContexts, seen[app.RoleTypeOrgReadOnly].Contexts)
	require.ElementsMatch(t, []string{app.RoleContextServiceAccount}, seen[app.RoleTypeRunner].Contexts)
	require.Empty(t, seen[app.RoleTypeOrgSupport].Contexts, "org_support is held-only")
	require.Empty(t, seen[app.RoleTypeInstaller].Contexts, "installer is held-only")

	readOnly := seen[app.RoleTypeOrgReadOnly]
	set := permissions.Set(permissions.NewSet())
	require.NoError(t, set.Add(readOnly.Policies[0].Permissions))
	require.NoError(t, set.CanPerform(orgID, permissions.PermissionRead))
	require.Error(t, set.CanPerform(orgID, permissions.PermissionCreate))
}
