package statusactivities_test

import (
	"context"
	"testing"
	"time"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/telemetry"
	workflowmetrics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/metrics"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type metricTestDeps struct {
	fx.In

	DB     *gorm.DB `name:"psql"`
	Seeder *testseed.Seeder
	MW     metrics.Writer
}

func TestPersistedWorkflowMetrics(t *testing.T) {
	// CI discovers PostgreSQL suites using the INTEGRATION marker.
	tests.SkipIfNotIntegration(t)
	var deps metricTestDeps
	fxApp := fxtest.New(t, append(tests.CtlApiFXOptions(t), fx.Populate(&deps))...)
	fxApp.RequireStart()
	t.Cleanup(fxApp.RequireStop)
	ctx := context.Background()
	testApp := deps.Seeder.CreateApp(ctx, t)
	ctx = cctx.SetAccountContext(ctx, &testApp.CreatedBy)
	ctx = cctx.SetOrgIDContext(ctx, testApp.OrgID)
	deps.Seeder.CreateAppConfig(ctx, t, testApp.ID)
	install := deps.Seeder.CreateInstall(ctx, t, testApp)
	queue := &app.Queue{OwnerID: install.ID, OwnerType: "installs"}
	require.NoError(t, deps.DB.WithContext(ctx).Create(queue).Error)

	reader := metric.NewManualReader()
	provider := metric.NewMeterProvider(metric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	var startActivities *activities.Activities
	newActivities := func(db *gorm.DB) *statusactivities.Activities {
		counters, err := workflowmetrics.NewCounters(workflowmetrics.CounterParams{
			Config: &telemetry.Config{Endpoint: "http://collector"}, Provider: provider, DB: db, L: zap.NewNop(),
		})
		require.NoError(t, err)
		startActivities = activities.New(activities.Params{DB: deps.DB, Metrics: counters})
		return statusactivities.New(statusactivities.Params{DB: deps.DB, MW: deps.MW, Counters: counters})
	}
	a := newActivities(deps.DB)
	newWorkflow := func(status app.Status) *app.Workflow {
		wf := deps.Seeder.CreateWorkflow(ctx, t, install.ID, app.WorkflowTypeManualDeploy)
		require.NoError(t, a.PkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{ID: wf.ID, Status: app.CompositeStatus{Status: status}}))
		return wf
	}
	newSignal := func(wf *app.Workflow) *app.QueueSignal {
		qs := &app.QueueSignal{QueueID: queue.ID, OwnerID: wf.ID, OwnerType: "install_workflows", Type: "execute-workflow", Status: app.CompositeStatus{Status: app.StatusInProgress}}
		require.NoError(t, deps.DB.WithContext(ctx).Create(qs).Error)
		return qs
	}
	updateSignal := func(qs *app.QueueSignal, status app.Status) {
		require.NoError(t, a.UpdateQueueSignalStatusV2(ctx, statusactivities.UpdateQueueSignalStatusV2Request{QueueSignalID: qs.ID, Status: status}))
	}
	values := func() map[string]int64 {
		var rm metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(ctx, &rm))
		result := make(map[string]int64)
		for _, scope := range rm.ScopeMetrics {
			for _, m := range scope.Metrics {
				sum, ok := m.Data.(metricdata.Sum[int64])
				if !ok {
					continue
				}
				for _, point := range sum.DataPoints {
					require.Equal(t, 2, point.Attributes.Len())
					if point.Value == 0 {
						continue
					}
					typ, _ := point.Attributes.Value("workflow.type")
					outcome, _ := point.Attributes.Value("workflow.outcome")
					source, _ := point.Attributes.Value("retry.source")
					result[m.Name+"/"+typ.AsString()+"/"+outcome.AsString()+source.AsString()] = point.Value
				}
			}
		}
		return result
	}
	histogram := func(name string) []metricdata.HistogramDataPoint[float64] {
		var rm metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(ctx, &rm))
		for _, scope := range rm.ScopeMetrics {
			for _, m := range scope.Metrics {
				if m.Name == name {
					return m.Data.(metricdata.Histogram[float64]).DataPoints
				}
			}
		}
		return nil
	}

	t.Run("first start survives duplicates and concurrent writers", func(t *testing.T) {
		wf := newWorkflow(app.StatusPending)
		created := time.Now().Add(-37 * time.Second).Truncate(time.Microsecond)
		require.NoError(t, deps.DB.Model(wf).Update("created_at", created).Error)
		errs := make(chan error, 8)
		for range 8 {
			go func() {
				errs <- startActivities.PkgWorkflowsFlowUpdateFlowStartedAt(ctx, activities.UpdateFlowStartedAtRequest{ID: wf.ID})
			}()
		}
		for range 8 {
			require.NoError(t, <-errs)
		}
		var started app.Workflow
		require.NoError(t, deps.DB.Where(app.Workflow{ID: wf.ID}).Take(&started).Error)
		points := histogram("nuon.workflow.start_delay")
		require.Len(t, points, 1)
		require.Equal(t, uint64(1), points[0].Count)
		require.Equal(t, 1, points[0].Attributes.Len())
		require.Equal(t, started.StartedAt.Sub(created).Seconds(), points[0].Sum)
		a = newActivities(deps.DB)
		require.NoError(t, startActivities.PkgWorkflowsFlowUpdateFlowStartedAt(ctx, activities.UpdateFlowStartedAtRequest{ID: wf.ID}))
		var retried app.Workflow
		require.NoError(t, deps.DB.Where(app.Workflow{ID: wf.ID}).Take(&retried).Error)
		require.Equal(t, started.StartedAt, retried.StartedAt)
		require.Equal(t, uint64(1), histogram("nuon.workflow.start_delay")[0].Count)
		for _, id := range []string{"", "wfl00000000000000000000000"} {
			require.Error(t, startActivities.PkgWorkflowsFlowUpdateFlowStartedAt(ctx, activities.UpdateFlowStartedAtRequest{ID: id}))
		}
	})

	wf := newWorkflow(app.StatusError)
	created := time.Now().Add(-2 * time.Hour).Truncate(time.Microsecond)
	require.NoError(t, deps.DB.Model(wf).Updates(map[string]any{
		"created_at": created, "started_at": created.Add(time.Hour), "finished_at": created.Add(90 * time.Minute),
	}).Error)
	qs := newSignal(wf)
	step := deps.Seeder.CreateWorkflowStep(ctx, t, wf.ID)
	stepUpdate := func(status app.Status, metadata map[string]any) {
		require.NoError(t, a.PkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{ID: step.ID, Status: app.CompositeStatus{Status: status, Metadata: metadata}}))
	}
	stepUpdate(app.StatusError, map[string]any{"auto_retried": true})
	stepUpdate(app.StatusError, map[string]any{"auto_retried": true})
	stepUpdate(app.StatusError, map[string]any{"reason": "later status update"})
	require.Equal(t, map[string]int64{"nuon.workflow.step.retries/manual_deploy/auto": 1}, values(), "retryable errors do not complete executions")
	require.Empty(t, histogram("nuon.workflow.elapsed_time"), "parked workflow with finished_at is not complete")

	step = deps.Seeder.CreateWorkflowStep(ctx, t, wf.ID)
	stepUpdate(app.StatusDiscarded, map[string]any{"retry_type": "manual"})
	stepUpdate(app.StatusDiscarded, map[string]any{"retry_type": "manual"})
	stepUpdate(app.StatusError, nil)
	stepUpdate(app.StatusDiscarded, nil)
	require.Equal(t, map[string]int64{
		"nuon.workflow.step.retries/manual_deploy/auto": 1, "nuon.workflow.step.retries/manual_deploy/manual": 1,
	}, values(), "carried metadata is not a new retry decision")
	require.NoError(t, a.PkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{ID: wf.ID, Status: app.CompositeStatus{Status: app.StatusSuccess}}))
	completedBefore := time.Now()
	updateSignal(qs, app.StatusSuccess)
	completedAfter := time.Now()
	elapsed := histogram("nuon.workflow.elapsed_time")
	require.Len(t, elapsed, 1)
	require.Equal(t, uint64(1), elapsed[0].Count)
	require.Equal(t, 2, elapsed[0].Attributes.Len())
	require.GreaterOrEqual(t, elapsed[0].Sum, completedBefore.Sub(created).Seconds())
	require.LessOrEqual(t, elapsed[0].Sum, completedAfter.Sub(created).Seconds())
	a = newActivities(deps.DB)
	updateSignal(qs, app.StatusSuccess)
	updateSignal(qs, app.StatusError)
	require.Equal(t, uint64(1), histogram("nuon.workflow.elapsed_time")[0].Count)
	for _, terminal := range []app.Status{app.StatusSuccess, app.StatusError} {
		updateSignal(newSignal(newWorkflow(app.StatusCancelled)), terminal)
	}
	updateSignal(newSignal(newWorkflow(app.StatusError)), app.StatusError)
	updateSignal(newSignal(newWorkflow(app.StatusInProgress)), app.StatusError)
	want := map[string]int64{
		"nuon.workflow.executions.completed/manual_deploy/success":   1,
		"nuon.workflow.executions.completed/manual_deploy/cancelled": 2,
		"nuon.workflow.executions.completed/manual_deploy/error":     1,
		"nuon.workflow.executions.completed/manual_deploy/unknown":   1,
		"nuon.workflow.step.retries/manual_deploy/auto":              1,
		"nuon.workflow.step.retries/manual_deploy/manual":            1,
	}
	require.Equal(t, want, values())

	for _, excluded := range []string{"plan-only", "other-type", "other-owner", "deleted-workflow", "deleted-install"} {
		wf := newWorkflow(app.StatusSuccess)
		qs := newSignal(wf)
		switch excluded {
		case "plan-only":
			require.NoError(t, deps.DB.Model(wf).Update("plan_only", true).Error)
		case "other-type":
			require.NoError(t, deps.DB.Model(wf).Update("type", "deprovision").Error)
		case "other-owner":
			require.NoError(t, deps.DB.Model(wf).Update("owner_type", "apps").Error)
		case "deleted-workflow":
			require.NoError(t, deps.DB.Model(wf).Update("deleted_at", 1).Error)
		case "deleted-install":
			require.NoError(t, deps.DB.Model(&app.Install{ID: install.ID}).Update("deleted_at", 1).Error)
		}
		startErr := startActivities.PkgWorkflowsFlowUpdateFlowStartedAt(ctx, activities.UpdateFlowStartedAtRequest{ID: wf.ID})
		if excluded == "deleted-workflow" {
			require.Error(t, startErr)
		} else {
			require.NoError(t, startErr)
		}
		updateSignal(qs, app.StatusSuccess)
	}
	require.Equal(t, want, values(), "excluded workflows do not add series or observations")
	require.NoError(t, deps.DB.Unscoped().Model(&app.Install{ID: install.ID}).Update("deleted_at", 0).Error)

	qs = newSignal(newWorkflow(app.StatusSuccess))
	err := a.UpdateQueueSignalStatusV2(context.Background(), statusactivities.UpdateQueueSignalStatusV2Request{QueueSignalID: qs.ID, Status: app.StatusSuccess})
	require.Error(t, err, "missing account must fail persistence")
	require.Equal(t, want, values(), "failed status writes do not emit")

	closedTx := deps.DB.Begin()
	require.NoError(t, closedTx.Rollback().Error)
	a = newActivities(closedTx)
	updateSignal(qs, app.StatusSuccess)
	var persisted app.QueueSignal
	require.NoError(t, deps.DB.Where(app.QueueSignal{ID: qs.ID}).Take(&persisted).Error)
	require.Equal(t, app.StatusSuccess, persisted.Status.Status, "telemetry lookup failure must not fail the persisted operation")
	require.Equal(t, want, values())
	wf = newWorkflow(app.StatusPending)
	require.NoError(t, startActivities.PkgWorkflowsFlowUpdateFlowStartedAt(ctx, activities.UpdateFlowStartedAtRequest{ID: wf.ID}), "lookup failures do not fail the first start")
	require.Equal(t, uint64(1), histogram("nuon.workflow.start_delay")[0].Count)
	elapsedCounts := make(map[string]uint64)
	for _, point := range histogram("nuon.workflow.elapsed_time") {
		outcome, _ := point.Attributes.Value("workflow.outcome")
		elapsedCounts[outcome.AsString()] = point.Count
	}
	require.Equal(t, map[string]uint64{"success": 1, "error": 1, "cancelled": 2, "unknown": 1}, elapsedCounts)
}
