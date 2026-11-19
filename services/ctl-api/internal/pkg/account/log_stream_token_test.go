package account_test

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

// A log stream account from before log streams had bindings keeps the API
// restriction of the role it already holds when it is adopted.
func (s *DeleteServiceAccountTestSuite) TestLogStreamAdoptionKeepsAPIRestriction() {
	t := s.T()
	ctx := context.Background()

	creatorCtx, creator := s.deps.Seeder.EnsureAccount(ctx, t)
	org := s.deps.Seeder.CreateOrg(ctx, t)
	stream := &app.LogStream{ID: domains.NewLogStreamID(), OrgID: org.ID, CreatedByID: creator.ID, Open: true}
	require.NoError(t, s.deps.DB.WithContext(creatorCtx).Create(stream).Error)

	acct := s.deps.Seeder.CreateServiceAccount(ctx, t, stream.ID)
	role := &app.Role{
		OrgID: generics.NewNullString(org.ID), RoleType: app.RoleTypeLogStream, Version: app.RoleVersionV2,
		Policies: []app.Policy{{OrgID: generics.NewNullString(org.ID), Name: app.PolicyNameLogStream, Permissions: pgtype.Hstore{
			permissions.APIPermissionKey:                                     generics.ToPtr(string(permissions.APICtl)),
			permissions.Object(org.ID, permissions.KindLogStream, stream.ID): permissions.PermissionCreate.ToStrPtr(),
		}}},
	}
	require.NoError(t, s.deps.DB.WithContext(creatorCtx).Create(role).Error)
	require.NoError(t, s.deps.DB.WithContext(creatorCtx).Create(&app.AccountRole{
		OrgID: generics.NewNullString(org.ID), AccountID: acct.ID, RoleID: role.ID,
	}).Error)

	_, err := s.deps.Client.IssueLogStreamToken(ctx, org.ID, stream.ID, time.Hour)
	require.NoError(t, err)

	var binding app.ManagedServiceAccount
	require.NoError(t, s.deps.DB.Where(app.ManagedServiceAccount{AccountID: acct.ID}).Take(&binding).Error)
	require.NotNil(t, binding.PrivateRoleID)
	assert.Equal(t, role.ID, *binding.PrivateRoleID, "the existing role must be adopted")

	var policy app.Policy
	require.NoError(t, s.deps.DB.Where(app.Policy{RoleID: role.ID}).Take(&policy).Error)
	api := policy.Permissions[permissions.APIPermissionKey]
	require.NotNil(t, api)
	assert.Equal(t, string(permissions.APICtl), *api, "adoption must not change the role's API restriction")
}
