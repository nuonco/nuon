// Integration tests: run with INTEGRATION=true against the migrated test database.
//
// External test package on purpose: the shared tests package reaches
// internal/pkg/authz, so an in-package test would be an import cycle.
package authz_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type stackRoleDeps struct {
	fx.In

	DB     *gorm.DB `name:"psql"`
	Client *authz.Client
	Seeder *testseed.Seeder
}

type EnsureStackInstallRoleTestSuite struct {
	tests.BaseDBTestSuite

	fxApp *fxtest.App
	deps  stackRoleDeps
}

func TestEnsureStackInstallRoleSuite(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	suite.Run(t, new(EnsureStackInstallRoleTestSuite))
}

func (s *EnsureStackInstallRoleTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	options := append(
		tests.CtlApiFXOptions(s.T()),
		fx.Populate(&s.deps),
	)
	s.fxApp = fxtest.New(s.T(), options...)
	s.fxApp.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *EnsureStackInstallRoleTestSuite) TearDownSuite() {
	s.fxApp.RequireStop()
}

func (s *EnsureStackInstallRoleTestSuite) rolesFor(accountID string) []app.Role {
	t := s.T()
	t.Helper()

	var roles []app.Role
	require.NoError(t, s.deps.DB.
		Joins("JOIN account_roles ON account_roles.role_id = roles.id AND account_roles.deleted_at = 0").
		Preload("Policies").
		Where("account_roles.account_id = ?", accountID).
		Where("roles.role_type = ?", app.RoleTypeStack).
		Find(&roles).Error)

	return roles
}

// Runs on every provision, so a second call must converge rather than pile up roles.
func (s *EnsureStackInstallRoleTestSuite) TestEnsureIsConvergent() {
	t := s.T()
	ctx := context.Background()

	org := s.deps.Seeder.CreateOrg(ctx, t)
	acct := s.deps.Seeder.CreateServiceAccount(ctx, t, generics.GetFakeObj[string]())
	installID := generics.GetFakeObj[string]()

	// No account in context: Role.CreatedByID is notnull.
	require.NoError(t, s.deps.Client.EnsureStackInstallRole(ctx, org.ID, installID, acct.ID))

	roles := s.rolesFor(acct.ID)
	require.Len(t, roles, 1)
	require.Len(t, roles[0].Policies, 1)
	assert.Equal(t, acct.ID, roles[0].CreatedByID, "the account is its own creator")

	set := permissions.Set(permissions.NewSet())
	require.NoError(t, set.Add(roles[0].Policies[0].Permissions))
	require.NoError(t, set.CanPerform(permissions.StackObject(org.ID, installID), permissions.PermissionRead))
	require.NoError(t, set.CanPerform(permissions.StackObject(org.ID, installID), permissions.PermissionCreate))
	require.Error(t, set.CanPerform(org.ID, permissions.PermissionRead), "the grant must not widen to the org")

	require.NoError(t, s.deps.Client.EnsureStackInstallRole(ctx, org.ID, installID, acct.ID))
	assert.Len(t, s.rolesFor(acct.ID), 1, "a second call must not create a second role")
}

// Roles predating the phone-home route hold `read`, which cannot report.
func (s *EnsureStackInstallRoleTestSuite) TestEnsureUpgradesReadToAll() {
	t := s.T()
	ctx := context.Background()

	org := s.deps.Seeder.CreateOrg(ctx, t)
	acct := s.deps.Seeder.CreateServiceAccount(ctx, t, generics.GetFakeObj[string]())
	installID := generics.GetFakeObj[string]()

	require.NoError(t, s.deps.Client.EnsureStackInstallRole(ctx, org.ID, installID, acct.ID))

	// Rewind to the old grant in place, the way an existing row looks.
	roles := s.rolesFor(acct.ID)
	require.Len(t, roles, 1)
	require.NoError(t, s.deps.DB.
		Model(&app.Policy{}).
		Where("id = ?", roles[0].Policies[0].ID).
		Update("permissions", pgtype.Hstore(map[string]*string{
			permissions.StackObject(org.ID, installID): permissions.PermissionRead.ToStrPtr(),
		})).Error)

	require.NoError(t, s.deps.Client.EnsureStackInstallRole(ctx, org.ID, installID, acct.ID))

	roles = s.rolesFor(acct.ID)
	require.Len(t, roles, 1, "converging must not create a second role")
	require.Len(t, roles[0].Policies, 1)

	set := permissions.Set(permissions.NewSet())
	require.NoError(t, set.Add(roles[0].Policies[0].Permissions))
	require.NoError(t, set.CanPerform(permissions.StackObject(org.ID, installID), permissions.PermissionCreate))
	require.Error(t, set.CanPerform(org.ID, permissions.PermissionRead), "converging must not widen to the org")
}

