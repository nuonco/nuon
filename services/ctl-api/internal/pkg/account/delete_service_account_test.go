// Integration tests: run with INTEGRATION=true against the migrated test database.
//
// External test package on purpose: the shared tests package imports
// internal/pkg/account, so an in-package test would be an import cycle. Everything is
// therefore exercised through the exported Client methods.
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
	"github.com/nuonco/nuon/pkg/shortid/domains"
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

// seedRoleBinding attaches the account to a real role — account_roles has a
// foreign key to roles.
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

// seedStackRole gives the account the install-scoped role a real stack account holds.
func (s *DeleteServiceAccountTestSuite) seedStackRole(ctx context.Context, orgID string, acct *app.Account) *app.Role {
	t := s.T()
	t.Helper()

	creatorCtx, _ := s.deps.Seeder.EnsureAccount(ctx, t)
	role := &app.Role{
		OrgID:    generics.NewNullString(orgID),
		RoleType: app.RoleTypeStack,
		Policies: []app.Policy{{OrgID: generics.NewNullString(orgID), Name: app.PolicyNameStack}},
	}
	require.NoError(t, s.deps.DB.WithContext(creatorCtx).Create(role).Error)

	binding := &app.AccountRole{OrgID: generics.NewNullString(orgID), AccountID: acct.ID, RoleID: role.ID}
	require.NoError(t, s.deps.DB.WithContext(creatorCtx).Create(binding).Error)

	return role
}

// A soft delete would keep the unique policy-per-role index occupied.
func (s *DeleteServiceAccountTestSuite) TestDeleteReapsStackRoles() {
	t := s.T()
	ctx := context.Background()

	org := s.deps.Seeder.CreateOrg(ctx, t)

	stackID := generics.GetFakeObj[string]()
	acct := s.deps.Seeder.CreateServiceAccount(ctx, t, stackID)
	role := s.seedStackRole(ctx, org.ID, acct)

	// A managed org role bound to the same account must survive the reap.
	s.seedRoleBinding(ctx, acct)

	otherAcct := s.deps.Seeder.CreateServiceAccount(ctx, t, generics.GetFakeObj[string]())
	otherRole := s.seedStackRole(ctx, org.ID, otherAcct)

	require.NoError(t, s.deps.Client.DeleteServiceAccount(ctx, stackID))

	var roles int64
	require.NoError(t, s.deps.DB.Unscoped().Model(&app.Role{}).
		Where("id = ?", role.ID).Count(&roles).Error)
	assert.Zero(t, roles, "the stack role must be hard-deleted")

	var policies int64
	require.NoError(t, s.deps.DB.Unscoped().Model(&app.Policy{}).
		Where("role_id = ?", role.ID).Count(&policies).Error)
	assert.Zero(t, policies, "the stack role's policy must be hard-deleted")

	require.NoError(t, s.deps.DB.Unscoped().Model(&app.Role{}).
		Where("id = ?", otherRole.ID).Count(&roles).Error)
	assert.EqualValues(t, 1, roles, "another account's stack role must survive")
}

// The account is tied to its stack only by naming convention, so a leftover
// binding or token is what this guards against.
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

	// Hard-deleted: OnDelete:CASCADE is a constraint a soft delete never fires.
	var roleRows int64
	require.NoError(t, s.deps.DB.Unscoped().Model(&app.AccountRole{}).
		Where("account_id = ?", acct.ID).Count(&roleRows).Error)
	assert.Zero(t, roleRows, "role bindings must be hard-deleted, not soft-deleted")

	// Soft-deleting these is enough: FindAccount cannot see a soft-deleted row.
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

// Delete workflows retry, and a stack with no account must not block teardown.
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
// unexpected could otherwise reach a real user.
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

func (s *DeleteServiceAccountTestSuite) seedLogStream(ctx context.Context, orgID string, open bool, age time.Duration) string {
	t := s.T()
	t.Helper()

	creatorCtx, _ := s.deps.Seeder.EnsureAccount(ctx, t)
	stream := &app.LogStream{ID: domains.NewLogStreamID(), OrgID: orgID, Open: open}
	require.NoError(t, s.deps.DB.WithContext(creatorCtx).Create(stream).Error)
	require.NoError(t, s.deps.DB.Model(&app.LogStream{}).Where("id = ?", stream.ID).
		UpdateColumn("created_at", time.Now().Add(-age)).Error)
	return stream.ID
}

