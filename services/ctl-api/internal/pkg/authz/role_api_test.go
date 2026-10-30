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

func TestCompatibleRoleAPIEligibility(t *testing.T) {
	for name, tc := range map[string]struct {
		version int
		mutate  func(*app.Role)
		deny    bool
	}{
		"v1 preserves untagged eligibility": {version: 1},
		"v2 preserves ctl-api eligibility":  {version: 2},
		"missing v2 tag cannot become v1": {version: 2, deny: true, mutate: func(r *app.Role) {
			delete(r.Policies[0].Permissions, "api")
		}},
		"wrong v2 API": {version: 2, deny: true, mutate: func(r *app.Role) {
			r.Policies[0].Permissions["api"] = generics.ToPtr("runner")
		}},
		"v1 cannot acquire a tag": {version: 1, deny: true, mutate: func(r *app.Role) {
			r.Policies[0].Permissions["api"] = generics.ToPtr("ctl-api")
		}},
		"no provenance":      {version: 0, deny: true},
		"unknown provenance": {version: 4, deny: true},
		"broader hstore": {version: 1, deny: true, mutate: func(r *app.Role) {
			r.Policies[0].Permissions["org_acme"] = permissions.PermissionAll.ToStrPtr()
		}},
		"extra wildcard": {version: 1, deny: true, mutate: func(r *app.Role) {
			r.Policies[0].Permissions["*"] = permissions.PermissionAll.ToStrPtr()
		}},
		"wrong policy tenant": {version: 1, deny: true, mutate: func(r *app.Role) {
			r.Policies[0].OrgID = generics.NewNullString("org_other")
		}},
		"missing structured policy": {version: 1, deny: true, mutate: func(r *app.Role) {
			r.Policies[0].Document = nil
		}},
		"narrow document is not legacy org read": {version: 1, deny: true, mutate: func(r *app.Role) {
			r.Policies[0].Document = []byte(`{"kind":"bundle","api":"ctl-api","binding_kind":"install","permissions":["install.read"]}`)
		}},
		"changed bundle cannot keep broad legacy meaning": {version: 1, deny: true, mutate: func(r *app.Role) {
			r.Policies[0].Document = []byte(`{"kind":"bundle","api":"ctl-api","binding_kind":"org","permissions":["org.read"]}`)
		}},
	} {
		t.Run(name, func(t *testing.T) {
			role := app.Role{
				ID: "role_reader", OrgID: generics.NewNullString("org_acme"), RoleType: app.RoleTypeOrgReadOnly,
				Version: 3, LegacyVersion: tc.version, CatalogRevision: 1, Managed: true,
				Policies: []app.Policy{{RoleID: "role_reader", OrgID: generics.NewNullString("org_acme"),
					Permissions: pgtype.Hstore{"org_acme": permissions.PermissionRead.ToStrPtr()},
					Document:    []byte(`{"kind":"bundle","api":"ctl-api","binding_kind":"org","permissions":["org.read","app.read","install.read","runner.read","runner_job.read","telemetry.read","log_stream.read"]}`),
				}},
			}
			if tc.version == 2 {
				role.Policies[0].Permissions["api"] = generics.ToPtr("ctl-api")
			}
			if tc.mutate != nil {
				tc.mutate(&role)
			}
			api, enforce, err := authz.APIForRoles([]app.Role{role})
			if tc.deny {
				require.Error(t, err)
				require.True(t, enforce)
				require.Empty(t, api)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.version == 2, enforce)
			if tc.version == 2 {
				require.Equal(t, permissions.APICtl, api)
			} else {
				require.Empty(t, api)
			}
			binding := app.AccountRole{RoleID: role.ID, Role: role, OrgID: role.OrgID, ScopeKind: permissions.KindOrg, ScopeID: "org_acme"}
			require.NoError(t, binding.ValidateLegacyCompatibility())
			binding.ScopeKind, binding.ScopeID = permissions.KindInstall, "install_one"
			require.Error(t, binding.ValidateLegacyCompatibility())
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
