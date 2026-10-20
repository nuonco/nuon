package authz

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func TestLogStreamWriteRole(t *testing.T) {
	const orgID, streamID = "org_one", "log_one"
	role := logStreamWriteRole(orgID, streamID)

	assert.Equal(t, app.RoleTypeLogStream, role.RoleType)
	assert.False(t, role.Managed)
	assert.Empty(t, role.Contexts)

	set := permissions.Set(permissions.NewSet())
	require.NoError(t, set.Add(role.Policies[0].Permissions))

	stream := permissions.Object(orgID, permissions.KindLogStream, streamID)
	assert.NoError(t, set.CanPerform(stream, permissions.PermissionCreate))
	assert.Error(t, set.CanPerform(stream, permissions.PermissionRead))
	assert.Error(t, set.CanPerform(permissions.Object(orgID, permissions.KindLogStream, "log_other"), permissions.PermissionCreate))
	assert.Error(t, set.CanPerform(orgID, permissions.PermissionCreate))
}

func TestRunnerOrgGrantCanWriteLogStream(t *testing.T) {
	const orgID, streamID = "org_one", "log_one"
	set := permissions.Set(permissions.NewSet())
	require.NoError(t, set.Add(map[string]*string{orgID: permissions.PermissionAll.ToStrPtr()}))

	assert.NoError(t, set.CanPerform(permissions.Object(orgID, permissions.KindLogStream, streamID), permissions.PermissionCreate))
}
