package installs

import (
	"context"
	"os"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	tclient "go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/metrics"
	temporal "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// syncBranchLabelsMockWorkflowRun implements tclient.WorkflowRun for mock queue creation.
type syncBranchLabelsMockWorkflowRun struct{}

func (m *syncBranchLabelsMockWorkflowRun) GetID() string    { return "mock-workflow-id" }
func (m *syncBranchLabelsMockWorkflowRun) GetRunID() string { return "mock-run-id" }
func (m *syncBranchLabelsMockWorkflowRun) Get(ctx context.Context, valuePtr interface{}) error {
	return nil
}
func (m *syncBranchLabelsMockWorkflowRun) GetWithOptions(ctx context.Context, valuePtr interface{}, options tclient.WorkflowRunGetOptions) error {
	return nil
}

type syncBranchLabelsDeps struct {
	fx.In

	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	MW             metrics.Writer
	InstallHelpers *installhelpers.Helpers
	Seeder         *testseed.Seeder
}

type SyncInstallBranchLabelsTestSuite struct {
	tests.BaseDBTestSuite

	fxApp         *fxtest.App
	deps          syncBranchLabelsDeps
	ctx           context.Context
	testOrg       *app.Org
	testAcc       *app.Account
	testApp       *app.App
	testAppConfig *app.AppConfig
	ctrl          *gomock.Controller
	mockTC        *temporal.MockClient
}

func TestSyncInstallBranchLabelsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(SyncInstallBranchLabelsTestSuite))
}

func (s *SyncInstallBranchLabelsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	s.ctrl = gomock.NewController(s.T())
	s.mockTC = temporal.NewMockClient(s.ctrl)

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T: s.T(),
			Mocks: &tests.TestMocks{
				MockTC: s.mockTC,
			},
			CustomValidator: true,
		}),
		fx.Populate(&s.deps),
	)

	s.fxApp = fxtest.New(s.T(), options...)
	s.fxApp.RequireStart()
	s.SetDB(s.deps.DB)
}

// expectQueueCreation sets up the mock expectation for queue creation via temporal.
func (s *SyncInstallBranchLabelsTestSuite) expectQueueCreation() {
	s.mockTC.EXPECT().ExecuteWorkflowInNamespace(
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Return(&syncBranchLabelsMockWorkflowRun{}, nil).AnyTimes()
}

func (s *SyncInstallBranchLabelsTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()

	s.ctx = context.Background()
	s.ctx, s.testAcc = s.deps.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.deps.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.deps.Seeder.CreateApp(s.ctx, s.T())
	s.testAppConfig = s.deps.Seeder.CreateAppConfig(s.ctx, s.T(), s.testApp.ID)
}

func (s *SyncInstallBranchLabelsTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *SyncInstallBranchLabelsTestSuite) TearDownSuite() {
	s.fxApp.RequireStop()
}

func (s *SyncInstallBranchLabelsTestSuite) setDisableAppSync(enabled bool) {
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.Org{}).
		Where(app.Org{ID: s.testOrg.ID}).
		Update("features", map[string]any{string(app.OrgFeatureDisableAppSync): enabled}).Error)
}

func (s *SyncInstallBranchLabelsTestSuite) seedBranch(name string, groups []app.AppBranchInstallGroup) *app.AppBranch {
	branch := &app.AppBranch{AppID: s.testApp.ID, Name: name}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(branch).Error)

	cfg := s.deps.Seeder.CreateAppConfig(s.ctx, s.T(), s.testApp.ID)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(cfg).
		Updates(map[string]any{"app_branch_id": branch.ID, "builds_pinned": true, "builds_locked": true}).Error)

	for i := range groups {
		groups[i].Order = i
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&app.AppBranchConfig{
		AppBranchID:   branch.ID,
		InstallGroups: groups,
	}).Error)
	return branch
}

func (s *SyncInstallBranchLabelsTestSuite) getInstallByName(name string) *app.Install {
	var install app.Install
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Where(app.Install{AppID: s.testApp.ID, Name: name}).First(&install).Error)
	return &install
}

func (s *SyncInstallBranchLabelsTestSuite) TestSyncCreatesOnMatchedBranchWithAppSyncEnabled() {
	s.expectQueueCreation()
	branch := s.seedBranch("prod", []app.AppBranchInstallGroup{{
		Name:          "prod",
		LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}},
	}})

	installCfg := &config.Install{
		Name:       "synced-matches-branch",
		Labels:     map[string]string{"env": "prod"},
		AWSAccount: &config.AWSAccount{Region: "us-west-2"},
	}
	_, err := SyncInstall(s.ctx, s.deps.DB, s.deps.InstallHelpers, s.testApp.ID, installCfg)
	require.NoError(s.T(), err)

	install := s.getInstallByName("synced-matches-branch")
	assert.True(s.T(), install.AppBranchID.Valid)
	assert.Equal(s.T(), branch.ID, install.AppBranchID.String)
}

