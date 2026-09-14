package helpers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type runnerActiveDeps struct {
	fx.In

	DB      *gorm.DB `name:"psql"`
	Seed    *testseed.Seeder
	Helpers *installhelpers.Helpers
}

type RunnerActiveTestSuite struct {
	tests.BaseDBTestSuite

	fxApp *fxtest.App
	deps  runnerActiveDeps
	ctx   context.Context
	app   *app.App
}

func TestRunnerActiveSuite(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	suite.Run(t, new(RunnerActiveTestSuite))
}

func (s *RunnerActiveTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	options := append(tests.CtlApiFXOptions(s.T()), fx.Populate(&s.deps))
	s.fxApp = fxtest.New(s.T(), options...)
	s.fxApp.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *RunnerActiveTestSuite) TearDownSuite() {
	s.fxApp.RequireStop()
}

func (s *RunnerActiveTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.ctx = context.Background()
	s.ctx, _ = s.deps.Seed.EnsureAccount(s.ctx, s.T())
	s.ctx, _ = s.deps.Seed.EnsureOrg(s.ctx, s.T())
	s.app = s.deps.Seed.CreateApp(s.ctx, s.T())
	s.deps.Seed.CreateAppConfig(s.ctx, s.T(), s.app.ID)
}

func (s *RunnerActiveTestSuite) TestNoRunner() {
	install := s.deps.Seed.CreateInstall(s.ctx, s.T(), s.app)
	active, err := s.deps.Helpers.HasActiveRunner(s.ctx, install.ID)
	require.NoError(s.T(), err)
	require.False(s.T(), active)
}

func (s *RunnerActiveTestSuite) TestRunnerWithoutProcess() {
	install, _ := s.seedRunner()
	active, err := s.deps.Helpers.HasActiveRunner(s.ctx, install.ID)
	require.NoError(s.T(), err)
	require.False(s.T(), active)
}

func (s *RunnerActiveTestSuite) TestInactiveProcess() {
	install, runner := s.seedRunner()
	s.seedProcess(runner.ID, app.RunnerProcessStatusInactive)
	active, err := s.deps.Helpers.HasActiveRunner(s.ctx, install.ID)
	require.NoError(s.T(), err)
	require.False(s.T(), active)
}

func (s *RunnerActiveTestSuite) TestActiveProcesses() {
	for _, status := range app.ActiveRunnerProcessStatuses() {
		s.Run(string(status), func() {
			install, runner := s.seedRunner()
			s.seedProcess(runner.ID, status)
			active, err := s.deps.Helpers.HasActiveRunner(s.ctx, install.ID)
			require.NoError(s.T(), err)
			require.True(s.T(), active)
		})
	}
}

func (s *RunnerActiveTestSuite) seedRunner() (*app.Install, *app.Runner) {
	install := s.deps.Seed.CreateInstall(s.ctx, s.T(), s.app)
	group := &app.RunnerGroup{
		OwnerID:   install.ID,
		OwnerType: "installs",
		Type:      app.RunnerGroupTypeInstall,
		Platform:  app.AppRunnerTypeAWS,
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(group).Error)
	runner := &app.Runner{
		RunnerGroupID:     group.ID,
		Name:              "runner-" + install.ID,
		DisplayName:       "runner",
		Status:            app.RunnerStatusActive,
		StatusDescription: "active",
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(runner).Error)
	return install, runner
}

func (s *RunnerActiveTestSuite) seedProcess(runnerID string, status app.RunnerProcessStatus) {
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&app.RunnerProcess{
		RunnerID: runnerID,
		Type:     app.RunnerProcessTypeInstall,
		CompositeStatus: app.CompositeStatus{
			Status: app.Status(status),
		},
	}).Error)
}
