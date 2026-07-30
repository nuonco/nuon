package runnerhealthcheck

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	basemetrics "github.com/nuonco/nuon/pkg/metrics"
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	runneractivities "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

func TestFirstFailedHealthCheckMarksRunnerOfflineWithoutAlerting(t *testing.T) {
	var workflowSuite testsuite.WorkflowTestSuite
	env := workflowSuite.NewTestWorkflowEnvironment()
	now := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	env.SetStartTime(now)

	runner := testRunner(app.RunnerStatusActive)
	sig := &Signal{RunnerID: runner.ID}
	var calls []string

	env.OnActivity((*statusactivities.Activities).UpdateRunnerStatusV2Metadata, mock.MatchedBy(func(req statusactivities.UpdateRunnerStatusV2MetadataRequest) bool {
		return len(req.Metadata) == 1 && req.Metadata[app.RunnerOfflineTSMetadataKey] == now.Unix()
	})).Run(func(mock.Arguments) { calls = append(calls, "offline-ts") }).Return(nil).Once()
	env.OnActivity((*runneractivities.Activities).UpdateStatus, mock.Anything, mock.Anything, mock.Anything).
		Run(func(mock.Arguments) { calls = append(calls, "status") }).
		Return(nil).
		Once()
	env.OnActivity((*statusactivities.Activities).UpdateRunnerStatusV2, mock.Anything, mock.Anything, mock.Anything).
		Run(func(mock.Arguments) { calls = append(calls, "status-v2") }).
		Return(nil).
		Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.handleRunnerOffline(ctx, nil, runner, "no active install process")
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	require.Equal(t, []string{"offline-ts", "status", "status-v2"}, calls)
	env.AssertExpectations(t)
}

func TestOfflineRunnerDoesNotAlertBeforeDelay(t *testing.T) {
	var workflowSuite testsuite.WorkflowTestSuite
	env := workflowSuite.NewTestWorkflowEnvironment()
	now := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	env.SetStartTime(now)

	runner := runnerWithOfflineMetadata(app.RunnerStatusOffline, now.Add(-runnerUnhealthyAlertDelay+time.Second))
	sig := &Signal{RunnerID: runner.ID}

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.handleRunnerOffline(ctx, nil, runner, "no active install process")
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	env.AssertExpectations(t)
}

func TestOfflineRunnerEnqueuesIdempotentAlertAfterDelay(t *testing.T) {
	var workflowSuite testsuite.WorkflowTestSuite
	env := workflowSuite.NewTestWorkflowEnvironment()
	env.SetDataConverter(signalDataConverter())
	now := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	env.SetStartTime(now)

	offlineAt := now.Add(-runnerUnhealthyAlertDelay)
	runner := runnerWithOfflineMetadata(app.RunnerStatusOffline, offlineAt)
	runner.RunnerGroup.Type = app.RunnerGroupTypeOrg
	runner.RunnerGroup.OwnerID = runner.OrgID
	runner.RunnerGroup.OwnerType = "orgs"
	sig := &Signal{RunnerID: runner.ID}
	tmw := testTemporalMetricsWriter(t)
	var calls []string

	env.OnActivity(new(sharedactivities.Activities).EnqueueSignalToOwner, mock.Anything, mock.MatchedBy(func(req *sharedactivities.EnqueueSignalToOwnerRequest) bool {
		return req.IdempotencyKey == fmt.Sprintf("runner-unhealthy:%s:%d", runner.ID, offlineAt.Unix())
	})).
		Run(func(mock.Arguments) { calls = append(calls, "notification") }).
		Return(&sharedactivities.EnqueueSignalToOwnerResponse{Deduplicated: false}, nil).
		Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.handleRunnerOffline(ctx, tmw, runner, "no active build process")
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	require.Equal(t, []string{"notification"}, calls)
	env.AssertExpectations(t)
}

func TestOfflineRunnerReusesAlertIdempotencyKey(t *testing.T) {
	var workflowSuite testsuite.WorkflowTestSuite
	env := workflowSuite.NewTestWorkflowEnvironment()
	env.SetDataConverter(signalDataConverter())
	now := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	env.SetStartTime(now)

	offlineAt := now.Add(-time.Hour)
	runner := runnerWithOfflineMetadata(app.RunnerStatusOffline, offlineAt)
	runner.RunnerGroup.Type = app.RunnerGroupTypeOrg
	runner.RunnerGroup.OwnerID = runner.OrgID
	runner.RunnerGroup.OwnerType = "orgs"
	sig := &Signal{RunnerID: runner.ID}

	env.OnActivity(new(sharedactivities.Activities).EnqueueSignalToOwner, mock.Anything, mock.MatchedBy(func(req *sharedactivities.EnqueueSignalToOwnerRequest) bool {
		return req.IdempotencyKey == fmt.Sprintf("runner-unhealthy:%s:%d", runner.ID, offlineAt.Unix())
	})).Return(&sharedactivities.EnqueueSignalToOwnerResponse{Deduplicated: true}, nil).Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.handleRunnerOffline(ctx, nil, runner, "no active install process")
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	env.AssertExpectations(t)
}

func TestOfflineRunnerWithoutTimestampArmsAlert(t *testing.T) {
	var workflowSuite testsuite.WorkflowTestSuite
	env := workflowSuite.NewTestWorkflowEnvironment()
	now := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	env.SetStartTime(now)

	runner := testRunner(app.RunnerStatusOffline)
	sig := &Signal{RunnerID: runner.ID}

	env.OnActivity((*statusactivities.Activities).UpdateRunnerStatusV2Metadata, mock.MatchedBy(func(req statusactivities.UpdateRunnerStatusV2MetadataRequest) bool {
		return req.Metadata[app.RunnerOfflineTSMetadataKey] == now.Unix()
	})).Return(nil).Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.handleRunnerOffline(ctx, nil, runner, "no active install process")
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	env.AssertExpectations(t)
}

