package helpers_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type migrateRolesDeps struct {
	fx.In

	DB      *gorm.DB `name:"psql"`
	Seed    *testseed.Seeder
	Helpers *installhelpers.Helpers
}

type MigrateInstallRolesTestSuite struct {
	tests.BaseDBTestSuite

	app  *fxtest.App
	deps migrateRolesDeps
}

func TestMigrateInstallRolesSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(MigrateInstallRolesTestSuite))
}

func (s *MigrateInstallRolesTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	options := append(tests.CtlApiFXOptionsWithMocks(tests.TestOpts{T: s.T()}), fx.Populate(&s.deps))
	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *MigrateInstallRolesTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *MigrateInstallRolesTestSuite) seedPermissions(ctx context.Context, appID, appConfigID string, roles map[string]app.AWSIAMRoleType) *app.AppPermissionsConfig {
	cfg := &app.AppPermissionsConfig{AppID: appID, AppConfigID: appConfigID}
	for name, typ := range roles {
		cfg.Roles = append(cfg.Roles, app.AppAWSIAMRoleConfig{
			AppConfigID: appConfigID,
			Type:        typ,
			Name:        name,
			DisplayName: name,
		})
	}
	s.Require().NoError(s.deps.DB.WithContext(ctx).Create(cfg).Error)
	return cfg
}

func (s *MigrateInstallRolesTestSuite) seedProvisionedInstallRoles(ctx context.Context, installID string, perm *app.AppPermissionsConfig) {
	for _, role := range perm.Roles {
		row := app.InstallRoles{
			InstallID:       installID,
			AppRoleConfigID: role.ID,
			Enabled:         true,
			Provisioned:     true,
			RoleID:          "arn:aws:iam::123456789012:role/" + role.Name,
		}
		s.Require().NoError(s.deps.DB.WithContext(ctx).Create(&row).Error)
	}
}

func (s *MigrateInstallRolesTestSuite) liveRolesByName(ctx context.Context, installID string) map[string]app.InstallRoles {
	var rows []app.InstallRoles
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Preload("AppRoleConfig").
		Where(app.InstallRoles{InstallID: installID}).
		Find(&rows).Error)
	byName := make(map[string]app.InstallRoles, len(rows))
	for _, r := range rows {
		byName[r.AppRoleConfig.Name] = r
	}
	return byName
}

func roleConfigID(perm *app.AppPermissionsConfig, name string) string {
	for _, r := range perm.Roles {
		if r.Name == name {
			return r.ID
		}
	}
	return ""
}

var baseRoles = map[string]app.AWSIAMRoleType{
	"provision":   app.AWSIAMRoleTypeRunnerProvision,
	"maintenance": app.AWSIAMRoleTypeRunnerMaintenance,
}

// A sync must not replace the live rows: install_role_usage points at their IDs,
// and the provisioned ARN is state about the customer account, not the config.
func (s *MigrateInstallRolesTestSuite) TestRepointsLiveRolesInPlace() {
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())

	testApp := s.deps.Seed.CreateApp(ctx, s.T())
	oldCfg := s.deps.Seed.CreateAppConfig(ctx, s.T(), testApp.ID)
	install := s.deps.Seed.CreateInstall(ctx, s.T(), testApp)
	oldPerm := s.seedPermissions(ctx, testApp.ID, oldCfg.ID, baseRoles)
	s.seedProvisionedInstallRoles(ctx, install.ID, oldPerm)
	before := s.liveRolesByName(ctx, install.ID)

	newCfg := s.deps.Seed.CreateBareAppConfig(ctx, s.T(), testApp.ID)
	newPerm := s.seedPermissions(ctx, testApp.ID, newCfg.ID, map[string]app.AWSIAMRoleType{
		"provision":   app.AWSIAMRoleTypeRunnerProvision,
		"maintenance": app.AWSIAMRoleTypeRunnerMaintenance,
		"read-only":   app.AWSIAMRoleTypeCustom,
	})
	s.Require().NoError(s.deps.Helpers.MigrateInstallRoles(ctx, s.deps.DB, testApp.ID, *newPerm))

	after := s.liveRolesByName(ctx, install.ID)
	s.Require().Len(after, 3)

	for name := range baseRoles {
		s.Equal(before[name].ID, after[name].ID, "%s must keep its row", name)
		s.Equal(roleConfigID(newPerm, name), after[name].AppRoleConfigID)
		s.True(after[name].Enabled)
		s.True(after[name].Provisioned)
		s.Equal(before[name].RoleID, after[name].RoleID)
	}

	added := after["read-only"]
	s.NotEmpty(added.ID)
	s.False(added.Enabled)
	s.False(added.Provisioned)
	s.Empty(added.RoleID)

	// The live set follows the newest permissions config even though the install
	// stays pinned to the config it was created against, so nothing reading the
	// live set may filter by install.app_config_id.
	var pinned app.Install
	s.Require().NoError(s.deps.DB.WithContext(ctx).Where(app.Install{ID: install.ID}).First(&pinned).Error)
	s.Equal(oldCfg.ID, pinned.AppConfigID)
	for _, r := range after {
		s.Equal(newCfg.ID, r.AppRoleConfig.AppConfigID)
	}
}

