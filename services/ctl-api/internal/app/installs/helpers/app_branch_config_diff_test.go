package helpers_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/sourceidentity"
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

	app    *fxtest.App
	deps   appBranchConfigDiffDeps
	ctx    context.Context
	builds map[string]string
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
	s.builds = map[string]string{}
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

	component := s.deps.Seed.CreateComponent(s.ctx, s.T(), a.ID, app.ComponentTypeTerraformModule)
	worker := s.deps.Seed.CreateComponent(s.ctx, s.T(), a.ID, app.ComponentTypeTerraformModule)
	s.connection(component.ID, oldCfg.ID, "checksum-same", "build-same")
	s.connection(component.ID, newCfg.ID, "checksum-same", "build-same")
	s.connection(worker.ID, oldCfg.ID, "worker-same", "worker-build")
	s.connection(worker.ID, newCfg.ID, "worker-same", "worker-build")
	s.deps.Seed.CreateAppRunnerConfig(s.ctx, s.T(), a.ID, oldCfg.ID)
	s.deps.Seed.CreateAppRunnerConfig(s.ctx, s.T(), a.ID, newCfg.ID)
	s.sandboxBuild(a.ID, oldCfg.ID, oldSandbox.ID)
	s.sandboxBuild(a.ID, newCfg.ID, newSandbox.ID)

	install := s.deps.Seed.CreateInstall(s.ctx, s.T(), a)
	s.pin(install, oldCfg.ID)
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

func (s *AppBranchConfigDiffTestSuite) connection(componentID, appConfigID, checksum, source string) {
	conn := s.deps.Seed.CreateTerraformComponentConfigConnection(s.ctx, s.T(), componentID, appConfigID)
	buildID, ok := s.builds[source]
	if !ok {
		buildID = s.componentBuild(conn.ID, source)
		s.builds[source] = buildID
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(conn).Updates(map[string]any{
		"checksum":        checksum,
		"latest_build_id": buildID,
	}).Error)
}

func (s *AppBranchConfigDiffTestSuite) sourceIdentity(source string) *sourceidentity.Identity {
	sum := sha256.Sum256([]byte(source))
	digest := hex.EncodeToString(sum[:])
	identity := sourceidentity.New("terraform", source)
	identity.Captured = true
	identity.Commit = digest[:40]
	identity.Repository = "https://example.com/acme/repo"
	identity.Files = []sourceidentity.File{{Path: "main.tf", SHA256: digest}}
	identity.Seal()
	return identity
}

func (s *AppBranchConfigDiffTestSuite) componentBuild(connID, source string) string {
	identity := s.sourceIdentity(source)
	build := app.ComponentBuild{
		ComponentConfigConnectionID: connID,
		Status:                      app.ComponentBuildStatusActive,
		StatusDescription:           "active",
		SourceIdentity:              identity,
		SourceChecksum:              identity.Checksum,
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&build).Error)
	return build.ID
}

func (s *AppBranchConfigDiffTestSuite) latestBuild(appConfigID, componentID string) (string, string) {
	var conn app.ComponentConfigConnection
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Where(app.ComponentConfigConnection{AppConfigID: appConfigID, ComponentID: componentID}).
		First(&conn).Error)
	return conn.ID, conn.LatestBuildID.String
}

func (s *AppBranchConfigDiffTestSuite) pin(install *app.Install, appConfigID string) {
	install.AppConfigID = appConfigID
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.Install{ID: install.ID}).
		Update("app_config_id", appConfigID).Error)
}

func (s *AppBranchConfigDiffTestSuite) diff(f changeClassFixture) *app.InstallConfigDiff {
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

	diff := s.diff(f)
	s.True(diff.StackChanged)
	s.False(diff.SandboxChanged)
	s.Empty(diff.Added)
	s.Empty(diff.Changed)
	s.Empty(diff.Removed)
}

