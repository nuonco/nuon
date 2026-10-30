package cctx_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func TestVerifiedCredentialOwnsItsCeiling(t *testing.T) {
	for name, mutate := range map[string]func(*cctx.VerifiedCredential){
		"permission mutation": func(c *cctx.VerifiedCredential) { c.Ceiling[0].Permissions[permissions.InstallWrite] = struct{}{} },
		"scope mutation":      func(c *cctx.VerifiedCredential) { c.Ceiling[0].Scope.ID = "install_other" },
		"slice replacement":   func(c *cctx.VerifiedCredential) { c.Ceiling[0] = permissions.ScopedPermissions{} },
	} {
		t.Run(name, func(t *testing.T) {
			scope := permissions.Scope{Kind: permissions.KindInstall, OrgID: "org_acme", ID: "install_one"}
			input := cctx.VerifiedCredential{
				AccountID: "account_one", TokenID: "token_one", Surface: cctx.SurfacePublicAPI,
				Profile: cctx.ProfileHuman, Audience: permissions.APICtl,
				Ceiling: []permissions.ScopedPermissions{{Scope: scope, Permissions: permissions.PermissionSet{permissions.InstallRead: {}}}},
			}
			ctx := cctx.SetVerifiedCredentialContext(context.Background(), input)
			mutate(&input)
			first, ok := cctx.VerifiedCredentialFromContext(ctx)
			require.True(t, ok)
			require.Equal(t, []permissions.ScopedPermissions{{Scope: scope, Permissions: permissions.PermissionSet{permissions.InstallRead: {}}}}, first.Ceiling)
			mutate(&first)
			second, ok := cctx.VerifiedCredentialFromContext(ctx)
			require.True(t, ok)
			require.Equal(t, []permissions.ScopedPermissions{{Scope: scope, Permissions: permissions.PermissionSet{permissions.InstallRead: {}}}}, second.Ceiling)
		})
	}
}
