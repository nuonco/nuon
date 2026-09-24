package activities

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type updateInstallAppConfigIDTestSuite struct {
	tests.BaseDBTestSuite

	db     *gorm.DB
	seeder *testseed.Seeder
}

func TestUpdateInstallAppConfigIDSuite(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	suite.Run(t, new(updateInstallAppConfigIDTestSuite))
}

func (s *updateInstallAppConfigIDTestSuite) SetupSuite() {
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

func (s *updateInstallAppConfigIDTestSuite) TearDownSuite() {
	db, err := s.db.DB()
	require.NoError(s.T(), err)
	require.NoError(s.T(), db.Close())
}

func (s *updateInstallAppConfigIDTestSuite) TestPointsInstallAtNewConfig() {
	t := s.T()
	ctx, _ := s.seeder.EnsureAccount(context.Background(), t)
	ctx, _ = s.seeder.EnsureOrg(ctx, t)

	a := s.seeder.CreateApp(ctx, t)
	s.seeder.CreateAppConfig(ctx, t, a.ID)
	install := s.seeder.CreateInstall(ctx, t, a)
	newCfg := s.seeder.CreateAppConfig(ctx, t, a.ID)

	acts := &Activities{db: s.db}
	require.NoError(t, acts.UpdateInstallAppConfigID(ctx, &UpdateInstallAppConfigIDInput{
		InstallID:      install.ID,
		NewAppConfigID: newCfg.ID,
	}))

	var got app.Install
	require.NoError(t, s.db.WithContext(ctx).First(&got, "id = ?", install.ID).Error)
	require.Equal(t, newCfg.ID, got.AppConfigID)

	var expected string
	require.NoError(t, s.db.WithContext(ctx).
		Raw(`SELECT app_config_ref->>'expected_config_id' FROM installs WHERE id = ?`, install.ID).
		Scan(&expected).Error)
	require.Equal(t, newCfg.ID, expected)
}
