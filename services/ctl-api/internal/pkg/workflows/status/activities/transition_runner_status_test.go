package statusactivities_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type transitionRunnerStatusDeps struct {
	fx.In

	DB         *gorm.DB `name:"psql"`
	Seed       *testseed.Seeder
	Activities *statusactivities.Activities
}

type transitionRunnerStatusSuite struct {
	tests.BaseDBTestSuite

	fxApp *fxtest.App
	deps  transitionRunnerStatusDeps
	ctx   context.Context
	orgID string
}

func TestTransitionRunnerStatusSuite(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	suite.Run(t, new(transitionRunnerStatusSuite))
}

func (s *transitionRunnerStatusSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	options := append(tests.CtlApiFXOptions(s.T()), fx.Provide(statusactivities.New), fx.Populate(&s.deps))
	s.fxApp = fxtest.New(s.T(), options...)
	s.fxApp.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *transitionRunnerStatusSuite) TearDownSuite() {
	s.fxApp.RequireStop()
}

func (s *transitionRunnerStatusSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.ctx = context.Background()
	s.ctx, _ = s.deps.Seed.EnsureAccount(s.ctx, s.T())
	var org *app.Org
	s.ctx, org = s.deps.Seed.EnsureOrg(s.ctx, s.T())
	s.orgID = org.ID
}

func (s *transitionRunnerStatusSuite) TestUpdatesLegacyAndV2Together() {
	runner := s.seedRunner(app.RunnerStatusOffline, app.RunnerStatusActive)
	runner.StatusV2.Metadata = map[string]any{
		"existing":                     "value",
		app.RunnerOfflineTSMetadataKey: 1,
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Save(runner).Error)

	updated, err := s.deps.Activities.TransitionRunnerStatus(s.ctx, statusactivities.TransitionRunnerStatusRequest{
		RunnerID:          runner.ID,
		Status:            app.RunnerStatusAwaitingHeartbeat,
		StatusDescription: "waiting for process",
		Metadata: map[string]any{
			"new":                          "metadata",
			app.RunnerOfflineTSMetadataKey: nil,
		},
	})
	require.NoError(s.T(), err)
	require.True(s.T(), updated)

	var got app.Runner
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).First(&got, "id = ?", runner.ID).Error)
	require.Equal(s.T(), app.RunnerStatusAwaitingHeartbeat, got.Status)
	require.Equal(s.T(), app.Status(app.RunnerStatusAwaitingHeartbeat), got.StatusV2.Status)
	require.Equal(s.T(), "waiting for process", got.StatusDescription)
	require.Equal(s.T(), "waiting for process", got.StatusV2.StatusHumanDescription)
	require.Equal(s.T(), "value", got.StatusV2.Metadata["existing"])
	require.Equal(s.T(), "metadata", got.StatusV2.Metadata["new"])
	_, hasOfflineTS := got.StatusV2.Metadata[app.RunnerOfflineTSMetadataKey]
	require.False(s.T(), hasOfflineTS)
	require.Len(s.T(), got.StatusV2.History, 1)
	require.Equal(s.T(), app.Status(app.RunnerStatusActive), got.StatusV2.History[0].Status)
}

func (s *transitionRunnerStatusSuite) TestDisabledGuardReconcilesNothing() {
	runner := s.seedRunner(app.RunnerStatusActive, app.RunnerStatusDisabled)

	updated, err := s.deps.Activities.TransitionRunnerStatus(s.ctx, statusactivities.TransitionRunnerStatusRequest{
		RunnerID:          runner.ID,
		Status:            app.RunnerStatusOffline,
		StatusDescription: "no active process",
		SkipIfDisabled:    true,
	})
	require.NoError(s.T(), err)
	require.False(s.T(), updated)

	var got app.Runner
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).First(&got, "id = ?", runner.ID).Error)
	require.Equal(s.T(), app.RunnerStatusActive, got.Status)
	require.Equal(s.T(), app.Status(app.RunnerStatusDisabled), got.StatusV2.Status)
}

