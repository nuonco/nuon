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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/configdiff"
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

type changeClassFixture struct {
	appID       string
	install     *app.Install
	oldID       string
	newID       string
	componentID string
	workerID    string
	billingID   string
	legacyID    string
	oldSandbox  *app.AppSandboxConfig
	newSandbox  *app.AppSandboxConfig
}

func (s *AppBranchConfigDiffTestSuite) seedPair() changeClassFixture {
	a := s.deps.Seed.CreateApp(s.ctx, s.T())
	oldCfg := s.deps.Seed.CreateBareAppConfig(s.ctx, s.T(), a.ID)
	newCfg := s.deps.Seed.CreateBareAppConfig(s.ctx, s.T(), a.ID)
	oldSandbox := s.deps.Seed.CreateAppSandboxConfig(s.ctx, s.T(), a.ID, oldCfg.ID)
	newSandbox := s.deps.Seed.CreateAppSandboxConfig(s.ctx, s.T(), a.ID, newCfg.ID)
	s.deps.Seed.CreateAppStackConfig(s.ctx, s.T(), a.ID, oldCfg.ID)
	s.deps.Seed.CreateAppStackConfig(s.ctx, s.T(), a.ID, newCfg.ID)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.AppStackConfig{}).
		Where("app_config_id IN ?", []string{oldCfg.ID, newCfg.ID}).
		Updates(map[string]any{"name": "shared-stack", "description": "shared"}).Error)

	component := s.deps.Seed.CreateComponent(s.ctx, s.T(), a.ID, app.ComponentTypeDockerBuild)
	worker := s.deps.Seed.CreateComponent(s.ctx, s.T(), a.ID, app.ComponentTypeDockerBuild)
	s.connection(component.ID, oldCfg.ID, "checksum-same", "build-same")
	s.connection(component.ID, newCfg.ID, "checksum-same", "build-same")
	s.connection(worker.ID, oldCfg.ID, "worker-same", "worker-build")
	s.connection(worker.ID, newCfg.ID, "worker-same", "worker-build")
	s.deps.Seed.CreateAppRunnerConfig(s.ctx, s.T(), a.ID, oldCfg.ID)
	s.deps.Seed.CreateAppRunnerConfig(s.ctx, s.T(), a.ID, newCfg.ID)

	install := s.deps.Seed.CreateInstall(s.ctx, s.T(), a)
	s.pin(install, newCfg.ID, newCfg.ID)
	return changeClassFixture{
		appID:       a.ID,
		install:     install,
		oldID:       oldCfg.ID,
		newID:       newCfg.ID,
		componentID: component.ID,
		workerID:    worker.ID,
		oldSandbox:  oldSandbox,
		newSandbox:  newSandbox,
	}
}

func (s *AppBranchConfigDiffTestSuite) connection(componentID, appConfigID, checksum, buildID string) {
	conn := s.deps.Seed.CreateDockerBuildComponentConfigConnection(s.ctx, s.T(), componentID, appConfigID)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(conn).Updates(map[string]any{
		"checksum":        checksum,
		"latest_build_id": buildID,
	}).Error)
}

func (s *AppBranchConfigDiffTestSuite) pin(install *app.Install, appConfigID, appliedID string) {
	install.AppConfigID = appConfigID
	install.AppConfigRef = app.AppConfigRef{AppliedConfigID: appliedID, ExpectedConfigID: appliedID}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.Install{ID: install.ID}).Updates(map[string]any{
		"app_config_id":  appConfigID,
		"app_config_ref": install.AppConfigRef,
	}).Error)
}

func (s *AppBranchConfigDiffTestSuite) stackRef(installID, appConfigID string) {
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Model(&app.InstallStack{}).
		Where("install_id = ?", installID).
		Update("app_config_ref", app.AppConfigRef{AppliedConfigID: appConfigID}).Error)
}

func (s *AppBranchConfigDiffTestSuite) sandboxRef(installID, appConfigID string) {
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Model(&app.InstallSandbox{}).
		Where("install_id = ?", installID).
		Update("app_config_ref", app.AppConfigRef{AppliedConfigID: appConfigID}).Error)
}

