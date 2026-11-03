package authz

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func TestStandardOrgRoles(t *testing.T) {
	const orgID = "org_one"

	roles, err := standardOrgRoles(orgID)
	require.NoError(t, err)

	seen := map[app.RoleType]app.Role{}
	for _, role := range roles {
		seen[role.RoleType] = role
	}

	require.Len(t, seen, len(roles), "duplicate role types in standardOrgRoles")
	require.NotContains(t, seen, app.RoleTypeOrgBuilder, "org_builder is deprecated and must not be created for new orgs")

	for _, roleType := range []app.RoleType{
		app.RoleTypeOrgAdmin,
		app.RoleTypeOrgReadOnly,
		"org_editor", "app_reader", "app_editor", "app_admin",
		"install_reader", "install_editor", "install_admin",
	} {
		require.Contains(t, seen, roleType)
	}

	for _, role := range roles {
		require.Equal(t, app.RoleVersionV3, role.Version, "role %s", role.RoleType)
		require.Len(t, role.Policies, 1, "role %s", role.RoleType)
		require.Empty(t, role.Policies[0].Permissions, "role %s must not carry legacy permissions", role.RoleType)
		require.Equal(t, orgID, role.Policies[0].OrgID.ValueString())
		policy, err := permissions.DecodePolicy(role.Version, role.Policies[0].Document)
		require.NoError(t, err)
		require.Equal(t, permissions.APICtl, policy.API())
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
	require.NotContains(t, seen, app.RoleTypeRunner)
	require.NotContains(t, seen, app.RoleTypeOrgSupport)
	require.NotContains(t, seen, app.RoleTypeInstaller)

	readOnly := seen[app.RoleTypeOrgReadOnly]
	policy, err := permissions.DecodePolicy(readOnly.Version, readOnly.Policies[0].Document)
	require.NoError(t, err)
	grants, err := policy.Compile(orgID, permissions.Scope{Kind: permissions.KindOrg, OrgID: orgID, ID: orgID})
	require.NoError(t, err)
	require.Contains(t, grants[0].Permissions, permissions.OrgRead)
	require.NotContains(t, grants[0].Permissions, permissions.AppWrite)
}