func TestActiveRunnerWithOfflineTimestampDoesNotResetDelay(t *testing.T) {
	var workflowSuite testsuite.WorkflowTestSuite
	env := workflowSuite.NewTestWorkflowEnvironment()
	now := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	env.SetStartTime(now)

	runner := runnerWithOfflineMetadata(app.RunnerStatusActive, now.Add(-time.Minute))
	sig := &Signal{RunnerID: runner.ID}

	env.OnActivity((*runneractivities.Activities).UpdateStatus, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Once()
	env.OnActivity((*statusactivities.Activities).UpdateRunnerStatusV2, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.handleRunnerOffline(ctx, nil, runner, "no active install process")
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	env.AssertExpectations(t)
}

func TestOfflineCheckRepairsStaleStatusV2(t *testing.T) {
	var workflowSuite testsuite.WorkflowTestSuite
	env := workflowSuite.NewTestWorkflowEnvironment()
	now := time.Date(2026, time.July, 30, 12, 0, 0, 0, time.UTC)
	env.SetStartTime(now)

	runner := runnerWithOfflineMetadata(app.RunnerStatusOffline, now.Add(-time.Minute))
	runner.StatusV2.Status = app.Status(app.RunnerStatusActive)
	sig := &Signal{RunnerID: runner.ID}

	env.OnActivity((*statusactivities.Activities).UpdateRunnerStatusV2, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.handleRunnerOffline(ctx, nil, runner, "no active install process")
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	env.AssertExpectations(t)
}

func TestHealthyCheckClearsOfflineMetadataAndRestoresActive(t *testing.T) {
	var workflowSuite testsuite.WorkflowTestSuite
	env := workflowSuite.NewTestWorkflowEnvironment()

	runner := runnerWithOfflineMetadata(app.RunnerStatusOffline, time.Now().Add(-time.Minute))
	sig := &Signal{RunnerID: runner.ID}
	var calls []string

	env.OnActivity((*statusactivities.Activities).UpdateRunnerStatusV2Metadata, mock.MatchedBy(func(req statusactivities.UpdateRunnerStatusV2MetadataRequest) bool {
		return len(req.Metadata) == 1 && req.Metadata[app.RunnerOfflineTSMetadataKey] == nil
	})).Run(func(mock.Arguments) { calls = append(calls, "clear") }).Return(nil).Once()
	env.OnActivity((*runneractivities.Activities).UpdateStatus, mock.Anything, mock.Anything, mock.Anything).
		Run(func(mock.Arguments) { calls = append(calls, "status") }).
		Return(nil).
		Once()
	env.OnActivity((*statusactivities.Activities).UpdateRunnerStatusV2, mock.Anything, mock.Anything, mock.Anything).
		Run(func(mock.Arguments) { calls = append(calls, "status-v2") }).
		Return(nil).
		Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.handleRunnerActive(ctx, runner)
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	require.Equal(t, []string{"clear", "status", "status-v2"}, calls)
	env.AssertExpectations(t)
}

func TestUpdateRunnerStatusStopsWhenLegacyUpdateFails(t *testing.T) {
	var workflowSuite testsuite.WorkflowTestSuite
	env := workflowSuite.NewTestWorkflowEnvironment()

	runner := testRunner(app.RunnerStatusActive)
	sig := &Signal{RunnerID: runner.ID}

	env.OnActivity((*runneractivities.Activities).UpdateStatus, mock.Anything, mock.Anything, mock.Anything).
		Return(temporal.NewNonRetryableApplicationError("update failed", "test", nil)).
		Once()

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		return sig.updateRunnerStatus(ctx, runner, app.RunnerStatusOffline, "no active install process")
	})

	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
	env.AssertExpectations(t)
}

func testRunner(status app.RunnerStatus) *app.Runner {
	return &app.Runner{
		ID:          "rnr_1",
		DisplayName: "Install runner",
		OrgID:       "org_1",
		Org: app.Org{
			Name: "Example org",
		},
		Status:        status,
		RunnerGroupID: "rng_1",
		RunnerGroup: app.RunnerGroup{
			Type:      app.RunnerGroupTypeInstall,
			OwnerID:   "ins_1",
			OwnerType: "installs",
		},
		StatusV2: app.CompositeStatus{
			Status: app.Status(status),
		},
	}
}

func runnerWithOfflineMetadata(status app.RunnerStatus, offlineAt time.Time) *app.Runner {
	runner := testRunner(status)
	runner.StatusV2.Metadata = map[string]any{
		app.RunnerOfflineTSMetadataKey: float64(offlineAt.Unix()),
	}
	return runner
}

func testTemporalMetricsWriter(t *testing.T) tmetrics.Writer {
	v := validator.New()
	mw, err := basemetrics.New(v, basemetrics.WithDisable(true))
	require.NoError(t, err)
	tmw, err := tmetrics.New(v, tmetrics.WithMetricsWriter(mw))
	require.NoError(t, err)
	return tmw
}

func signalDataConverter() converter.DataConverter {
	return converter.NewCompositeDataConverter(
		signaldb.NewPayloadConverter(),
		converter.NewNilPayloadConverter(),
		converter.NewByteSlicePayloadConverter(),
		converter.NewJSONPayloadConverter(),
	)
}