func (s *AppBranchConfigDiffTestSuite) TestSandboxConfig() {
	f := s.seedPair()
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.PublicGitVCSConfig{}).
		Where("id = ?", f.oldSandbox.PublicGitVCSConfig.ID).
		Update("branch", "release").Error)

	diff := s.diff(f)
	s.False(diff.StackChanged)
	s.True(diff.SandboxChanged)
	s.True(diff.SandboxBuildChanged)
	s.Empty(diff.Changed)
}

func (s *AppBranchConfigDiffTestSuite) TestSandboxSource() {
	f := s.seedPair()

	diff := s.diff(f)
	s.False(diff.StackChanged)
	s.False(diff.SandboxChanged)
	s.True(diff.SandboxBuildChanged)
	s.NotEqual(diff.SandboxBuildOldID, diff.SandboxBuildNewID)
}

func (s *AppBranchConfigDiffTestSuite) sandboxBuild(appID, appConfigID, sandboxConfigID string) {
	identity := s.sourceIdentity("sandbox")
	build := app.AppSandboxBuild{
		AppID:              appID,
		AppConfigID:        appConfigID,
		AppSandboxConfigID: sandboxConfigID,
		Status:             app.AppSandboxBuildStatusActive,
		StatusDescription:  "active",
		SourceIdentity:     identity,
		SourceChecksum:     identity.Checksum,
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&build).Error)
}

func (s *AppBranchConfigDiffTestSuite) TestComponentConfig() {
	f := s.seedPair()
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.ComponentConfigConnection{}).
		Where("app_config_id = ? AND component_id = ?", f.oldID, f.componentID).
		Update("checksum", "checksum-old").Error)

	diff := s.diff(f)
	s.False(diff.StackChanged)
	s.False(diff.SandboxChanged)
	s.Equal([]string{f.componentID}, s.componentIDs(diff.Changed))
	s.False(diff.Changed[0].BuildChanged)
}

func (s *AppBranchConfigDiffTestSuite) TestComponentSource() {
	f := s.seedPair()
	oldConnID, _ := s.latestBuild(f.oldID, f.componentID)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.ComponentConfigConnection{ID: oldConnID}).
		Update("latest_build_id", s.componentBuild(oldConnID, "build-same")).Error)

	diff := s.diff(f)
	s.Empty(diff.Changed)
	s.ElementsMatch([]string{f.componentID, f.workerID}, s.componentIDs(diff.Unchanged))
	for _, entry := range diff.Unchanged {
		if entry.ComponentID == f.componentID {
			s.True(entry.BuildChanged)
		}
	}
	s.False(diff.StackChanged)
}

func (s *AppBranchConfigDiffTestSuite) TestComponentAdded() {
	f := s.seedPair()
	billing := s.deps.Seed.CreateComponent(s.ctx, s.T(), f.appID, app.ComponentTypeTerraformModule)
	f.billingID = billing.ID
	s.connection(billing.ID, f.newID, "billing", "billing-build")

	diff := s.diff(f)
	s.Equal([]string{billing.ID}, s.componentIDs(diff.Added))
	s.Empty(diff.Changed)
	s.Empty(diff.Removed)
	s.False(diff.StackChanged)
	s.False(diff.SandboxChanged)
}

func (s *AppBranchConfigDiffTestSuite) TestComponentRemoved() {
	f := s.seedPair()
	legacy := s.deps.Seed.CreateComponent(s.ctx, s.T(), f.appID, app.ComponentTypeTerraformModule)
	s.connection(legacy.ID, f.oldID, "legacy", "legacy-build")

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

	diff := s.diff(f)
	s.False(diff.StackChanged)
	s.True(diff.SandboxChanged)
	s.Equal([]string{f.componentID}, s.componentIDs(diff.Changed))
	s.NotContains(s.componentIDs(diff.Changed), f.workerID)
	s.NotContains(s.componentIDs(diff.Added), f.workerID)
}

func (s *AppBranchConfigDiffTestSuite) TestNeverApplied() {
	f := s.seedPair()

	diff := s.diff(f)
	s.False(diff.StackChanged)
	s.False(diff.SandboxChanged)
	s.Empty(diff.Added)
	s.Empty(diff.Changed)
	s.Empty(diff.Removed)
	s.ElementsMatch([]string{f.componentID, f.workerID}, s.componentIDs(diff.Unchanged))
}

