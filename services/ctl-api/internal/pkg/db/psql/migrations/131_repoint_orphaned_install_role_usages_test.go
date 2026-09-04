package migrations_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	psqlmigrations "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql/migrations"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type migration131TestSuite struct {
	tests.BaseDBTestSuite

	db     *gorm.DB
	seeder *testseed.Seeder
}

func TestMigration131Suite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(migration131TestSuite))
}

func (s *migration131TestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	cfg, err := tests.LoadDBConfig()
	require.NoError(s.T(), err)
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	s.db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(s.T(), err)
	s.seeder = testseed.New(testseed.Params{DB: s.db})
	s.SetDB(s.db)
}

func (s *migration131TestSuite) TearDownSuite() {
	db, err := s.db.DB()
	require.NoError(s.T(), err)
	require.NoError(s.T(), db.Close())
}

func (s *migration131TestSuite) seedPermissions(ctx context.Context, appID, appConfigID string, names ...string) *app.AppPermissionsConfig {
	cfg := &app.AppPermissionsConfig{AppID: appID, AppConfigID: appConfigID}
	for _, name := range names {
		cfg.Roles = append(cfg.Roles, app.AppAWSIAMRoleConfig{
			AppConfigID: appConfigID,
			Type:        app.AWSIAMRoleTypeCustom,
			Name:        name,
			DisplayName: name,
		})
	}
	require.NoError(s.T(), s.db.WithContext(ctx).Create(cfg).Error)
	return cfg
}

func (s *migration131TestSuite) seedInstallRoles(ctx context.Context, installID string, perm *app.AppPermissionsConfig) map[string]app.InstallRoles {
	rows := make(map[string]app.InstallRoles, len(perm.Roles))
	for _, role := range perm.Roles {
		row := app.InstallRoles{InstallID: installID, AppRoleConfigID: role.ID}
		require.NoError(s.T(), s.db.WithContext(ctx).Create(&row).Error)
		rows[role.Name] = row
	}
	return rows
}

func (s *migration131TestSuite) seedUsage(ctx context.Context, installID, installRoleID string) app.InstallRoleUsage {
	job := s.seeder.CreateRunnerJob(ctx, s.T(), installID, "installs")
	usage := app.InstallRoleUsage{InstallRoleID: installRoleID, RunnerJobID: job.ID, RoleName: "r"}
	require.NoError(s.T(), s.db.WithContext(ctx).Create(&usage).Error)
	return usage
}

func (s *migration131TestSuite) usageRoleID(ctx context.Context, usageID string) string {
	var got app.InstallRoleUsage
	require.NoError(s.T(), s.db.WithContext(ctx).Where(app.InstallRoleUsage{ID: usageID}).First(&got).Error)
	return got.InstallRoleID
}

func (s *migration131TestSuite) TestRepointsUsagesToLiveRowOfSameName() {
	ctx := context.Background()
	ctx, _ = s.seeder.EnsureAccount(ctx, s.T())
	ctx, _ = s.seeder.EnsureOrg(ctx, s.T())

	testApp := s.seeder.CreateApp(ctx, s.T())
	oldCfg := s.seeder.CreateAppConfig(ctx, s.T(), testApp.ID)
	install := s.seeder.CreateInstall(ctx, s.T(), testApp)

	oldPerm := s.seedPermissions(ctx, testApp.ID, oldCfg.ID, "provision", "maintenance", "retired")
	dead := s.seedInstallRoles(ctx, install.ID, oldPerm)
	for _, row := range dead {
		require.NoError(s.T(), s.db.WithContext(ctx).Delete(&app.InstallRoles{}, "id = ?", row.ID).Error)
	}

	newCfg := s.seeder.CreateBareAppConfig(ctx, s.T(), testApp.ID)
	newPerm := s.seedPermissions(ctx, testApp.ID, newCfg.ID, "provision", "maintenance")
	live := s.seedInstallRoles(ctx, install.ID, newPerm)

	orphanProvision := s.seedUsage(ctx, install.ID, dead["provision"].ID)
	orphanMaintenance := s.seedUsage(ctx, install.ID, dead["maintenance"].ID)
	orphanRetired := s.seedUsage(ctx, install.ID, dead["retired"].ID)
	alreadyLive := s.seedUsage(ctx, install.ID, live["provision"].ID)

	var before app.InstallRoleUsage
	require.NoError(s.T(), s.db.WithContext(ctx).Where(app.InstallRoleUsage{ID: alreadyLive.ID}).First(&before).Error)

	migrations := psqlmigrations.New(psqlmigrations.Params{L: zap.NewNop()})
	require.NoError(s.T(), migrations.Migration131RepointOrphanedInstallRoleUsages(ctx, s.db))

	require.Equal(s.T(), live["provision"].ID, s.usageRoleID(ctx, orphanProvision.ID))
	require.Equal(s.T(), live["maintenance"].ID, s.usageRoleID(ctx, orphanMaintenance.ID))
	require.Equal(s.T(), dead["retired"].ID, s.usageRoleID(ctx, orphanRetired.ID),
		"a role with no live counterpart has nowhere to go and must be left alone")

	var after app.InstallRoleUsage
	require.NoError(s.T(), s.db.WithContext(ctx).Where(app.InstallRoleUsage{ID: alreadyLive.ID}).First(&after).Error)
	require.Equal(s.T(), live["provision"].ID, after.InstallRoleID)
	require.Equal(s.T(), before.UpdatedAt.UnixNano(), after.UpdatedAt.UnixNano(), "live usages must not be rewritten")

	require.NoError(s.T(), migrations.Migration131RepointOrphanedInstallRoleUsages(ctx, s.db))
	require.Equal(s.T(), live["provision"].ID, s.usageRoleID(ctx, orphanProvision.ID))
}
