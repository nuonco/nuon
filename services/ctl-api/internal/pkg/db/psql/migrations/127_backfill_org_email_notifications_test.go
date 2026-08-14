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

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	psqlmigrations "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql/migrations"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type migration127TestSuite struct {
	tests.BaseDBTestSuite

	db     *gorm.DB
	seeder *testseed.Seeder
}

func TestMigration127Suite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(migration127TestSuite))
}

func (s *migration127TestSuite) SetupSuite() {
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

func (s *migration127TestSuite) TearDownSuite() {
	db, err := s.db.DB()
	require.NoError(s.T(), err)
	require.NoError(s.T(), db.Close())
}

func (s *migration127TestSuite) seedOrgWithNotificationsConfig(
	ctx context.Context,
	acctType app.AccountType,
	ownerType string,
	emailEnabled bool,
) (*app.Org, *app.NotificationsConfig) {
	acct := &app.Account{
		ID:          domains.NewAccountID(),
		Subject:     domains.NewAccountID(),
		Email:       fmt.Sprintf("%s@test.nuon.co", domains.NewAccountID()),
		AccountType: acctType,
	}
	require.NoError(s.T(), s.db.WithContext(ctx).Create(acct).Error)

	org := &app.Org{
		ID:          domains.NewOrgID(),
		Name:        fmt.Sprintf("org-%s", domains.NewOrgID()),
		OrgType:     app.OrgTypeSandbox,
		Status:      app.OrgStatusActive,
		CreatedByID: acct.ID,
	}
	require.NoError(s.T(), s.db.WithContext(ctx).Create(org).Error)

	cfg := &app.NotificationsConfig{
		CreatedByID:              acct.ID,
		OrgID:                    org.ID,
		OwnerID:                  org.ID,
		OwnerType:                ownerType,
		EnableEmailNotifications: emailEnabled,
	}
	require.NoError(s.T(), s.db.WithContext(ctx).Create(cfg).Error)

	return org, cfg
}

func (s *migration127TestSuite) emailEnabled(ctx context.Context, id string) bool {
	var cfg app.NotificationsConfig
	require.NoError(s.T(), s.db.WithContext(ctx).First(&cfg, "id = ?", id).Error)
	return cfg.EnableEmailNotifications
}

func (s *migration127TestSuite) TestBackfillsOnlyUserCreatedOrgConfigs() {
	ctx := context.Background()

	// the Restate Cloud case: org created by a new-auth-service user while
	// create_org still gated the flag on auth0 only
	_, authOrgCfg := s.seedOrgWithNotificationsConfig(ctx, app.AccountTypeAuth, "orgs", false)
	_, auth0OrgCfg := s.seedOrgWithNotificationsConfig(ctx, app.AccountTypeAuth0, "orgs", false)
	_, serviceOrgCfg := s.seedOrgWithNotificationsConfig(ctx, app.AccountTypeService, "orgs", false)
	_, canaryOrgCfg := s.seedOrgWithNotificationsConfig(ctx, app.AccountTypeCanary, "orgs", false)
	_, appCfg := s.seedOrgWithNotificationsConfig(ctx, app.AccountTypeAuth, "apps", false)

	migrations := psqlmigrations.New(psqlmigrations.Params{L: zap.NewNop()})
	require.NoError(s.T(), migrations.Migration127BackfillOrgEmailNotifications(ctx, s.db))

	require.True(s.T(), s.emailEnabled(ctx, authOrgCfg.ID), "auth-created org should be backfilled")
	require.True(s.T(), s.emailEnabled(ctx, auth0OrgCfg.ID), "auth0-created org should be backfilled")
	require.False(s.T(), s.emailEnabled(ctx, serviceOrgCfg.ID), "service-account org must stay disabled")
	require.False(s.T(), s.emailEnabled(ctx, canaryOrgCfg.ID), "canary org must stay disabled")
	require.False(s.T(), s.emailEnabled(ctx, appCfg.ID), "app-owned config must be untouched")
}

func (s *migration127TestSuite) TestIsIdempotentAndLeavesEnabledRowsAlone() {
	ctx := context.Background()

	_, disabled := s.seedOrgWithNotificationsConfig(ctx, app.AccountTypeAuth, "orgs", false)
	_, alreadyEnabled := s.seedOrgWithNotificationsConfig(ctx, app.AccountTypeAuth, "orgs", true)

	var before app.NotificationsConfig
	require.NoError(s.T(), s.db.WithContext(ctx).First(&before, "id = ?", alreadyEnabled.ID).Error)

	migrations := psqlmigrations.New(psqlmigrations.Params{L: zap.NewNop()})
	require.NoError(s.T(), migrations.Migration127BackfillOrgEmailNotifications(ctx, s.db))
	require.NoError(s.T(), migrations.Migration127BackfillOrgEmailNotifications(ctx, s.db))

	require.True(s.T(), s.emailEnabled(ctx, disabled.ID))

	var after app.NotificationsConfig
	require.NoError(s.T(), s.db.WithContext(ctx).First(&after, "id = ?", alreadyEnabled.ID).Error)
	require.True(s.T(), after.EnableEmailNotifications)
	require.Equal(s.T(), before.UpdatedAt.UnixNano(), after.UpdatedAt.UnixNano(),
		"already-enabled rows must not be rewritten")
}