func (s *transitionRunnerStatusSuite) TestLegacyDisabledGuardReconcilesNothing() {
	runner := s.seedRunner(app.RunnerStatusDisabled, app.RunnerStatusActive)

	updated, err := s.deps.Activities.TransitionRunnerStatus(s.ctx, statusactivities.TransitionRunnerStatusRequest{
		RunnerID:          runner.ID,
		Status:            app.RunnerStatusOffline,
		StatusDescription: "no active process",
		SkipIfDisabled:    true,
	})
	require.NoError(s.T(), err)
	require.False(s.T(), updated)
}

func (s *transitionRunnerStatusSuite) TestMatchingStatusDoesNotGrowHistory() {
	runner := s.seedRunner(app.RunnerStatusActive, app.RunnerStatusActive)

	updated, err := s.deps.Activities.TransitionRunnerStatus(s.ctx, statusactivities.TransitionRunnerStatusRequest{
		RunnerID:          runner.ID,
		Status:            app.RunnerStatusActive,
		StatusDescription: string(app.RunnerStatusActive),
	})
	require.NoError(s.T(), err)
	require.False(s.T(), updated)

	var got app.Runner
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).First(&got, "id = ?", runner.ID).Error)
	require.Empty(s.T(), got.StatusV2.History)
}

func (s *transitionRunnerStatusSuite) TestLegacyRepairDoesNotDuplicateV2History() {
	runner := s.seedRunner(app.RunnerStatusOffline, app.RunnerStatusActive)
	runner.StatusV2.StatusHumanDescription = string(app.RunnerStatusActive)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Save(runner).Error)

	updated, err := s.deps.Activities.TransitionRunnerStatus(s.ctx, statusactivities.TransitionRunnerStatusRequest{
		RunnerID:          runner.ID,
		Status:            app.RunnerStatusActive,
		StatusDescription: string(app.RunnerStatusActive),
	})
	require.NoError(s.T(), err)
	require.True(s.T(), updated)

	var got app.Runner
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).First(&got, "id = ?", runner.ID).Error)
	require.Equal(s.T(), app.RunnerStatusActive, got.Status)
	require.Empty(s.T(), got.StatusV2.History)
}

func (s *transitionRunnerStatusSuite) TestMissingAccountUsesExistingProvenance() {
	runner := s.seedRunner(app.RunnerStatusOffline, app.RunnerStatusOffline)
	runner.StatusV2.CreatedByID = runner.CreatedByID
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Save(runner).Error)

	updated, err := s.deps.Activities.TransitionRunnerStatus(context.Background(), statusactivities.TransitionRunnerStatusRequest{
		RunnerID:          runner.ID,
		Status:            app.RunnerStatusActive,
		StatusDescription: "active",
	})
	require.NoError(s.T(), err)
	require.True(s.T(), updated)

	var got app.Runner
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).First(&got, "id = ?", runner.ID).Error)
	require.Equal(s.T(), runner.CreatedByID, got.StatusV2.CreatedByID)
}

func (s *transitionRunnerStatusSuite) TestWriteFailurePropagates() {
	_, err := s.deps.Activities.TransitionRunnerStatus(s.ctx, statusactivities.TransitionRunnerStatusRequest{
		RunnerID:          "rnr_missing",
		Status:            app.RunnerStatusActive,
		StatusDescription: "active",
	})
	require.Error(s.T(), err)
}

func (s *transitionRunnerStatusSuite) seedRunner(status, statusV2 app.RunnerStatus) *app.Runner {
	group := &app.RunnerGroup{
		OrgID:     s.orgID,
		OwnerID:   s.orgID,
		OwnerType: "orgs",
		Type:      app.RunnerGroupTypeOrg,
		Platform:  app.AppRunnerTypeAWS,
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(group).Error)

	runner := &app.Runner{
		OrgID:             s.orgID,
		RunnerGroupID:     group.ID,
		Name:              "runner-" + group.ID,
		DisplayName:       "runner",
		Status:            status,
		StatusDescription: string(status),
		StatusV2: app.CompositeStatus{
			Status:                 app.Status(statusV2),
			StatusHumanDescription: string(statusV2),
		},
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(runner).Error)
	return runner
}
