package activities

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type getLatestDeployTestSuite struct {
	tests.BaseDBTestSuite

	db     *gorm.DB
	seeder *testseed.Seeder
}

func TestGetLatestDeploySuite(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	suite.Run(t, new(getLatestDeployTestSuite))
}

func (s *getLatestDeployTestSuite) SetupSuite() {
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

func (s *getLatestDeployTestSuite) TearDownSuite() {
	db, err := s.db.DB()
	require.NoError(s.T(), err)
	require.NoError(s.T(), db.Close())
}

type deployFixture struct {
	ctx         context.Context
	installID   string
	componentID string
	icID        string
	cccID       string
}

func (s *getLatestDeployTestSuite) fixture() deployFixture {
	t := s.T()
	ctx, _ := s.seeder.EnsureAccount(context.Background(), t)
	ctx, _ = s.seeder.EnsureOrg(ctx, t)

	a := s.seeder.CreateApp(ctx, t)
	cfg := s.seeder.CreateAppConfig(ctx, t, a.ID)
	install := s.seeder.CreateInstall(ctx, t, a)
	comp := s.seeder.CreateComponent(ctx, t, a.ID, app.ComponentTypeHelmChart)
	ccc := s.seeder.CreateHelmComponentConfigConnection(ctx, t, comp.ID, cfg.ID)
	ic := s.seeder.CreateInstallComponent(ctx, t, install.ID, comp.ID)

	return deployFixture{ctx: ctx, installID: install.ID, componentID: comp.ID, icID: ic.ID, cccID: ccc.ID}
}

func (s *getLatestDeployTestSuite) deploy(f deployFixture, status app.InstallDeployStatus, createdAt time.Time) string {
	t := s.T()
	bld := s.seeder.CreateComponentBuild(f.ctx, t, f.cccID)
	d := s.seeder.CreateInstallDeploy(f.ctx, t, f.icID, bld.ID)
	require.NoError(t, s.db.WithContext(f.ctx).Model(&app.InstallDeploy{}).Where("id = ?", d.ID).
		Updates(map[string]any{"status": status, "created_at": createdAt}).Error)
	return bld.ID
}

func (s *getLatestDeployTestSuite) TestPreferAppliedSkipsNewerFailedApply() {
	t := s.T()
	f := s.fixture()
	now := time.Now()
	applied := s.deploy(f, app.InstallDeployStatusActive, now.Add(-time.Hour))
	failed := s.deploy(f, app.InstallDeployStatusError, now)

	acts := &Activities{db: s.db}

	got, err := acts.GetLatestDeploy(f.ctx, GetLatestDeployRequest{InstallID: f.installID, ComponentID: f.componentID, PreferApplied: true})
	require.NoError(t, err)
	require.Equal(t, applied, got.ComponentBuildID)

	got, err = acts.GetLatestDeploy(f.ctx, GetLatestDeployRequest{InstallID: f.installID, ComponentID: f.componentID})
	require.NoError(t, err)
	require.Equal(t, failed, got.ComponentBuildID)
}

func (s *getLatestDeployTestSuite) TestPreferAppliedFallsBackWhenNothingApplied() {
	t := s.T()
	f := s.fixture()
	failed := s.deploy(f, app.InstallDeployStatusError, time.Now())

	acts := &Activities{db: s.db}
	got, err := acts.GetLatestDeploy(f.ctx, GetLatestDeployRequest{InstallID: f.installID, ComponentID: f.componentID, PreferApplied: true})
	require.NoError(t, err)
	require.Equal(t, failed, got.ComponentBuildID)
}