func (s *SyncInstallBranchLabelsTestSuite) TestSyncCreatesOnMatchedBranchWithAppSyncDisabled() {
	s.expectQueueCreation()
	s.setDisableAppSync(true)
	branch := s.seedBranch("prod", []app.AppBranchInstallGroup{{
		Name:          "prod",
		LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}},
	}})

	installCfg := &config.Install{
		Name:       "disabled-sync-matches-branch",
		Labels:     map[string]string{"env": "prod"},
		AWSAccount: &config.AWSAccount{Region: "us-west-2"},
	}
	_, err := SyncInstall(s.ctx, s.deps.DB, s.deps.InstallHelpers, s.testApp.ID, installCfg)
	require.NoError(s.T(), err)

	install := s.getInstallByName("disabled-sync-matches-branch")
	assert.True(s.T(), install.AppBranchID.Valid)
	assert.Equal(s.T(), branch.ID, install.AppBranchID.String)
}

func (s *SyncInstallBranchLabelsTestSuite) TestSyncWithAppSyncDisabledRejectsInstallWithNoMatchingBranch() {
	s.setDisableAppSync(true)

	installCfg := &config.Install{
		Name:       "disabled-sync-no-branch",
		AWSAccount: &config.AWSAccount{Region: "us-west-2"},
	}
	_, err := SyncInstall(s.ctx, s.deps.DB, s.deps.InstallHelpers, s.testApp.ID, installCfg)
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "has no app_branch and its labels")
}

func (s *SyncInstallBranchLabelsTestSuite) TestSyncDoesNotMoveExistingOffBranchInstallWhenLabelsStartMatching() {
	s.expectQueueCreation()

	installCfg := &config.Install{
		Name:       "existing-off-branch-stays",
		AWSAccount: &config.AWSAccount{Region: "us-west-2"},
	}
	_, err := SyncInstall(s.ctx, s.deps.DB, s.deps.InstallHelpers, s.testApp.ID, installCfg)
	require.NoError(s.T(), err)

	install := s.getInstallByName("existing-off-branch-stays")
	require.False(s.T(), install.AppBranchID.Valid)

	s.seedBranch("prod", []app.AppBranchInstallGroup{{
		Name:          "prod",
		LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}},
	}})

	installCfg = &config.Install{
		Name:       "existing-off-branch-stays",
		Labels:     map[string]string{"env": "prod"},
		AWSAccount: &config.AWSAccount{Region: "us-west-2"},
	}
	_, err = SyncInstall(s.ctx, s.deps.DB, s.deps.InstallHelpers, s.testApp.ID, installCfg)
	require.NoError(s.T(), err)

	install = s.getInstallByName("existing-off-branch-stays")
	assert.False(s.T(), install.AppBranchID.Valid)
}

func (s *SyncInstallBranchLabelsTestSuite) TestSyncKeepsExistingInstallOnItsBranchWhenTOMLOmitsAppBranch() {
	s.expectQueueCreation()
	branch := s.seedBranch("prod", []app.AppBranchInstallGroup{{
		Name: "default",
		LabelSelector: &labels.Selector{MatchLabels: labels.Labels{
			"keep": "me",
		}},
		Default: true,
	}})

	installCfg := &config.Install{
		Name:       "existing-on-branch-stays",
		AppBranch:  branch.Name,
		Labels:     map[string]string{"keep": "me"},
		AWSAccount: &config.AWSAccount{Region: "us-west-2"},
	}
	_, err := SyncInstall(s.ctx, s.deps.DB, s.deps.InstallHelpers, s.testApp.ID, installCfg)
	require.NoError(s.T(), err)

	install := s.getInstallByName("existing-on-branch-stays")
	require.True(s.T(), install.AppBranchID.Valid)
	require.Equal(s.T(), branch.ID, install.AppBranchID.String)

	installCfg = &config.Install{
		Name:       "existing-on-branch-stays",
		Labels:     map[string]string{"keep": "me", "extra": "value"},
		AWSAccount: &config.AWSAccount{Region: "us-west-2"},
	}
	_, err = SyncInstall(s.ctx, s.deps.DB, s.deps.InstallHelpers, s.testApp.ID, installCfg)
	require.NoError(s.T(), err)

	install = s.getInstallByName("existing-on-branch-stays")
	assert.True(s.T(), install.AppBranchID.Valid)
	assert.Equal(s.T(), branch.ID, install.AppBranchID.String)
}

func (s *SyncInstallBranchLabelsTestSuite) TestSyncWithAppSyncDisabledRejectsInstallMatchingTwoBranches() {
	s.setDisableAppSync(true)
	branchA := s.seedBranch("prod", []app.AppBranchInstallGroup{{
		Name:          "prod",
		LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}},
	}})
	branchB := s.seedBranch("acme", []app.AppBranchInstallGroup{{
		Name:          "acme",
		LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"customer": "acme"}},
	}})

	installCfg := &config.Install{
		Name:       "disabled-sync-ambiguous",
		Labels:     map[string]string{"env": "prod", "customer": "acme"},
		AWSAccount: &config.AWSAccount{Region: "us-west-2"},
	}
	_, err := SyncInstall(s.ctx, s.deps.DB, s.deps.InstallHelpers, s.testApp.ID, installCfg)
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), branchA.Name)
	assert.Contains(s.T(), err.Error(), branchB.Name)
}