func (s *DeleteServiceAccountTestSuite) seedIdleLogStreamAccount(ctx context.Context, streamID string) *app.Account {
	t := s.T()
	t.Helper()

	acct := s.deps.Seeder.CreateServiceAccount(ctx, t, streamID)
	require.NoError(t, s.deps.DB.Model(&app.Account{}).Where("id = ?", acct.ID).
		UpdateColumn("updated_at", time.Now().Add(-3*time.Hour)).Error)
	return acct
}

func (s *DeleteServiceAccountTestSuite) seedPrivateRole(ctx context.Context, orgID string, managed bool, accts ...*app.Account) *app.Role {
	t := s.T()
	t.Helper()

	creatorCtx, _ := s.deps.Seeder.EnsureAccount(ctx, t)
	role := &app.Role{OrgID: generics.NewNullString(orgID), RoleType: app.RoleTypeLogStream, Managed: managed}
	require.NoError(t, s.deps.DB.WithContext(creatorCtx).Create(role).Error)
	for _, acct := range accts {
		binding := &app.AccountRole{OrgID: generics.NewNullString(orgID), AccountID: acct.ID, RoleID: role.ID}
		require.NoError(t, s.deps.DB.WithContext(creatorCtx).Create(binding).Error)
	}
	return role
}

func (s *DeleteServiceAccountTestSuite) sweepLogStreamAccounts(ctx context.Context, dryRun bool) *account.SweepLogStreamAccountsResult {
	t := s.T()
	t.Helper()

	now := time.Now()
	res, err := s.deps.Client.SweepLogStreamAccounts(ctx, account.SweepLogStreamAccountsRequest{
		Limit:            1000,
		IdleBefore:       now.Add(-2 * time.Hour),
		OpenStreamBefore: now.Add(-24 * time.Hour),
		DryRun:           dryRun,
	})
	require.NoError(t, err)
	return res
}

func (s *DeleteServiceAccountTestSuite) accountLive(id string) bool {
	var n int64
	require.NoError(s.T(), s.deps.DB.Model(&app.Account{}).Where("id = ?", id).Count(&n).Error)
	return n == 1
}

func (s *DeleteServiceAccountTestSuite) roleExists(id string) bool {
	var n int64
	require.NoError(s.T(), s.deps.DB.Unscoped().Model(&app.Role{}).Where("id = ?", id).Count(&n).Error)
	return n == 1
}

func (s *DeleteServiceAccountTestSuite) TestSweepDeletesIdleLogStreamAccounts() {
	t := s.T()
	ctx := context.Background()
	org := s.deps.Seeder.CreateOrg(ctx, t)

	closed := s.seedIdleLogStreamAccount(ctx, s.seedLogStream(ctx, org.ID, false, time.Hour))
	missing := s.seedIdleLogStreamAccount(ctx, domains.NewLogStreamID())
	oldOpenStream := s.seedLogStream(ctx, org.ID, true, 25*time.Hour)
	oldOpen := s.seedIdleLogStreamAccount(ctx, oldOpenStream)
	closedRole := s.seedPrivateRole(ctx, org.ID, false, closed)
	missingRole := s.seedPrivateRole(ctx, org.ID, false, missing)
	oldOpenRole := s.seedPrivateRole(ctx, org.ID, false, oldOpen)

	sharedA := s.seedIdleLogStreamAccount(ctx, s.seedLogStream(ctx, org.ID, false, time.Hour))
	sharedB := s.seedIdleLogStreamAccount(ctx, s.seedLogStream(ctx, org.ID, false, time.Hour))
	sharedRole := s.seedPrivateRole(ctx, org.ID, false, sharedA, sharedB)

	youngOpen := s.seedIdleLogStreamAccount(ctx, s.seedLogStream(ctx, org.ID, true, 3*time.Hour))
	recent := s.deps.Seeder.CreateServiceAccount(ctx, t, domains.NewLogStreamID())
	withToken := s.seedIdleLogStreamAccount(ctx, s.seedLogStream(ctx, org.ID, false, time.Hour))
	s.deps.Seeder.CreateToken(ctx, t, withToken, time.Now().Add(time.Hour))
	managed := s.seedIdleLogStreamAccount(ctx, s.seedLogStream(ctx, org.ID, false, time.Hour))
	creatorCtx, _ := s.deps.Seeder.EnsureAccount(ctx, t)
	require.NoError(t, s.deps.DB.WithContext(creatorCtx).Create(&app.ManagedServiceAccount{
		AccountID:   managed.ID,
		OrgID:       org.ID,
		OwnerType:   "runner_jobs",
		OwnerID:     generics.GetFakeObj[string](),
		Purpose:     app.ManagedServiceAccountPurposeRunnerJob,
		InstanceKey: "default",
	}).Error)

	deleted := []*app.Account{closed, missing, oldOpen, sharedA, sharedB}
	kept := []*app.Account{youngOpen, recent, withToken, managed}

	dry := s.sweepLogStreamAccounts(ctx, true)
	assert.GreaterOrEqual(t, dry.Selected, len(deleted))
	assert.Zero(t, dry.Deleted)
	for _, acct := range deleted {
		assert.True(t, s.accountLive(acct.ID), "dry-run must not delete %s", acct.Subject)
	}

	s.sweepLogStreamAccounts(ctx, false)

	for _, acct := range deleted {
		assert.False(t, s.accountLive(acct.ID), "idle account %s must be deleted", acct.Subject)
	}
	for _, acct := range kept {
		assert.True(t, s.accountLive(acct.ID), "account %s must be kept", acct.Subject)
	}
	for _, role := range []*app.Role{closedRole, missingRole, oldOpenRole, sharedRole} {
		assert.False(t, s.roleExists(role.ID), "private role %s must be deleted", role.ID)
	}

	var stream app.LogStream
	require.NoError(t, s.deps.DB.Where("id = ?", oldOpenStream).First(&stream).Error)
	assert.True(t, stream.Open, "the sweep must not close streams")
}

