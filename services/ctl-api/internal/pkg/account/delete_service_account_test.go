// Integration tests: run with INTEGRATION=true against the migrated test database.
//
// In the external test package on purpose — the shared tests package imports
// internal/pkg/account, so an in-package test here would be an import cycle. That
// also means everything is exercised through the exported DeleteServiceAccount,
// FindAccount lookup included, rather than through the unexported record helper.
package account_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type deleteServiceAccountDeps struct {
	fx.In

	DB     *gorm.DB `name:"psql"`
	Client *account.Client
	Seeder *testseed.Seeder
}

type DeleteServiceAccountTestSuite struct {
	tests.BaseDBTestSuite

	fxApp *fxtest.App
	deps  deleteServiceAccountDeps
}

func TestDeleteServiceAccountSuite(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	suite.Run(t, new(DeleteServiceAccountTestSuite))
}

func (s *DeleteServiceAccountTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	options := append(
		tests.CtlApiFXOptions(s.T()),
		fx.Populate(&s.deps),
	)
	s.fxApp = fxtest.New(s.T(), options...)
	s.fxApp.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *DeleteServiceAccountTestSuite) TearDownSuite() {
	s.fxApp.RequireStop()
}

// seedRoleBinding attaches the account to a freshly created role. The role has to be
// real: account_roles carries a foreign key to roles in Postgres, unlike the sqlite
// version of this test, which could invent a role ID.
func (s *DeleteServiceAccountTestSuite) seedRoleBinding(ctx context.Context, acct *app.Account) {
	t := s.T()
	t.Helper()

	// Role.BeforeCreate fills CreatedByID from the account ID in the context.
	creatorCtx, _ := s.deps.Seeder.EnsureAccount(ctx, t)
	role := &app.Role{RoleType: app.RoleTypeOrgAdmin}
	require.NoError(t, s.deps.DB.WithContext(creatorCtx).Create(role).Error)

	binding := &app.AccountRole{AccountID: acct.ID, RoleID: role.ID}
	require.NoError(t, s.deps.DB.WithContext(creatorCtx).Create(binding).Error)
}

// A leftover role binding or token is the whole reason this exists: the account is
// tied to its stack only by a naming convention, so nothing else will ever collect
// it, and what survives is an org-admin credential with no owner.
func (s *DeleteServiceAccountTestSuite) TestDeleteRemovesCredentialRecords() {
	t := s.T()
	ctx := context.Background()
	future := time.Now().Add(time.Hour)

	stackID := generics.GetFakeObj[string]()
	acct := s.deps.Seeder.CreateServiceAccount(ctx, t, stackID)
	s.seedRoleBinding(ctx, acct)
	s.seedRoleBinding(ctx, acct)
	s.deps.Seeder.CreateToken(ctx, t, acct, future)

	// A second account proves the deletes are scoped rather than table-wide.
	otherStackID := generics.GetFakeObj[string]()
	otherAcct := s.deps.Seeder.CreateServiceAccount(ctx, t, otherStackID)
	s.seedRoleBinding(ctx, otherAcct)
	s.deps.Seeder.CreateToken(ctx, t, otherAcct, future)

	require.NoError(t, s.deps.Client.DeleteServiceAccount(ctx, stackID))

	// Role bindings are hard-deleted: the many2many declares OnDelete:CASCADE, which
	// is a foreign-key constraint that a soft delete would never fire.
	var roleRows int64
	require.NoError(t, s.deps.DB.Unscoped().Model(&app.AccountRole{}).
		Where("account_id = ?", acct.ID).Count(&roleRows).Error)
	assert.Zero(t, roleRows, "role bindings must be hard-deleted, not soft-deleted")

	// Tokens and the account are soft-deleted, which is enough: the auth middleware
	// resolves a token through FindAccount, and that cannot see a soft-deleted row.
	var tokens []app.Token
	require.NoError(t, s.deps.DB.Where("account_id = ?", acct.ID).Find(&tokens).Error)
	assert.Empty(t, tokens, "tokens must be invisible to a normal query")

	var gone app.Account
	err := s.deps.DB.Where("id = ?", acct.ID).First(&gone).Error
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound, "the account must be invisible to a normal query")

	var stillThere int64
	require.NoError(t, s.deps.DB.Unscoped().Model(&app.AccountRole{}).
		Where("account_id = ?", otherAcct.ID).Count(&stillThere).Error)
	assert.EqualValues(t, 1, stillThere, "another account's bindings must survive")

	var otherTokens []app.Token
	require.NoError(t, s.deps.DB.Where("account_id = ?", otherAcct.ID).Find(&otherTokens).Error)
	assert.Len(t, otherTokens, 1, "another account's tokens must survive")
}

// Delete workflows retry, and a stack that never had a service account must not be
// able to block its own install from being torn down.
func (s *DeleteServiceAccountTestSuite) TestDeleteIsIdempotent() {
	t := s.T()
	ctx := context.Background()

	require.NoError(t, s.deps.Client.DeleteServiceAccount(ctx, generics.GetFakeObj[string]()))

	stackID := generics.GetFakeObj[string]()
	acct := s.deps.Seeder.CreateServiceAccount(ctx, t, stackID)
	s.deps.Seeder.CreateToken(ctx, t, acct, time.Now().Add(time.Hour))

	require.NoError(t, s.deps.Client.DeleteServiceAccount(ctx, stackID))
	require.NoError(t, s.deps.Client.DeleteServiceAccount(ctx, stackID),
		"a second delete must be a no-op, not an error")
}

// FindAccount matches on email, subject, or ID, so a caller passing something
// unexpected could otherwise reach a real user through a path meant for machine
// identities.
func (s *DeleteServiceAccountTestSuite) TestDeleteRejectsNonServiceAccount() {
	t := s.T()
	ctx := context.Background()

	// A human account squatting on the service-account email convention.
	stackID := generics.GetFakeObj[string]()
	human := testseed.BuildAccount()
	human.Email = account.ServiceAccountEmail(stackID)
	require.NoError(t, s.deps.DB.WithContext(ctx).Create(human).Error)

	err := s.deps.Client.DeleteServiceAccount(ctx, stackID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a service account")

	var survivor app.Account
	require.NoError(t, s.deps.DB.Where("id = ?", human.ID).First(&survivor).Error,
		"the account must survive a rejected delete")
}
