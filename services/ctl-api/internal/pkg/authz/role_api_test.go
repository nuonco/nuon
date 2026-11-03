package authz_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func TestLegacyAuthorizationRejectsNewRoleFormats(t *testing.T) {
	for name, tc := range map[string]struct {
		version int
		enforce bool
		deny    bool
	}{
		"v1 remains legacy":                         {app.RoleVersionV1, false, false},
		"v2 retains API eligibility":                {app.RoleVersionV2, true, false},
		"v3 cannot fall back to hstore":             {app.RoleVersionV3, true, true},
		"unknown format cannot fall back to hstore": {4, true, true},
	} {
		t.Run(name, func(t *testing.T) {
			role := app.Role{Version: tc.version, Policies: []app.Policy{{
				Permissions: pgtype.Hstore{
					"org_acme":                   permissions.PermissionAll.ToStrPtr(),
					permissions.APIPermissionKey: generics.ToPtr(string(permissions.APICtl)),
				},
				Document: []byte(`{"kind":"bundle","api":"ctl-api","binding_kind":"org","permissions":["org.read"]}`),
			}}}
			api, enforce, err := authz.APIForRoles([]app.Role{role})
			require.Equal(t, tc.enforce, enforce)
			account := app.Account{Roles: []app.Role{role}}
			require.NoError(t, account.AfterQuery(nil))
			grantErr := account.AllPermissions.CanPerform("org_acme", permissions.PermissionRead)
			if tc.deny {
				require.Error(t, err)
				require.Empty(t, api)
				require.Error(t, grantErr)
				return
			}
			require.NoError(t, err)
			require.NoError(t, grantErr)
			if tc.enforce {
				require.Equal(t, permissions.APICtl, api)
			}
		})
	}
}

func TestScopedBindingRequiresExplicitTenant(t *testing.T) {
	for name, tc := range map[string]struct {
		binding app.AccountRole
		deny    bool
	}{
		"scoped binding cannot borrow request tenant": {
			app.AccountRole{ScopeKind: permissions.KindApp, ScopeID: "app_one"}, true,
		},
		"missing scope kind": {
			app.AccountRole{OrgID: generics.NewNullString("org_acme"), ScopeID: "app_one"}, true,
		},
		"missing scope ID": {
			app.AccountRole{OrgID: generics.NewNullString("org_acme"), ScopeKind: permissions.KindApp}, true,
		},
		"exact workload resource is not a customer binding": {
			app.AccountRole{OrgID: generics.NewNullString("org_acme"), ScopeKind: permissions.KindRunnerJob, ScopeID: "job_one"}, true,
		},
		"explicit tenant remains independent of request tenant": {
			app.AccountRole{OrgID: generics.NewNullString("org_acme"), ScopeKind: permissions.KindApp, ScopeID: "app_one"}, false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctx := cctx.SetOrgIDContext(context.Background(), "org_other")
			err := tc.binding.BeforeSave(&gorm.DB{Statement: &gorm.Statement{Context: ctx}})
			if tc.deny {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, "org_acme", tc.binding.OrgID.ValueString())
		})
	}
}