func (s *MigrateInstallRolesTestSuite) TestRemovesRolesDroppedFromConfig() {
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())

	testApp := s.deps.Seed.CreateApp(ctx, s.T())
	oldCfg := s.deps.Seed.CreateAppConfig(ctx, s.T(), testApp.ID)
	install := s.deps.Seed.CreateInstall(ctx, s.T(), testApp)
	oldPerm := s.seedPermissions(ctx, testApp.ID, oldCfg.ID, baseRoles)
	s.seedProvisionedInstallRoles(ctx, install.ID, oldPerm)
	before := s.liveRolesByName(ctx, install.ID)

	newCfg := s.deps.Seed.CreateBareAppConfig(ctx, s.T(), testApp.ID)
	newPerm := s.seedPermissions(ctx, testApp.ID, newCfg.ID, map[string]app.AWSIAMRoleType{
		"provision": app.AWSIAMRoleTypeRunnerProvision,
	})
	s.Require().NoError(s.deps.Helpers.MigrateInstallRoles(ctx, s.deps.DB, testApp.ID, *newPerm))

	after := s.liveRolesByName(ctx, install.ID)
	s.Require().Len(after, 1)
	s.Equal(before["provision"].ID, after["provision"].ID)
	s.NotContains(after, "maintenance")

	var all []app.InstallRoles
	s.Require().NoError(s.deps.DB.WithContext(ctx).Unscoped().
		Where(app.InstallRoles{InstallID: install.ID}).
		Find(&all).Error)
	s.Len(all, 2, "dropped roles are soft-deleted, not hard-deleted")
}

func (s *MigrateInstallRolesTestSuite) TestResyncOfSameConfigIsANoop() {
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())

	testApp := s.deps.Seed.CreateApp(ctx, s.T())
	cfg := s.deps.Seed.CreateAppConfig(ctx, s.T(), testApp.ID)
	install := s.deps.Seed.CreateInstall(ctx, s.T(), testApp)
	perm := s.seedPermissions(ctx, testApp.ID, cfg.ID, baseRoles)
	s.seedProvisionedInstallRoles(ctx, install.ID, perm)
	before := s.liveRolesByName(ctx, install.ID)

	s.Require().NoError(s.deps.Helpers.MigrateInstallRoles(ctx, s.deps.DB, testApp.ID, *perm))
	s.Require().NoError(s.deps.Helpers.MigrateInstallRoles(ctx, s.deps.DB, testApp.ID, *perm))

	after := s.liveRolesByName(ctx, install.ID)
	s.Require().Len(after, len(before))
	for name, row := range before {
		s.Equal(row.ID, after[name].ID)
		s.Equal(row.UpdatedAt.Unix(), after[name].UpdatedAt.Unix())
	}

	var all []app.InstallRoles
	s.Require().NoError(s.deps.DB.WithContext(ctx).Unscoped().
		Where(app.InstallRoles{InstallID: install.ID}).
		Find(&all).Error)
	s.Len(all, len(before))
}
