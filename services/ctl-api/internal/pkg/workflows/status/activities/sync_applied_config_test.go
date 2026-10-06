package statusactivities_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

type syncAppliedConfigSuite struct {
	tests.BaseDBTestSuite

	fxApp *fxtest.App
	deps  transitionRunnerStatusDeps
	ctx   context.Context
}

func TestSyncAppliedConfigSuite(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	suite.Run(t, new(syncAppliedConfigSuite))
}

func (s *syncAppliedConfigSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	mockTC := temporalclient.NewMockClient(gomock.NewController(s.T()))
	options := append(tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
		T:               s.T(),
		Mocks:           &tests.TestMocks{MockTC: mockTC},
		CustomValidator: true,
	}), fx.Provide(statusactivities.New), fx.Populate(&s.deps))
	s.fxApp = fxtest.New(s.T(), options...)
	s.fxApp.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *syncAppliedConfigSuite) TearDownSuite() {
	s.fxApp.RequireStop()
}

func (s *syncAppliedConfigSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.ctx = context.Background()
	s.ctx, _ = s.deps.Seed.EnsureAccount(s.ctx, s.T())
	s.ctx, _ = s.deps.Seed.EnsureOrg(s.ctx, s.T())
}

type seededStep struct {
	idx, groupIdx int
	status        app.Status
	retried       bool
}

// runFlow seeds an install on oldCfg, an app-branch rollout to a new config with the given steps, marks the flow
// successful and returns the install's applied config afterwards along with the new config ID.
func (s *syncAppliedConfigSuite) runFlow(steps ...seededStep) (applied, newCfgID string) {
	a := s.deps.Seed.CreateApp(s.ctx, s.T())
	oldCfg := s.deps.Seed.CreateAppConfig(s.ctx, s.T(), a.ID)
	newCfg := s.deps.Seed.CreateBareAppConfig(s.ctx, s.T(), a.ID)
	install := s.deps.Seed.CreateInstall(s.ctx, s.T(), a)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.Install{ID: install.ID}).
		Update("app_config_ref", app.AppConfigRef{AppliedConfigID: oldCfg.ID, ExpectedConfigID: oldCfg.ID}).Error)

	flw := app.Workflow{
		OwnerID:   install.ID,
		OwnerType: "installs",
		Type:      app.WorkflowTypeAppBranchConfigUpdate,
		Metadata:  pgtype.Hstore{"new_app_config_id": &newCfg.ID},
		Status:    app.CompositeStatus{Status: app.StatusInProgress},
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&flw).Error)
	for _, st := range steps {
		step := app.WorkflowStep{
			InstallWorkflowID: flw.ID,
			OwnerID:           install.ID,
			OwnerType:         "installs",
			Name:              "step",
			Idx:               st.idx,
			GroupIdx:          st.groupIdx,
			Status:            app.CompositeStatus{Status: st.status},
			Retried:           st.retried,
		}
		require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Omit("WorkflowStepGroupID").Create(&step).Error)
	}

	require.NoError(s.T(), s.deps.Activities.PkgStatusUpdateFlowStatus(s.ctx, statusactivities.UpdateStatusRequest{
		ID:     flw.ID,
		Status: app.CompositeStatus{Status: app.StatusSuccess},
	}))

	var got app.Install
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).First(&got, "id = ?", install.ID).Error)
	return got.AppConfigRef.AppliedConfigID, newCfg.ID
}

func (s *syncAppliedConfigSuite) TestAllSuccessAdvances() {
	applied, newCfg := s.runFlow(
		seededStep{idx: 100, groupIdx: 1, status: app.StatusSuccess},
		seededStep{idx: 200, groupIdx: 2, status: app.StatusSuccess},
	)
	s.Equal(newCfg, applied)
}

func (s *syncAppliedConfigSuite) TestNoopSandboxPlanAdvances() {
	applied, newCfg := s.runFlow(
		seededStep{idx: 300, groupIdx: 3, status: app.StatusSuccess},
		seededStep{idx: 400, groupIdx: 4, status: app.StatusAutoSkipped},
		seededStep{idx: 500, groupIdx: 4, status: app.StatusDiscarded},
		seededStep{idx: 600, groupIdx: 5, status: app.StatusSuccess},
	)
	s.Equal(newCfg, applied)
}

func (s *syncAppliedConfigSuite) TestStopDiscardDoesNotAdvance() {
	applied, newCfg := s.runFlow(
		seededStep{idx: 400, groupIdx: 4, status: app.StatusSuccess},
		seededStep{idx: 500, groupIdx: 4, status: app.StatusDiscarded},
	)
	s.NotEqual(newCfg, applied)
}

func (s *syncAppliedConfigSuite) TestErroredStepDoesNotAdvance() {
	applied, newCfg := s.runFlow(
		seededStep{idx: 400, groupIdx: 4, status: app.StatusAutoSkipped},
		seededStep{idx: 500, groupIdx: 5, status: app.StatusError},
	)
	s.NotEqual(newCfg, applied)
}

func (s *syncAppliedConfigSuite) TestRetriedDiscardAdvances() {
	applied, newCfg := s.runFlow(
		seededStep{idx: 500, groupIdx: 5, status: app.StatusDiscarded, retried: true},
		seededStep{idx: 501, groupIdx: 5, status: app.StatusSuccess},
	)
	s.Equal(newCfg, applied)
}

func (s *syncAppliedConfigSuite) TestRetriedErrorThenSuccessAdvances() {
	applied, newCfg := s.runFlow(
		seededStep{idx: 400, groupIdx: 4, status: app.StatusError, retried: true},
		seededStep{idx: 401, groupIdx: 4, status: app.StatusError, retried: true},
		seededStep{idx: 402, groupIdx: 4, status: app.StatusSuccess},
		seededStep{idx: 500, groupIdx: 5, status: app.StatusSuccess},
	)
	s.Equal(newCfg, applied)
}
