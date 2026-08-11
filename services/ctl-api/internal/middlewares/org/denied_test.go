package org

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func TestPermissionDeniedError(t *testing.T) {
	orgOne := &app.Org{ID: "org_one"}
	orgTwo := &app.Org{ID: "org_two"}

	acct := &app.Account{
		Roles: []app.Role{
			{Org: orgOne, RoleType: app.RoleTypeOrgReadOnly, Title: "Read-only"},
			{Org: orgTwo, RoleType: app.RoleTypeOrgAdmin, Title: "Admin"},
		},
	}

	err := permissionDeniedError(acct, "org_one", permissions.PermissionCreate, "installs in this organization")
	require.Equal(t, "this action requires write access to installs in this organization", err.Error())
	require.Equal(t,
		"Your role (Read-only) does not have write access to installs in this organization. Ask an organization admin to assign a role that does.",
		err.Description)

	err = permissionDeniedError(acct, "org_one", permissions.PermissionRead, "resources in this organization")
	require.Equal(t, "this action requires read access to resources in this organization", err.Error())

	multi := &app.Account{
		Roles: []app.Role{
			{Org: orgOne, RoleType: app.RoleTypeOrgReadOnly, Title: "Read-only"},
			{Org: orgOne, RoleType: app.RoleTypeInstaller},
		},
	}
	err = permissionDeniedError(multi, "org_one", permissions.PermissionDelete, "resources in this organization")
	require.Equal(t,
		"Your roles (Read-only, installer) do not have write access to resources in this organization. Ask an organization admin to assign a role that does.",
		err.Description)

	scopes := map[string]string{
		"/v1/service-accounts":                       "service accounts in this organization",
		"/v1/runner-jobs/:runner_job_id/cancel":      "runner jobs in this organization",
		"/v1/orgs/current":                           "organization settings",
		"/v1/orgs/current/webhooks":                  "webhooks in this organization",
		"/v1/orgs/current/webhooks/:webhook_id":      "webhooks in this organization",
		"/v1/orgs/current/invites":                   "team members in this organization",
		"/v1/orgs/current/invites/:invite_id/revoke": "team members in this organization",
		"/v1/orgs/current/accounts/:account_id/role": "team members in this organization",
		"/v1/orgs/current/remove-user":               "team members in this organization",
		"/v1/orgs/current/features":                  "organization settings",
		"/v1/vcs/connections":                        "VCS connections in this organization",
		"/v1/oidc/trust-policies":                    "OIDC federation in this organization",
		"/v1/account/static-token":                   "API tokens in this organization",
		"/v1/account/static-tokens/:token_id":        "API tokens in this organization",
		"/v1/account":                                "your account",
		"/v1/account/user-journeys":                  "your account",
		"":                                           "this organization",
	}
	for path, want := range scopes {
		require.Equal(t, want, scopeFromPath(path), "path %q", path)
	}

	noRoles := &app.Account{}
	err = permissionDeniedError(noRoles, "org_one", permissions.PermissionUpdate, "apps in this organization")
	require.Equal(t,
		"Ask an organization admin to assign you a role with write access to apps in this organization.",
		err.Description)
}
