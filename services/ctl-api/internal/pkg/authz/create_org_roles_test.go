package authz

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func TestStandardOrgRolesBuilderPermissions(t *testing.T) {
	const orgID = "org_one"

	var builder *app.Role
	roles := standardOrgRoles(orgID)
	for i := range roles {
		role := roles[i]
		if role.RoleType == app.RoleTypeOrgBuilder {
			builder = &role
			break
		}
	}

	require.NotNil(t, builder)
	require.Len(t, builder.Policies, 1)
	require.Equal(t, app.PolicyNameOrgBuilder, builder.Policies[0].Name)
	require.Equal(t, permissions.PermissionRead.ToStrPtr(), builder.Policies[0].Permissions[orgID])
	require.Equal(t, permissions.PermissionCreate.ToStrPtr(), builder.Policies[0].Permissions[permissions.ComponentBuildsObject(orgID)])
	require.Len(t, builder.Policies[0].Permissions, 2)

	set := permissions.Set(permissions.NewSet())
	require.NoError(t, set.Add(builder.Policies[0].Permissions))
	require.NoError(t, set.CanPerform(orgID, permissions.PermissionRead))
	require.NoError(t, set.CanPerform(permissions.ComponentBuildsObject(orgID), permissions.PermissionCreate))
	require.Error(t, set.CanPerform(orgID, permissions.PermissionCreate))
	require.Error(t, set.CanPerform(permissions.ComponentBuildsObject(orgID), permissions.PermissionUpdate))
	require.Error(t, set.CanPerform(permissions.ComponentBuildsObject("org_two"), permissions.PermissionCreate))
}