func (s *DeleteServiceAccountTestSuite) TestSweepSkipsAccountsBlockedByASharedRole() {
	t := s.T()
	ctx := context.Background()
	org := s.deps.Seeder.CreateOrg(ctx, t)

	sharesWithBlocked := s.seedIdleLogStreamAccount(ctx, domains.NewLogStreamID())
	blocked := s.seedIdleLogStreamAccount(ctx, domains.NewLogStreamID())
	sharesWithOutsider := s.seedIdleLogStreamAccount(ctx, domains.NewLogStreamID())
	outsider := s.deps.Seeder.CreateServiceAccount(ctx, t, generics.GetFakeObj[string]())
	unrelated := s.seedIdleLogStreamAccount(ctx, domains.NewLogStreamID())
	sharedRole := s.seedPrivateRole(ctx, org.ID, false, sharesWithBlocked, blocked)
	s.seedPrivateRole(ctx, org.ID, true, blocked)
	outsiderRole := s.seedPrivateRole(ctx, org.ID, false, sharesWithOutsider, outsider)
	s.seedPrivateRole(ctx, org.ID, false, unrelated)

	res := s.sweepLogStreamAccounts(ctx, false)

	assert.GreaterOrEqual(t, res.Skipped, 3)
	assert.True(t, s.accountLive(sharesWithOutsider.ID), "an account sharing a role outside the batch must be skipped")
	assert.True(t, s.roleExists(outsiderRole.ID))
	assert.True(t, s.accountLive(blocked.ID), "an account holding a managed role must be skipped")
	assert.True(t, s.accountLive(sharesWithBlocked.ID), "an account sharing a role with a skipped account must be skipped")
	assert.True(t, s.roleExists(sharedRole.ID))
	assert.False(t, s.accountLive(unrelated.ID), "a skipped account must not stop the rest of the batch")
}

func (s *DeleteServiceAccountTestSuite) TestDeleteRejectsPrivateRoleSharedWithAnotherAccount() {
	t := s.T()
	ctx := context.Background()
	org := s.deps.Seeder.CreateOrg(ctx, t)

	stackID := generics.GetFakeObj[string]()
	acct := s.deps.Seeder.CreateServiceAccount(ctx, t, stackID)
	other := s.deps.Seeder.CreateServiceAccount(ctx, t, generics.GetFakeObj[string]())
	role := s.seedPrivateRole(ctx, org.ID, false, acct, other)

	require.Error(t, s.deps.Client.DeleteServiceAccount(ctx, stackID))

	assert.True(t, s.accountLive(acct.ID), "the account must survive a rejected delete")
	assert.True(t, s.roleExists(role.ID), "a shared private role must survive")
	var bindings int64
	require.NoError(t, s.deps.DB.Model(&app.AccountRole{}).Where("role_id = ?", role.ID).Count(&bindings).Error)
	assert.EqualValues(t, 2, bindings, "both bindings must survive a rejected delete")
}