func (s *AppBranchConfigDiffTestSuite) componentRef(installID, componentID, appConfigID string) {
	component := app.InstallComponent{
		InstallID:    installID,
		ComponentID:  componentID,
		Status:       app.InstallComponentStatusActive,
		AppConfigRef: app.AppConfigRef{AppliedConfigID: appConfigID},
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&component).Error)
}

func (s *AppBranchConfigDiffTestSuite) diff(f changeClassFixture) *app.InstallConfigDiff {
	single, err := configdiff.ComputeInstallConfigDiff(s.ctx, s.deps.DB, f.install.AppConfigID, f.newID)
	require.NoError(s.T(), err)
	s.False(single.StackChanged)
	s.False(single.SandboxChanged)
	s.False(single.SandboxBuildChanged)
	s.Empty(single.Added)
	s.Empty(single.Changed)
	s.Empty(single.Removed)

	diff, err := s.deps.Helpers.AppBranchConfigDiff(s.ctx, f.install, f.newID)
	require.NoError(s.T(), err)
	return diff
}

func (s *AppBranchConfigDiffTestSuite) componentIDs(entries []app.ComponentDiffEntry) []string {
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.ComponentID)
	}
	return ids
}

func (s *AppBranchConfigDiffTestSuite) TestStackConfig() {
	f := s.seedPair()
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.AppStackConfig{}).
		Where("app_config_id = ?", f.oldID).
		Update("name", "previous-stack").Error)
	s.stackRef(f.install.ID, f.oldID)
	s.sandboxRef(f.install.ID, f.newID)
	s.componentRef(f.install.ID, f.componentID, f.newID)

	diff := s.diff(f)
	s.True(diff.StackChanged)
	s.False(diff.SandboxChanged)
	s.False(diff.SandboxBuildChanged)
	s.Empty(diff.Added)
	s.Empty(diff.Changed)
	s.Empty(diff.Removed)
}

func (s *AppBranchConfigDiffTestSuite) TestSandboxConfig() {
	f := s.seedPair()
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.PublicGitVCSConfig{}).
		Where("id = ?", f.oldSandbox.PublicGitVCSConfig.ID).
		Update("branch", "release").Error)
	s.stackRef(f.install.ID, f.newID)
	s.sandboxRef(f.install.ID, f.oldID)

	diff := s.diff(f)
	s.False(diff.StackChanged)
	s.True(diff.SandboxChanged)
	s.False(diff.SandboxBuildChanged)
	s.Empty(diff.Changed)
}

func (s *AppBranchConfigDiffTestSuite) TestSandboxSource() {
	f := s.seedPair()
	s.sandboxBuild(f.appID, f.oldID, f.oldSandbox.ID)
	s.sandboxBuild(f.appID, f.newID, f.newSandbox.ID)
	s.stackRef(f.install.ID, f.newID)
	s.sandboxRef(f.install.ID, f.oldID)

	diff := s.diff(f)
	s.False(diff.StackChanged)
	s.False(diff.SandboxChanged)
	s.True(diff.SandboxBuildChanged)
	s.NotEqual(diff.SandboxBuildOldID, diff.SandboxBuildNewID)
}

func (s *AppBranchConfigDiffTestSuite) sandboxBuild(appID, appConfigID, sandboxConfigID string) {
	build := app.AppSandboxBuild{
		AppID:              appID,
		AppConfigID:        appConfigID,
		AppSandboxConfigID: sandboxConfigID,
		Status:             app.AppSandboxBuildStatusActive,
		StatusDescription:  "active",
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&build).Error)
}

