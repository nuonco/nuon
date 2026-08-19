package syncer

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	testseedconfig "github.com/nuonco/nuon/services/ctl-api/tests/testseed/config"
)

// syncWithBuildDispatch runs a sync over cfg into a fresh app config on an
// existing app, the way the CLI path does.
func (s *SyncFieldsTestSuite) syncWithBuildDispatch(ctx context.Context, appID string, cfg *config.AppConfig) *app.AppConfig {
	appCfg := s.deps.Seed.CreateBareAppConfig(ctx, s.T(), appID)

	err := s.deps.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		syncer := NewDBSyncer(
			tx,
			s.deps.AppsHelpers,
			s.deps.ComponentHelpers,
			s.deps.ActionsHelpers,
			s.deps.RunbooksHelpers,
			s.deps.InstallHelpers,
			s.deps.VCSHelpers,
			s.deps.TFClient,
			appID,
			cfg,
			appCfg.ID,
			WithComponentBuildDispatch(),
		)
		if err := syncer.Sync(ctx); err != nil {
			return err
		}
		s.scheduled = syncer.GetComponentsScheduled()
		return nil
	})
	s.Require().NoError(err, "sync should succeed")

	return appCfg
}

// markBuilt gives the component's latest config connection a non-failed build,
// which is what makes it reusable on the next sync.
func (s *SyncFieldsTestSuite) markBuilt(ctx context.Context, componentID string) {
	var ccc app.ComponentConfigConnection
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.ComponentConfigConnection{ComponentID: componentID}).
		Order("created_at DESC").
		First(&ccc).Error)

	bld := s.deps.Seed.CreateComponentBuild(ctx, s.T(), ccc.ID)
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Model(&app.ComponentConfigConnection{}).
		Where("id = ?", ccc.ID).
		Update("latest_build_id", bld.ID).Error)
}

func (s *SyncFieldsTestSuite) componentConfigCount(ctx context.Context, componentID string) int64 {
	var count int64
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Model(&app.ComponentConfigConnection{}).
		Where(app.ComponentConfigConnection{ComponentID: componentID}).
		Count(&count).Error)
	return count
}

// terraformComponent pins a concrete terraform version — "latest" is only
// resolvable by reaching the real terraform registry.
func terraformComponent(name string) *config.Component {
	cmp := testseedconfig.BuildTerraformComponent(name)
	cmp.TerraformModule.TerraformVersion = "1.9.0"
	return cmp
}

func (s *SyncFieldsTestSuite) componentID(ctx context.Context, appID, name string) string {
	var cmp app.Component
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.Component{AppID: appID, Name: name}).
		First(&cmp).Error)
	return cmp.ID
}

// A component whose checksum has not moved must not get a new config connection
// or a build. Without this the CLI would rebuild every component on every sync.
func (s *SyncFieldsTestSuite) TestBuildDispatchReusesUnchangedComponent() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	cmp := terraformComponent("unchanged-component")
	cfg.Components = config.ComponentList{cmp}

	ctx, testApp, _ := s.syncEmpty()

	s.syncWithBuildDispatch(ctx, testApp.ID, cfg)
	cmpID := s.componentID(ctx, testApp.ID, "unchanged-component")
	s.Len(s.scheduled, 1, "first sync must schedule the component")
	s.Equal(int64(1), s.componentConfigCount(ctx, cmpID))

	s.markBuilt(ctx, cmpID)

	s.syncWithBuildDispatch(ctx, testApp.ID, cfg)
	s.Empty(s.scheduled, "unchanged component must not be scheduled again")
	s.Equal(int64(1), s.componentConfigCount(ctx, cmpID),
		"unchanged component must reuse its existing config connection")
}

