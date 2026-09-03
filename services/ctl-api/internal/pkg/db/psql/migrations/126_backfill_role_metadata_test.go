package migrations_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	psqlmigrations "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql/migrations"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type migration125TestSuite struct {
	tests.BaseDBTestSuite

	db     *gorm.DB
	seeder *testseed.Seeder
}

func TestMigration125Suite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(migration125TestSuite))
}

func (s *migration125TestSuite) SetupSuite() {
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

func (s *migration125TestSuite) TearDownSuite() {
	db, err := s.db.DB()
	require.NoError(s.T(), err)
	require.NoError(s.T(), db.Close())
}

func (s *migration125TestSuite) TestBackfillsMetadataAndRecreatesMissingRoles() {
	ctx, _ := s.seeder.EnsureAccount(context.Background(), s.T())
	ctx, org := s.seeder.EnsureOrg(ctx, s.T())

	migrations := new(psqlmigrations.Migrations)
	require.NoError(s.T(), migrations.Migration126BackfillRoleMetadata(ctx, s.db),
		"first run creates the org's missing roles")

	var admin app.Role
	require.NoError(s.T(), s.db.WithContext(ctx).
		Preload("Policies").
		Where(app.Role{OrgID: generics.NewNullString(org.ID), RoleType: app.RoleTypeOrgAdmin}).
		First(&admin).Error)
	require.Len(s.T(), admin.Policies, 1)
	permissionsBefore := admin.Policies[0].Permissions

	require.NoError(s.T(), s.db.WithContext(ctx).
		Model(&app.Role{}).
		Where(app.Role{ID: admin.ID}).
		Select("title", "description", "contexts", "managed").
		Updates(app.Role{}).Error)

	var runner app.Role
	require.NoError(s.T(), s.db.WithContext(ctx).
		Where(app.Role{OrgID: generics.NewNullString(org.ID), RoleType: app.RoleTypeRunner}).
		First(&runner).Error)
	require.NoError(s.T(), s.db.WithContext(ctx).Delete(&runner).Error)

	require.NoError(s.T(), migrations.Migration126BackfillRoleMetadata(ctx, s.db))
	require.NoError(s.T(), migrations.Migration126BackfillRoleMetadata(ctx, s.db))

	var reconciledAdmin app.Role
	require.NoError(s.T(), s.db.WithContext(ctx).
		Preload("Policies").
		Where(app.Role{ID: admin.ID}).
		First(&reconciledAdmin).Error)
	require.Equal(s.T(), "Admin", reconciledAdmin.Title)
	require.NotEmpty(s.T(), reconciledAdmin.Description)
	require.True(s.T(), reconciledAdmin.Managed)
	require.ElementsMatch(s.T(), []string{
		app.RoleContextTeam,
		app.RoleContextServiceAccount,
		app.RoleContextAPIToken,
		app.RoleContextTrustPolicy,
	}, reconciledAdmin.Contexts)
	require.Len(s.T(), reconciledAdmin.Policies, 1)
	require.Equal(s.T(), permissionsBefore, reconciledAdmin.Policies[0].Permissions,
		"reconciler must never modify existing policies")

	var reconciledRunner app.Role
	require.NoError(s.T(), s.db.WithContext(ctx).
		Preload("Policies").
		Where(app.Role{OrgID: generics.NewNullString(org.ID), RoleType: app.RoleTypeRunner}).
		First(&reconciledRunner).Error)
	require.NotEqual(s.T(), runner.ID, reconciledRunner.ID)
	require.True(s.T(), reconciledRunner.Managed)
	require.Empty(s.T(), reconciledRunner.Contexts)
	require.Len(s.T(), reconciledRunner.Policies, 1)
}