func (s *EnsureStackInstallRoleTestSuite) boundStackAccount() (orgID, installID, accountID string) {
	t := s.T()
	ctx := context.Background()
	ctx, _ = s.deps.Seeder.EnsureAccount(ctx, t)
	ctx, org := s.deps.Seeder.EnsureOrg(ctx, t)
	a := s.deps.Seeder.CreateApp(ctx, t)
	s.deps.Seeder.CreateAppConfig(ctx, t, a.ID)
	install := s.deps.Seeder.CreateInstall(ctx, t, a)
	stack := install.InstallStack

	acct := s.deps.Seeder.CreateServiceAccount(ctx, t, stack.ID)
	require.NoError(t, s.deps.DB.Create(&app.ManagedServiceAccount{
		AccountID:   acct.ID,
		CreatedByID: acct.ID,
		OrgID:       org.ID,
		OwnerType:   plugins.TableName(s.deps.DB, app.InstallStack{}),
		OwnerID:     stack.ID,
		Purpose:     app.ManagedServiceAccountPurposeStack,
		InstanceKey: "default",
	}).Error)

	return org.ID, install.ID, acct.ID
}

func (s *EnsureStackInstallRoleTestSuite) requireStackAPI(accountID string) app.Role {
	roles := s.rolesFor(accountID)
	require.Len(s.T(), roles, 1)
	api, enforced, err := authz.APIFromRole(&roles[0])
	require.NoError(s.T(), err)
	require.True(s.T(), enforced)
	require.Equal(s.T(), permissions.APIStack, api)
	return roles[0]
}

func (s *EnsureStackInstallRoleTestSuite) TestEnsureKeepsAPIForBoundAccount() {
	t := s.T()
	ctx := context.Background()
	orgID, installID, accountID := s.boundStackAccount()

	require.NoError(t, s.deps.Client.EnsureStackInstallRole(ctx, orgID, installID, accountID))
	first := s.requireStackAPI(accountID)

	require.NoError(t, s.deps.Client.EnsureStackInstallRole(ctx, orgID, installID, accountID))
	second := s.requireStackAPI(accountID)
	assert.Equal(t, first.ID, second.ID)

	set := permissions.Set(permissions.NewSet())
	require.NoError(t, set.Add(second.Policies[0].Permissions))
	require.NoError(t, set.CanPerform(permissions.StackObject(orgID, installID), permissions.PermissionCreate))
	require.Error(t, set.CanPerform(orgID, permissions.PermissionRead))
}

func (s *EnsureStackInstallRoleTestSuite) TestEnsureRestoresStrippedAPIForBoundAccount() {
	t := s.T()
	ctx := context.Background()
	orgID, installID, accountID := s.boundStackAccount()

	require.NoError(t, s.deps.Client.EnsureStackInstallRole(ctx, orgID, installID, accountID))
	role := s.requireStackAPI(accountID)
	require.NoError(t, s.deps.DB.
		Model(&app.Policy{}).
		Where("id = ?", role.Policies[0].ID).
		Update("permissions", pgtype.Hstore(map[string]*string{
			permissions.StackObject(orgID, installID): permissions.PermissionAll.ToStrPtr(),
		})).Error)

	require.NoError(t, s.deps.Client.EnsureStackInstallRole(ctx, orgID, installID, accountID))
	s.requireStackAPI(accountID)
}
