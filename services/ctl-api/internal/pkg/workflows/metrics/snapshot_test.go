package workflowmetrics

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type snapshotTestDeps struct {
	fx.In

	Config *internal.Config
	DB     *gorm.DB `name:"psql"`
	Seeder *testseed.Seeder
}

func TestPostgresSnapshot(t *testing.T) {
	// CI discovers database-backed packages by the literal INTEGRATION marker.
	tests.SkipIfNotIntegration(t)

	var deps snapshotTestDeps
	options := append(tests.CtlApiFXOptions(t), fx.Populate(&deps))
	fxApp := fxtest.New(t, options...)
	fxApp.RequireStart()
	t.Cleanup(fxApp.RequireStop)

	ctx := context.Background()
	connect := func() *pgx.Conn {
		conn, err := psql.NewPrimaryListenerConn(ctx, deps.Config)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, conn.Close(ctx)) })
		return conn
	}
	conn := connect()
	baseline, err := readSnapshot(ctx, conn)
	require.NoError(t, err)

	testApp := deps.Seeder.CreateApp(ctx, t)
	ctx = cctx.SetAccountContext(ctx, &testApp.CreatedBy)
	ctx = cctx.SetOrgIDContext(ctx, testApp.OrgID)
	deps.Seeder.CreateAppConfig(ctx, t, testApp.ID)

	oldest := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	fixtures := []struct{ id, status string }{
		{"queued", `{"status":"queued"}`},
		{"pending", `{"status":"pending"}`},
		{"executing", `{"status":"in-progress","metadata":{"awaiting_retry":true}}`},
		{"retrying", `{"status":"retrying"}`},
		{"approval", `{"status":"approval-awaiting"}`},
		{"retry", `{"status":"failed-pending-retry"}`},
		{"legacy-retry", `{"status":"error","metadata":{"awaiting_retry":true}}`},
		{"terminal-error", `{"status":"error"}`},
		{"stopped", `{"status":"error","metadata":{"awaiting_retry":true,"stopped":true}}`},
		{"exhausted", `{"status":"error","metadata":{"awaiting_retry":true,"retries_exhausted":true}}`},
		{"success", `{"status":"success","metadata":{"awaiting_retry":true}}`},
		{"cancelled", `{"status":"cancelled","metadata":{"awaiting_retry":true}}`},
		{"unknown", `{"status":"future-state"}`},
		{"deleted-workflow", `{"status":"queued"}`},
		{"deleted-install", `{"status":"queued"}`},
		{"other-owner", `{"status":"queued"}`},
		{"plan-only", `{"status":"queued"}`},
		{"other-type", `{"status":"queued"}`},
		{"retry-abandoned", `{"status":"failed-pending-retry"}`},
		{"legacy-abandoned", `{"status":"error","metadata":{"awaiting_retry":true}}`},
		{"retry-no-signal", `{"status":"failed-pending-retry"}`},
		{"legacy-no-signal", `{"status":"error","metadata":{"awaiting_retry":true}}`},
		{"retry-wrong-signal", `{"status":"failed-pending-retry"}`},
		{"retry-wrong-owner", `{"status":"failed-pending-retry"}`},
		{"retry-deleted-signal", `{"status":"failed-pending-retry"}`},
		{"retry-stopped", `{"status":"failed-pending-retry","metadata":{"stopped":true}}`},
		{"retry-exhausted", `{"status":"failed-pending-retry","metadata":{"retries_exhausted":true}}`},
	}
	workflowIDs := make(map[string]string, len(fixtures))
	installIDs := make(map[string]string, len(fixtures))
	t.Cleanup(func() {
		require.NoError(t, deps.DB.Exec("UPDATE install_workflows SET deleted_at=1 WHERE id IN ?", mapValues(workflowIDs)).Error)
		require.NoError(t, deps.DB.Exec("UPDATE installs SET deleted_at=1 WHERE id IN ?", mapValues(installIDs)).Error)
	})
	for i, fixture := range fixtures {
		install := deps.Seeder.CreateInstall(ctx, t, testApp)
		workflow := deps.Seeder.CreateWorkflow(ctx, t, install.ID, app.WorkflowTypeManualDeploy)
		workflowIDs[fixture.id] = workflow.ID
		installIDs[fixture.id] = install.ID
		require.NoError(t, deps.DB.Exec(`UPDATE install_workflows SET status=?::jsonb, created_at=? WHERE id=?`, fixture.status, oldest.Add(time.Duration(i)*time.Minute), workflow.ID).Error)
	}
	queue := &app.Queue{OwnerID: testApp.ID, OwnerType: "apps"}
	require.NoError(t, deps.DB.WithContext(ctx).Create(queue).Error)
	for _, name := range []string{"retry", "legacy-retry", "retry-abandoned", "legacy-abandoned", "retry-wrong-signal", "retry-wrong-owner", "retry-deleted-signal", "retry-stopped", "retry-exhausted", "stopped", "exhausted"} {
		sig := &app.QueueSignal{
			QueueID: queue.ID, OwnerID: workflowIDs[name], OwnerType: "install_workflows",
			Type: "execute-workflow", Status: app.NewCompositeStatus(ctx, app.StatusInProgress),
		}
		switch name {
		case "retry-abandoned":
			sig.Status.Status = app.StatusError
		case "legacy-abandoned":
			sig.Status.Status = app.StatusCancelled
		case "retry-wrong-signal":
			sig.Type = "generate-steps"
		case "retry-wrong-owner":
			sig.OwnerType = "install_workflow_steps"
		case "retry-deleted-signal":
			sig.DeletedAt = 1
		}
		require.NoError(t, deps.DB.WithContext(ctx).Create(sig).Error)
	}
	newer := deps.Seeder.CreateWorkflow(ctx, t, installIDs["legacy-retry"], app.WorkflowTypeManualDeploy)
	workflowIDs["newer"] = newer.ID
	require.NoError(t, deps.DB.Exec(`UPDATE install_workflows SET status='{"status":"success"}'::jsonb WHERE id=?`, newer.ID).Error)

	require.NoError(t, deps.DB.Exec("UPDATE install_workflows SET deleted_at=1 WHERE id=?", workflowIDs["deleted-workflow"]).Error)
	require.NoError(t, deps.DB.Exec("UPDATE installs SET deleted_at=1 WHERE id=?", installIDs["deleted-install"]).Error)
	require.NoError(t, deps.DB.Exec("UPDATE install_workflows SET owner_type='apps' WHERE id=?", workflowIDs["other-owner"]).Error)
	require.NoError(t, deps.DB.Exec("UPDATE install_workflows SET plan_only=true WHERE id=?", workflowIDs["plan-only"]).Error)
	require.NoError(t, deps.DB.Exec("UPDATE install_workflows SET type='drift_run' WHERE id=?", workflowIDs["other-type"]).Error)
	require.NoError(t, deps.DB.Exec("UPDATE install_workflows SET type='provision' WHERE id=?", workflowIDs["pending"]).Error)
	require.NoError(t, deps.DB.Exec("UPDATE install_workflows SET type='deploy_components', plan_only=NULL WHERE id=?", workflowIDs["retrying"]).Error)
	require.NoError(t, deps.DB.Exec("UPDATE install_workflows SET finished_at=now() WHERE id=?", workflowIDs["legacy-retry"]).Error)

	values, err := readSnapshot(ctx, conn)
	require.NoError(t, err)
	wants := map[bucket]value{
		{"manual_deploy", "queued"}:            {1, oldest},
		{"provision", "queued"}:                {1, oldest.Add(time.Minute)},
		{"manual_deploy", "executing"}:         {1, oldest.Add(2 * time.Minute)},
		{"deploy_components", "executing"}:     {1, oldest.Add(3 * time.Minute)},
		{"manual_deploy", "awaiting_approval"}: {1, oldest.Add(4 * time.Minute)},
		{"manual_deploy", "awaiting_retry"}:    {2, oldest.Add(5 * time.Minute)},
	}
	expected := make(map[bucket]value, len(baseline)+len(wants))
	for key, v := range baseline {
		expected[key] = v
	}
	for key, fixtureValue := range wants {
		want := fixtureValue
		if existing, ok := baseline[key]; ok {
			want.count += existing.count
			if existing.oldest.Before(want.oldest) {
				want.oldest = existing.oldest
			}
		}
		expected[key] = want
	}
	require.Len(t, values, len(expected))
	for key, want := range expected {
		require.Equal(t, want.count, values[key].count, key)
		require.True(t, want.oldest.Equal(values[key].oldest), key)
	}

	t.Run("partial index supports generic plans", func(t *testing.T) {
		tx, err := conn.Begin(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx)
		_, err = tx.Exec(ctx, "SET LOCAL enable_seqscan=off; SET LOCAL plan_cache_mode=force_generic_plan")
		require.NoError(t, err)
		rows, err := tx.Query(ctx, "EXPLAIN "+snapshotQuery, workflowTypes)
		require.NoError(t, err)
		defer rows.Close()
		var plan string
		for rows.Next() {
			var line string
			require.NoError(t, rows.Scan(&line))
			plan += line + "\n"
		}
		require.NoError(t, rows.Err())
		require.Contains(t, plan, "idx_install_workflows_metrics_current")
	})

	r, reader := testReporter(t)
	require.NoError(t, r.refresh(ctx, conn))
	previous := r.collectedAt
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	require.Error(t, r.refresh(canceled, conn))
	require.Equal(t, previous, r.collectedAt, "failed refresh must not advance freshness")
	require.Equal(t, values, r.values)

	require.NoError(t, deps.DB.Exec(`UPDATE queue_signals SET status='{"status":"success"}'::jsonb WHERE owner_id=?`, workflowIDs["retry"]).Error)
	require.NoError(t, deps.DB.Exec(`UPDATE queue_signals SET status='{"status":"error"}'::jsonb WHERE owner_id=?`, workflowIDs["legacy-retry"]).Error)
	require.NoError(t, r.refresh(ctx, conn))
	require.Equal(t, baseline[bucket{"manual_deploy", "awaiting_retry"}], r.values[bucket{"manual_deploy", "awaiting_retry"}], "terminal signals clear stale retry metadata without changing workflow status")

	require.NoError(t, deps.DB.Exec("UPDATE install_workflows SET status='{\"status\":\"success\"}'::jsonb WHERE id IN ?", mapValues(workflowIDs)).Error)
	require.NoError(t, r.refresh(ctx, conn))
	require.Equal(t, baseline, r.values, "a successful snapshot with no fixture categories replaces old fixture state")
	require.NotEmpty(t, collect(t, reader))

	first, err := acquire(ctx, conn)
	require.NoError(t, err)
	require.True(t, first)
	peer := connect()
	second, err := acquire(ctx, peer)
	require.NoError(t, err)
	require.False(t, second, "a second reporter cannot acquire the same database lock")
	r.release(conn)
	require.Eventually(t, func() bool {
		second, err = acquire(ctx, peer)
		require.NoError(t, err)
		return second
	}, time.Second, 10*time.Millisecond, "releasing the leader session permits failover")
}

func mapValues(values map[string]string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	return result
}
