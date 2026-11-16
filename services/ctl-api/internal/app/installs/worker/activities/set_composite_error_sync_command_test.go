package activities

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type compositeErrorSyncCommandTestSuite struct {
	tests.BaseDBTestSuite

	db     *gorm.DB
	seeder *testseed.Seeder
}

func TestCompositeErrorSyncCommandSuite(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	suite.Run(t, new(compositeErrorSyncCommandTestSuite))
}

func (s *compositeErrorSyncCommandTestSuite) SetupSuite() {
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

func (s *compositeErrorSyncCommandTestSuite) TearDownSuite() {
	db, err := s.db.DB()
	require.NoError(s.T(), err)
	require.NoError(s.T(), db.Close())
}

func (s *compositeErrorSyncCommandTestSuite) seedInstall(appSyncDisabled bool) (context.Context, *app.Install, *app.AppConfig) {
	t := s.T()
	ctx, _ := s.seeder.EnsureAccount(context.Background(), t)
	ctx, testOrg := s.seeder.EnsureOrg(ctx, t)

	a := s.seeder.CreateApp(ctx, t)
	cfg := s.seeder.CreateAppConfig(ctx, t, a.ID)
	install := s.seeder.CreateInstall(ctx, t, a)
	if appSyncDisabled {
		require.NoError(t, s.db.WithContext(ctx).
			Model(&app.Org{ID: testOrg.ID}).
			Update("features", types.StringBoolMap{string(app.OrgFeatureDisableAppSync): true}).Error)
	}
	return ctx, install, cfg
}

func (s *compositeErrorSyncCommandTestSuite) activityEnv(ctx context.Context) *testsuite.TestActivityEnvironment {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestActivityEnvironment()
	env.SetWorkerOptions(worker.Options{BackgroundActivityContext: ctx})
	return env
}

func howToFix(data *compositeerrors.CompositeErrorData) string {
	for _, section := range data.Sections {
		if section.Heading == "How to fix" {
			return section.Body
		}
	}
	return ""
}

func (s *compositeErrorSyncCommandTestSuite) deployPlanHowToFix(appSyncDisabled bool) string {
	t := s.T()
	ctx, install, cfg := s.seedInstall(appSyncDisabled)
	comp := s.seeder.CreateComponent(ctx, t, install.AppID, app.ComponentTypeHelmChart)
	ccc := s.seeder.CreateHelmComponentConfigConnection(ctx, t, comp.ID, cfg.ID)
	ic := s.seeder.CreateInstallComponent(ctx, t, install.ID, comp.ID)
	bld := s.seeder.CreateComponentBuild(ctx, t, ccc.ID)
	deploy := s.seeder.CreateInstallDeploy(ctx, t, ic.ID, bld.ID)

	acts := &Activities{db: s.db, features: features.New(features.Params{DB: s.db})}
	env := s.activityEnv(ctx)
	env.RegisterActivity(acts.SetInstallDeployPlanCompositeError)
	_, err := env.ExecuteActivity(acts.SetInstallDeployPlanCompositeError, SetInstallDeployPlanCompositeErrorRequest{
		InstallDeployID: deploy.ID,
		Detail:          "missing output",
	})
	require.NoError(t, err)

	var got app.InstallDeploy
	require.NoError(t, s.db.WithContext(ctx).First(&got, "id = ?", deploy.ID).Error)
	require.NotNil(t, got.CompositeError)
	return howToFix(got.CompositeError)
}

func (s *compositeErrorSyncCommandTestSuite) stackVersionHowToFix(appSyncDisabled bool) string {
	t := s.T()
	ctx, install, cfg := s.seedInstall(appSyncDisabled)
	version := s.seeder.CreateInstallStackVersion(ctx, t, install.ID, install.InstallStack.ID, cfg.ID)

	acts := &Activities{db: s.db, features: features.New(features.Params{DB: s.db})}
	env := s.activityEnv(ctx)
	env.RegisterActivity(acts.SetInstallStackVersionCompositeError)
	_, err := env.ExecuteActivity(acts.SetInstallStackVersionCompositeError, SetInstallStackVersionCompositeErrorRequest{
		StackVersionID: version.ID,
		Platform:       "aws",
		Detail:         "bad template",
	})
	require.NoError(t, err)

	var got app.InstallStackVersion
	require.NoError(t, s.db.WithContext(ctx).First(&got, "id = ?", version.ID).Error)
	require.NotNil(t, got.CompositeError)
	return howToFix(got.CompositeError)
}

func (s *compositeErrorSyncCommandTestSuite) TestDeployPlanErrorWithAppSyncEnabledSuggestsAppsSync() {
	require.Equal(s.T(), "Check the referenced value exists in your app config — a missing sandbox output, action workflow output, or input is the usual cause. Fix the reference, re-run `nuon apps sync` or run your app branch with the fix, then retry the deploy.", s.deployPlanHowToFix(false))
}

func (s *compositeErrorSyncCommandTestSuite) TestDeployPlanErrorWithAppSyncDisabledSuggestsAppBranch() {
	require.Equal(s.T(), "Check the referenced value exists in your app config — a missing sandbox output, action workflow output, or input is the usual cause. Fix the reference, run your app branch with the fix, then retry the deploy.", s.deployPlanHowToFix(true))
}

func (s *compositeErrorSyncCommandTestSuite) TestStackVersionErrorWithAppSyncEnabledSuggestsAppsSync() {
	require.Equal(s.T(), "Fix the app stack config, then re-run `nuon apps sync` or run your app branch with the fix to regenerate the template.", s.stackVersionHowToFix(false))
}

func (s *compositeErrorSyncCommandTestSuite) TestStackVersionErrorWithAppSyncDisabledSuggestsAppBranch() {
	require.Equal(s.T(), "Fix the app stack config, then run your app branch with the fix to regenerate the template.", s.stackVersionHowToFix(true))
}
