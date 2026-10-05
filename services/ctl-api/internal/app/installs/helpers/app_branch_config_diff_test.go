package helpers_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type appBranchConfigDiffDeps struct {
	fx.In

	DB      *gorm.DB `name:"psql"`
	Seed    *testseed.Seeder
	Helpers *installhelpers.Helpers
}

type AppBranchConfigDiffTestSuite struct {
	tests.BaseDBTestSuite

	app  *fxtest.App
	deps appBranchConfigDiffDeps
	ctx  context.Context
}

func TestAppBranchConfigDiffSuite(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	suite.Run(t, new(AppBranchConfigDiffTestSuite))
}

func (s *AppBranchConfigDiffTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	options := append(tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
		T:               s.T(),
		Mocks:           &tests.TestMocks{MockTC: temporalclient.NewMockClient(gomock.NewController(s.T()))},
		CustomValidator: true,
	}), fx.Populate(&s.deps))
	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *AppBranchConfigDiffTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AppBranchConfigDiffTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.ctx = context.Background()
	s.ctx, _ = s.deps.Seed.EnsureAccount(s.ctx, s.T())
	s.ctx, _ = s.deps.Seed.EnsureOrg(s.ctx, s.T())
}

type fixture struct {
	install                    *app.Install
	applied, live, next, other string
}

// seed builds four app configs: applied (what the install last recorded), live (what its stack runs), next (the new
// config, whose stack matches live) and other (a config whose stack matches neither).
func (s *AppBranchConfigDiffTestSuite) seed() fixture {
	a := s.deps.Seed.CreateApp(s.ctx, s.T())
	applied := s.deps.Seed.CreateAppConfig(s.ctx, s.T(), a.ID)
	live := s.deps.Seed.CreateAppConfig(s.ctx, s.T(), a.ID)
	other := s.deps.Seed.CreateAppConfig(s.ctx, s.T(), a.ID)
	next := s.deps.Seed.CreateAppConfig(s.ctx, s.T(), a.ID)

	var liveStack app.AppStackConfig
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).First(&liveStack, "app_config_id = ?", live.ID).Error)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.AppStackConfig{}).
		Where("app_config_id = ?", next.ID).
		Updates(map[string]any{"name": liveStack.Name, "description": liveStack.Description}).Error)

	install := s.deps.Seed.CreateInstall(s.ctx, s.T(), a)
	install.AppConfigRef = app.AppConfigRef{AppliedConfigID: applied.ID, ExpectedConfigID: applied.ID}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.Install{ID: install.ID}).
		Update("app_config_ref", install.AppConfigRef).Error)

	return fixture{install: install, applied: applied.ID, live: live.ID, next: next.ID, other: other.ID}
}

func (s *AppBranchConfigDiffTestSuite) activeStackVersion(installID, appConfigID string) {
	stack := app.InstallStack{InstallID: installID}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&stack).Error)
	version := app.InstallStackVersion{
		InstallID:      installID,
		InstallStackID: stack.ID,
		AppConfigID:    appConfigID,
		Status:         app.CompositeStatus{Status: app.InstallStackVersionStatusActive},
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&version).Error)
}

func (s *AppBranchConfigDiffTestSuite) stackChanged(f fixture, newAppConfigID string) bool {
	diff, err := s.deps.Helpers.AppBranchConfigDiff(s.ctx, f.install, newAppConfigID)
	require.NoError(s.T(), err)
	return diff.StackChanged
}

func (s *AppBranchConfigDiffTestSuite) TestNoActiveStackVersionUsesAppliedConfig() {
	f := s.seed()
	s.True(s.stackChanged(f, f.next))
}

func (s *AppBranchConfigDiffTestSuite) TestActiveStackVersionAlreadyRunsNewStack() {
	f := s.seed()
	s.activeStackVersion(f.install.ID, f.live)
	s.False(s.stackChanged(f, f.next))
}

func (s *AppBranchConfigDiffTestSuite) TestActiveStackVersionOnAppliedConfig() {
	f := s.seed()
	s.activeStackVersion(f.install.ID, f.applied)
	s.True(s.stackChanged(f, f.next))
}

func (s *AppBranchConfigDiffTestSuite) TestActiveStackVersionDiffersFromNewStack() {
	f := s.seed()
	s.activeStackVersion(f.install.ID, f.other)
	s.True(s.stackChanged(f, f.next))
}
