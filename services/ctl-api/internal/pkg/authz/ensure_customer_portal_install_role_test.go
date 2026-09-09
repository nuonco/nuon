package authz

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func TestCustomerPortalInstallRole(t *testing.T) {
	const orgID = "org_one"
	const installID = "install_one"

	role := customerPortalInstallRole(orgID, installID)

	assert.Equal(t, app.RoleTypeCustomerPortal, role.RoleType)
	assert.Equal(t, orgID, role.OrgID.String)
	assert.NotEmpty(t, role.Title)
	assert.NotEmpty(t, role.Description)
	assert.False(t, role.Managed)
	assert.Empty(t, role.Contexts)

	require.Len(t, role.Policies, 1)
	require.Len(t, role.Policies[0].Permissions, 1)
	assert.Equal(t, app.PolicyNameCustomerPortal, role.Policies[0].Name)
	assert.Equal(t, permissions.PermissionAll.ToStrPtr(),
		role.Policies[0].Permissions[permissions.Object(orgID, permissions.KindInstall, installID)])

	set := permissions.Set(permissions.NewSet())
	require.NoError(t, set.Add(role.Policies[0].Permissions))
	require.NoError(t, set.CanPerform(permissions.Object(orgID, permissions.KindInstall, installID), permissions.PermissionRead))
	require.NoError(t, set.CanPerform(permissions.Object(orgID, permissions.KindInstall, installID), permissions.PermissionCreate))
	require.Error(t, set.CanPerform(permissions.Object(orgID, permissions.KindInstall, "install_two"), permissions.PermissionRead))
	require.Error(t, set.CanPerform(orgID, permissions.PermissionRead))
}