type updateFixture struct {
	appID    string
	install  *app.Install
	producer string
	consumer string
	initial  string
}

func (s *AppBranchConfigDiffTestSuite) seedUpdates() updateFixture {
	a := s.deps.Seed.CreateApp(s.ctx, s.T())
	producer := s.deps.Seed.CreateComponent(s.ctx, s.T(), a.ID, app.ComponentTypeTerraformModule)
	consumer := s.deps.Seed.CreateComponent(s.ctx, s.T(), a.ID, app.ComponentTypeTerraformModule)
	f := updateFixture{appID: a.ID, producer: producer.ID, consumer: consumer.ID}
	f.initial = s.branchConfig(f, "1", "1")
	f.install = s.deps.Seed.CreateInstall(s.ctx, s.T(), a)
	return f
}

func (s *AppBranchConfigDiffTestSuite) branchConfig(f updateFixture, producer, consumer string) string {
	cfg := s.deps.Seed.CreateBareAppConfig(s.ctx, s.T(), f.appID)
	s.deps.Seed.CreateAppSandboxConfig(s.ctx, s.T(), f.appID, cfg.ID)
	s.deps.Seed.CreateAppStackConfig(s.ctx, s.T(), f.appID, cfg.ID)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.AppStackConfig{}).
		Where(app.AppStackConfig{AppConfigID: cfg.ID}).
		Updates(map[string]any{"name": "shared-stack", "description": "shared"}).Error)
	s.deps.Seed.CreateAppRunnerConfig(s.ctx, s.T(), f.appID, cfg.ID)
	s.connection(f.producer, cfg.ID, "producer-"+producer, "producer-build-"+producer)
	s.connection(f.consumer, cfg.ID, "consumer-"+consumer, "consumer-build-"+consumer)
	return cfg.ID
}

func (s *AppBranchConfigDiffTestSuite) configUpdate(f updateFixture, appConfigID string, started ...app.WorkflowStepTargetType) {
	wf := s.deps.Seed.CreateWorkflow(s.ctx, s.T(), f.install.ID, app.WorkflowTypeAppBranchConfigUpdate)
	pending := s.deps.Seed.CreateWorkflowStep(s.ctx, s.T(), wf.ID)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(pending).
		Update("step_target_type", string(app.WorkflowStepTargetTypeInstallDeploys)).Error)
	for _, target := range started {
		step := s.deps.Seed.CreateWorkflowStep(s.ctx, s.T(), wf.ID)
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(step).Updates(map[string]any{
			"step_target_type": string(target),
			"started_at":       time.Now(),
		}).Error)
	}
	version := app.InstallAppConfigVersion{InstallID: f.install.ID, NewAppConfigID: appConfigID, WorkflowID: &wf.ID}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&version).Error)
}

func (s *AppBranchConfigDiffTestSuite) changedSince(f updateFixture, newAppConfigID string) []string {
	diff, err := s.deps.Helpers.AppBranchConfigDiff(s.ctx, f.install, newAppConfigID)
	require.NoError(s.T(), err)
	s.Empty(diff.Added)
	s.Empty(diff.Removed)
	return s.componentIDs(diff.Changed)
}

func (s *AppBranchConfigDiffTestSuite) TestDiffsFromTheDesiredConfigWhateverDeployed() {
	f := s.seedUpdates()
	s.configUpdate(f, f.initial, app.WorkflowStepTargetTypeInstallDeploys)
	skipped := s.branchConfig(f, "2", "1")
	s.configUpdate(f, skipped)
	s.pin(f.install, skipped)

	s.Equal([]string{f.consumer}, s.changedSince(f, s.branchConfig(f, "2", "2")))
}

func (s *AppBranchConfigDiffTestSuite) TestComponentChangedAgainRedeploys() {
	f := s.seedUpdates()
	s.pin(f.install, s.branchConfig(f, "2", "1"))

	s.Equal([]string{f.producer}, s.changedSince(f, s.branchConfig(f, "3", "1")))
}
