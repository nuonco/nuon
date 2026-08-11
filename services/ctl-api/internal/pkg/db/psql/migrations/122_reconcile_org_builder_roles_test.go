package migrations_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	psqlmigrations "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql/migrations"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type migration122TestSuite struct {
	tests.BaseDBTestSuite

	db     *gorm.DB
	seeder *testseed.Seeder
}

func TestMigration122Suite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(migration122TestSuite))
}

func (s *migration122TestSuite) SetupSuite() {
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

func (s *migration122TestSuite) TearDownSuite() {
	db, err := s.db.DB()
	require.NoError(s.T(), err)
	require.NoError(s.T(), db.Close())
}

func (s *migration122TestSuite) TestReconcilesMissingAndMalformedBuilderRole() {
	ctx, account := s.seeder.EnsureAccount(context.Background(), s.T())
	ctx, org := s.seeder.EnsureOrg(ctx, s.T())

	migrations := new(psqlmigrations.Migrations)
	require.NoError(s.T(), migrations.Migration122ReconcileOrgBuilderRoles(ctx, s.db))
	role := s.requireCanonicalBuilderRole(ctx, org.ID)
	roleID := role.ID
	s.createRoleAssignment(ctx, org.ID, role.ID, account.ID)

	var policy app.Policy
	require.NoError(s.T(), s.db.WithContext(ctx).Where(app.Policy{RoleID: role.ID}).First(&policy).Error)
	require.NoError(s.T(), s.db.WithContext(ctx).Model(&policy).Updates(map[string]any{
		"name":        app.PolicyNameOrgAdmin,
		"permissions": pgtype.Hstore(map[string]*string{org.ID: permissions.PermissionAll.ToStrPtr()}),
	}).Error)
	require.NoError(s.T(), s.db.WithContext(ctx).Delete(&policy).Error)

	duplicateRole := app.Role{
		OrgID:       generics.NewNullString(org.ID),
		CreatedByID: account.ID,
		RoleType:    app.RoleTypeOrgBuilder,
		Policies: []app.Policy{
			{
				OrgID:       generics.NewNullString(org.ID),
				CreatedByID: account.ID,
				Name:        app.PolicyNameOrgAdmin,
				Permissions: pgtype.Hstore(map[string]*string{org.ID: permissions.PermissionAll.ToStrPtr()}),
			},
		},
	}
	require.NoError(s.T(), s.db.WithContext(ctx).Create(&duplicateRole).Error)
	s.createRoleAssignment(ctx, org.ID, duplicateRole.ID, account.ID)

	require.NoError(s.T(), migrations.Migration122ReconcileOrgBuilderRoles(ctx, s.db))
	require.NoError(s.T(), migrations.Migration122ReconcileOrgBuilderRoles(ctx, s.db))
	role = s.requireBuilderRole(ctx, roleID, org.ID)
	require.Equal(s.T(), roleID, role.ID)
	s.requireBuilderRole(ctx, duplicateRole.ID, org.ID)
	s.requireRoleAssignment(ctx, role.ID, account.ID)
	s.requireRoleAssignment(ctx, duplicateRole.ID, account.ID)
	require.NoError(s.T(), s.db.WithContext(ctx).Unscoped().
		Where(app.Policy{RoleID: duplicateRole.ID}).
		Delete(&app.Policy{}).Error)
	require.NoError(s.T(), migrations.Migration122ReconcileOrgBuilderRoles(ctx, s.db))
	s.requireBuilderRole(ctx, duplicateRole.ID, org.ID)
	s.requireRoleAssignment(ctx, duplicateRole.ID, account.ID)

	var roleCount int64
	require.NoError(s.T(), s.db.WithContext(ctx).
		Model(&app.Role{}).
		Where(app.Role{OrgID: generics.NewNullString(org.ID), RoleType: app.RoleTypeOrgBuilder}).
		Count(&roleCount).Error)
	require.EqualValues(s.T(), 2, roleCount)
}

func (s *migration122TestSuite) requireCanonicalBuilderRole(ctx context.Context, orgID string) app.Role {
	var role app.Role
	require.NoError(s.T(), s.db.WithContext(ctx).
		Where(app.Role{OrgID: generics.NewNullString(orgID), RoleType: app.RoleTypeOrgBuilder}).
		First(&role).Error)
	return s.requireBuilderRole(ctx, role.ID, orgID)
}

func (s *migration122TestSuite) requireBuilderRole(ctx context.Context, roleID, orgID string) app.Role {
	var role app.Role
	require.NoError(s.T(), s.db.WithContext(ctx).
		Preload("Policies").
		Where(app.Role{ID: roleID, OrgID: generics.NewNullString(orgID), RoleType: app.RoleTypeOrgBuilder}).
		First(&role).Error)
	require.Len(s.T(), role.Policies, 1)
	require.Equal(s.T(), app.PolicyNameOrgBuilder, role.Policies[0].Name)
	require.Equal(s.T(), permissions.PermissionRead.ToStrPtr(), role.Policies[0].Permissions[orgID])
	require.Equal(s.T(), permissions.PermissionCreate.ToStrPtr(), role.Policies[0].Permissions[orgID+":component_builds"])
	require.Len(s.T(), role.Policies[0].Permissions, 2)

	return role
}

func (s *migration122TestSuite) createRoleAssignment(ctx context.Context, orgID, roleID, accountID string) {
	require.NoError(s.T(), s.db.WithContext(ctx).Create(&app.AccountRole{
		OrgID:     generics.NewNullString(orgID),
		RoleID:    roleID,
		AccountID: accountID,
	}).Error)
}

func (s *migration122TestSuite) requireRoleAssignment(ctx context.Context, roleID, accountID string) {
	var assignmentCount int64
	require.NoError(s.T(), s.db.WithContext(ctx).
		Model(&app.AccountRole{}).
		Where(app.AccountRole{RoleID: roleID, AccountID: accountID}).
		Count(&assignmentCount).Error)
	require.EqualValues(s.T(), 1, assignmentCount)
}
