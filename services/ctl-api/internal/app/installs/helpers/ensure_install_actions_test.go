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

type ensureInstallActionsDeps struct {
	fx.In

	DB      *gorm.DB `name:"psql"`
	Seed    *testseed.Seeder
	Helpers *installhelpers.Helpers
}

type EnsureInstallActionsTestSuite struct {
	tests.BaseDBTestSuite

	fxApp *fxtest.App
	deps  ensureInstallActionsDeps
	ctx   context.Context

	install   *app.Install
	newConfig *app.AppConfig
	kept      *app.ActionWorkflow
	oldOnly   *app.ActionWorkflow
	added     *app.ActionWorkflow
}

func TestEnsureInstallActionsSuite(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	suite.Run(t, new(EnsureInstallActionsTestSuite))
}

func (s *EnsureInstallActionsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	options := append(tests.CtlApiFXOptions(s.T()), fx.Populate(&s.deps))
	s.fxApp = fxtest.New(s.T(), options...)
	s.fxApp.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *EnsureInstallActionsTestSuite) TearDownSuite() {
	s.fxApp.RequireStop()
}

// SetupTest leaves the install on an old config with kept and oldOnly, and a
// new config with kept and added.
func (s *EnsureInstallActionsTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.ctx = context.Background()
	s.ctx, _ = s.deps.Seed.EnsureAccount(s.ctx, s.T())
	s.ctx, _ = s.deps.Seed.EnsureOrg(s.ctx, s.T())
	a := s.deps.Seed.CreateApp(s.ctx, s.T())

	// CreateAppConfig seeds one action workflow, which becomes oldOnly.
	oldConfig := s.deps.Seed.CreateAppConfig(s.ctx, s.T(), a.ID)
	var seeded app.ActionWorkflowConfig
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Where(app.ActionWorkflowConfig{AppConfigID: oldConfig.ID}).
		First(&seeded).Error)
	s.oldOnly = &app.ActionWorkflow{ID: seeded.ActionWorkflowID}

	s.kept = s.deps.Seed.CreateActionWorkflow(s.ctx, s.T(), a.ID)
	s.added = s.deps.Seed.CreateActionWorkflow(s.ctx, s.T(), a.ID)
	s.deps.Seed.CreateActionWorkflowConfig(s.ctx, s.T(), a.ID, oldConfig.ID, s.kept.ID)
	s.install = s.deps.Seed.CreateInstall(s.ctx, s.T(), a)
	require.Equal(s.T(), oldConfig.ID, s.install.AppConfigID)
	require.NoError(s.T(), s.deps.Helpers.ReconcileInstallActions(s.ctx, s.install.ID))

	s.newConfig = s.deps.Seed.CreateBareAppConfig(s.ctx, s.T(), a.ID)
	for _, aw := range []*app.ActionWorkflow{s.kept, s.added} {
		s.deps.Seed.CreateActionWorkflowConfig(s.ctx, s.T(), a.ID, s.newConfig.ID, aw.ID)
	}
}

func (s *EnsureInstallActionsTestSuite) installActions() map[string]app.InstallActionWorkflow {
	var rows []app.InstallActionWorkflow
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Where(app.InstallActionWorkflow{InstallID: s.install.ID}).
		Find(&rows).Error)
	byAction := make(map[string]app.InstallActionWorkflow, len(rows))
	for _, row := range rows {
		byAction[row.ActionWorkflowID] = row
	}
	return byAction
}

func removed(iaw app.InstallActionWorkflow) bool {
	_, ok := iaw.StatusV2.Metadata["removed_at_app_config_id"]
	return ok
}

// The install is still on its old config, so the old config's actions must be
// left alone.
func (s *EnsureInstallActionsTestSuite) TestCreatesOnlyMissingActions() {
	before := s.installActions()
	require.NoError(s.T(), s.deps.Helpers.EnsureInstallActions(s.ctx, s.install.ID, s.newConfig.ID))

	got := s.installActions()
	s.Require().Len(got, 3)
	s.Equal(before[s.kept.ID].ID, got[s.kept.ID].ID)
	s.Contains(got, s.added.ID)
	s.False(removed(got[s.oldOnly.ID]))
}

func (s *EnsureInstallActionsTestSuite) TestIsIdempotent() {
	for range 2 {
		require.NoError(s.T(), s.deps.Helpers.EnsureInstallActions(s.ctx, s.install.ID, s.newConfig.ID))
	}
	s.Len(s.installActions(), 3)
}

// Once the install moves to the new config, reconcile keeps the row created
// ahead of time and marks the old config's action removed.
func (s *EnsureInstallActionsTestSuite) TestReconcileAfterMovingToNewConfig() {
	require.NoError(s.T(), s.deps.Helpers.EnsureInstallActions(s.ctx, s.install.ID, s.newConfig.ID))
	added := s.installActions()[s.added.ID]

	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Model(&app.Install{ID: s.install.ID}).
		Update("app_config_id", s.newConfig.ID).Error)
	require.NoError(s.T(), s.deps.Helpers.ReconcileInstallActions(s.ctx, s.install.ID))

	got := s.installActions()
	s.Equal(added.ID, got[s.added.ID].ID)
	s.False(removed(got[s.added.ID]))
	s.False(removed(got[s.kept.ID]))
	s.True(removed(got[s.oldOnly.ID]))
}