func (s *AppBranchConfigDiffTestSuite) TestComponentConfig() {
	f := s.seedPair()
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.ComponentConfigConnection{}).
		Where("app_config_id = ? AND component_id = ?", f.oldID, f.componentID).
		Update("checksum", "checksum-old").Error)
	s.stackRef(f.install.ID, f.newID)
	s.sandboxRef(f.install.ID, f.newID)
	s.componentRef(f.install.ID, f.componentID, f.oldID)

	diff := s.diff(f)
	s.False(diff.StackChanged)
	s.False(diff.SandboxChanged)
	s.Equal([]string{f.componentID}, s.componentIDs(diff.Changed))
	s.False(diff.Changed[0].BuildChanged)
}

func (s *AppBranchConfigDiffTestSuite) TestComponentSource() {
	f := s.seedPair()
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.ComponentConfigConnection{}).
		Where("app_config_id = ? AND component_id = ?", f.oldID, f.componentID).
		Update("latest_build_id", "build-old").Error)
	s.componentRef(f.install.ID, f.componentID, f.oldID)

	diff := s.diff(f)
	s.Equal([]string{f.componentID}, s.componentIDs(diff.Changed))
	s.True(diff.Changed[0].BuildChanged)
	s.False(diff.StackChanged)
}

func (s *AppBranchConfigDiffTestSuite) TestComponentAdded() {
	f := s.seedPair()
	billing := s.deps.Seed.CreateComponent(s.ctx, s.T(), f.appID, app.ComponentTypeDockerBuild)
	f.billingID = billing.ID
	s.connection(billing.ID, f.newID, "billing", "billing-build")
	s.pin(f.install, f.newID, f.oldID)

	diff := s.diff(f)
	s.Equal([]string{billing.ID}, s.componentIDs(diff.Added))
	s.Empty(diff.Changed)
	s.Empty(diff.Removed)
	s.False(diff.StackChanged)
	s.False(diff.SandboxChanged)
}

func (s *AppBranchConfigDiffTestSuite) TestComponentRemoved() {
	f := s.seedPair()
	legacy := s.deps.Seed.CreateComponent(s.ctx, s.T(), f.appID, app.ComponentTypeDockerBuild)
	s.connection(legacy.ID, f.oldID, "legacy", "legacy-build")
	s.componentRef(f.install.ID, legacy.ID, f.oldID)

	diff := s.diff(f)
	s.Equal([]string{legacy.ID}, s.componentIDs(diff.Removed))
	s.Empty(diff.Added)
	s.Empty(diff.Changed)
}

func (s *AppBranchConfigDiffTestSuite) TestMixedPartial() {
	f := s.seedPair()
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.PublicGitVCSConfig{}).
		Where("id = ?", f.oldSandbox.PublicGitVCSConfig.ID).
		Update("branch", "release").Error)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.ComponentConfigConnection{}).
		Where("app_config_id = ? AND component_id = ?", f.oldID, f.componentID).
		Update("checksum", "checksum-old").Error)
	s.stackRef(f.install.ID, f.newID)
	s.sandboxRef(f.install.ID, f.oldID)
	s.componentRef(f.install.ID, f.componentID, f.oldID)
	s.componentRef(f.install.ID, f.workerID, f.newID)

	diff := s.diff(f)
	s.False(diff.StackChanged)
	s.True(diff.SandboxChanged)
	s.Equal([]string{f.componentID}, s.componentIDs(diff.Changed))
	s.NotContains(s.componentIDs(diff.Changed), f.workerID)
	s.NotContains(s.componentIDs(diff.Added), f.workerID)
}

func (s *AppBranchConfigDiffTestSuite) TestEmptyStackRefUsesActiveStackVersion() {
	f := s.seedPair()
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.AppStackConfig{}).
		Where("app_config_id = ?", f.oldID).
		Update("name", "previous-stack").Error)
	var stack app.InstallStack
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Where(app.InstallStack{InstallID: f.install.ID}).
		First(&stack).Error)
	version := app.InstallStackVersion{
		InstallID:      f.install.ID,
		InstallStackID: stack.ID,
		AppConfigID:    f.oldID,
		Status:         app.CompositeStatus{Status: app.InstallStackVersionStatusActive},
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&version).Error)

	diff := s.diff(f)
	s.True(diff.StackChanged)
}