// A checksum change must produce a fresh config connection and a scheduled
// build, even though the previous config connection has a healthy build.
func (s *SyncFieldsTestSuite) TestBuildDispatchSchedulesChangedComponent() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	cmp := terraformComponent("changed-component")
	cfg.Components = config.ComponentList{cmp}

	ctx, testApp, _ := s.syncEmpty()

	s.syncWithBuildDispatch(ctx, testApp.ID, cfg)
	cmpID := s.componentID(ctx, testApp.ID, "changed-component")
	s.markBuilt(ctx, cmpID)

	cmp.TerraformModule.EnvVarMap["CHANGED"] = "yes"
	s.syncWithBuildDispatch(ctx, testApp.ID, cfg)

	s.Len(s.scheduled, 1, "changed component must be scheduled")
	s.Equal(cmpID, s.scheduled[0].ID)
	s.Equal(int64(2), s.componentConfigCount(ctx, cmpID),
		"changed component must get a fresh config connection")
}

// A component whose last build failed must be rebuilt even when its checksum is
// unchanged, otherwise a transient build failure is unrecoverable by re-syncing.
func (s *SyncFieldsTestSuite) TestBuildDispatchRetriesFailedBuild() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	cmp := terraformComponent("failed-component")
	cfg.Components = config.ComponentList{cmp}

	ctx, testApp, _ := s.syncEmpty()

	s.syncWithBuildDispatch(ctx, testApp.ID, cfg)
	cmpID := s.componentID(ctx, testApp.ID, "failed-component")
	s.markBuilt(ctx, cmpID)

	var ccc app.ComponentConfigConnection
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.ComponentConfigConnection{ComponentID: cmpID}).
		Order("created_at DESC").
		First(&ccc).Error)
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Model(&app.ComponentBuild{}).
		Where("id = ?", ccc.LatestBuildID.String).
		Update("status", "error").Error)

	s.syncWithBuildDispatch(ctx, testApp.ID, cfg)

	s.Len(s.scheduled, 1, "component with a failed build must be scheduled again")
	s.Equal(int64(2), s.componentConfigCount(ctx, cmpID))
}

// Branch sync leaves build dispatch off and must keep its always-fresh config
// connection behaviour, which strict config-connection pinning relies on.
func (s *SyncFieldsTestSuite) TestWithoutBuildDispatchAlwaysCreatesFreshConfig() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	cmp := terraformComponent("branch-component")
	cfg.Components = config.ComponentList{cmp}

	ctx, testApp, _ := s.syncEmpty()

	s.syncInto(ctx, testApp.ID, cfg)
	cmpID := s.componentID(ctx, testApp.ID, "branch-component")
	s.markBuilt(ctx, cmpID)

	s.syncInto(ctx, testApp.ID, cfg)

	s.Equal(int64(2), s.componentConfigCount(ctx, cmpID),
		"branch sync must create a config connection per sync")
}

// Branch sync pre-creates exactly one queued build per fresh CCC for image
// components, which the branch run's builds step adopts and executes via
// queuebuild — never a duplicate.
func (s *SyncFieldsTestSuite) TestWithoutBuildDispatchPrecreatesOneImageBuild() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	cmp := testseedconfig.BuildExternalImageComponent("plain-image")
	cfg.Components = config.ComponentList{cmp}

	ctx, testApp, _ := s.syncEmpty()

	appCfg := s.syncInto(ctx, testApp.ID, cfg)
	cmpID := s.componentID(ctx, testApp.ID, "plain-image")

	var ccc app.ComponentConfigConnection
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.ComponentConfigConnection{ComponentID: cmpID, AppConfigID: appCfg.ID}).
		First(&ccc).Error)

	var builds []app.ComponentBuild
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.ComponentBuild{ComponentConfigConnectionID: ccc.ID}).
		Find(&builds).Error)
	s.Require().Len(builds, 1, "branch sync must pre-create exactly one build")
	s.Equal(app.ComponentBuildStatusQueued, builds[0].Status)
	s.Require().True(ccc.LatestBuildID.Valid)
	s.Equal(builds[0].ID, ccc.LatestBuildID.String)
}

