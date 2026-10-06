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
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// updateDeployStatusTestSuite covers how a deploy's status reaches its install
// component: always, never (SkipStatusSync), or unless the component is live
// (SyncStatusUnlessLive, which the sync-and-plan step uses for failures).
type updateDeployStatusTestSuite struct {
	tests.BaseDBTestSuite

	db     *gorm.DB
	seeder *testseed.Seeder
}

func TestUpdateDeployStatusSuite(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	suite.Run(t, new(updateDeployStatusTestSuite))
}

func (s *updateDeployStatusTestSuite) SetupSuite() {
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

func (s *updateDeployStatusTestSuite) TearDownSuite() {
	db, err := s.db.DB()
	require.NoError(s.T(), err)
	require.NoError(s.T(), db.Close())
}

// seedDeploy returns a pending deploy of a component whose install component
// has componentStatus.
func (s *updateDeployStatusTestSuite) seedDeploy(componentStatus app.InstallComponentStatus) (context.Context, *app.InstallDeploy, *app.InstallComponent) {
	t := s.T()
	ctx, _ := s.seeder.EnsureAccount(context.Background(), t)
	ctx, _ = s.seeder.EnsureOrg(ctx, t)

	a := s.seeder.CreateApp(ctx, t)
	cfg := s.seeder.CreateAppConfig(ctx, t, a.ID)
	install := s.seeder.CreateInstall(ctx, t, a)
	comp := s.seeder.CreateComponent(ctx, t, a.ID, app.ComponentTypeHelmChart)
	ccc := s.seeder.CreateHelmComponentConfigConnection(ctx, t, comp.ID, cfg.ID)
	build := s.seeder.CreateComponentBuild(ctx, t, ccc.ID)

	ic := s.seeder.CreateInstallComponent(ctx, t, install.ID, comp.ID)
	require.NoError(t, s.db.WithContext(ctx).Model(ic).Updates(map[string]any{
		"status":    componentStatus,
		"status_v2": app.NewCompositeStatus(ctx, app.Status(componentStatus)),
	}).Error)

	deploy := s.seeder.CreateInstallDeploy(ctx, t, ic.ID, build.ID)
	require.NoError(t, s.db.WithContext(ctx).Model(deploy).Update("status", app.InstallDeployStatusPlanning).Error)
	return ctx, deploy, ic
}

// fail marks the deploy failed through both status activities, the way the
// sync-and-plan step does.
func (s *updateDeployStatusTestSuite) fail(ctx context.Context, deployID string, skip, syncUnlessLive bool) {
	t := s.T()
	acts := &Activities{db: s.db}
	require.NoError(t, acts.UpdateDeployStatus(ctx, UpdateDeployStatusRequest{
		DeployID:             deployID,
		Status:               app.InstallDeployStatusError,
		StatusDescription:    "plan job failed",
		SkipStatusSync:       skip,
		SyncStatusUnlessLive: syncUnlessLive,
	}))

	statusActs := statusactivities.New(statusactivities.Params{DB: s.db})
	require.NoError(t, statusActs.UpdateDeployStatusV2(ctx, statusactivities.UpdateDeployStatusV2Request{
		DeployID:             deployID,
		Status:               app.Status(app.InstallDeployStatusError),
		StatusDescription:    "plan job failed",
		SkipStatusSync:       skip,
		SyncStatusUnlessLive: syncUnlessLive,
	}))
}

// componentStatus is an install component's stored status columns. Loading the
// model would report status_v2 as both (InstallComponent.AfterQuery), but the
// installs view rolls up the legacy status column, so both are checked.
type componentStatus struct {
	Status            app.InstallComponentStatus
	StatusDescription string
	StatusV2          app.Status
}

func (s *updateDeployStatusTestSuite) component(ctx context.Context, id string) componentStatus {
	var got componentStatus
	require.NoError(s.T(), s.db.WithContext(ctx).Raw(
		`SELECT status, status_description, status_v2->>'status' AS status_v2 FROM install_components WHERE id = ?`, id,
	).Scan(&got).Error)
	return got
}

func (s *updateDeployStatusTestSuite) deploy(ctx context.Context, id string) app.InstallDeploy {
	var d app.InstallDeploy
	require.NoError(s.T(), s.db.WithContext(ctx).First(&d, "id = ?", id).Error)
	return d
}

func (s *updateDeployStatusTestSuite) TestSkipLeavesComponent() {
	ctx, deploy, ic := s.seedDeploy(app.InstallComponentStatusPending)
	s.fail(ctx, deploy.ID, true, false)

	s.Equal(app.InstallDeployStatusError, s.deploy(ctx, deploy.ID).Status)
	got := s.component(ctx, ic.ID)
	s.Equal(app.InstallComponentStatusPending, got.Status)
	s.Equal(app.Status(app.InstallComponentStatusPending), got.StatusV2)
}

func (s *updateDeployStatusTestSuite) TestSyncUnlessLiveFailsUndeployedComponent() {
	for _, status := range []app.InstallComponentStatus{
		"", // created with the install, never deployed
		app.InstallComponentStatusPending,
		app.InstallComponentStatusError,
		app.InstallComponentStatusInactive,
	} {
		s.Run("status="+string(status), func() {
			ctx, deploy, ic := s.seedDeploy(status)
			s.fail(ctx, deploy.ID, true, true)

			got := s.component(ctx, ic.ID)
			s.Equal(app.InstallComponentStatusError, got.Status)
			s.Equal("plan job failed", got.StatusDescription)
			s.Equal(app.Status(app.InstallDeployStatusError), got.StatusV2)
		})
	}
}

func (s *updateDeployStatusTestSuite) TestSyncUnlessLiveKeepsLiveComponent() {
	for _, status := range []app.InstallComponentStatus{
		app.InstallComponentStatusActive,
		app.InstallComponentStatusNoop,
		app.InstallComponentStatusDisabled,
	} {
		s.Run("status="+string(status), func() {
			ctx, deploy, ic := s.seedDeploy(status)
			s.fail(ctx, deploy.ID, true, true)

			s.Equal(app.InstallDeployStatusError, s.deploy(ctx, deploy.ID).Status)
			got := s.component(ctx, ic.ID)
			s.Equal(status, got.Status)
			s.Equal(app.Status(status), got.StatusV2)
		})
	}
}

func (s *updateDeployStatusTestSuite) TestNoSkipAlwaysSyncs() {
	ctx, deploy, ic := s.seedDeploy(app.InstallComponentStatusActive)
	s.fail(ctx, deploy.ID, false, false)

	got := s.component(ctx, ic.ID)
	s.Equal(app.InstallComponentStatusError, got.Status)
	s.Equal(app.Status(app.InstallDeployStatusError), got.StatusV2)
}
