package authz

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func TestStackInstallRole(t *testing.T) {
	const orgID = "org_one"
	const installID = "install_one"

	role := stackInstallRole(orgID, installID)

	assert.Equal(t, app.RoleTypeStack, role.RoleType)
	assert.Equal(t, orgID, role.OrgID.String)
	assert.NotEmpty(t, role.Title)
	assert.NotEmpty(t, role.Description)
	assert.False(t, role.Managed, "per-install roles are not reconciled by standardOrgRoles")

	// Contextless keeps it out of every role picker: it is held, never assigned.
	assert.Empty(t, role.Contexts)

	require.Len(t, role.Policies, 1, "policies carry a unique index on role_id")
	require.Len(t, role.Policies[0].Permissions, 1)
	assert.Equal(t, app.PolicyNameStack, role.Policies[0].Name)
	assert.Equal(t, permissions.PermissionAll.ToStrPtr(),
		role.Policies[0].Permissions[permissions.StackObject(orgID, installID)])

	set := permissions.Set(permissions.NewSet())
	require.NoError(t, set.Add(role.Policies[0].Permissions))

	// Every verb on its own install and nothing more: not another install, not the
	// org. Create is what the phone-home route declares.
	require.NoError(t, set.CanPerform(permissions.StackObject(orgID, installID), permissions.PermissionRead))
	require.NoError(t, set.CanPerform(permissions.StackObject(orgID, installID), permissions.PermissionCreate))
	require.Error(t, set.CanPerform(permissions.StackObject(orgID, "install_two"), permissions.PermissionRead))
	require.Error(t, set.CanPerform(permissions.StackObject(orgID, "install_two"), permissions.PermissionCreate))
	require.Error(t, set.CanPerform(orgID, permissions.PermissionRead))
}
