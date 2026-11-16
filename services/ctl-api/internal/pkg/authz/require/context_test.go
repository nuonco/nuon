package require

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

func TestOrg(t *testing.T) {
	_, err := Org(context.Background())
	assert.Error(t, err)

	ctx := context.WithValue(context.Background(), keys.OrgIDCtxKey, "org_abc")
	orgID, err := Org(ctx)
	require.NoError(t, err)
	assert.Equal(t, "org_abc", orgID)
}

func TestWriteOAuthReadOnly(t *testing.T) {
	acct := &app.Account{
		ID: "acct_1",
		Authority: &permissions.AccountPermissions{
			Sources: map[string]permissions.PermissionSource{"org_abc": {Permission: permissions.PermissionAll}},
		},
	}
	ctx := cctx.SetAccountContext(context.Background(), acct)
	ctx = context.WithValue(ctx, keys.OrgIDCtxKey, "org_abc")
	ctx = keys.WithTokenRole(ctx, string(app.RoleTypeOrgReadOnly))

	_, err := Write(ctx)
	assert.Error(t, err)
}

func TestWriteRBAC(t *testing.T) {
	acct := &app.Account{
		ID: "acct_1",
		Authority: &permissions.AccountPermissions{
			Sources: map[string]permissions.PermissionSource{"org_abc": {Permission: permissions.PermissionRead}},
		},
	}
	ctx := cctx.SetAccountContext(context.Background(), acct)
	ctx = context.WithValue(ctx, keys.OrgIDCtxKey, "org_abc")

	_, err := Write(ctx)
	assert.Error(t, err)

	acct.Authority.Sources["org_abc"] = permissions.PermissionSource{Permission: permissions.PermissionCreate}
	_, err = Write(ctx)
	assert.NoError(t, err)
}

func TestReadRBAC(t *testing.T) {
	acct := &app.Account{
		ID: "acct_1",
		Authority: &permissions.AccountPermissions{
			Sources: map[string]permissions.PermissionSource{"org_abc": {Permission: permissions.PermissionRead}},
		},
	}
	ctx := cctx.SetAccountContext(context.Background(), acct)
	ctx = context.WithValue(ctx, keys.OrgIDCtxKey, "org_abc")

	orgID, err := Read(ctx)
	require.NoError(t, err)
	assert.Equal(t, "org_abc", orgID)

	acct.Authority.Sources["org_abc"] = permissions.PermissionSource{Permission: permissions.PermissionDelete}
	_, err = Read(ctx)
	assert.Error(t, err)
}
