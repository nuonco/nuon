package authz

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func TestScopedRolePolicyRunnerJob(t *testing.T) {
	const orgID, jobID, installID = "orgone", "jobone", "installone"
	jobRole := func(grants pgtype.Hstore) *app.Role {
		grants[permissions.APIPermissionKey] = generics.ToPtr(string(permissions.APIRunner))
		return &app.Role{OrgID: generics.NewNullString(orgID), RoleType: app.RoleTypeRunnerJob, Version: app.RoleVersionV2,
			Policies: []app.Policy{{Permissions: grants}}}
	}
	job := permissions.Object(orgID, permissions.KindRunnerJob, jobID)
	install := permissions.Object(orgID, permissions.KindInstall, installID)
	all := permissions.PermissionAll.ToStrPtr()

	withInstall, err := ScopedRolePolicy(jobRole(pgtype.Hstore{job: all, install: all}))
	require.NoError(t, err)
	assert.Equal(t, permissions.APIRunner, withInstall.API())
	doc, err := withInstall.MarshalJSON()
	require.NoError(t, err)
	assert.Contains(t, string(doc), jobID)
	assert.Contains(t, string(doc), installID)

	jobOnly, err := ScopedRolePolicy(jobRole(pgtype.Hstore{job: all}))
	require.NoError(t, err)
	doc, err = jobOnly.MarshalJSON()
	require.NoError(t, err)
	assert.Contains(t, string(doc), jobID)
	assert.NotContains(t, string(doc), installID)

	_, err = ScopedRolePolicy(jobRole(pgtype.Hstore{install: all}))
	assert.ErrorContains(t, err, "requires a resource grant", "the job grant is required")

	_, err = ScopedRolePolicy(jobRole(pgtype.Hstore{job: all, install: all, permissions.Object(orgID, permissions.KindInstall, "installtwo"): all}))
	assert.ErrorContains(t, err, "unexpected policy shape", "a job covers at most one install")

	_, err = ScopedRolePolicy(jobRole(pgtype.Hstore{job: all, install: permissions.PermissionCreate.ToStrPtr()}))
	assert.ErrorContains(t, err, "unexpected resource grant")

	_, err = ScopedRolePolicy(jobRole(pgtype.Hstore{job: all, permissions.Object(orgID, permissions.KindStack, installID): all}))
	assert.ErrorContains(t, err, "unexpected policy shape", "grants outside the role type are rejected")
}