// An unchanged image component whose previous build is Active must not get a
// new build on re-sync — the fresh CCC is pinned to the previous Active build.
func (s *SyncFieldsTestSuite) TestWithoutBuildDispatchReusesActiveImageBuild() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	cmp := testseedconfig.BuildExternalImageComponent("stable-image")
	cfg.Components = config.ComponentList{cmp}

	ctx, testApp, _ := s.syncEmpty()

	s.syncInto(ctx, testApp.ID, cfg)
	cmpID := s.componentID(ctx, testApp.ID, "stable-image")

	var firstBuild app.ComponentBuild
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Joins("JOIN component_config_connections ccc ON ccc.id = component_builds.component_config_connection_id").
		Where("ccc.component_id = ?", cmpID).
		First(&firstBuild).Error)
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Model(&app.ComponentBuild{}).
		Where("id = ?", firstBuild.ID).
		Update("status", app.ComponentBuildStatusActive).Error)

	secondCfg := s.syncInto(ctx, testApp.ID, cfg)

	var secondCCC app.ComponentConfigConnection
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Where(app.ComponentConfigConnection{ComponentID: cmpID, AppConfigID: secondCfg.ID}).
		First(&secondCCC).Error)

	var count int64
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Model(&app.ComponentBuild{}).
		Joins("JOIN component_config_connections ccc ON ccc.id = component_builds.component_config_connection_id").
		Where("ccc.component_id = ?", cmpID).
		Count(&count).Error)
	s.Equal(int64(1), count, "unchanged image must not get a second build")
	s.Require().True(secondCCC.LatestBuildID.Valid)
	s.Equal(firstBuild.ID, secondCCC.LatestBuildID.String,
		"fresh CCC must be pinned to the previous Active build")
}

// An update_policy image resolves tags at build time, so it must pre-create a
// fresh build every branch sync even when the config is unchanged.
func (s *SyncFieldsTestSuite) TestWithoutBuildDispatchAlwaysBuildsUpdatePolicyImage() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	cmp := testseedconfig.BuildExternalImageComponent("tracked-branch-image")
	cmp.ExternalImage.PublicImageConfig.UpdatePolicy = ">= 1.0.0"
	cfg.Components = config.ComponentList{cmp}

	ctx, testApp, _ := s.syncEmpty()

	s.syncInto(ctx, testApp.ID, cfg)
	cmpID := s.componentID(ctx, testApp.ID, "tracked-branch-image")
	s.markBuilt(ctx, cmpID)

	s.syncInto(ctx, testApp.ID, cfg)

	var count int64
	s.Require().NoError(s.deps.DB.WithContext(ctx).
		Model(&app.ComponentBuild{}).
		Joins("JOIN component_config_connections ccc ON ccc.id = component_builds.component_config_connection_id").
		Where("ccc.component_id = ?", cmpID).
		Count(&count).Error)
	s.GreaterOrEqual(count, int64(2), "update_policy image must pre-create a build per sync")
}

// An external image with an update_policy resolves its tag against the registry
// at build time, so an unchanged config does not mean an unchanged artifact. It
// must rebuild every sync or installs silently stop picking up new tags.
func (s *SyncFieldsTestSuite) TestBuildDispatchAlwaysRebuildsUpdatePolicyImage() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	cmp := testseedconfig.BuildExternalImageComponent("tracked-image")
	cmp.ExternalImage.PublicImageConfig.UpdatePolicy = ">= 1.0.0"
	cfg.Components = config.ComponentList{cmp}

	ctx, testApp, _ := s.syncEmpty()

	s.syncWithBuildDispatch(ctx, testApp.ID, cfg)
	cmpID := s.componentID(ctx, testApp.ID, "tracked-image")
	s.markBuilt(ctx, cmpID)

	s.syncWithBuildDispatch(ctx, testApp.ID, cfg)

	s.Len(s.scheduled, 1, "update_policy image must rebuild despite unchanged config")
	s.Equal(int64(2), s.componentConfigCount(ctx, cmpID))
}
