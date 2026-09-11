package workflowmetrics

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/psql/migrations"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

func TestDeploymentObservation(t *testing.T) {
	r, reader := testReporter(t)
	c := deploymentComponent{"org-example", "app-example", "component-example"}
	key := deploymentBucket{c, "error"}
	r.deployments = deploymentSnapshot{
		attempts: map[deploymentBucket]int64{key: 7},
		applies:  map[deploymentComponent]int64{c: 2},
		latest:   map[deploymentBucket]int64{key: 3},
	}
	require.Empty(t, collect(t, reader))
	r.deploymentsCollected, r.deploymentsActive = time.Now(), true
	metrics := collect(t, reader)
	for name, want := range map[string]int64{
		"nuon.deployment.attempts.recent": 7,
		"nuon.deployment.applies.recent":  2,
		"nuon.deployment.latest":          3,
	} {
		points := metrics[name].Data.(metricdata.Gauge[int64]).DataPoints
		require.Len(t, points, 1)
		require.Equal(t, want, points[0].Value)
		attributes := points[0].Attributes
		for name, want := range map[string]string{"org.id": c.org, "app.id": c.app, "component.id": c.component} {
			value, exists := attributes.Value(attribute.Key(name))
			require.True(t, exists)
			require.Equal(t, want, value.AsString())
		}
		if name == "nuon.deployment.applies.recent" {
			require.Equal(t, 3, attributes.Len())
		} else {
			require.Equal(t, 4, attributes.Len())
			state, _ := attributes.Value("deployment.state")
			require.Equal(t, "error", state.AsString())
		}
	}
	r.collectedAt, r.active = time.Now(), true
	r.deploymentsCollected = time.Now().Add(-deploymentSnapshotTTL - time.Second)
	metrics = collect(t, reader)
	require.Contains(t, metrics, "nuon.workflow.current")
	require.Contains(t, metrics, "nuon.deployment.snapshot.collected_at")
	require.NotContains(t, metrics, "nuon.deployment.attempts.recent")
	r.deploymentsCollected = time.Now()
	r.release(nil)
	metrics = collect(t, reader)
	require.NotContains(t, metrics, "nuon.deployment.latest")
	require.Contains(t, metrics, "nuon.deployment.snapshot.collected_at")
}

func TestDeploymentZerosAndBudget(t *testing.T) {
	c := deploymentComponent{"org-example", "app-example", "component-example"}
	old := deploymentBucket{c, "error"}
	newKey := deploymentBucket{c, "applied"}
	deleted := deploymentBucket{deploymentComponent{"org-example", "app-example", "component-deleted"}, "pending"}
	snapshot := deploymentSnapshot{
		attempts: map[deploymentBucket]int64{},
		applies:  map[deploymentComponent]int64{c: 0},
		latest:   map[deploymentBucket]int64{newKey: 1},
	}
	require.NoError(t, snapshot.retainZeros(deploymentSnapshot{attempts: map[deploymentBucket]int64{old: 3, deleted: 2}}))
	require.Equal(t, map[deploymentBucket]int64{old: 0, newKey: 0}, snapshot.attempts)
	require.Equal(t, map[deploymentBucket]int64{old: 0, newKey: 1}, snapshot.latest)
	for i := 0; i < maxDeploymentSeries-2; i++ {
		component := deploymentComponent{"org-example", "app-example", fmt.Sprint(i)}
		snapshot.applies[component] = 1
		snapshot.latest[deploymentBucket{component, "applied"}] = 1
	}
	require.NoError(t, snapshot.retainZeros(deploymentSnapshot{}))
	snapshot.latest[deploymentBucket{c, "retried"}] = 1
	require.ErrorContains(t, snapshot.retainZeros(deploymentSnapshot{}), "series per metric")
}

func TestPostgresDeploymentSnapshot(t *testing.T) {
	// INTEGRATION marks packages for the PostgreSQL CI lane.
	tests.SkipIfNotIntegration(t)
	var deps snapshotTestDeps
	fxApp := fxtest.New(t, append(tests.CtlApiFXOptions(t), fx.Populate(&deps))...)
	fxApp.RequireStart()
	t.Cleanup(fxApp.RequireStop)
	ctx := context.Background()
	migration := &migrations.Migrations{}
	require.NoError(t, migration.Migration133DeploymentMetricsIndexes(ctx, deps.DB))
	require.NoError(t, migration.Migration133DeploymentMetricsIndexes(ctx, deps.DB))
	conn, err := psql.NewPrimaryListenerConn(ctx, deps.Config)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close(ctx)) })
	testApp := deps.Seeder.CreateApp(ctx, t)
	ctx = cctx.SetAccountContext(ctx, &testApp.CreatedBy)
	ctx = cctx.SetOrgIDContext(ctx, testApp.OrgID)
	config := deps.Seeder.CreateAppConfig(ctx, t, testApp.ID)
	install := deps.Seeder.CreateInstall(ctx, t, testApp)
	now := time.Now().UTC().Truncate(time.Second)
	fixture := func(t *testing.T, raw, v2 string, age, appliedAge time.Duration) (deploymentComponent, *app.InstallDeploy, *app.Workflow) {
		t.Helper()
		component := deps.Seeder.CreateComponent(ctx, t, testApp.ID, app.ComponentTypeHelmChart)
		connection := deps.Seeder.CreateHelmComponentConfigConnection(ctx, t, component.ID, config.ID)
		build := deps.Seeder.CreateComponentBuild(ctx, t, connection.ID)
		ic := deps.Seeder.CreateInstallComponent(ctx, t, install.ID, component.ID)
		wf := deps.Seeder.CreateWorkflow(ctx, t, install.ID, app.WorkflowTypeManualDeploy)
		d := deps.Seeder.CreateInstallDeploy(ctx, t, ic.ID, build.ID)
		var applied *time.Time
		if appliedAge != 0 {
			at := now.Add(-appliedAge)
			applied = &at
		}
		require.NoError(t, deps.DB.Exec(`UPDATE install_deploys SET status=?, status_v2=jsonb_build_object('status', ?::text),
created_at=?, applied_at=?, install_workflow_id=? WHERE id=?`, raw, v2, now.Add(-age), applied, wf.ID, d.ID).Error)
		return deploymentComponent{testApp.OrgID, testApp.ID, component.ID}, d, wf
	}
	for _, tc := range []struct {
		name, raw, v2, state string
		age, appliedAge      time.Duration
		attempts, applies    int64
	}{
		{"legacy", "active", "", "applied", time.Hour, time.Hour, 1, 1},
		{"v2 wins", "active", "error", "error", time.Hour, 0, 1, 0},
		{"health gate", "health-failed", "error", "health_failed", time.Hour, time.Hour, 1, 1},
		{"retry", "error", "retried", "retried", time.Hour, 0, 1, 0},
		{"no-op", "noop", "noop", "skipped", time.Hour, 0, 1, 0},
		{"cancel", "cancelled", "cancelled", "cancelled", time.Hour, 0, 1, 0},
		{"approval", "pending-approval", "pending-approval", "awaiting_approval", time.Hour, 0, 1, 0},
		{"planning", "planning", "planning", "pending", time.Hour, 0, 1, 0},
		{"executing", "executing", "executing", "executing", time.Hour, 0, 1, 0},
		{"unknown", "future-state", "future-state", "unknown", time.Hour, 0, 1, 0},
		{"inclusive start", "active", "active", "applied", 24 * time.Hour, 24 * time.Hour, 1, 1},
		{"outside start", "active", "active", "applied", 24*time.Hour + time.Second, 24*time.Hour + time.Second, 0, 0},
		{"old attempt recent apply", "active", "active", "applied", 48 * time.Hour, time.Hour, 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, _, _ := fixture(t, tc.raw, tc.v2, tc.age, tc.appliedAge)
			values, err := readDeploymentSnapshot(ctx, conn, now)
			require.NoError(t, err)
			key := deploymentBucket{c, tc.state}
			require.Equal(t, tc.attempts, values.attempts[key])
			require.Equal(t, tc.applies, values.applies[c])
			require.EqualValues(t, 1, values.latest[key])
		})
	}
	for _, exclusion := range []string{"plan-only", "teardown", "recover", "sync-image", "deploy", "workflow", "component", "install-component", "install", "app", "org", "missing-workflow", "window-end"} {
		t.Run("exclude "+exclusion, func(t *testing.T) {
			c, d, wf := fixture(t, "active", "active", time.Hour, time.Hour)
			var restore func()
			switch exclusion {
			case "plan-only":
				require.NoError(t, deps.DB.Exec("UPDATE install_workflows SET plan_only=true WHERE id=?", wf.ID).Error)
			case "teardown", "recover", "sync-image":
				require.NoError(t, deps.DB.Exec("UPDATE install_deploys SET type=? WHERE id=?", exclusion, d.ID).Error)
			case "missing-workflow":
				require.NoError(t, deps.DB.Exec("UPDATE install_deploys SET install_workflow_id=NULL WHERE id=?", d.ID).Error)
			case "window-end":
				require.NoError(t, deps.DB.Exec("UPDATE install_deploys SET created_at=?, applied_at=? WHERE id=?", now, now, d.ID).Error)
			default:
				target := map[string]struct{ table, id string }{
					"deploy": {"install_deploys", d.ID}, "workflow": {"install_workflows", wf.ID},
					"component": {"components", c.component}, "install-component": {"install_components", d.InstallComponentID},
					"install": {"installs", install.ID}, "app": {"apps", c.app}, "org": {"orgs", c.org},
				}[exclusion]
				require.NoError(t, deps.DB.Exec("UPDATE "+target.table+" SET deleted_at=1 WHERE id=?", target.id).Error)
				restore = func() {
					require.NoError(t, deps.DB.Exec("UPDATE "+target.table+" SET deleted_at=0 WHERE id=?", target.id).Error)
				}
			}
			if restore != nil {
				defer restore()
			}
			values, err := readDeploymentSnapshot(ctx, conn, now)
			require.NoError(t, err)
			require.NotContains(t, values.applies, c)
		})
	}

	c, d, wf := fixture(t, "error", "error", 2*time.Hour, 0)
	newer := deps.Seeder.CreateInstallDeploy(ctx, t, d.InstallComponentID, d.ComponentBuildID)
	plan := deps.Seeder.CreateWorkflow(ctx, t, install.ID, app.WorkflowTypeManualDeploy)
	require.NoError(t, deps.DB.Exec("UPDATE install_workflows SET plan_only=true WHERE id=?", plan.ID).Error)
	require.NoError(t, deps.DB.Exec("UPDATE install_deploys SET install_workflow_id=?, created_at=? WHERE id=?", plan.ID, now.Add(-time.Hour), newer.ID).Error)
	values, err := readDeploymentSnapshot(ctx, conn, now)
	require.NoError(t, err)
	require.EqualValues(t, 1, values.latest[deploymentBucket{c, "error"}], "newer plan must not hide real deployment")
	require.NoError(t, deps.DB.Exec("UPDATE install_deploys SET install_workflow_id=?, status='active', status_v2='{}' WHERE id=?", wf.ID, newer.ID).Error)
	require.NoError(t, deps.DB.Exec("UPDATE install_deploys SET status='retried', status_v2='{}' WHERE id=?", d.ID).Error)
	values, err = readDeploymentSnapshot(ctx, conn, now)
	require.NoError(t, err)
	require.EqualValues(t, 1, values.attempts[deploymentBucket{c, "retried"}])
	require.EqualValues(t, 1, values.attempts[deploymentBucket{c, "applied"}])
	require.EqualValues(t, 1, values.latest[deploymentBucket{c, "applied"}])
	require.Zero(t, values.latest[deploymentBucket{c, "retried"}])

	secondInstall := deps.Seeder.CreateInstall(ctx, t, testApp)
	secondComponent := deps.Seeder.CreateInstallComponent(ctx, t, secondInstall.ID, c.component)
	secondWorkflow := deps.Seeder.CreateWorkflow(ctx, t, secondInstall.ID, app.WorkflowTypeManualDeploy)
	secondDeploy := deps.Seeder.CreateInstallDeploy(ctx, t, secondComponent.ID, d.ComponentBuildID)
	require.NoError(t, deps.DB.Exec("UPDATE install_deploys SET install_workflow_id=?, status='executing', status_v2='{}', created_at=? WHERE id=?", secondWorkflow.ID, now.Add(-time.Minute), secondDeploy.ID).Error)
	values, err = readDeploymentSnapshot(ctx, conn, now)
	require.NoError(t, err)
	require.EqualValues(t, 1, values.latest[deploymentBucket{c, "applied"}])
	require.EqualValues(t, 1, values.latest[deploymentBucket{c, "executing"}])
	values, err = readDeploymentSnapshot(ctx, conn, now.Add(25*time.Hour))
	require.NoError(t, err)
	require.Zero(t, values.attempts[deploymentBucket{c, "applied"}])
	require.Zero(t, values.applies[c])
	require.EqualValues(t, 1, values.latest[deploymentBucket{c, "applied"}])

	t.Run("indexes support generic plans", func(t *testing.T) {
		tx, err := conn.Begin(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx)
		// Historical rows make the time-range indexes selective rather than relying
		// on arbitrary planner choices between single-page fixture indexes.
		_, err = tx.Exec(ctx, `INSERT INTO install_deploys
(id, created_by_id, created_at, updated_at, deleted_at, org_id, component_build_id,
 install_component_id, status, status_description, type, applied_at, install_workflow_id)
SELECT lpad(n::text, 26, 'x'), created_by_id, created_at - n * interval '1 day', updated_at,
 deleted_at, org_id, component_build_id, install_component_id, status, status_description,
 type, created_at - n * interval '1 day', install_workflow_id
FROM install_deploys CROSS JOIN generate_series(1, 5000) n WHERE id = $1`, d.ID)
		require.NoError(t, err)
		_, err = tx.Exec(ctx, "ANALYZE install_deploys; ANALYZE install_components; ANALYZE install_workflows; ANALYZE installs; ANALYZE components; ANALYZE apps; ANALYZE orgs")
		require.NoError(t, err)
		_, err = tx.Exec(ctx, "SET LOCAL enable_seqscan=off; SET LOCAL plan_cache_mode=force_generic_plan")
		require.NoError(t, err)
		_, err = tx.Prepare(ctx, "deployment_metrics_plan", deploymentSnapshotQuery)
		require.NoError(t, err)
		rows, err := tx.Query(ctx, fmt.Sprintf("EXPLAIN EXECUTE deployment_metrics_plan('%s', '%s', %d)",
			now.Add(-24*time.Hour).Format(time.RFC3339), now.Format(time.RFC3339), 3*maxDeploymentSeries+1))
		require.NoError(t, err)
		defer rows.Close()
		var plan string
		for rows.Next() {
			var line string
			require.NoError(t, rows.Scan(&line))
			plan += line + "\n"
		}
		require.NoError(t, rows.Err())
		for _, index := range []string{"idx_install_deploys_metrics_created", "idx_install_deploys_metrics_applied"} {
			require.Contains(t, plan, index)
		}
		// The latest lookup may use another component-keyed index depending on
		// join costs. Require an indexed lookup, not a specific index choice.
		require.Regexp(t, `Index Cond: .*install_component_id = c_[0-9]+\.id`, plan)
		var valid bool
		require.NoError(t, tx.QueryRow(ctx, `SELECT indisvalid FROM pg_index
WHERE indexrelid = 'idx_install_deploys_metrics_latest'::regclass`).Scan(&valid))
		require.True(t, valid)
	})

	r, _ := testReporter(t)
	require.NoError(t, r.refresh(ctx, conn))
	require.NoError(t, r.refreshDeployments(ctx, conn))
	stamp, workflowStamp := r.deploymentsCollected, r.collectedAt
	locker, err := psql.NewPrimaryListenerConn(ctx, deps.Config)
	require.NoError(t, err)
	defer locker.Close(ctx)
	tx, err := locker.Begin(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, "LOCK TABLE install_deploys IN ACCESS EXCLUSIVE MODE")
	require.NoError(t, err)
	started := time.Now()
	require.Error(t, r.refreshDeploymentsIfDue(ctx, conn))
	require.Less(t, time.Since(started), 2*time.Second, "lock wait is bounded")
	attempted := r.deploymentAttempt
	require.NoError(t, r.refreshDeploymentsIfDue(ctx, nil), "a failed query is not retried on the 30-second workflow cadence")
	require.Equal(t, attempted, r.deploymentAttempt)
	require.Equal(t, stamp, r.deploymentsCollected)
	require.Equal(t, workflowStamp, r.collectedAt)
	require.False(t, conn.IsClosed(), "server timeout must preserve the reporter session")
	require.NoError(t, r.refresh(ctx, conn), "deployment lock contention does not break workflow collection")
	require.NoError(t, tx.Rollback(ctx))
	require.NoError(t, r.refreshDeployments(ctx, conn))
	require.True(t, r.deploymentsCollected.After(stamp))
}
